package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

const scaleBoard = "Scale replay.canvas"
const scaleReaderBytes = 4096
const fixedDesignID = "a-fixed"

type creativeScaleResult struct {
	Unrelated       int              `json:"unrelated_designs"`
	Edit            string           `json:"edit"`
	Method          string           `json:"method"`
	Required        int              `json:"required_reviews"`
	Candidates      int              `json:"candidate_reviews_before_ceiling"`
	CandidateMisses int              `json:"missed_before_ceiling"`
	CandidateExtras int              `json:"unnecessary_before_ceiling"`
	Packed          int              `json:"reviews_in_packet"`
	Missed          int              `json:"missed_after_ceiling"`
	Unnecessary     int              `json:"unnecessary_after_ceiling"`
	PacketBytes     int              `json:"packet_bytes"`
	Opened          int              `json:"historical_bodies_reopened"`
	HistoricalBytes int              `json:"exact_historical_excerpt_bytes"`
	Passages        int              `json:"old_and_new_passages_in_packet"`
	LookupNS        int64            `json:"selection_ns"`
	TotalNS         int64            `json:"selection_and_packet_ns"`
	Repair          reservationProbe `json:"scripted_repair"`
}

// Reservation misses grow the allocation until its configured reserve is full,
// then replace in FIFO order. With a stale reserve the physical allocator can
// refuse growth before the design reaches its own replacement threshold.
func serveReservationRequests(physical, reserve int, requests []int) (served, replacements int) {
	resident := make([]int, 0, physical)
	victim := 0
	for _, passage := range requests {
		if !slices.Contains(resident, passage) {
			if len(resident) < reserve {
				if len(resident) == physical {
					continue
				}
				resident = append(resident, passage)
			} else {
				resident[victim] = passage
				victim = (victim + 1) % len(resident)
				replacements++
			}
		}
		served++
	}
	return served, replacements
}

type reservationProbe struct {
	Design             string `json:"design"`
	Admitted           bool   `json:"admitted_for_review"`
	Physical           int    `json:"physical_slots"`
	ReserveBefore      int    `json:"reserve_before"`
	ReserveAfter       int    `json:"reserve_after"`
	Requests           int    `json:"requests"`
	ServedBefore       int    `json:"served_before"`
	ServedAfter        int    `json:"served_after"`
	ReplacementsBefore int    `json:"replacements_before"`
	ReplacementsAfter  int    `json:"replacements_after"`
}

func probeReservationRepair(physical int, admitted bool) reservationProbe {
	// The two passes exercise full allocations, replacement, and hits on the
	// four resident pages of a design that is still waiting for more slots.
	requests := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	r := reservationProbe{Design: fixedDesignID, Admitted: admitted, Physical: physical,
		ReserveBefore: 8, ReserveAfter: 8, Requests: len(requests)}
	if admitted {
		r.ReserveAfter = physical
	}
	r.ServedBefore, r.ReplacementsBefore = serveReservationRequests(physical, r.ReserveBefore, requests)
	r.ServedAfter, r.ReplacementsAfter = serveReservationRequests(physical, r.ReserveAfter, requests)
	return r
}

type creativeScaleReport struct {
	Method       string                `json:"methodology"`
	InputCeiling int                   `json:"reader_bytes_per_edit"`
	Results      []creativeScaleResult `json:"results"`
}

func scaleCard(key, role, body string) CreativeCard {
	return CreativeCard{Key: key, Kind: "file", Source: key + ".md", Role: role,
		Relation: role, Body: body, BodySHA: vault.SHA(body), Start: 0, End: len(body), Excerpt: body}
}

type scaleEdit struct {
	label, key, message, needs string
	physicalSlots              int
	snapshot                   CreativeCanvas
	old                        CreativeCard
	replacement                *CreativeCard
}

