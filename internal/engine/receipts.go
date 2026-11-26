package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// StaleAlert reports one saved Query Receipt that became stale because a new
// turn or superseding ledger record satisfied one of its watch predicates.
type StaleAlert struct {
	ReceiptID   string   `json:"receipt_id"`
	Question    string   `json:"question"`
	TriggerTurn string   `json:"trigger_turn"`
	Reason      string   `json:"reason"`
	MatchedKeys []string `json:"matched_keys,omitempty"`
}

// ReceiptDiff is a deterministic, model-free comparison between a receipt's
// original snapshot and the current archive state.
type ReceiptDiff struct {
	Receipt           store.Receipt `json:"receipt"`
	OriginalVerified  bool          `json:"original_verified"`
	SnapshotMatch     bool          `json:"snapshot_match"`
	CurrentSeq        int           `json:"current_seq"`
	CurrentRoot       string        `json:"current_root"`
	DeltaTurns        []vault.Turn  `json:"delta_turns,omitempty"`
	VerifiedQuotes    []string      `json:"verified_quotes,omitempty"`
	VerificationNotes []string      `json:"verification_notes,omitempty"`
}

// CreateReceipt saves a snapshot-bound Query Receipt in SQLite and under
// artifacts/receipts/<id>.json. Set asOfSeq > 0 for a historical cutoff
// (never invalidated by later turns); asOfSeq == 0 creates a live watched
// receipt that reactive invalidation monitors as new turns arrive.
func (e *Engine) CreateReceipt(question, answerTurn, answerText string, asOfSeq int, evidence []store.ReceiptEvidence, extraPreds []store.ReceiptPredicate, unresolved, parentID string) (store.Receipt, error) {
	if e.S == nil || e.S.ReadOnly {
		return store.Receipt{}, fmt.Errorf("receipts require a writable workspace store")
	}
	snapSeq, snapRoot, err := e.S.SnapshotRoot(asOfSeq)
	if err != nil {
		return store.Receipt{}, err
	}
	if snapSeq == 0 {
		if last, lerr := e.V.LastSeq(); lerr == nil {
			snapSeq = last
		}
	}

	// Automatically derive watch predicates from the question's content stems,
	// relevant facts/decisions/entities, and any explicit absence/custom checks.
	preds := derivePredicates(e, question, answerText, evidence, extraPreds)
	now := time.Now().UTC().Truncate(time.Second)
	r := store.Receipt{
		ID:              vault.NewID("rcpt"),
		Question:        strings.TrimSpace(question),
		AnswerTurn:      answerTurn,
		AnswerText:      strings.TrimSpace(answerText),
		SnapshotSeq:     snapSeq,
		SnapshotRoot:    snapRoot,
		AsOfSeq:         asOfSeq,
		Evidence:        evidence,
		Predicates:      preds,
		UnresolvedScope: strings.TrimSpace(unresolved),
		ParentID:        parentID,
		Status:          "current",
		Created:         now,
		Updated:         now,
	}
	if err := e.S.PutReceipt(r); err != nil {
		return store.Receipt{}, err
	}
	e.writeReceiptArtifact(r)
	return r, nil
}

func (e *Engine) writeReceiptArtifact(r store.Receipt) {
	dir := filepath.Join(e.V.Dir, "artifacts", "receipts")
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	if b, err := json.MarshalIndent(r, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(dir, r.ID+".json"), append(b, '\n'), 0o644)
	}
}

