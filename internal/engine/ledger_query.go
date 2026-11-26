package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// Exact resumable scoped ledger queries (graduated from
// docs/experiments/coverage/receipts.py and examples/answer_watch).
//
// Unlike prose answers, a LedgerQuery executes deterministically over the
// immutable vault with zero model calls:
//   - Binds its operator, kind, filter, status, and AsOfSeq cutoff to a
//     cryptographic SnapshotRoot.
//   - Reconstructs true point-in-time active/superseded status at AsOfSeq
//     even when later turns supersede earlier records.
//   - Conservatively returns Verdict="withheld" (instead of a false zero or
//     misleading count) whenever Strict=true and unindexed or unreadable turns
//     exist within the query's scope.
//   - Supports incremental ResumeLedgerQuery(id) returning an exact LedgerDelta
//     (Added, Removed, Superseded witnesses + ExaminedNew count) with byte-
//     verified EvidenceAnchors.

// LedgerQuerySpec specifies one exact deterministic query over the vault.
type LedgerQuerySpec struct {
	Kind    string `json:"kind"`                // decisions | facts | turns (default decisions)
	Op      string `json:"op"`                  // count | list | latest | presence | absent (default list)
	Filter  string `json:"filter,omitempty"`    // optional keyword/substring scope
	Status  string `json:"status,omitempty"`    // active (default) | superseded | all
	AsOfSeq int    `json:"as_of_seq,omitempty"` // 0 = live; >0 = immutable historical cutoff
	Strict  bool   `json:"strict"`              // when true, withhold verdict if unindexed turns exist in scope
}

// LedgerWitness is one verified record backing a LedgerQueryResult.
type LedgerWitness struct {
	RecordID   string                `json:"record_id"`
	Kind       string                `json:"kind"`
	SourceTurn string                `json:"source_turn"`
	Seq        int                   `json:"seq"`
	Statement  string                `json:"statement"`
	Quote      string                `json:"quote,omitempty"`
	Status     string                `json:"status"`
	Anchor     *vault.EvidenceAnchor `json:"anchor,omitempty"`
	Verified   bool                  `json:"verified"`
}

// LedgerDelta reports the exact difference between a prior saved snapshot and
// the resumed snapshot.
type LedgerDelta struct {
	FromSeq     int             `json:"from_seq"`
	ToSeq       int             `json:"to_seq"`
	ExaminedNew int             `json:"examined_new"`
	Added       []LedgerWitness `json:"added,omitempty"`
	Removed     []LedgerWitness `json:"removed,omitempty"`
	Superseded  []string        `json:"superseded,omitempty"`
	Unchanged   bool            `json:"unchanged"`
}

// LedgerQueryResult is a persisted, snapshot-bound deterministic query result.
type LedgerQueryResult struct {
	ID             string          `json:"id"`
	Spec           LedgerQuerySpec `json:"spec"`
	SnapshotSeq    int             `json:"snapshot_seq"`
	SnapshotRoot   string          `json:"snapshot_root"`
	Verdict        string          `json:"verdict"` // exact | withheld
	WithheldReason string          `json:"withheld_reason,omitempty"`
	Count          int             `json:"count"`
	Present        bool            `json:"present"`
	Witnesses      []LedgerWitness `json:"witnesses,omitempty"`
	Delta          *LedgerDelta    `json:"delta,omitempty"`
	Created        time.Time       `json:"created"`
	Updated        time.Time       `json:"updated"`
}

func (e *Engine) ledgerQueryDir() string {
	return filepath.Join(e.V.Dir, ".cx", "ledger_queries")
}

// ledgerQueryIDRe is the shape of the ids vault.NewID("lq") makes. An id
// reaches loadLedgerQuery from the MCP server and the web UI, and is joined
// into a file path, so anything else is refused: with no check a resume id
// of "../../x" read a file outside the vault, and the id inside the file
// read chose where the result was written.
var ledgerQueryIDRe = regexp.MustCompile(`^lq-[a-z0-9]{1,32}$`)

func (e *Engine) saveLedgerQuery(res LedgerQueryResult) error {
	if !ledgerQueryIDRe.MatchString(res.ID) {
		return fmt.Errorf("ledger query: bad id %q", res.ID)
	}
	dir := e.ledgerQueryDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, res.ID+".json"), append(b, '\n'), 0o644)
}