func scaleFixtures(e *Engine) ([]scaleEdit, error) {
	original := CreativeCanvas{Question: "How should the passage cache reserve slots?", Cards: []CreativeCard{
		scaleCard("contract", "constraint", "Eight slots are always available for pinning."),
		scaleCard("fallback", "constraint", "The spill file independently guarantees eight additional resident slots."),
		scaleCard("display", "inspiration", "The cache display uses blue labels for reserved slots."),
	}}
	type fixture struct {
		id, title, question string
		edges               []string
	}
	for _, f := range []fixture{
		{fixedDesignID, "Reserve eight internal slots", "How should the passage cache reserve slots?", []string{"contract"}},
		{"b-backed", "Borrow slots from the spill file if needed", "What can support the passage cache reserve?", []string{"contract", "fallback"}},
		{"c-display", "Color the meter from the display palette", "How should the cache display mark its slots?", []string{"display"}},
	} {
		spec := CreativeDesign{ID: f.id, Title: f.title, Prediction: "The existing mechanism holds under the cited conditions.",
			Choices: []CreativeChoice{{Choice: f.title, AnyOf: f.edges}}}
		if _, err := SaveCreativeDesign(e.S, scaleBoard, original, spec); err != nil {
			return nil, err
		}
		// Pin the same original designs and derive the production receipt's
		// actual content predicates without issuing a second generation.
		r := store.Receipt{ID: f.id, Question: f.question, AnswerText: f.title, Status: "current",
			Predicates: derivePredicates(e, f.question, f.title, nil, nil)}
		if err := e.S.PutReceipt(r); err != nil {
			return nil, err
		}
	}

	var edits []scaleEdit
	for _, change := range []struct {
		label, key, replacement, message, needs string
		physicalSlots                           int
	}{
		{"constraint-tightened", "contract", "Only four entries can remain resident.", "Only four entries can remain resident.", fixedDesignID, 4},
		{"unrelated-edit", "display", "The cache display uses green labels for reserved slots.", "The cache display uses green labels for reserved slots.", "", 8},
		{"one-support-withdrawn", "fallback", "", "The spill file guarantee has been withdrawn.", "", 8},
	} {
		edit := scaleEdit{label: change.label, key: change.key, message: change.message, needs: change.needs,
			physicalSlots: change.physicalSlots, snapshot: CreativeCanvas{Question: original.Question}}
		for _, card := range original.Cards {
			if card.Key != change.key {
				edit.snapshot.Cards = append(edit.snapshot.Cards, card)
				continue
			}
			edit.old = card
			if change.replacement != "" {
				newCard := scaleCard(card.Key, card.Role, change.replacement)
				edit.replacement = &newCard
				edit.snapshot.Cards = append(edit.snapshot.Cards, newCard)
			}
		}
		edits = append(edits, edit)
	}
	return edits, nil
}

