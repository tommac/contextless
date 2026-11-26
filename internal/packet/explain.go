package packet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/store"
)

// QueryClass classifies a user request by deterministic rules so the planner
// can report recommended budget shape and output contracts.
type QueryClass string

const (
	ClassLookup    QueryClass = "lookup"
	ClassTimeline  QueryClass = "timeline"
	ClassCode      QueryClass = "code"
	ClassSynthesis QueryClass = "synthesis"
	ClassGeneral   QueryClass = "general"
)

// ClassifyQuery determines a request's class using fast deterministic rules.
func ClassifyQuery(msg string) QueryClass {
	low := strings.ToLower(strings.TrimSpace(msg))
	for _, kw := range []string{"timeline", "history of", "changed", "change ", "over time", "evolve", "used to be", "before we", "supersede", "when did we", "evolution"} {
		if strings.Contains(low, kw) {
			return ClassTimeline
		}
	}
	for _, kw := range []string{"implement", "refactor", "fix the bug", "write a function", "compile", "test fails", "stack trace", "diff", "pull request", "repo", "codebase"} {
		if strings.Contains(low, kw) {
			return ClassCode
		}
	}
	for _, kw := range []string{"compare", "summarize all", "every ", "list all", "across the conversation", "synthesize", "overview"} {
		if strings.Contains(low, kw) {
			return ClassSynthesis
		}
	}
	if len(low) < 180 && (strings.HasSuffix(low, "?") || strings.HasPrefix(low, "what ") || strings.HasPrefix(low, "which ") || strings.HasPrefix(low, "who ") || strings.HasPrefix(low, "where ")) {
		return ClassLookup
	}
	return ClassGeneral
}

// SectionCost summarizes one section's items and token footprint.
type SectionCost struct {
	Key    string `json:"key"`
	Count  int    `json:"count"`
	Tokens int    `json:"tokens"`
}

// ProfileEstimate predicts billed input cost and wall latency for one routed
// provider profile using its rate card and historical runs from store.Stats().
type ProfileEstimate struct {
	Profile       string  `json:"profile"`
	Fits          bool    `json:"fits"`
	PlannedTokens int     `json:"planned_tokens"`
	CapTokens     int     `json:"cap_tokens"`
	EstUSD        float64 `json:"est_usd"`
	EstCachedUSD  float64 `json:"est_cached_usd"`
	MedianWallMs  int64   `json:"median_wall_ms,omitempty"`
	SuccessRate   float64 `json:"success_rate,omitempty"`
	ObservedRuns  int     `json:"observed_runs,omitempty"`
}

// LaneHitEntry is one ranked atom shown in the explain plan.
type LaneHitEntry struct {
	ID         string  `json:"id"`
	Turn       string  `json:"turn"`
	Score      float64 `json:"score"`
	Included   bool    `json:"included"`
	Reason     string  `json:"reason,omitempty"`
	Provenance string  `json:"provenance,omitempty"`
}

// ExplainPlan is the query-planner report produced by `cx explain`.
type ExplainPlan struct {
	Query             string            `json:"query"`
	Class             QueryClass        `json:"class"`
	SelectedProfile   string            `json:"selected_profile"`
	WholeHistory      bool              `json:"whole_history"`
	AdaptivePack      bool              `json:"adaptive_pack"`
	Budget            Budget            `json:"budget"`
	Sections          []SectionCost     `json:"sections"`
	OmittedCount      int               `json:"omitted_count"`
	OmittedHandle     string            `json:"omitted_handle,omitempty"`
	CachePrefixTokens int               `json:"cache_prefix_tokens"`
	CacheSharePct     int               `json:"cache_share_pct"`
	LaneCounts        map[string]int    `json:"lane_counts"`
	TopCandidates     []LaneHitEntry    `json:"top_candidates"`
	ProfileEstimates  []ProfileEstimate `json:"profile_estimates"`
	CacheAdvice       CachePolicyAdvice `json:"cache_advice"`
	PrefixOpt         *PrefixOptResult    `json:"prefix_opt,omitempty"`
	SubmodularCert    *SubmodularCert     `json:"submodular_cert,omitempty"`
	ActiveKeyViews    []store.KeyStateRow `json:"active_key_views,omitempty"`
	BridgeHits        []store.BridgeHit   `json:"bridge_hits,omitempty"`
	DroppedForTest    []string            `json:"dropped_for_test,omitempty"`
	Index             IndexStatus         `json:"index"`
}

