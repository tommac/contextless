package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// ExecuteRecallOps runs typed, deterministic, zero-LLM recall operations
// (exchange, open_block, timeline, grep) pinned to turns strictly before
// focus.Seq. Returns newly pinned turn/atom ids, structured supplied items
// (for timeline/grep summaries), and diagnostic notes for the next run.
func (e *Engine) ExecuteRecallOps(ops []packet.RecallOp, focus vault.Turn, pinned []string) (add []string, extras []packet.Item, notes []string) {
	if len(ops) == 0 {
		return nil, nil, nil
	}
	have := map[string]bool{}
	for _, p := range pinned {
		have[p] = true
	}
	maxSeq := focus.Seq - 1
	if maxSeq < 0 {
		if last, err := e.V.LastSeq(); err == nil {
			maxSeq = last
		}
	}
	addID := func(id string) bool {
		if have[id] || len(add) >= maxRecallIDs {
			return false
		}
		have[id] = true
		add = append(add, id)
		return true
	}

	const maxOps = 6
	for idx, op := range ops {
		if idx >= maxOps {
			notes = append(notes, fmt.Sprintf("Only %d recall_ops run at once; remaining operations were skipped.", maxOps))
			break
		}
		kind := strings.ToLower(strings.TrimSpace(op.Op))
		target := strings.TrimSpace(op.Target)
		if target == "" && kind != "list_omitted" {
			continue
		}
		// Enforce snapshot pinning: if the caller pinned a snapshot root,
		// verify the archive's snapshot at maxSeq still matches verbatim so
		// recall never silently mixes revisions.
		if op.Snapshot != "" && e.S != nil {
			_, curRoot, serr := e.S.SnapshotRoot(maxSeq)
			if serr != nil || curRoot != op.Snapshot {
				notes = append(notes, fmt.Sprintf("%s rejected (snapshot mismatch: op pinned %q, archive at seq %d has %q).",
					kind, op.Snapshot, maxSeq, curRoot))
				continue
			}
		}
		switch kind {
		case "notes":
			// The map lines of the notes under a topic, or matching search
			// words: how a reader opens what an overfull map folded away.
			item, note := e.opNotes(target, op.Limit, maxSeq, idx)
			if item != nil {
				extras = append(extras, *item)
			}
			notes = append(notes, note)

		case "exchange":
			// Recall predecessor + target + successor turns to preserve
			// speaker attribution and question-answer reply windows.
			ref, ok := vault.ParseRef(target)
			if !ok {
				notes = append(notes, fmt.Sprintf("exchange %q: not a valid turn id.", target))
				continue
			}
			seq, ok := vault.TurnSeq(ref.Turn)
			if !ok || seq < 1 || (focus.Seq > 0 && seq >= focus.Seq) {
				notes = append(notes, fmt.Sprintf("exchange %q: outside earlier turns 1..%d.", target, maxSeq))
				continue
			}
			radius := op.Radius
			if radius <= 0 {
				radius = 1
			}
			if radius > 3 {
				radius = 3
			}
			var got []string
			for s := max(1, seq-radius); s <= min(maxSeq, seq+radius); s++ {
				tid := vault.TurnID(s)
				if _, err := e.V.Turn(tid); err == nil {
					if addID(tid) {
						got = append(got, tid)
					}
				}
			}
			notes = append(notes, fmt.Sprintf("exchange(%s, radius=%d) pinned %s.", ref.Turn, radius, strings.Join(got, ", ")))

		case "open_block":
			// Expand an atom to its complete enclosing Markdown code block,
			// table, or countable list across atom boundaries.
			ref, ok := vault.ParseRef(target)
			if !ok || ref.Atom <= 0 {
				notes = append(notes, fmt.Sprintf("open_block %q: provide an atom id such as t000123#a4.", target))
				continue
			}
			seq, _ := vault.TurnSeq(ref.Turn)
			if focus.Seq > 0 && seq >= focus.Seq {
				notes = append(notes, fmt.Sprintf("open_block %q: not before focus turn.", target))
				continue
			}
			t, err := e.V.Turn(ref.Turn)
			if err != nil {
				notes = append(notes, fmt.Sprintf("open_block %q: %v.", target, err))
				continue
			}
			lo, hi, bkind := vault.EnclosingBlock(t.ID, t.Body, ref.Atom)
			if lo == 0 || hi == 0 {
				notes = append(notes, fmt.Sprintf("open_block %q: paragraph out of range.", target))
				continue
			}
			id := vault.AtomRangeID(t.ID, lo, hi)
			if lo == 1 && hi == len(packet.AtomsOf(t)) {
				id = t.ID
			}
			addID(id)
			notes = append(notes, fmt.Sprintf("open_block(%s) expanded %s block to %s.", target, bkind, id))

		case "timeline":
			// Project a dated chronological chain of active + superseded facts
			// and decisions matching target (ProgMem / CogCanvas state view).
			item, found := e.buildTimelineItem(target, focus.Seq, idx)
			if !found {
				notes = append(notes, fmt.Sprintf("timeline(%q) found no matching facts or decisions.", target))
				continue
			}
			extras = append(extras, item)
			notes = append(notes, fmt.Sprintf("timeline(%q) attached chronological ledger view (%s).", target, item.ID))

		case "grep":
			// Exact literal or /regex/ search across earlier turns returning
			// line-numbered snippets and pinning the top matching atoms.
			limit := op.Limit
			if limit <= 0 {
				limit = 8
			}
			if limit > 20 {
				limit = 20
			}
			hits, item, err := e.grepVault(target, focus.Seq, limit, idx)
			if err != nil {
				notes = append(notes, fmt.Sprintf("grep(%q) failed: %v.", target, err))
				continue
			}
			if len(hits) == 0 {
				notes = append(notes, fmt.Sprintf("grep(%q) matched 0 lines across turns 1..%d.", target, maxSeq))
				continue
			}
			var pinnedHits []string
			for _, id := range hits {
				if addID(id) {
					pinnedHits = append(pinnedHits, id)
				}
			}
			if item != nil {
				extras = append(extras, *item)
			}
			notes = append(notes, fmt.Sprintf("grep(%q) matched %d span(s): %s.", target, len(hits), strings.Join(hits, ", ")))

		case "list_omitted":
			// Snapshot-verified paginated enumeration of an external .cxi omitted catalog.
			item, note := e.opListOmitted(op, maxSeq, idx)
			if item != nil {
				extras = append(extras, *item)
			}
			notes = append(notes, note)

		case "check_omitted":
			// Snapshot-verified O(1) point membership check in an external .cxi omitted catalog.
			notes = append(notes, e.opCheckOmitted(op, maxSeq))

		case "key_history":
			// Materialized bitemporal key=value evolution lookup (0% stale exposure).
			if e.S == nil {
				notes = append(notes, "key_history: store unavailable.")
				continue
			}
			rows, kerr := e.S.QueryKeyHistory(target, maxSeq)
			if kerr != nil || len(rows) == 0 {
				notes = append(notes, fmt.Sprintf("key_history(%q) found no recorded assertions.", target))
				continue
			}
			var kb strings.Builder
			fmt.Fprintf(&kb, "Materialized key_history for %q (%d assertion(s), oldest first):\n", rows[0].Key, len(rows))
			for _, r := range rows {
				span := fmt.Sprintf("t%06d..current", r.ValidFromSeq)
				state := "ACTIVE"
				if r.ValidUntilSeq > 0 {
					span = fmt.Sprintf("t%06d..t%06d (superseded at t%06d)", r.ValidFromSeq, r.ValidUntilSeq, r.SupersededBySeq)
					state = "SUPERSEDED"
				}
				fmt.Fprintf(&kb, "- [%s] %s = %s (%s) · quote: %q\n", state, r.Key, r.RawValue, span, r.Quote)
			}
			extras = append(extras, packet.Item{
				Key:    packet.KeyExtra,
				ID:     fmt.Sprintf("recall-kv-%d", idx+1),
				Seq:    98 + idx,
				Header: fmt.Sprintf("Materialized key history for %q (bitemporal validity chain)", rows[0].Key),
				Text:   strings.TrimSpace(kb.String()),
			})
			notes = append(notes, fmt.Sprintf("key_history(%q) attached %d bitemporal state row(s).", rows[0].Key, len(rows)))

		case "bridge":
			// 2-hop rare-key Personalized PageRank bridge from seed turn(s).
			if e.S == nil {
				notes = append(notes, "bridge: store unavailable.")
				continue
			}
			var seeds []int
			for _, part := range strings.Split(target, ",") {
				if ref, ok := vault.ParseRef(strings.TrimSpace(part)); ok {
					if seq, ok2 := vault.TurnSeq(ref.Turn); ok2 {
						seeds = append(seeds, seq)
					}
				}
			}
			lim := op.Limit
			if lim <= 0 {
				lim = 4
			}
			bhits, berr := e.S.BridgeRareKeysPPR(seeds, maxSeq+1, lim)
			if berr != nil || len(bhits) == 0 {
				notes = append(notes, fmt.Sprintf("bridge(%q) found no 2-hop rare-key neighbors.", target))
				continue
			}
			var pinnedB []string
			for _, bh := range bhits {
				tid := vault.TurnID(bh.TurnSeq)
				if addID(tid) {
					pinnedB = append(pinnedB, fmt.Sprintf("%s (via %s)", tid, strings.Join(bh.ViaKeys, "+")))
				}
			}
			notes = append(notes, fmt.Sprintf("bridge(%q) pinned %d 2-hop rare-key turn(s): %s.", target, len(pinnedB), strings.Join(pinnedB, ", ")))

		case "cover_window":
			// Clarke-Cormack minimal contiguous multi-atom interval covering dispersed query terms.
			turns, terr := e.V.Turns()
			if terr != nil {
				notes = append(notes, fmt.Sprintf("cover_window failed: %v.", terr))
				continue
			}
			lim := op.Limit
			if lim <= 0 {
				lim = 3
			}
			wins := packet.BestCoverWindows(turns, target, maxSeq+1, max(2, op.Radius), lim)
			if len(wins) == 0 {
				notes = append(notes, fmt.Sprintf("cover_window(%q) found no multi-term span intervals.", target))
				continue
			}
			var pinnedW []string
			for _, w := range wins {
				if addID(w.ID) {
					pinnedW = append(pinnedW, fmt.Sprintf("%s (covers %s)", w.ID, strings.Join(w.CoveredTerms, "+")))
				}
			}
			notes = append(notes, fmt.Sprintf("cover_window(%q) pinned %d minimal interval(s): %s.", target, len(pinnedW), strings.Join(pinnedW, ", ")))

		default:
			notes = append(notes, fmt.Sprintf("Unknown recall_op %q (supported: exchange, open_block, timeline, grep, notes, list_omitted, check_omitted, key_history, bridge, cover_window).", op.Op))
		}
	}
	return add, extras, notes
}