func derivePredicates(e *Engine, question, answerText string, evidence []store.ReceiptEvidence, extra []store.ReceiptPredicate) []store.ReceiptPredicate {
	seen := map[string]bool{}
	var out []store.ReceiptPredicate
	add := func(kind, key, val string) {
		key = strings.ToLower(strings.TrimSpace(key))
		if kind == "" || key == "" {
			return
		}
		k := kind + "\x00" + key
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, store.ReceiptPredicate{Kind: kind, Key: key, Value: val})
	}
	for _, p := range extra {
		add(p.Kind, p.Key, p.Value)
	}
	// Content words of the question watch for future turns discussing the same subject.
	qw := packet.ContentWords(question)
	words := make([]string, 0, len(qw))
	for w := range qw {
		words = append(words, w)
	}
	sort.Strings(words)
	for _, w := range words {
		if len(w) >= 3 {
			add("term", w, "")
		}
	}
	// Detect absence-style questions automatically ("nobody", "has anyone", "is there any", "not yet").
	lowQ := strings.ToLower(question + " " + answerText)
	if strings.Contains(lowQ, "nobody") || strings.Contains(lowQ, "no one") || strings.Contains(lowQ, "not yet") || strings.Contains(lowQ, "none found") || strings.Contains(lowQ, "has anyone") {
		if len(words) > 0 {
			add("absence", strings.Join(words[:min(4, len(words))], " "), "no matching record at snapshot")
		}
	}
	// Link matching active facts, decisions, and entities in the store.
	if e.S != nil {
		if hits, err := e.S.SearchE(question, 6, vault.TypeFact); err == nil {
			for _, h := range hits {
				add("fact", h.ID, "")
			}
		}
		if hits, err := e.S.SearchE(question, 6, vault.TypeDecision); err == nil {
			for _, h := range hits {
				add("decision", h.ID, "")
			}
		}
		if hits, err := e.S.SearchE(question, 6, vault.TypeEntity); err == nil {
			for _, h := range hits {
				add("entity", h.ID, "")
			}
		}
	}
	// Watch values stated in the answer so a later value change stands out.
	for _, v := range packet.Values(answerText) {
		add("value", v, "")
	}
	for _, ev := range evidence {
		if ev.Turn != "" {
			add("source_turn", ev.Turn, ev.BodySHA)
		}
	}
	return out
}

