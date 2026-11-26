package packet

import (
	"math"
	"sort"
	"time"
)

// PrefixCandidate is one reusable cache-prefix boundary P_k from an earlier
// call's prompt (e.g., rules+session, +catalog spine, +aligned recent epoch).
type PrefixCandidate struct {
	Name      string   `json:"name"`
	Tokens    int      `json:"tokens"`
	AtomIDs   []string `json:"atom_ids,omitempty"`
	TermCount map[string]float64
}

// SuffixCandidate is one atom that can be appended after the chosen prefix at
// fresh (uncached/write) input price.
type SuffixCandidate struct {
	ID        string
	Tokens    int
	RankPrior float64
	Terms     map[string]float64
}

// PrefixOptInput specifies one instance of the prefix-closed budgeted
// selection problem:
//
//	maximize  f(P_k ∪ S)
//	subject to  ReadPrice * Tokens(P_k) + FreshPrice * Tokens(S) <= BilledBudget
//
// where f is monotone submodular (saturating IDF-weighted query-term coverage
// plus diminishing rank prior). For each k, gain-per-billed-token greedy with
// best-single-item fallback achieves at least 0.405 of optimal; taking the
// best k preserves the 0.405 factor overall.
type PrefixOptInput struct {
	QueryIDF     map[string]float64
	Prefixes     []PrefixCandidate
	Suffixes     []SuffixCandidate
	BilledBudget float64
	ReadPrice    float64 // e.g., 0.10
	FreshPrice   float64 // e.g., 1.00 or 1.25
}

// PrefixOptResult is the winning prefix boundary and suffix selection.
type PrefixOptResult struct {
	PrefixName   string   `json:"prefix_name"`
	PrefixIndex  int      `json:"prefix_index"`
	PrefixTokens int      `json:"prefix_tokens"`
	SuffixIDs    []string `json:"suffix_ids"`
	SuffixTokens int      `json:"suffix_tokens"`
	BilledTokens float64  `json:"billed_tokens"`
	Utility      float64  `json:"utility"`
	UsedFallback bool     `json:"used_fallback"`
}

// OptimizePrefixClosed solves the prefix-closed budgeted selection problem
// deterministically under a monotone submodular saturating coverage objective.
func OptimizePrefixClosed(in PrefixOptInput) PrefixOptResult {
	readP := in.ReadPrice
	if readP <= 0 {
		readP = 0.10
	}
	freshP := in.FreshPrice
	if freshP <= 0 {
		freshP = 1.0
	}
	prefixes := in.Prefixes
	if len(prefixes) == 0 {
		prefixes = []PrefixCandidate{{Name: "empty", Tokens: 0}}
	}
	// Deterministic sorted query words for exact float summation.
	qwords := make([]string, 0, len(in.QueryIDF))
	for w := range in.QueryIDF {
		qwords = append(qwords, w)
	}
	sort.Strings(qwords)

	// Saturating monotone submodular coverage: sum_w idf[w] * sqrt(cov[w])
	// (concave, strictly non-decreasing, diminishing marginal gains).
	evalCoverage := func(cov map[string]float64, priorSum float64) float64 {
		u := priorSum
		for _, w := range qwords {
			if c := cov[w]; c > 0 {
				u += in.QueryIDF[w] * math.Sqrt(c)
			}
		}
		return u
	}
	marginalGain := func(cov map[string]float64, s SuffixCandidate) float64 {
		g := s.RankPrior
		for _, w := range qwords {
			if add := s.Terms[w]; add > 0 {
				cur := cov[w]
				g += in.QueryIDF[w] * (math.Sqrt(cur+add) - math.Sqrt(cur))
			}
		}
		return g
	}

	var best PrefixOptResult
	bestSet := false

	for k, pk := range prefixes {
		prefixCost := float64(pk.Tokens) * readP
		if prefixCost > in.BilledBudget {
			continue
		}
		rem := in.BilledBudget - prefixCost
		inPrefix := map[string]bool{}
		for _, id := range pk.AtomIDs {
			inPrefix[id] = true
		}

		// Candidate 1: greedy gain-per-billed-token.
		cov := map[string]float64{}
		for w, c := range pk.TermCount {
			cov[w] = c
		}
		chosen := map[int]bool{}
		var greedyIDs []string
		greedyToks := 0
		greedyBilled := 0.0
		greedyPrior := 0.0
		for {
			bestIdx := -1
			bestRatio := 0.0
			bestGain := 0.0
			for i, s := range in.Suffixes {
				if chosen[i] || inPrefix[s.ID] || s.Tokens <= 0 {
					continue
				}
				scost := float64(s.Tokens) * freshP
				if greedyBilled+scost > rem {
					continue
				}
				g := marginalGain(cov, s)
				ratio := g / scost
				if ratio > bestRatio || (ratio == bestRatio && bestIdx >= 0 && s.ID < in.Suffixes[bestIdx].ID) {
					bestIdx = i
					bestRatio = ratio
					bestGain = g
				}
			}
			if bestIdx < 0 || bestGain <= 0 {
				break
			}
			s := in.Suffixes[bestIdx]
			chosen[bestIdx] = true
			greedyIDs = append(greedyIDs, s.ID)
			greedyToks += s.Tokens
			greedyBilled += float64(s.Tokens) * freshP
			greedyPrior += s.RankPrior
			for w, c := range s.Terms {
				cov[w] += c
			}
		}
		greedyUtil := evalCoverage(cov, greedyPrior)

		// Candidate 2: best single suffix item fallback (preserves 0.405 bound).
		baseCov := map[string]float64{}
		for w, c := range pk.TermCount {
			baseCov[w] = c
		}
		singleIdx := -1
		singleUtil := evalCoverage(baseCov, 0)
		for i, s := range in.Suffixes {
			if inPrefix[s.ID] || s.Tokens <= 0 {
				continue
			}
			if float64(s.Tokens)*freshP > rem {
				continue
			}
			u := evalCoverage(baseCov, 0) + marginalGain(baseCov, s)
			if u > singleUtil {
				singleUtil = u
				singleIdx = i
			}
		}

		res := PrefixOptResult{
			PrefixName:   pk.Name,
			PrefixIndex:  k,
			PrefixTokens: pk.Tokens,
			SuffixIDs:    greedyIDs,
			SuffixTokens: greedyToks,
			BilledTokens: prefixCost + greedyBilled,
			Utility:      greedyUtil,
		}
		if singleIdx >= 0 && singleUtil > greedyUtil {
			s := in.Suffixes[singleIdx]
			res.SuffixIDs = []string{s.ID}
			res.SuffixTokens = s.Tokens
			res.BilledTokens = prefixCost + float64(s.Tokens)*freshP
			res.Utility = singleUtil
			res.UsedFallback = true
		}
		if !bestSet || res.Utility > best.Utility || (res.Utility == best.Utility && res.BilledTokens < best.BilledTokens) {
			best = res
			bestSet = true
		}
	}
	return best
}