func (e *Engine) openOmittedForOp(op packet.RecallOp, maxSeq int) (*packet.OmittedCatalog, packet.OmittedHandle, string, error) {
	if e.S == nil {
		return nil, packet.OmittedHandle{}, "", fmt.Errorf("store unavailable")
	}
	snapSeq, curRoot, err := e.S.SnapshotRoot(maxSeq)
	if err != nil {
		return nil, packet.OmittedHandle{}, "", err
	}
	var h packet.OmittedHandle
	raw := strings.TrimSpace(op.Target)
	if strings.HasPrefix(raw, "{") {
		if uerr := json.Unmarshal([]byte(raw), &h); uerr != nil {
			return nil, h, curRoot, fmt.Errorf("invalid omitted handle JSON: %w", uerr)
		}
	} else {
		indexID := strings.TrimSpace(op.Index)
		if indexID == "" {
			indexID = raw
		}
		if indexID == "" {
			return nil, h, curRoot, fmt.Errorf("provide index digest or omitted handle JSON")
		}
		snap := op.Snapshot
		if snap == "" {
			snap = curRoot
		}
		h = packet.OmittedHandle{
			Format:   "cx-omitted/v1",
			Snapshot: snap,
			Index:    indexID,
			Last:     snapSeq,
		}
		// Read exact metadata from the catalog sidecar if present.
		if metaBytes, merr := os.ReadFile(filepath.Join(e.V.Dir, ".cx", "catalogs", indexID+".json")); merr == nil {
			_ = json.Unmarshal(metaBytes, &h)
		}
	}
	cat, err := packet.OpenOmittedCatalog(filepath.Join(e.V.Dir, ".cx", "catalogs"), h, curRoot)
	return cat, h, curRoot, err
}

