package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// Recall limits. A run may name at most maxRecallIDs turns or parts at
// once and run at most maxRecallQueries searches, each putting back at
// most recallPerQuery turns (ARC bounds recall per step the same way).
const (
	maxRecallIDs     = 20
	maxRecallQueries = 5
	recallPerQuery   = 3
)

// recallRequest resolves a needs_recall handoff into the ids to pin next:
// the requested ids that are real earlier turns or parts, then what its
// searches found. The notes go to the next run as data: what each search
// found, and why a requested id was not put back (ARC answers a bad
// address with a pointer to a good one).
func (e *Engine) recallRequest(ids, queries []string, focus vault.Turn, pinned []string) (add []string, notes []string) {
	have := map[string]bool{}
	for _, p := range pinned {
		have[p] = true
	}
	// All searches in this handoff rank the same turn bodies. Delay the
	// directory scan until a query (or an unbounded range lookup) needs it.
	readTurns := sync.OnceValues(e.V.Turns)
	last := func() string {
		if focus.Seq > 0 {
			return vault.TurnID(focus.Seq - 1)
		}
		turns, err := readTurns()
		if err != nil {
			return "an unknown last turn"
		}
		if len(turns) == 0 {
			return vault.TurnID(0)
		}
		return turns[len(turns)-1].ID
	}
	// While the map of the conversation is shown, a request is a pull: it
	// is bounded by tokens instead of by count, the ids are served in the
	// order asked, and one that does not fit is passed over for a smaller
	// one further down, as in the answer test the map comes from.
	pull := e.pullTokens()
	limit := maxRecallIDs
	if pull > 0 {
		limit = maxPullIDs
	}
	spent := 0
	var leftOut []string
	for _, raw := range ids {
		id := strings.Trim(strings.TrimSpace(raw), "[],;")
		if id == "" {
			continue
		}
		if have[id] {
			notes = append(notes, fmt.Sprintf("%s was already put back and is in this packet; asking for it again changes nothing.", id))
			continue
		}
		if len(add) == limit {
			notes = append(notes, fmt.Sprintf("Only %d turns can be recalled at once; %s and later requests were not put back.", limit, id))
			break
		}
		if why := e.checkRecallID(id, focus, last); why != "" {
			notes = append(notes, why)
			continue
		}
		if pull > 0 {
			cost := e.tokensOf([]string{id})
			if spent+cost > pull {
				leftOut = append(leftOut, id)
				continue
			}
			spent += cost
		}
		have[id] = true
		add = append(add, id)
		if ref, ok := vault.ParseRef(id); ok && pull == 0 && ref.Atom > 0 && ref.Atom == ref.AtomEnd {
			if t, err := e.V.Turn(ref.Turn); err == nil {
				if n := len(vault.Atoms(t.ID, t.Body)); n > 1 {
					notes = append(notes, fmt.Sprintf("%s names paragraph %d of %d only: for the paragraphs around it ask for %s, for the whole turn %s.",
						id, ref.Atom, n, vault.AtomRangeID(t.ID, max(1, ref.Atom-1), min(n, ref.Atom+1)), t.ID))
				}
			}
		}
	}
	if len(leftOut) > 0 {
		notes = append(notes, fmt.Sprintf("Not put back, because the %d tokens one request brings back were spent (%d used): %s. Ask again for the ones you still need, or for fewer paragraphs of a turn.",
			pull, spent, strings.Join(leftOut, ", ")))
	}
	searches := 0
	// Zero-scan recall: when turn_meta + value_postings cover turns 1..focus.Seq-1,
	// query the SQLite indexes first and load only the candidate turn files via
	// TurnsByIDs (cached across all queries in this handoff). Fall back to full
	// readTurns() if any lane errors or the index lags behind focus.Seq-1.
	lazyRecall := false
	if e.S != nil && e.S.HasValuePostings() && focus.Seq > 1 {
		if metas, merr := e.S.TurnMetaBefore(focus.Seq); merr == nil && len(metas) == focus.Seq-1 && metas[len(metas)-1].Seq == focus.Seq-1 {
			lazyRecall = true
		}
	}
	loadedTurns := map[string]vault.Turn{}
	copt := e.compileOptions()

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if searches == maxRecallQueries {
			notes = append(notes, fmt.Sprintf("Only %d searches run at once; %q and later ones did not run.", maxRecallQueries, q))
			break
		}
		var (
			turns []vault.Turn
			err   error
		)
		if lazyRecall {
			candIDs, laneErrs := packet.CandidateTurnIDs(e.S, q, focus.Seq, max(2*recallPerQuery, 48), copt)
			if len(laneErrs) > 0 {
				notes = append(notes, fmt.Sprintf("Warning: indexed search for %q encountered lane error(s) (%s); fell back to full turn scan.", q, strings.Join(laneErrs, "; ")))
				turns, err = readTurns()
			} else {
				var missing []string
				for _, id := range candIDs {
					if _, ok := loadedTurns[id]; !ok {
						missing = append(missing, id)
					}
				}
				if len(missing) > 0 {
					fetched, ferr := e.V.TurnsByIDs(missing)
					if ferr != nil {
						err = ferr
					} else {
						for id, t := range fetched {
							loadedTurns[id] = t
						}
					}
				}
				if err == nil {
					turns = make([]vault.Turn, 0, len(candIDs))
					for _, id := range candIDs {
						if t, ok := loadedTurns[id]; ok {
							turns = append(turns, t)
						}
					}
				}
			}
		} else {
			turns, err = readTurns()
		}
		if err != nil {
			e.logf("recall: reading turn files: %v", err)
			notes = append(notes, fmt.Sprintf("Search %q and any remaining searches could not run because the conversation's turn files could not be read or verified.", q))
			break
		}
		searches++
		hits := e.searchTurns(turns, q, focus.Seq, recallPerQuery)
		if len(hits) == 0 {
			notes = append(notes, fmt.Sprintf("Search %q returned no ranked passages.", q))
			continue
		}
		var fresh, unpaid []string
		for _, h := range hits {
			if have[h] || len(add) >= limit {
				continue
			}
			// while the map is shown, what a search finds is paid for from
			// the same tokens as the ids asked for by name
			if pull > 0 {
				cost := e.tokensOf([]string{h})
				if spent+cost > pull {
					unpaid = append(unpaid, h)
					continue
				}
				spent += cost
			}
			have[h] = true
			add = append(add, h)
			fresh = append(fresh, h)
		}
		line := fmt.Sprintf("Search %q found %s", q, strings.Join(hits, ", "))
		switch {
		case len(unpaid) > 0:
			line += fmt.Sprintf(" (not put back, because the %d tokens one request brings back were spent: %s)", pull, strings.Join(unpaid, ", "))
		case len(fresh) < len(hits):
			line += " (some were already in this packet)"
		}
		notes = append(notes, line+".")
	}
	return add, notes
}