// Each filler is a live design revision with one genuine content-addressed
// dependency and one keyword receipt. Identical text in separate notes shares
// one stored body, while identifiers and indexed keys remain distinct.
func addScaleDesigns(s *store.Store, start, end int) error {
	const marker = "ROWMARK"
	card := scaleCard("e"+marker, "constraint", "An archived sample is kept in numeric drawer order.")
	version := designVersion{
		CreativeDesign: CreativeDesign{ID: "f" + marker, Title: "Keep drawer " + marker, Prediction: "The sample remains available.",
			Choices: []CreativeChoice{{Choice: "Keep the drawer", AnyOf: []string{card.Key}}}},
		Revision: 1, Question: "Locate drawer" + marker, Cards: map[string]CreativeCard{card.Key: card},
	}
	template, err := json.Marshal(version)
	if err != nil {
		return err
	}
	if _, err := s.DB.Exec("INSERT OR IGNORE INTO creative_bodies(sha,body) VALUES(?,?)", card.BodySHA, []byte(card.Body)); err != nil {
		return err
	}

	const numbers = `WITH RECURSIVE seq(n) AS (SELECT ? UNION ALL SELECT n+1 FROM seq WHERE n < ?),
	ids(i) AS (SELECT printf('%07d',n) FROM seq) `
	for lo := start + 1; lo <= end; lo += 10000 {
		hi := min(lo+9999, end)
		tx, err := s.DB.Begin()
		if err != nil {
			return err
		}
		statements := []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO creative_revisions(board,design,revision,content) SELECT ?, 'f'||i,1,replace(?,?,i) FROM ids`, []any{scaleBoard, string(template), marker}},
			{`INSERT INTO creative_refs(board,edge,design) SELECT ?, 'e'||i,'f'||i FROM ids`, []any{scaleBoard}},
			{`INSERT INTO receipts(id,question,answer_text,snapshot_seq,snapshot_root,created,updated)
			SELECT 'f'||i, 'Locate drawer'||i,'Keep drawer '||i,0,'','','' FROM ids`, nil},
			{`INSERT INTO receipt_predicates(receipt_id,kind,key) SELECT 'f'||i,'term','drawer'||i FROM ids`, nil},
		}
		for _, stmt := range statements {
			if _, err := tx.Exec(numbers+stmt.sql, append([]any{lo, hi}, stmt.args...)...); err != nil {
				tx.Rollback()
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func scaleCandidateIDs(e *Engine, edit scaleEdit, method string) ([]string, error) {
	var ids []string
	switch method {
	case "source-linked":
		changes, err := ReviewCreativeDesigns(e.S, scaleBoard, edit.snapshot, []string{edit.key})
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		for _, c := range changes {
			if c.Status == "review" && !seen[c.ID] {
				seen[c.ID] = true
				ids = append(ids, c.ID)
			}
		}
	case "keyword-watcher":
		turn := vault.Turn{ID: "t000002", Seq: 2, Body: edit.message}
		alerts, err := e.CheckStaleReceipts(turn, nil, nil)
		if err != nil {
			return nil, err
		}
		for _, a := range alerts {
			ids = append(ids, a.ReceiptID)
		}
	}
	sort.Strings(ids)
	return ids, nil
}

func measureScale(e *Engine, unrelated int, edit scaleEdit, method string) (creativeScaleResult, error) {
	r := creativeScaleResult{Unrelated: unrelated, Edit: edit.label, Method: method}
	if edit.needs != "" {
		r.Required = 1
	}
	// Each edit starts from the same current receipts. Synthetic edits cannot
	// match a filler predicate. Reset real alerts outside the measurement.
	if _, err := e.S.DB.Exec(`UPDATE receipts SET status='current',stale_reason='',stale_turns_json='[]' WHERE id IN ('a-fixed','b-backed','c-display')`); err != nil {
		return r, err
	}
	start := time.Now()
	ids, err := scaleCandidateIDs(e, edit, method)
	if err != nil {
		return r, err
	}
	query := "SELECT id,question,answer_text FROM receipts"
	var args []any
	if method == "review-all" {
		// All live designs are candidates. Stream just as much of the actual
		// ordered result as fits, instead of materializing a million prompts.
		if err := e.S.DB.QueryRow("SELECT COUNT(*) FROM receipts").Scan(&r.Candidates); err != nil {
			return r, err
		}
	} else {
		r.Candidates = len(ids)
		if r.Required != 0 && !slices.Contains(ids, edit.needs) {
			r.CandidateMisses = 1
		}
		if len(ids) > 0 {
			query += " WHERE id IN (" + strings.TrimRight(strings.Repeat("?,", len(ids)), ",") + ")"
			for _, id := range ids {
				args = append(args, id)
			}
		}
	}
	r.LookupNS = time.Since(start).Nanoseconds()
	r.CandidateExtras = r.Candidates - (r.Required - r.CandidateMisses)
	r.Missed = r.Required
	if r.Candidates == 0 {
		r.TotalNS = time.Since(start).Nanoseconds()
		r.Repair = probeReservationRepair(edit.physicalSlots, false)
		return r, nil
	}

	oldBody, err := CreativeOriginal(e.S, edit.old.BodySHA)
	if err != nil {
		return r, err
	}
	oldText := oldBody[edit.old.Start:edit.old.End]
	if oldText != edit.old.Excerpt {
		return r, fmt.Errorf("historical excerpt changed for %s", edit.key)
	}
	r.Opened, r.HistoricalBytes = 1, len(oldText)
	newText := ""
	if edit.replacement != nil {
		newText = edit.replacement.Excerpt
	}
	rows, err := e.S.DB.Query(query+" ORDER BY id", args...)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	fixedAdmitted := false
	for rows.Next() {
		var id, question, design string
		if err := rows.Scan(&id, &question, &design); err != nil {
			return r, err
		}
		// This is the identical packet serialization for all three selectors;
		// one JSON object plus a newline per candidate.
		packet := struct {
			ID       string `json:"design"`
			Question string `json:"question"`
			Policy   string `json:"mechanism"`
			Edited   string `json:"edited_source"`
			Old      string `json:"original"`
			New      string `json:"replacement,omitempty"`
		}{id, question, design, edit.old.Source, oldText, newText}
		raw, err := json.Marshal(packet)
		if err != nil {
			return r, err
		}
		if r.PacketBytes+len(raw)+1 > scaleReaderBytes {
			break
		}
		r.Packed++
		if id == fixedDesignID {
			fixedAdmitted = true
		}
		r.PacketBytes += len(raw) + 1
		r.Passages++
		if newText != "" {
			r.Passages++
		}
		if id == edit.needs {
			r.Missed = 0
		} else {
			r.Unnecessary++
		}
	}
	r.TotalNS = time.Since(start).Nanoseconds()
	r.Repair = probeReservationRepair(edit.physicalSlots, fixedAdmitted)
	return r, rows.Err()
}

// RunCreativeScale writes a complete model-free report and removes the isolated
// workspace (including its SQLite file, WAL and synthetic receipt artifacts).
func RunCreativeScale(w io.Writer) error {
	dir, err := os.MkdirTemp("", "cx-creative-scale-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	report, err := runCreativeScale(dir, []int{1000, 100000, 1000000})
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func runCreativeScale(dir string, sizes []int) (creativeScaleReport, error) {
	report := creativeScaleReport{
		InputCeiling: scaleReaderBytes,
		Method: "Model-free replay on a scratch disk-backed SQLite WAL database. Each background design has a distinct indexed source edge and lexical receipt. " +
			"The three test designs sort before background designs, favoring review-all. The source selector receives the edited edge ID. The keyword watcher receives the edit text, " +
			"uses its ordinary invalidation rules with indexed absence discovery, and persists its alerts. Review-all counts every receipt and streams in ID order until full. " +
			"Every selector uses the same per-design JSONL encoding. Packet bytes count the actual serialized records admitted to a simulated reader under one ceiling per edit. " +
			"No model is called. Each nonempty selection fetches and verifies the old body once; repeated excerpts count toward the byte ceiling every time. " +
			"Timings are single observations on one connection after each bulk load, with uncontrolled caches. Selection and packet timing includes enumeration and serialization; fixture creation and repair replay are excluded. " +
			"Scripted repair replays two cycles over twelve passage IDs on the edited hardware with a fresh allocation before each trial. The fixed design grows until its configured reserve is full, " +
			"then replaces FIFO. An unavailable physical slot rejects a growth request. When its ID is admitted to the packet, every selector applies the same repair: match the reserve to the physical capacity.",
	}
	s, err := store.Open(dir)
	if err != nil {
		return report, err
	}
	defer s.Close()
	e := &Engine{S: s, V: &vault.Vault{Dir: dir}}
	edits, err := scaleFixtures(e)
	if err != nil {
		return report, err
	}
	previous := 0
	for _, n := range sizes {
		if n < previous || n < 0 {
			return report, fmt.Errorf("scale sizes must be nonnegative and increasing")
		}
		if err := addScaleDesigns(s, previous, n); err != nil {
			return report, err
		}
		for _, edit := range edits {
			for _, method := range []string{"source-linked", "keyword-watcher", "review-all"} {
				result, err := measureScale(e, n, edit, method)
				if err != nil {
					return report, err
				}
				report.Results = append(report.Results, result)
			}
		}
		previous = n
	}
	return report, nil
}