func (e *Engine) opListOmitted(op packet.RecallOp, maxSeq, idx int) (*packet.Item, string) {
	cat, h, _, err := e.openOmittedForOp(op, maxSeq)
	if err != nil {
		return nil, fmt.Sprintf("list_omitted failed: %v.", err)
	}
	defer cat.Close()
	limit := op.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 250 {
		limit = 250
	}
	cur := &packet.OmittedCursor{Index: h.Index, After: op.After}
	seqs, next, err := cat.Enumerate(cur, limit)
	if err != nil {
		return nil, fmt.Sprintf("list_omitted enumerate failed: %v.", err)
	}
	ids := make([]string, 0, len(seqs))
	for _, s := range seqs {
		ids = append(ids, vault.TurnID(s))
	}
	nextAfter := 0
	if next != nil {
		nextAfter = next.After
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Verified omitted catalog %s (snapshot %s, total omitted %d, page after=%d, returned %d, next_after=%d):\n",
		h.Index[:min(12, len(h.Index))], h.Snapshot[:min(12, len(h.Snapshot))], h.Omitted, op.After, len(ids), nextAfter)
	b.WriteString(strings.Join(ids, ", "))
	item := &packet.Item{
		Key:    packet.KeyExtra,
		ID:     fmt.Sprintf("recall-omitted-%d", idx+1),
		Seq:    96 + idx,
		Header: fmt.Sprintf("Omitted catalog page (%d turn IDs, next_after=%d)", len(ids), nextAfter),
		Text:   b.String(),
	}
	return item, fmt.Sprintf("list_omitted returned %d verified turn IDs (next_after=%d).", len(ids), nextAfter)
}