func (e *Engine) loadLedgerQuery(id string) (LedgerQueryResult, error) {
	var res LedgerQueryResult
	if !ledgerQueryIDRe.MatchString(id) {
		return res, fmt.Errorf("ledger query: bad id %q (ids look like lq-abc123de)", id)
	}
	b, err := os.ReadFile(filepath.Join(e.ledgerQueryDir(), id+".json"))
	if err != nil {
		return res, err
	}
	if err = json.Unmarshal(b, &res); err != nil {
		return res, err
	}
	if res.ID != id {
		return res, fmt.Errorf("ledger query: file %s.json holds id %q", id, res.ID)
	}
	return res, nil
}

// RunLedgerQuery executes spec deterministically against the vault, persists
// its snapshot-pinned checkpoint under .cx/ledger_queries/<id>.json, and
// returns the verified result.
func (e *Engine) RunLedgerQuery(spec LedgerQuerySpec) (LedgerQueryResult, error) {
	if spec.Kind == "" {
		spec.Kind = "decisions"
	}
	if spec.Op == "" {
		spec.Op = "list"
	}
	if spec.Status == "" {
		spec.Status = "active"
	}
	now := time.Now().UTC().Truncate(time.Second)
	res, err := e.evalLedgerQuery(vault.NewID("lq"), spec, now, now)
	if err != nil {
		return res, err
	}
	if e.S != nil && !e.S.ReadOnly {
		_ = e.saveLedgerQuery(res)
	}
	return res, nil
}

