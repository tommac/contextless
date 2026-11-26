package packet

import (
	"math"
	"sort"

	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// PickAtomsAdaptive implements the empirically validated conditional packing
// policy from docs/experiments/longmem/excerpt_growth.py:
//
//  1. Rank candidate atoms across lanes (plus optional working-set locality)
//     and interleave thread/turn-level matches.
//  2. When the top candidate pool fits inside budget, deliver whole atoms
//     cleanly (preserving full context as on LoCoMo).
//  3. When the pool overflows budget, visit distinct source turns first in
//     round-robin order (source diversity), seed each turn with its highest
//     IDF-weighted query-term atom, and grow contiguous neighbor atoms
//     outward in alternating left/right rounds until budget is spent.
//
// Adjacent selected atoms of the same turn merge into single AtomRuns so
// headers are never duplicated and every byte remains 100% verbatim.
func PickAtomsAdaptive(turns []vault.Turn, s *store.Store, query string, beforeSeq, budget int, have func(turn string, k int) bool, opt CompileOptions) (runs []AtomRun, hit map[string]float64) {
	hit = map[string]float64{}
	if budget <= 0 {
		return nil, hit
	}
	byID := make(map[string]vault.Turn, len(turns))
	for _, t := range turns {
		byID[t.ID] = t
	}
	diag := RankAtomsDiag(turns, s, query, beforeSeq, max(atomCandidates, budget/40), opt)
	ranked := diag.Fused
	if len(ranked) == 0 {
		return nil, hit
	}
	tokens := func(n int) int { return Tokens(providers.Profile{}, n) }
	headerCost := func(t vault.Turn) int {
		var replyTo *vault.Turn
		if q, ok := byID[t.ReplyTo]; ok {
			replyTo = &q
		}
		return tokens(len(AtomHeader(t, 1, 1, len(atomsOf(t)), replyTo)) + 24)
	}

	// Build source-diverse candidate pool (round-robin across turns, up to 24
	// top candidates) and check whether the entire pool fits whole.
	const maxPool = 24
	byTurn := map[string][]RankedAtom{}
	var turnOrder []string
	seenText := map[string]bool{}
	for _, r := range ranked {
		if r.CardOnly {
			if _, ok := byTurn[r.Turn.ID]; !ok {
				turnOrder = append(turnOrder, r.Turn.ID)
			}
			byTurn[r.Turn.ID] = append(byTurn[r.Turn.ID], r)
			continue
		}
		txt := r.Atom.Text(r.Turn.Body)
		if have != nil && have(r.Turn.ID, r.Atom.K) {
			if r.Score > hit[r.Turn.ID] {
				hit[r.Turn.ID] = r.Score
			}
			continue
		}
		if len(txt) >= atomRepeatBytes && seenText[txt] {
			continue
		}
		seenText[txt] = true
		if _, ok := byTurn[r.Turn.ID]; !ok {
			turnOrder = append(turnOrder, r.Turn.ID)
		}
		byTurn[r.Turn.ID] = append(byTurn[r.Turn.ID], r)
	}

	// Round-robin interleave: one candidate per distinct turn per pass.
	var diverse []RankedAtom
	for pass := 0; len(diverse) < maxPool; pass++ {
		added := false
		for _, tid := range turnOrder {
			list := byTurn[tid]
			if pass < len(list) {
				diverse = append(diverse, list[pass])
				added = true
				if len(diverse) >= maxPool {
					break
				}
			}
		}
		if !added {
			break
		}
	}

	// Compute IDF weights of query terms across the candidate turns so anchor
	// seeding prioritizes distinctive terms over common ones.
	qw := ContentWords(query)
	df := map[string]int{}
	for _, tid := range turnOrder {
		t := byID[tid]
		tw := ContentWords(t.Body)
		for w := range qw {
			if tw[w] {
				df[w]++
			}
		}
	}
	idf := map[string]float64{}
	for w := range qw {
		idf[w] = 1 + math.Log(float64(1+len(turnOrder))/float64(1+df[w]))
	}
	atomRelevance := func(t vault.Turn, a vault.Atom) float64 {
		tw := ContentWords(a.Text(t.Body))
		sum := 0.0
		for w, wt := range idf {
			if tw[w] {
				sum += wt
			}
		}
		return sum
	}

	// Seed phase: pick one highest-utility anchor atom per distinct turn in
	// rank order so broad multi-turn questions never starve behind one long turn.
	type spanState struct {
		turn  vault.Turn
		atoms []vault.Atom
		lo    int // 1-based inclusive
		hi    int // 1-based inclusive
		score float64
		step  int // alternates left/right outward growth
	}
	chosen := map[string]map[int]float64{}
	var states []*spanState
	stateByTurn := map[string]*spanState{}
	used := 0

	costAtom := func(t vault.Turn, a vault.Atom) int {
		c := tokens(len(a.Text(t.Body)) + 2)
		_, left := chosen[t.ID][a.K-1]
		_, right := chosen[t.ID][a.K+1]
		switch {
		case !left && !right:
			c += headerCost(t)
		case left && right:
			c -= headerCost(t)
		}
		return c
	}
	takeAtom := func(t vault.Turn, a vault.Atom, sc float64) bool {
		if have != nil && have(t.ID, a.K) {
			return false
		}
		if _, ok := chosen[t.ID][a.K]; ok {
			return false
		}
		c := costAtom(t, a)
		if used+c > budget {
			return false
		}
		used += c
		if chosen[t.ID] == nil {
			chosen[t.ID] = map[int]float64{}
		}
		chosen[t.ID][a.K] = sc
		return true
	}

	// Pass 1: seed distinct turns using their best query-anchored atom.
	for _, tid := range turnOrder {
		t := byID[tid]
		atoms := atomsOf(t)
		if len(atoms) == 0 {
			continue
		}
		list := byTurn[tid]
		best := list[0]
		if best.CardOnly {
			// Card-only match: try whole turn if it fits, else seed first atom.
			whole := tokens(len(t.Body)) + headerCost(t)
			if used+whole <= budget {
				for _, a := range atoms {
					takeAtom(t, a, best.Score)
				}
				continue
			}
		}
		bestRel := atomRelevance(t, best.Atom)
		for _, cand := range list[1:] {
			if rel := atomRelevance(t, cand.Atom); rel > bestRel {
				bestRel = rel
				best = cand
			}
		}
		if takeAtom(t, best.Atom, best.Score) {
			st := &spanState{turn: t, atoms: atoms, lo: best.Atom.K, hi: best.Atom.K, score: best.Score}
			states = append(states, st)
			stateByTurn[tid] = st
		}
	}

	// Pass 2: add secondary high-ranked atoms from the round-robin pool if room remains.
	for _, r := range diverse {
		if r.CardOnly {
			continue
		}
		if takeAtom(r.Turn, r.Atom, r.Score) {
			if st := stateByTurn[r.Turn.ID]; st != nil {
				st.lo = min(st.lo, r.Atom.K)
				st.hi = max(st.hi, r.Atom.K)
			}
		}
	}

	// Pass 3: grow each seeded span outward (alternating left/right neighbors)
	// in rounds so cut-off answers and surrounding context are recovered without
	// any single turn monopolizing the budget.
	for len(states) > 0 {
		grew := false
		for _, st := range states {
			n := len(st.atoms)
			var tryOrder []int
			if st.step%2 == 0 {
				if st.lo > 1 {
					tryOrder = append(tryOrder, st.lo-1)
				}
				if st.hi < n {
					tryOrder = append(tryOrder, st.hi+1)
				}
			} else {
				if st.hi < n {
					tryOrder = append(tryOrder, st.hi+1)
				}
				if st.lo > 1 {
					tryOrder = append(tryOrder, st.lo-1)
				}
			}
			st.step++
			for _, k := range tryOrder {
				if takeAtom(st.turn, st.atoms[k-1], st.score*0.95) {
					if k < st.lo {
						st.lo = k
					}
					if k > st.hi {
						st.hi = k
					}
					grew = true
					break
				}
			}
		}
		if !grew {
			break
		}
	}

	for id, ks := range chosen {
		t := byID[id]
		n := len(atomsOf(t))
		for _, x := range JoinAtoms(ks) {
			run := AtomRun{Turn: t, A: x[0], B: x[1], N: n}
			for k := x[0]; k <= x[1]; k++ {
				run.Score = max(run.Score, ks[k])
			}
			runs = append(runs, run)
		}
	}
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Turn.Seq != runs[j].Turn.Seq {
			return runs[i].Turn.Seq < runs[j].Turn.Seq
		}
		return runs[i].A < runs[j].A
	})
	return runs, hit
}