func (e *Engine) opCheckOmitted(op packet.RecallOp, maxSeq int) string {
	// Target may be a turn id ("t000042") or integer seq ("42"); Index carries the .cxi digest.
	seq := 0
	if s, ok := vault.TurnSeq(strings.TrimSpace(op.Target)); ok {
		seq = s
	} else if n, err := strconv.Atoi(strings.TrimSpace(op.Target)); err == nil {
		seq = n
	} else {
		return fmt.Sprintf("check_omitted %q: provide a turn id (tNNNNNN) or turn seq.", op.Target)
	}
	cat, h, _, err := e.openOmittedForOp(packet.RecallOp{Index: op.Index, Snapshot: op.Snapshot}, maxSeq)
	if err != nil {
		return fmt.Sprintf("check_omitted(%s) failed: %v.", op.Target, err)
	}
	defer cat.Close()
	ok, err := cat.Contains(seq)
	if err != nil {
		return fmt.Sprintf("check_omitted(%s) failed: %v.", op.Target, err)
	}
	return fmt.Sprintf("check_omitted(%s, index=%s): omitted=%v.", vault.TurnID(seq), h.Index[:min(12, len(h.Index))], ok)
}

func (e *Engine) buildTimelineItem(query string, beforeSeq, idx int) (packet.Item, bool) {
	qw := packet.ContentWords(query)
	ql := strings.ToLower(query)
	matches := func(s string) bool {
		sl := strings.ToLower(s)
		if strings.Contains(sl, ql) {
			return true
		}
		if len(qw) == 0 {
			return false
		}
		sw := packet.ContentWords(s)
		hits := 0
		for w := range qw {
			if sw[w] {
				hits++
			}
		}
		return hits >= max(1, (len(qw)+1)/2)
	}
	type entry struct {
		seq        int
		date       string
		kind       string
		id         string
		status     string
		source     string
		statement  string
		quote      string
		supersedes []string
	}
	turnDate := func(id string) (int, string) {
		seq, _ := vault.TurnSeq(id)
		if t, err := e.V.Turn(id); err == nil {
			return t.Seq, t.Created.Format("2006-01-02")
		}
		return seq, ""
	}
	var rows []entry
	for _, f := range e.V.Facts() {
		seq, date := turnDate(f.Source)
		if beforeSeq > 0 && seq >= beforeSeq {
			continue
		}
		if matches(f.Statement + " " + f.Quote) {
			st := f.Status
			if st == "" {
				st = "active"
			}
			rows = append(rows, entry{seq: seq, date: date, kind: "fact", id: f.ID, status: st, source: f.Source, statement: f.Statement, quote: f.Quote, supersedes: f.Supersedes})
		}
	}
	for _, d := range e.V.Decisions() {
		seq, date := turnDate(d.Source)
		if beforeSeq > 0 && seq >= beforeSeq {
			continue
		}
		if matches(d.Statement + " " + d.Quote) {
			st := d.Status
			if st == "" {
				st = "active"
			}
			rows = append(rows, entry{seq: seq, date: date, kind: "decision", id: d.ID, status: st, source: d.Source, statement: d.Statement, quote: d.Quote, supersedes: d.Supersedes})
		}
	}
	if len(rows) == 0 {
		return packet.Item{}, false
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].seq != rows[j].seq {
			return rows[i].seq < rows[j].seq
		}
		return rows[i].id < rows[j].id
	})
	var b strings.Builder
	fmt.Fprintf(&b, "Chronological ledger timeline for %q (%d entries, oldest first):\n", query, len(rows))
	for _, r := range rows {
		at := r.source
		if r.date != "" {
			at += " · " + r.date
		}
		sup := ""
		if len(r.supersedes) > 0 {
			sup = " [replaces " + strings.Join(r.supersedes, ", ") + "]"
		}
		fmt.Fprintf(&b, "- [%s] %s %s (%s)%s: %s\n  Quote: %q\n", r.status, r.kind, r.id, at, sup, r.statement, r.quote)
	}
	return packet.Item{
		Key:    packet.KeyExtra,
		ID:     fmt.Sprintf("recall-timeline-%d", idx+1),
		Seq:    91 + idx,
		Header: fmt.Sprintf("Timeline for %q (verbatim ledger entries, oldest first)", query),
		Text:   strings.TrimSpace(b.String()),
	}, true
}