// ResumeLedgerQuery reloads a previously saved query checkpoint and evaluates
// only what changed since its recorded SnapshotSeq. Historical AsOfSeq queries
// never drift and verify byte-for-byte against their original snapshot.
func (e *Engine) ResumeLedgerQuery(queryID string) (LedgerQueryResult, error) {
	prior, err := e.loadLedgerQuery(queryID)
	if err != nil {
		return LedgerQueryResult{}, fmt.Errorf("ledger query %q: %w", queryID, err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	cur, err := e.evalLedgerQuery(prior.ID, prior.Spec, prior.Created, now)
	if err != nil {
		return LedgerQueryResult{}, err
	}
	examined := 0
	if prior.Spec.AsOfSeq == 0 && cur.SnapshotSeq > prior.SnapshotSeq {
		examined = cur.SnapshotSeq - prior.SnapshotSeq
	}
	priorByID := map[string]LedgerWitness{}
	for _, w := range prior.Witnesses {
		priorByID[w.RecordID] = w
	}
	curByID := map[string]LedgerWitness{}
	for _, w := range cur.Witnesses {
		curByID[w.RecordID] = w
	}
	delta := &LedgerDelta{
		FromSeq:     prior.SnapshotSeq,
		ToSeq:       cur.SnapshotSeq,
		ExaminedNew: examined,
	}
	for _, w := range cur.Witnesses {
		if _, ok := priorByID[w.RecordID]; !ok {
			delta.Added = append(delta.Added, w)
		}
	}
	for _, w := range prior.Witnesses {
		if _, ok := curByID[w.RecordID]; !ok {
			delta.Removed = append(delta.Removed, w)
			delta.Superseded = append(delta.Superseded, w.RecordID)
		}
	}
	delta.Unchanged = len(delta.Added) == 0 && len(delta.Removed) == 0 && prior.Verdict == cur.Verdict && prior.Count == cur.Count
	cur.Delta = delta
	if e.S != nil && !e.S.ReadOnly {
		_ = e.saveLedgerQuery(cur)
	}
	return cur, nil
}

func (e *Engine) evalLedgerQuery(id string, spec LedgerQuerySpec, created, updated time.Time) (LedgerQueryResult, error) {
	lastDiskSeq, err := e.V.LastSeq()
	if err != nil {
		return LedgerQueryResult{}, err
	}
	effectiveSeq := lastDiskSeq
	if spec.AsOfSeq > 0 && spec.AsOfSeq < effectiveSeq {
		effectiveSeq = spec.AsOfSeq
	}
	snapSeq, snapRoot := effectiveSeq, ""
	if e.S != nil {
		s, r, serr := e.S.SnapshotRoot(effectiveSeq)
		if serr == nil {
			snapSeq, snapRoot = s, r
		}
	}
	res := LedgerQueryResult{
		ID:           id,
		Spec:         spec,
		SnapshotSeq:  effectiveSeq,
		SnapshotRoot: snapRoot,
		Verdict:      "exact",
		Created:      created,
		Updated:      updated,
	}

	// Strict completeness gate: if the SQLite metadata index lags behind
	// effectiveSeq (e.g., turns on disk not yet indexed or store unavailable),
	// conservatively withhold an exact count/absence verdict rather than lying.
	if spec.Strict && effectiveSeq > 0 {
		if e.S == nil {
			res.Verdict = "withheld"
			res.WithheldReason = "SQLite index unavailable; exact completeness cannot be guaranteed"
			return res, nil
		}
		if snapSeq < effectiveSeq {
			res.Verdict = "withheld"
			res.WithheldReason = fmt.Sprintf("index lags archive (indexed up to seq %d, archive at seq %d); run `cx reindex` before asserting exact %s", snapSeq, effectiveSeq, spec.Op)
			return res, nil
		}
	}

	filterWords := packet.ContentWords(spec.Filter)
	filterLow := strings.ToLower(strings.TrimSpace(spec.Filter))
	matchesFilter := func(s string) bool {
		if filterLow == "" {
			return true
		}
		if strings.Contains(strings.ToLower(s), filterLow) {
			return true
		}
		if len(filterWords) == 0 {
			return false
		}
		sw := packet.ContentWords(s)
		hits := 0
		for w := range filterWords {
			if sw[w] {
				hits++
			}
		}
		return hits >= max(1, (len(filterWords)+1)/2)
	}

	anchorQuote := func(turnID, quote string) (*vault.EvidenceAnchor, bool) {
		t, terr := e.V.Turn(turnID)
		if terr != nil {
			return nil, false
		}
		// A witness is verified only when its quote is in the turn. The
		// anchor of a quote that is not found still spans the whole body,
		// so a reader can open the turn, but it is reported unverified: it
		// used to count as verified, which made "SHA-verified" true of any
		// quote at all.
		start, end := 0, len(t.Body)
		found := quote == ""
		if quote != "" {
			if idx := strings.Index(t.Body, quote); idx >= 0 {
				start, end, found = idx, idx+len(quote), true
			} else if idx := strings.Index(strings.ToLower(t.Body), strings.ToLower(quote)); idx >= 0 && idx+len(quote) <= len(t.Body) {
				start, end, found = idx, idx+len(quote), true
			}
		}
		anc := &vault.EvidenceAnchor{Turn: t.ID, BodySHA: t.BodySHA, Start: start, End: end, Seg: vault.AtomSeg}
		_, ok := vault.VerifyAnchor(t.Body, t.BodySHA, *anc)
		return anc, ok && found
	}

	var witnesses []LedgerWitness
	switch strings.ToLower(spec.Kind) {
	case "decisions":
		decs := e.V.Decisions()
		// Reconstruct true point-in-time supersession at effectiveSeq:
		// a decision is superseded at effectiveSeq iff some decision with
		// source seq <= effectiveSeq names it in Supersedes.
		supersededAt := map[string]bool{}
		for _, d := range decs {
			seq, _ := vault.TurnSeq(d.Source)
			if effectiveSeq > 0 && seq > effectiveSeq {
				continue
			}
			for _, old := range d.Supersedes {
				supersededAt[old] = true
			}
		}
		for _, d := range decs {
			seq, _ := vault.TurnSeq(d.Source)
			if effectiveSeq > 0 && seq > effectiveSeq {
				continue
			}
			st := "active"
			if supersededAt[d.ID] || (spec.AsOfSeq == 0 && d.Status == "superseded") {
				st = "superseded"
			}
			if spec.Status != "all" && st != spec.Status {
				continue
			}
			if !matchesFilter(d.Statement + " " + d.Quote) {
				continue
			}
			anc, ok := anchorQuote(d.Source, d.Quote)
			witnesses = append(witnesses, LedgerWitness{
				RecordID: d.ID, Kind: "decision", SourceTurn: d.Source, Seq: seq,
				Statement: d.Statement, Quote: d.Quote, Status: st, Anchor: anc, Verified: ok,
			})
		}

	case "facts":
		facts := e.V.Facts()
		supersededAt := map[string]bool{}
		for _, f := range facts {
			seq, _ := vault.TurnSeq(f.Source)
			if effectiveSeq > 0 && seq > effectiveSeq {
				continue
			}
			for _, old := range f.Supersedes {
				supersededAt[old] = true
			}
		}
		for _, f := range facts {
			seq, _ := vault.TurnSeq(f.Source)
			if effectiveSeq > 0 && seq > effectiveSeq {
				continue
			}
			st := "active"
			if supersededAt[f.ID] || (spec.AsOfSeq == 0 && (f.Status == "superseded" || f.Status == "retracted")) {
				st = "superseded"
				if f.Status == "retracted" {
					st = "retracted"
				}
			}
			if spec.Status != "all" && st != spec.Status {
				continue
			}
			if !matchesFilter(f.Statement + " " + f.Quote) {
				continue
			}
			anc, ok := anchorQuote(f.Source, f.Quote)
			witnesses = append(witnesses, LedgerWitness{
				RecordID: f.ID, Kind: "fact", SourceTurn: f.Source, Seq: seq,
				Statement: f.Statement, Quote: f.Quote, Status: st, Anchor: anc, Verified: ok,
			})
		}

	case "turns":
		turns, terr := e.V.Turns()
		if terr != nil {
			res.Verdict = "withheld"
			res.WithheldReason = "turn archive verification error: " + terr.Error()
			return res, nil
		}
		for _, t := range turns {
			if effectiveSeq > 0 && t.Seq > effectiveSeq {
				continue
			}
			if !matchesFilter(t.Body) {
				continue
			}
			anc := &vault.EvidenceAnchor{Turn: t.ID, BodySHA: t.BodySHA, Start: 0, End: len(t.Body), Seg: vault.AtomSeg}
			_, ok := vault.VerifyAnchor(t.Body, t.BodySHA, *anc)
			witnesses = append(witnesses, LedgerWitness{
				RecordID: t.ID, Kind: "turn", SourceTurn: t.ID, Seq: t.Seq,
				Statement: oneLine(t.Body, 160), Status: t.Role, Anchor: anc, Verified: ok,
			})
		}
	default:
		return res, fmt.Errorf("unknown ledger query kind %q (decisions, facts, turns)", spec.Kind)
	}

	sort.SliceStable(witnesses, func(i, j int) bool {
		if witnesses[i].Seq != witnesses[j].Seq {
			return witnesses[i].Seq < witnesses[j].Seq
		}
		return witnesses[i].RecordID < witnesses[j].RecordID
	})

	res.Count = len(witnesses)
	res.Present = len(witnesses) > 0
	switch strings.ToLower(spec.Op) {
	case "latest":
		if len(witnesses) > 0 {
			res.Witnesses = witnesses[len(witnesses)-1:]
		}
	case "count", "presence", "absent":
		// Keep up to 3 representative verified witnesses alongside the exact count/boolean.
		if len(witnesses) > 3 {
			res.Witnesses = witnesses[len(witnesses)-3:]
		} else {
			res.Witnesses = witnesses
		}
		if strings.ToLower(spec.Op) == "absent" {
			res.Present = len(witnesses) == 0
		}
	default: // "list"
		res.Witnesses = witnesses
	}
	return res, nil
}

// FormatLedgerQuery renders a LedgerQueryResult cleanly for CLI output.
func FormatLedgerQuery(r LedgerQueryResult) string {
	var b strings.Builder
	scope := fmt.Sprintf("live#%d", r.SnapshotSeq)
	if r.Spec.AsOfSeq > 0 {
		scope = fmt.Sprintf("as-of#%d", r.Spec.AsOfSeq)
	}
	fmt.Fprintf(&b, "LEDGER QUERY %s  kind=%s  op=%s  status=%s  scope=%s  verdict=%s\n",
		r.ID, r.Spec.Kind, r.Spec.Op, r.Spec.Status, scope, r.Verdict)
	if r.SnapshotRoot != "" {
		fmt.Fprintf(&b, "  snapshot_root: %s\n", r.SnapshotRoot)
	}
	if r.Verdict == "withheld" {
		fmt.Fprintf(&b, "  WITHHELD: %s\n", r.WithheldReason)
		return b.String()
	}
	fmt.Fprintf(&b, "  count=%d  present=%v\n", r.Count, r.Present)
	if r.Delta != nil {
		fmt.Fprintf(&b, "  delta: seq %d -> %d (examined %d new, unchanged=%v, added=%d, removed=%d)\n",
			r.Delta.FromSeq, r.Delta.ToSeq, r.Delta.ExaminedNew, r.Delta.Unchanged, len(r.Delta.Added), len(r.Delta.Removed))
	}
	if len(r.Witnesses) > 0 {
		b.WriteString("\nWITNESSES (SHA-verified)\n")
		for _, w := range r.Witnesses {
			mark := "ok"
			if !w.Verified {
				mark = "UNVERIFIED"
			}
			fmt.Fprintf(&b, "  - [%s|%s] %s (%s): %s\n", w.Status, mark, w.RecordID, w.SourceTurn, w.Statement)
		}
	}
	return b.String()
}