// CheckStaleReceipts runs reactive invalidation when a new turn arrives or
// when existing facts/decisions are superseded. Historical-cutoff receipts
// (AsOfSeq > 0) are never marked stale.
func (e *Engine) CheckStaleReceipts(newTurn vault.Turn, supersededFacts, supersededDecisions []string) ([]StaleAlert, error) {
	if e.S == nil || e.S.ReadOnly {
		return nil, nil
	}
	var pairs [][2]string
	turnWords := packet.ContentWords(newTurn.Body)
	for w := range turnWords {
		pairs = append(pairs, [2]string{"term", w})
	}
	for _, fid := range supersededFacts {
		pairs = append(pairs, [2]string{"fact", fid})
	}
	for _, did := range supersededDecisions {
		pairs = append(pairs, [2]string{"decision", did})
	}
	// Also check any active absence predicates (small set of live receipts).
	ids, err := e.S.MatchingReceiptIDs(newTurn.Seq, pairs)
	if err != nil {
		return nil, err
	}
	live, err := e.S.ListReceipts("current")
	if err != nil {
		return nil, err
	}
	staleExisting, _ := e.S.ListReceipts("stale")
	candSet := map[string]bool{}
	for _, id := range ids {
		candSet[id] = true
	}
	for _, r := range append(live, staleExisting...) {
		if r.AsOfSeq > 0 || r.SnapshotSeq >= newTurn.Seq {
			continue
		}
		for _, p := range r.Predicates {
			if p.Kind == "absence" {
				candSet[r.ID] = true
				break
			}
		}
	}

	supFactSet := map[string]bool{}
	for _, f := range supersededFacts {
		supFactSet[strings.ToLower(f)] = true
	}
	supDecSet := map[string]bool{}
	for _, d := range supersededDecisions {
		supDecSet[strings.ToLower(d)] = true
	}
	turnVals := packet.Values(newTurn.Body)

	var alerts []StaleAlert
	candIDs := make([]string, 0, len(candSet))
	for id := range candSet {
		candIDs = append(candIDs, id)
	}
	sort.Strings(candIDs)

	for _, id := range candIDs {
		r, ok, err := e.S.GetReceipt(id)
		if err != nil || !ok || r.AsOfSeq > 0 || r.Status == "superseded" || r.SnapshotSeq >= newTurn.Seq {
			continue
		}
		var reasons []string
		var matchedKeys []string
		matchedTerms := 0
		totalTerms := 0
		for _, p := range r.Predicates {
			switch p.Kind {
			case "fact":
				if supFactSet[p.Key] {
					reasons = append(reasons, fmt.Sprintf("watched fact %s was superseded by %s", p.Key, newTurn.ID))
					matchedKeys = append(matchedKeys, "fact:"+p.Key)
				}
			case "decision":
				if supDecSet[p.Key] {
					reasons = append(reasons, fmt.Sprintf("watched decision %s was superseded by %s", p.Key, newTurn.ID))
					matchedKeys = append(matchedKeys, "decision:"+p.Key)
				}
			case "absence":
				req := strings.Fields(p.Key)
				hit := 0
				for _, w := range req {
					if turnWords[w] {
						hit++
					}
				}
				if len(req) > 0 && hit >= max(1, (len(req)+1)/2) {
					reasons = append(reasons, fmt.Sprintf("absence condition %q may now be satisfied by %s", p.Key, newTurn.ID))
					matchedKeys = append(matchedKeys, "absence:"+p.Key)
				}
			case "term":
				totalTerms++
				if turnWords[p.Key] {
					matchedTerms++
					matchedKeys = append(matchedKeys, "term:"+p.Key)
				}
			}
		}
		// Require meaningful topical overlap (>=2 watched stems, or all watched stems
		// when the query had only 1 stem alongside a concrete value) so unrelated
		// chatter never false-triggers staleness.
		if len(reasons) == 0 && totalTerms > 0 {
			needTerms := min(2, totalTerms)
			if matchedTerms >= needTerms && (matchedTerms >= 2 || len(turnVals) > 0) {
				reasons = append(reasons, fmt.Sprintf("new turn %s matches %d/%d watched terms (values: %s)",
					newTurn.ID, matchedTerms, totalTerms, strings.Join(turnVals[:min(4, len(turnVals))], ", ")))
			}
		}
		if len(reasons) == 0 {
			continue
		}
		already := false
		for _, tid := range r.StaleTurns {
			if tid == newTurn.ID {
				already = true
				break
			}
		}
		if !already {
			r.StaleTurns = append(r.StaleTurns, newTurn.ID)
		}
		r.Status = "stale"
		r.StaleReason = strings.Join(reasons, "; ")
		r.Updated = time.Now().UTC().Truncate(time.Second)
		if err := e.S.PutReceipt(r); err != nil {
			return alerts, err
		}
		e.writeReceiptArtifact(r)
		alerts = append(alerts, StaleAlert{
			ReceiptID:   r.ID,
			Question:    r.Question,
			TriggerTurn: newTurn.ID,
			Reason:      r.StaleReason,
			MatchedKeys: matchedKeys,
		})
	}
	return alerts, nil
}