// FormatExplain renders an ExplainPlan in a readable database-planner style.
func FormatExplain(p ExplainPlan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "CONTEXT PLAN  class=%s  profile=%s  whole_history=%v  adaptive_pack=%v\n",
		p.Class, p.SelectedProfile, p.WholeHistory, p.AdaptivePack)
	fmt.Fprintf(&b, "  budget: %d / %d tokens (brief %d, sys %d, output %d, kill %d, escalated=%v)\n",
		p.Budget.Planned, p.Budget.Cap, p.Budget.Brief, p.Budget.SysOverhead, p.Budget.Output, p.Budget.KillLine, p.Budget.Escalated)
	fmt.Fprintf(&b, "  cache:  %d prefix tokens shared with previous message (%d%% of brief)\n",
		p.CachePrefixTokens, p.CacheSharePct)
	if p.CacheAdvice.SampleGaps > 0 {
		fmt.Fprintf(&b, "  cadence: %d gaps, median %ds, 5m-hit %.0f%%, expected prefix price %.2fx (%s)\n",
			p.CacheAdvice.SampleGaps, p.CacheAdvice.MedianGapSeconds, p.CacheAdvice.HitRate5m*100,
			p.CacheAdvice.ExpectedMultiplier, p.CacheAdvice.Reason)
	}
	if p.PrefixOpt != nil {
		fmt.Fprintf(&b, "  prefix-opt: best=%s (%d cached + %d fresh tokens, billed %.0f, utility %.2f, fallback=%v)\n",
			p.PrefixOpt.PrefixName, p.PrefixOpt.PrefixTokens, p.PrefixOpt.SuffixTokens,
			p.PrefixOpt.BilledTokens, p.PrefixOpt.Utility, p.PrefixOpt.UsedFallback)
	}
	if p.SubmodularCert != nil {
		fmt.Fprintf(&b, "  optimality: >= %.1f%% of theoretical OPT (utility %.3f / LP upper bound %.3f; worst-case theorem >= %.1f%%)\n",
			p.SubmodularCert.OptimalityPct, p.SubmodularCert.Utility, p.SubmodularCert.UpperBoundOPT, p.SubmodularCert.WorstCaseBoundPct)
	}
	if len(p.DroppedForTest) > 0 {
		fmt.Fprintf(&b, "  causal ablation: dropped %s\n", strings.Join(p.DroppedForTest, ", "))
	}
	if p.Index.SnapshotSeq > 0 || len(p.Index.LaneErrors) > 0 || !p.Index.Complete {
		root := p.Index.SnapshotRoot
		if len(root) > 12 {
			root = root[:12]
		}
		fmt.Fprintf(&b, "  index:  snapshot=#%d (%s)  indexed_up_to=%d  complete=%v\n",
			p.Index.SnapshotSeq, root, p.Index.IndexedUpTo, p.Index.Complete)
		for _, le := range p.Index.LaneErrors {
			fmt.Fprintf(&b, "  lane warning: %s\n", le)
		}
	}

	b.WriteString("\nSECTIONS\n")
	for _, s := range p.Sections {
		fmt.Fprintf(&b, "  %-14s %4d item(s)  %7d tokens\n", s.Key, s.Count, s.Tokens)
	}
	if p.OmittedCount > 0 {
		fmt.Fprintf(&b, "  %-14s %4d item(s)", "omitted", p.OmittedCount)
		if p.OmittedHandle != "" {
			b.WriteString("  (external paged handle)")
		}
		b.WriteString("\n")
	}

	if len(p.LaneCounts) > 0 {
		b.WriteString("\nCANDIDATE LANES\n")
		keys := make([]string, 0, len(p.LaneCounts))
		for k := range p.LaneCounts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%d", k, p.LaneCounts[k]))
		}
		fmt.Fprintf(&b, "  hits: %s\n", strings.Join(parts, "  "))
	}
	if len(p.TopCandidates) > 0 {
		b.WriteString("\nTOP RANKED ATOMS\n")
		for i, c := range p.TopCandidates {
			mark := "skip"
			if c.Included {
				mark = "IN  "
			}
			fmt.Fprintf(&b, "  %2d. [%s] %-16s score=%.4f  %s\n", i+1, mark, c.ID, c.Score, c.Reason)
			if c.Provenance != "" {
				fmt.Fprintf(&b, "      provenance: %s\n", c.Provenance)
			}
		}
	}
	if len(p.ActiveKeyViews) > 0 {
		b.WriteString("\nMATERIALIZED KEY STATE (0% STALE EXPOSURE)\n")
		for _, kv := range p.ActiveKeyViews {
			fmt.Fprintf(&b, "  %-22s = %-20s (valid from t%06d)\n", kv.Key, kv.RawValue, kv.ValidFromSeq)
		}
	}
	if len(p.BridgeHits) > 0 {
		b.WriteString("\n2-HOP RARE-KEY PPR BRIDGES\n")
		for _, bh := range p.BridgeHits {
			fmt.Fprintf(&b, "  t%06d  mass=%.4f  via=%s\n", bh.TurnSeq, bh.Score, strings.Join(bh.ViaKeys, ", "))
		}
	}

	if len(p.ProfileEstimates) > 0 {
		b.WriteString("\nROUTED PROFILES (COST & LATENCY)\n")
		for _, pe := range p.ProfileEstimates {
			fit := "fits"
			if !pe.Fits {
				fit = "OVER"
			}
			hist := "no runs yet"
			if pe.ObservedRuns > 0 {
				hist = fmt.Sprintf("%d runs, %.0f%% ok, median %.1fs", pe.ObservedRuns, pe.SuccessRate*100, float64(pe.MedianWallMs)/1000)
			}
			fmt.Fprintf(&b, "  %-38s %4s  %6d/%-6d tok  ~$%.4f (cached ~$%.4f)  [%s]\n",
				pe.Profile, fit, pe.PlannedTokens, pe.CapTokens, pe.EstUSD, pe.EstCachedUSD, hist)
		}
	}
	return b.String()
}