func (e *Engine) grepVault(pattern string, beforeSeq, limit, idx int) ([]string, *packet.Item, error) {
	var re *regexp.Regexp
	literal := pattern
	isRegex := false
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") && len(pattern) > 2 {
		var err error
		re, err = regexp.Compile(pattern[1 : len(pattern)-1])
		if err != nil {
			return nil, nil, err
		}
		isRegex = true
	}
	// Load candidate turns: for literal patterns when FTS5 atom index is built,
	// prefilter using phrase search; otherwise scan turns in reverse (newest first).
	var candidates []vault.Turn
	if !isRegex && e.S != nil && e.S.HasValuePostings() {
		if phits, err := e.S.SearchAtomsPhraseE(literal, limit*8, beforeSeq); err == nil && len(phits) > 0 {
			seen := map[string]bool{}
			var ids []string
			for _, h := range phits {
				if !seen[h.Turn] {
					seen[h.Turn] = true
					ids = append(ids, h.Turn)
				}
			}
			if loaded, lerr := e.V.TurnsByIDs(ids); lerr == nil {
				for _, id := range ids {
					candidates = append(candidates, loaded[id])
				}
			}
		}
	}
	if len(candidates) == 0 {
		turns, err := e.V.Turns()
		if err != nil {
			return nil, nil, err
		}
		for i := len(turns) - 1; i >= 0; i-- {
			if beforeSeq > 0 && turns[i].Seq >= beforeSeq {
				continue
			}
			candidates = append(candidates, turns[i])
		}
	}
	lowLit := strings.ToLower(literal)
	var ids []string
	seenID := map[string]bool{}
	var lines []string
	for _, t := range candidates {
		if beforeSeq > 0 && t.Seq >= beforeSeq {
			continue
		}
		atoms := packet.AtomsOf(t)
		bodyLines := strings.Split(t.Body, "\n")
		offset := 0
		for lineNo, ln := range bodyLines {
			matched := false
			if isRegex {
				matched = re.MatchString(ln)
			} else {
				matched = strings.Contains(strings.ToLower(ln), lowLit)
			}
			if matched {
				ref := t.ID
				for _, a := range atoms {
					if offset >= a.Start && offset < a.End {
						if len(atoms) > 1 {
							ref = a.ID()
						}
						break
					}
				}
				if !seenID[ref] {
					seenID[ref] = true
					ids = append(ids, ref)
				}
				lines = append(lines, fmt.Sprintf("- %s line %d: %s", ref, lineNo+1, oneLine(ln, 160)))
				if len(lines) >= limit {
					break
				}
			}
			offset += len(ln) + 1
		}
		if len(lines) >= limit {
			break
		}
	}
	if len(ids) == 0 {
		return nil, nil, nil
	}
	item := &packet.Item{
		Key:    packet.KeyExtra,
		ID:     fmt.Sprintf("recall-grep-%d", idx+1),
		Seq:    95 + idx,
		Header: fmt.Sprintf("Grep matches for %q (%d lines)", pattern, len(lines)),
		Text:   strings.Join(lines, "\n"),
	}
	return ids, item, nil
}