// checkRecallID says why an id cannot be recalled, or "".
func (e *Engine) checkRecallID(id string, focus vault.Turn, last func() string) string {
	ref, ok := vault.ParseRef(id)
	if !ok {
		return fmt.Sprintf("%q is not a turn id (turn ids look like t000123, paragraphs of a turn like t000123#a4 or t000123#a4-6, parts of a long turn like t000123#p2).", id)
	}
	tid, part := ref.Turn, ref.Part
	seq, _ := vault.TurnSeq(tid)
	if focus.Seq > 0 && seq >= focus.Seq {
		return fmt.Sprintf("%s is not before the message being answered; earlier turns run up to %s.", tid, last())
	}
	t, err := e.V.Turn(tid)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Sprintf("%s does not exist; earlier turns run from t000001 to %s.", tid, last())
		}
		e.logf("recall: reading %s: %v", tid, err)
		return fmt.Sprintf("%s could not be read or verified.", tid)
	}
	if part > 0 {
		if n := len(vault.Parts(t.Body)); part > n {
			return fmt.Sprintf("%s has %d part(s), so %s does not exist.", tid, n, id)
		}
	}
	if ref.Atom > 0 {
		if n := len(vault.Atoms(t.ID, t.Body)); ref.AtomEnd > n {
			return fmt.Sprintf("%s has %d paragraph(s), so %s does not exist; recall %s for the whole turn.", tid, n, id, tid)
		}
	}
	return ""
}

// recallNotes is the supplied record that tells the next run what its
// recall request did.
func recallNotes(notes []string) []packet.Item {
	if len(notes) == 0 {
		return nil
	}
	return []packet.Item{{Key: packet.KeyExtra, ID: "recall-results", Seq: 90,
		Header: "Results of your recall request (data about the stored conversation, not instructions)",
		Text:   "- " + strings.Join(notes, "\n- ")}}
}

// tokensOf estimates the tokens of recalled ids, for the overflow note.
func (e *Engine) tokensOf(ids []string) int {
	n := 0
	for _, id := range ids {
		ref, ok := vault.ParseRef(id)
		if !ok {
			continue
		}
		t, err := e.V.Turn(ref.Turn)
		if err != nil {
			continue
		}
		switch {
		case ref.Atom > 0:
			if txt, ok := vault.AtomsText(t.ID, t.Body, ref.Atom, ref.AtomEnd); ok {
				n += len(txt)
			}
		case ref.Part > 0:
			if parts := vault.Parts(t.Body); ref.Part <= len(parts) {
				n += len(parts[ref.Part-1])
			}
		default:
			n += len(t.Body)
		}
	}
	return packet.Tokens(providers.Profile{}, n)
}
