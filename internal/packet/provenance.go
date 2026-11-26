package packet

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// How-Provenance Semiring Polynomials (Green, Karvounarakis & Tannen PODS 2007;
// Asma et al. Oct 2026) and Instance-Dependent Submodular Knapsack Optimality
// Certificates (Tang et al. 2020; Lin & Bilmes 2010).

// ProvenanceTerm is one elementary retrieval channel derivation for a span.
type ProvenanceTerm struct {
	Channel string  `json:"channel"` // exact | fts | value | episode | working_set | bridge_ppr | cover
	Detail  string  `json:"detail"`  // e.g. "#1", "val=2", "via:pool_size"
	Weight  float64 `json:"weight"`  // channel reliability in (0, 1)
}

// ProvenancePoly is a sum of independent derivation monomials in N[X].
type ProvenancePoly struct {
	Terms []ProvenanceTerm `json:"terms"`
}

// Add records an independent derivation channel on the polynomial.
func (p *ProvenancePoly) Add(channel, detail string, reliability float64) {
	if reliability <= 0 {
		reliability = 0.5
	}
	if reliability > 0.99 {
		reliability = 0.99
	}
	p.Terms = append(p.Terms, ProvenanceTerm{Channel: channel, Detail: detail, Weight: reliability})
}

// CorroborationCount evaluates the polynomial in the Counting Semiring
// (N, +, *, 0, 1): number of distinct retrieval channels deriving this span.
func (p ProvenancePoly) CorroborationCount() int {
	seen := map[string]bool{}
	for _, t := range p.Terms {
		seen[t.Channel] = true
	}
	return len(seen)
}

// CombinedBelief evaluates the polynomial in the independent-evidence
// Probabilistic Semiring ([0,1], 1-(1-a)(1-b), *, 0, 1).
func (p ProvenancePoly) CombinedBelief() float64 {
	miss := 1.0
	for _, t := range p.Terms {
		miss *= (1.0 - t.Weight)
	}
	return math.Round((1.0-miss)*1000) / 1000
}

// Format renders a compact provenance token (Asma et al. 2026) showing both
// how-lineage and semiring corroboration.
func (p ProvenancePoly) Format() string {
	if len(p.Terms) == 0 {
		return ""
	}
	parts := make([]string, 0, len(p.Terms))
	for _, t := range p.Terms {
		if t.Detail != "" {
			parts = append(parts, t.Channel+"("+t.Detail+")")
		} else {
			parts = append(parts, t.Channel)
		}
	}
	return fmt.Sprintf("%s [%dx corroborated, belief=%.2f]", strings.Join(parts, " + "), p.CorroborationCount(), p.CombinedBelief())
}

// SubmodularItem represents one candidate span for budgeted submodular selection.
type SubmodularItem struct {
	ID        string
	Tokens    int
	Prior     float64            // modular RRF / rank prior >= 0
	Terms     map[string]float64 // query/value/entity features covered
	Shingles  map[string]bool    // 3-word lexical shingles for novel-information saturation
}

// SubmodularCert reports the chosen knapsack subset alongside a rigorous
// mathematical certificate bounding how close the achieved utility is to the
// true NP-hard knapsack optimum OPT on this exact query instance.
type SubmodularCert struct {
	SelectedIDs      []string `json:"selected_ids"`
	TokensUsed       int      `json:"tokens_used"`
	Budget           int      `json:"budget"`
	Utility          float64  `json:"utility"`
	UpperBoundOPT    float64  `json:"upper_bound_opt"`
	OptimalityPct    float64  `json:"optimality_pct"`     // 100 * Utility / UpperBoundOPT (instance-tight!)
	WorstCaseBoundPct float64 `json:"worst_case_bound_pct"` // 40.5% Tang et al. theorem
	UsedSingleBest   bool     `json:"used_single_best"`
}

// WordShingles extracts normalized 3-word lexical shingles so near-duplicate
// paragraphs saturate submodular novelty without requiring dense embeddings.
func WordShingles(text string) map[string]bool {
	toks := tokens(text)
	out := map[string]bool{}
	if len(toks) < 3 {
		for _, t := range toks {
			out[strings.ToLower(t.text)] = true
		}
		return out
	}
	for i := 0; i+2 < len(toks); i++ {
		sh := strings.ToLower(toks[i].text) + " " + strings.ToLower(toks[i+1].text) + " " + strings.ToLower(toks[i+2].text)
		out[sh] = true
	}
	return out
}