// DiffReceiptSince verifies a receipt's original evidence byte-for-byte against
// the immutable vault and returns only the new turns arriving after its
// snapshot that affect its predicates (zero model calls).
func (e *Engine) DiffReceiptSince(receiptID string) (ReceiptDiff, error) {
	r, ok, err := e.S.GetReceipt(receiptID)
	if err != nil {
		return ReceiptDiff{}, err
	}
	if !ok {
		return ReceiptDiff{}, fmt.Errorf("receipt %q not found", receiptID)
	}
	d := ReceiptDiff{Receipt: r, OriginalVerified: true}
	// Verify historical snapshot root matches the immutable archive at SnapshotSeq.
	_, histRoot, err := e.S.SnapshotRoot(r.SnapshotSeq)
	if err == nil && histRoot == r.SnapshotRoot {
		d.SnapshotMatch = true
	} else {
		d.VerificationNotes = append(d.VerificationNotes, "snapshot root differs from recorded receipt root")
	}
	// Verify every recorded evidence anchor byte-for-byte on disk.
	for _, ev := range r.Evidence {
		t, terr := e.V.Turn(ev.Turn)
		if terr != nil {
			d.OriginalVerified = false
			d.VerificationNotes = append(d.VerificationNotes, fmt.Sprintf("%s unreadable: %v", ev.Turn, terr))
			continue
		}
		anc := vault.EvidenceAnchor{Turn: ev.Turn, BodySHA: ev.BodySHA, Start: ev.Start, End: ev.End, Seg: ev.Seg, AtomRef: ev.AtomRef}
		// Evidence is verified only if it names the body it was taken from
		// and the quote recorded with it is still inside its span (the span
		// may be the whole paragraph around the quote). A span with no hash
		// and any quote at all used to pass on its byte range alone.
		txt, ok := vault.VerifyAnchor(t.Body, t.BodySHA, anc)
		if ok && (ev.BodySHA == "" || (strings.TrimSpace(ev.Quote) != "" &&
			!strings.Contains(strings.Join(strings.Fields(txt), " "), strings.Join(strings.Fields(ev.Quote), " ")))) {
			ok = false
		}
		if ok {
			d.VerifiedQuotes = append(d.VerifiedQuotes, fmt.Sprintf("%s [%d:%d]: %s", ev.Turn, ev.Start, ev.End, oneLine(txt, 140)))
		} else {
			d.OriginalVerified = false
			d.VerificationNotes = append(d.VerificationNotes, fmt.Sprintf("anchor %s[%d:%d] failed SHA/range verification", ev.Turn, ev.Start, ev.End))
		}
	}
	d.CurrentSeq, d.CurrentRoot, _ = e.S.SnapshotRoot(0)
	// Collect delta turns after r.SnapshotSeq (respecting historical AsOfSeq).
	if r.AsOfSeq == 0 && d.CurrentSeq > r.SnapshotSeq {
		wantIDs := map[string]bool{}
		for _, tid := range r.StaleTurns {
			wantIDs[tid] = true
		}
		// Also include any post-snapshot turns matching the question's watched terms.
		var qTerms []string
		for _, p := range r.Predicates {
			if p.Kind == "term" || p.Kind == "absence" {
				qTerms = append(qTerms, p.Key)
			}
		}
		if len(qTerms) > 0 {
			if hits, herr := e.S.SearchE(strings.Join(qTerms, " "), 20, vault.TypeTurn); herr == nil {
				for _, h := range hits {
					if seq, ok := vault.TurnSeq(h.ID); ok && seq > r.SnapshotSeq {
						wantIDs[h.ID] = true
					}
				}
			}
		}
		ids := make([]string, 0, len(wantIDs))
		for id := range wantIDs {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		if loaded, lerr := e.V.TurnsByIDs(ids); lerr == nil {
			for _, id := range ids {
				d.DeltaTurns = append(d.DeltaTurns, loaded[id])
			}
		}
	}
	return d, nil
}

// WatchQuestion answers a question (live or as-of a historical seq) and binds
// a durable Query Receipt so future turns automatically flag staleness.
func (e *Engine) WatchQuestion(ctx context.Context, question string, asOfSeq int, absenceKeys []string) (store.Receipt, Reply, error) {
	rep, err := e.Ask(ctx, question)
	if err != nil {
		return store.Receipt{}, rep, err
	}
	var extraPreds []store.ReceiptPredicate
	for _, ak := range absenceKeys {
		extraPreds = append(extraPreds, store.ReceiptPredicate{Kind: "absence", Key: ak, Value: "absent at watch time"})
	}
	// Pin evidence anchors from top ranked atoms backing this question.
	var evs []store.ReceiptEvidence
	if turns, terr := e.V.Turns(); terr == nil {
		for _, r := range packet.RankAtoms(turns, e.S, question, rep.User.Seq, 4) {
			evs = append(evs, store.ReceiptEvidence{
				Turn:    r.Turn.ID,
				BodySHA: r.Turn.BodySHA,
				Start:   r.Atom.Start,
				End:     r.Atom.End,
				Seg:     vault.AtomSeg,
				AtomRef: r.Atom.ID(),
				Quote:   oneLine(r.Atom.Text(r.Turn.Body), 180),
			})
		}
	}
	rcpt, err := e.CreateReceipt(question, rep.Assistant.ID, rep.Assistant.Body, asOfSeq, evs, extraPreds, "", "")
	return rcpt, rep, err
}

// RefreshReceipt updates a stale receipt: it feeds a fresh model call only the
// prior verified receipt (original answer + verified evidence) plus the delta
// turns that arrived after the receipt's snapshot, keeping context constant as
// the archive grows. Chains a new revision receipt and supersedes the old one.
func (e *Engine) RefreshReceipt(ctx context.Context, receiptID string) (store.Receipt, outcome, error) {
	diff, err := e.DiffReceiptSince(receiptID)
	if err != nil {
		return store.Receipt{}, outcome{}, err
	}
	old := diff.Receipt
	if old.AsOfSeq > 0 {
		return old, outcome{}, fmt.Errorf("receipt %s is pinned to historical cutoff seq %d and does not refresh", old.ID, old.AsOfSeq)
	}
	var deltaText strings.Builder
	fmt.Fprintf(&deltaText, "Prior verified answer (snapshot seq %d, root %s):\n%s\n\nVerified original evidence:\n",
		old.SnapshotSeq, short(old.SnapshotRoot), old.AnswerText)
	for _, q := range diff.VerifiedQuotes {
		fmt.Fprintf(&deltaText, "- %s\n", q)
	}
	if old.UnresolvedScope != "" {
		fmt.Fprintf(&deltaText, "\nPrior unresolved scope: %s\n", old.UnresolvedScope)
	}
	fmt.Fprintf(&deltaText, "\nStaleness trigger: %s\n", old.StaleReason)
	if len(diff.DeltaTurns) == 0 {
		deltaText.WriteString("\nNo new turns matched since snapshot.\n")
	} else {
		deltaText.WriteString("\nNew verbatim turns since snapshot:\n")
		for _, dt := range diff.DeltaTurns {
			fmt.Fprintf(&deltaText, "\n### %s\n\n%s\n", packet.TurnHeader(dt), dt.Body)
		}
	}

	focus := e.nextTurn(old.Question)
	deltaItem := packet.Item{
		Key:    packet.KeyExtra,
		ID:     "receipt-delta",
		Seq:    1,
		Header: fmt.Sprintf("Delta resume for receipt %s (rev %d -> %d)", old.ID, old.SnapshotSeq, diff.CurrentSeq),
		Text:   deltaText.String(),
	}
	spec := callSpec{
		Kind:      packet.KindRespond,
		Role:      config.RoleResponder,
		Focus:     focus,
		NoHistory: true,
		Extra:     []packet.Item{deltaItem},
		Instructions: func(id string) string {
			return fmt.Sprintf(`You are resuming a previously answered question after new records arrived in the archive.
The prior verified answer, its verified evidence, and the exact new turns since its snapshot are under "Supplied records".
State clearly whether the answer changed, what exact new turn/statement changed it (or why the prior answer still stands), and note any remaining unresolved scope.
If you need an earlier turn not shown here, return status needs_recall with recall_ids, recall_queries, or recall_ops.
Status completed with the updated reply in `+"`answer`"+`. Set run_id to "%s". Fill every field you do not use with an empty string or an empty array.`, id)
		},
	}
	o, err := e.runToCompletion(ctx, spec)
	if err != nil {
		return store.Receipt{}, o, err
	}
	newEvs := append([]store.ReceiptEvidence{}, old.Evidence...)
	for _, dt := range diff.DeltaTurns {
		newEvs = append(newEvs, store.ReceiptEvidence{
			Turn:    dt.ID,
			BodySHA: dt.BodySHA,
			Start:   0,
			End:     len(dt.Body),
			Seg:     vault.AtomSeg,
			AtomRef: dt.ID,
			Quote:   oneLine(dt.Body, 180),
		})
	}
	updated, err := e.CreateReceipt(old.Question, "", o.Answer, 0, newEvs, old.Predicates, old.UnresolvedScope, old.ID)
	if err != nil {
		return store.Receipt{}, o, err
	}
	old.Status = "superseded"
	old.Updated = time.Now().UTC().Truncate(time.Second)
	_ = e.S.PutReceipt(old)
	e.writeReceiptArtifact(old)
	return updated, o, nil
}