// CachePolicyAdvice summarizes the online rent-or-buy caching decision from
// a workspace's empirical inter-message gaps.
type CachePolicyAdvice struct {
	SampleGaps         int     `json:"sample_gaps"`
	MedianGapSeconds   int64   `json:"median_gap_seconds"`
	HitRate5m          float64 `json:"hit_rate_5m"`
	HitRate30m         float64 `json:"hit_rate_30m"`
	ExpectedMultiplier float64 `json:"expected_multiplier"` // billed prefix price relative to 1.0 uncached
	RecommendStable    bool    `json:"recommend_stable"`
	Reason             string  `json:"reason"`
}

// AdviseCachePolicy evaluates whether keeping a stable prompt prefix (catalog
// spine + epoch alignment + stable cwd) pays off given recent inter-message
// gaps, assuming a 5-minute TTL at 1.25x write / 0.10x read (or 30-minute TTL).
func AdviseCachePolicy(gaps []time.Duration, readPrice, writePrice float64, ttl time.Duration) CachePolicyAdvice {
	if readPrice <= 0 {
		readPrice = 0.10
	}
	if writePrice <= 0 {
		writePrice = 1.25
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if len(gaps) == 0 {
		return CachePolicyAdvice{
			ExpectedMultiplier: 1.0,
			Reason:             "no inter-message timing history yet; default sliding window",
		}
	}
	sorted := append([]time.Duration(nil), gaps...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	hits, hits5, hits30 := 0, 0, 0
	for _, g := range sorted {
		if g <= ttl {
			hits++
		}
		if g <= 5*time.Minute {
			hits5++
		}
		if g <= 30*time.Minute {
			hits30++
		}
	}
	n := float64(len(sorted))
	hitRate := float64(hits) / n
	mult := hitRate*readPrice + (1-hitRate)*writePrice
	// Stable prefix adds ~8-12% overhead on catalog spine + epoch alignment, so
	// require expected multiplier <= 0.85 to recommend enabling it.
	rec := mult <= 0.85
	reason := "cadence is sparse; prefix writes would exceed cache read savings"
	if rec {
		reason = "interactive cadence beats cache break-even; enable catalog spine + epoch alignment"
	}
	return CachePolicyAdvice{
		SampleGaps:         len(sorted),
		MedianGapSeconds:   int64(sorted[len(sorted)/2].Seconds()),
		HitRate5m:          math.Round(float64(hits5)/n*1000) / 1000,
		HitRate30m:         math.Round(float64(hits30)/n*1000) / 1000,
		ExpectedMultiplier: math.Round(mult*1000) / 1000,
		RecommendStable:    rec,
		Reason:             reason,
	}
}