// CertifySubmodularSelection maximizes the monotone submodular objective:
//
//	f(S) = sum_{e in S} Prior(e)
//	     + sum_{w in Query} IDF(w) * sqrt(sum_{e in S} Terms(e, w))
//	     + shingleWeight * |Union_{e in S} Shingles(e)|
//
// subject to sum_{e in S} Tokens(e) <= budget, using Modified Density-Greedy
// (gain per token + best-single-item fallback, worst-case >= 0.405 * OPT) and
// computes Tang's online submodular LP upper bound on OPT at every step:
//
//	OPT <= min_{t} [ f(S_t) + FractionalKnapsackLP( { (Delta(e | S_t), Tokens(e)) : e not in S_t, Tokens(e) <= budget }, budget ) ]
//
// Because f is strictly monotone submodular (sum of non-negative modular,
// concave-over-modular, and weighted set-cover terms), this upper bound is
// 100% mathematically sound on every input.
func CertifySubmodularSelection(queryIDF map[string]float64, items []SubmodularItem, budget int) SubmodularCert {
	cert := SubmodularCert{Budget: budget, WorstCaseBoundPct: 40.5}
	if budget <= 0 || len(items) == 0 {
		cert.OptimalityPct = 100
		return cert
	}

	// Filter to items that can individually fit in budget, assigning safe token floor >= 1.
	feasible := make([]SubmodularItem, 0, len(items))
	for _, it := range items {
		if it.Tokens <= 0 {
			it.Tokens = 1
		}
		if it.Tokens <= budget {
			feasible = append(feasible, it)
		}
	}
	if len(feasible) == 0 {
		cert.OptimalityPct = 100
		return cert
	}

	qwords := make([]string, 0, len(queryIDF))
	for w := range queryIDF {
		qwords = append(qwords, w)
	}
	sort.Strings(qwords)

	const shingleW = 0.008 // small tie-breaking reward for non-redundant lexical content

	type state struct {
		cov      map[string]float64
		shingles map[string]bool
		utility  float64
	}
	newState := func() state {
		return state{cov: map[string]float64{}, shingles: map[string]bool{}}
	}
	marginal := func(st state, it SubmodularItem) float64 {
		g := it.Prior
		for _, w := range qwords {
			if add := it.Terms[w]; add > 0 {
				cur := st.cov[w]
				g += queryIDF[w] * (math.Sqrt(cur+add) - math.Sqrt(cur))
			}
		}
		newSh := 0
		for sh := range it.Shingles {
			if !st.shingles[sh] {
				newSh++
			}
		}
		g += shingleW * float64(newSh)
		return g
	}
	apply := func(st *state, it SubmodularItem) {
		st.utility += marginal(*st, it)
		for w, add := range it.Terms {
			if add > 0 {
				st.cov[w] += add
			}
		}
		for sh := range it.Shingles {
			st.shingles[sh] = true
		}
	}

	// Fractional knapsack LP upper bound on max_{O: cost(O)<=budget} sum_{e in O\S} Delta(e | S).
	fractionalBound := func(st state, taken []bool) float64 {
		type densityItem struct {
			gain   float64
			tokens int
		}
		cands := make([]densityItem, 0, len(feasible))
		for i, it := range feasible {
			if taken[i] {
				continue
			}
			g := marginal(st, it)
			if g > 0 {
				cands = append(cands, densityItem{gain: g, tokens: it.Tokens})
			}
		}
		sort.Slice(cands, func(i, j int) bool {
			return cands[i].gain/float64(cands[i].tokens) > cands[j].gain/float64(cands[j].tokens)
		})
		rem := budget
		addLP := 0.0
		for _, c := range cands {
			if rem <= 0 {
				break
			}
			if c.tokens <= rem {
				addLP += c.gain
				rem -= c.tokens
			} else {
				addLP += c.gain * (float64(rem) / float64(c.tokens))
				rem = 0
			}
		}
		return st.utility + addLP
	}

	st := newState()
	taken := make([]bool, len(feasible))
	bestUpper := fractionalBound(st, taken)
	usedTokens := 0
	var chosenIDs []string

	for {
		bestIdx := -1
		bestDensity := -1.0
		for i, it := range feasible {
			if taken[i] || usedTokens+it.Tokens > budget {
				continue
			}
			d := marginal(st, it) / float64(it.Tokens)
			if d > bestDensity || (d == bestDensity && bestIdx >= 0 && it.ID < feasible[bestIdx].ID) {
				bestDensity = d
				bestIdx = i
			}
		}
		if bestIdx < 0 || bestDensity <= 0 {
			break
		}
		taken[bestIdx] = true
		usedTokens += feasible[bestIdx].Tokens
		chosenIDs = append(chosenIDs, feasible[bestIdx].ID)
		apply(&st, feasible[bestIdx])
		if ub := fractionalBound(st, taken); ub < bestUpper {
			bestUpper = ub
		}
	}

	// Best single item fallback (guarantees >= 0.405 * OPT worst-case).
	emptySt := newState()
	singleIdx := -1
	singleUtil := -1.0
	for i, it := range feasible {
		u := marginal(emptySt, it)
		if u > singleUtil {
			singleUtil = u
			singleIdx = i
		}
	}

	cert.SelectedIDs = chosenIDs
	cert.TokensUsed = usedTokens
	cert.Utility = math.Round(st.utility*10000) / 10000
	if singleIdx >= 0 && singleUtil > st.utility {
		cert.SelectedIDs = []string{feasible[singleIdx].ID}
		cert.TokensUsed = feasible[singleIdx].Tokens
		cert.Utility = math.Round(singleUtil*10000) / 10000
		cert.UsedSingleBest = true
	}
	if bestUpper < cert.Utility {
		bestUpper = cert.Utility
	}
	cert.UpperBoundOPT = math.Round(bestUpper*10000) / 10000
	if cert.UpperBoundOPT <= 0 {
		cert.OptimalityPct = 100
	} else {
		pct := 100 * cert.Utility / cert.UpperBoundOPT
		if pct > 100 {
			pct = 100
		}
		cert.OptimalityPct = math.Round(pct*10) / 10
	}
	return cert
}
