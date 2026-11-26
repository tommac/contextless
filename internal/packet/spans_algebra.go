package packet

import (
	"math"
	"sort"

	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// Region-algebra cover-density contiguous span windows (Clarke, Cormack &
// Burkowski, The Computer Journal 1995, lifted from token positions to
// conversation atoms).
//
// When a technical answer or multi-constraint requirement is dispersed across
// 2-4 consecutive paragraphs of a turn, no single paragraph matches all query
// terms, so top-1 atom ranking under-scores the passage. CoverWindow finds
// minimal contiguous atom intervals [AtomStart..AtomEnd] whose union covers
// >= 2 distinct query stems with neither endpoint redundant, ranked by
// (IDF-weighted covered mass)^2 / interval_tokens.

// CoverWindow is a minimal contiguous run of atoms covering multiple query stems.
type CoverWindow struct {
	Turn         vault.Turn `json:"-"`
	ID           string     `json:"id"`
	TurnID       string     `json:"turn_id"`
	AtomStart    int        `json:"atom_start"`
	AtomEnd      int        `json:"atom_end"`
	Tokens       int        `json:"tokens"`
	CoveredTerms []string   `json:"covered_terms"`
	DensityScore float64    `json:"density_score"`
}

// BestCoverWindows finds up to topK minimal multi-term contiguous atom windows
// (up to maxWindowAtoms wide, default 4) across turns strictly before beforeSeq.
func BestCoverWindows(turns []vault.Turn, query string, beforeSeq, maxWindowAtoms, topK int) []CoverWindow {
	qw := ContentWords(query)
	if len(qw) < 2 {
		return nil
	}
	if maxWindowAtoms <= 0 {
		maxWindowAtoms = 4
	}
	if topK <= 0 {
		topK = 6
	}

	// Compute corpus IDF across candidate turns.
	df := map[string]int{}
	nTurns := 0
	for _, t := range turns {
		if beforeSeq > 0 && t.Seq >= beforeSeq {
			continue
		}
		nTurns++
		tw := ContentWords(t.Body)
		for w := range qw {
			if tw[w] {
				df[w]++
			}
		}
	}
	if nTurns == 0 {
		return nil
	}
	idf := map[string]float64{}
	for w := range qw {
		idf[w] = 1 + math.Log(float64(1+nTurns)/float64(1+df[w]))
	}

	var windows []CoverWindow
	for _, t := range turns {
		if beforeSeq > 0 && t.Seq >= beforeSeq {
			continue
		}
		atoms := atomsOf(t)
		if len(atoms) == 0 {
			continue
		}
		// Precompute query stems and token cost per atom.
		atomTerms := make([]map[string]bool, len(atoms))
		atomToks := make([]int, len(atoms))
		anyHit := 0
		for i, a := range atoms {
			txt := a.Text(t.Body)
			atomToks[i] = max(1, Tokens(providers.Profile{}, len(txt)))
			tw := ContentWords(txt)
			m := map[string]bool{}
			for w := range qw {
				if tw[w] {
					m[w] = true
				}
			}
			if len(m) > 0 {
				anyHit++
			}
			atomTerms[i] = m
		}
		if anyHit == 0 {
			continue
		}

		for i := 0; i < len(atoms); i++ {
			if len(atomTerms[i]) == 0 {
				continue
			}
			maxJ := min(len(atoms)-1, i+maxWindowAtoms-1)
			for j := i; j <= maxJ; j++ {
				if len(atomTerms[j]) == 0 {
					continue
				}
				// Check minimality (Clarke-Cormack rule): removing endpoint i or j
				// must strictly shrink the covered query-term set!
				union := map[string]bool{}
				tokSum := 0
				for k := i; k <= j; k++ {
					tokSum += atomToks[k]
					for w := range atomTerms[k] {
						union[w] = true
					}
				}
				if len(union) < 2 {
					continue
				}
				if j > i {
					withoutLeft := map[string]bool{}
					withoutRight := map[string]bool{}
					for k := i + 1; k <= j; k++ {
						for w := range atomTerms[k] {
							withoutLeft[w] = true
						}
					}
					for k := i; k <= j-1; k++ {
						for w := range atomTerms[k] {
							withoutRight[w] = true
						}
					}
					if len(withoutLeft) == len(union) || len(withoutRight) == len(union) {
						continue // non-minimal interval
					}
				}
				mass := 0.0
				covList := make([]string, 0, len(union))
				for w := range union {
					mass += idf[w]
					covList = append(covList, w)
				}
				sort.Strings(covList)
				// Reward multi-term coverage quadratically, normalized by soft token cost.
				density := (mass * mass) / math.Pow(float64(tokSum+20), 0.65)
				windows = append(windows, CoverWindow{
					Turn:         t,
					ID:           vault.AtomRangeID(t.ID, atoms[i].K, atoms[j].K),
					TurnID:       t.ID,
					AtomStart:    atoms[i].K,
					AtomEnd:      atoms[j].K,
					Tokens:       tokSum,
					CoveredTerms: covList,
					DensityScore: math.Round(density*10000) / 10000,
				})
			}
		}
	}

	sort.Slice(windows, func(i, j int) bool {
		if windows[i].DensityScore != windows[j].DensityScore {
			return windows[i].DensityScore > windows[j].DensityScore
		}
		if len(windows[i].CoveredTerms) != len(windows[j].CoveredTerms) {
			return len(windows[i].CoveredTerms) > len(windows[j].CoveredTerms)
		}
		return windows[i].Turn.Seq > windows[j].Turn.Seq
	})

	// Deduplicate overlapping windows in the same turn (keep highest-density window).
	var out []CoverWindow
	for _, w := range windows {
		overlaps := false
		for _, kept := range out {
			if kept.TurnID == w.TurnID && !(w.AtomEnd < kept.AtomStart || w.AtomStart > kept.AtomEnd) {
				overlaps = true
				break
			}
		}
		if !overlaps {
			out = append(out, w)
			if len(out) >= topK {
				break
			}
		}
	}
	return out
}
