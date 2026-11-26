package engine

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
)

// Empirical multi-agent scaling gate (Kim et al., "Towards a Science of Scaling
// Agent Systems", 2026), barrier-free reactive DAG scheduler (eliminating
// Pregel superstep straggler barriers), and Verified-Quote Dominance guard
// (preventing ungrounded claims from flipping quote-verified facts).

// PlanTopologyMetrics quantifies a proposed multi-unit DAG before execution.
type PlanTopologyMetrics struct {
	UnitCount         int     `json:"unit_count"`
	CriticalPathDepth int     `json:"critical_path_depth"`
	MaxWidth          int     `json:"max_width"`
	ParallelismRatio  float64 `json:"parallelism_ratio"`  // UnitCount / CriticalPathDepth
	EstTokenOverhead  float64 `json:"est_token_overhead"` // relative to 1.0 single call
	HasRepoUnits      bool    `json:"has_repo_units"`
	AllowFanOut       bool    `json:"allow_fan_out"`
	RefuseReason      string  `json:"refuse_reason,omitempty"`
}

// EvaluatePlanTopology analyzes a proposed unit DAG. Near-sequential multi-unit
// plans (e.g. >=3 non-repo units with critical path >= N-1 and width < 2, where
// Kim et al. 2026 measured -39% to -70% sequential planning degradation and
// ~2.8x token overhead) are collapsed back to a single strong reasoning call.
func EvaluatePlanTopology(units []packet.Unit) PlanTopologyMetrics {
	n := len(units)
	if n == 0 {
		return PlanTopologyMetrics{AllowFanOut: false, RefuseReason: "empty plan"}
	}
	waves := planWaves(units)
	depth := len(waves)
	maxW := 0
	for _, w := range waves {
		if len(w) > maxW {
			maxW = len(w)
		}
	}
	hasRepo := false
	for _, u := range units {
		if u.Repo != "" {
			hasRepo = true
			break
		}
	}
	ratio := float64(n) / float64(max(1, depth))
	overhead := 1.0 + 0.55*float64(n)
	m := PlanTopologyMetrics{
		UnitCount:         n,
		CriticalPathDepth: depth,
		MaxWidth:          maxW,
		ParallelismRatio:  math.Round(ratio*100) / 100,
		EstTokenOverhead:  math.Round(overhead*100) / 100,
		HasRepoUnits:      hasRepo,
		AllowFanOut:       true,
	}
	if hasRepo {
		return m
	}
	if n == 1 {
		m.AllowFanOut = false
		m.RefuseReason = "a plan of one unit is one call"
		return m
	}
	if maxW < 2 {
		m.AllowFanOut = false
		m.RefuseReason = "every unit waits for the one before it, so nothing would run in parallel"
		return m
	}
	// Scaling-law gate: >=4 non-repo units whose critical path is >= N-1
	// (e.g. a long serial chain with only 1 tiny side-branch) burns ~3x tokens
	// and amplifies serial handoff drift for negligible parallelism (ratio <= 1.34).
	if n >= 4 && depth >= n-1 && ratio < 1.35 {
		m.AllowFanOut = false
		m.RefuseReason = fmt.Sprintf("plan has %d units across %d serial waves (parallelism %.2fx < 1.35x); collapsing to one call to avoid serial handoff drift", n, depth, ratio)
		return m
	}
	return m
}

// runDAG executes n indexed nodes in a barrier-free reactive dataflow DAG:
// each node i starts immediately once all indices in deps[i] have completed
// (bounded by maxPar concurrent workers), eliminating Pregel wave barriers
// where fast branches stall behind an unrelated slow sibling.
func runDAG(ctx context.Context, n int, deps [][]int, maxPar int, fn func(i int)) {
	if n == 0 {
		return
	}
	if maxPar <= 0 {
		maxPar = 4
	}
	remDeps := make([]int, n)
	children := make([][]int, n)
	for i := 0; i < n; i++ {
		seen := map[int]bool{}
		for _, d := range deps[i] {
			if d >= 0 && d < n && d != i && !seen[d] {
				seen[d] = true
				remDeps[i]++
				children[d] = append(children[d], i)
			}
		}
	}

	sem := make(chan struct{}, maxPar)
	var mu sync.Mutex
	var wg sync.WaitGroup
	launched := make([]bool, n)

	var launch func(i int)
	launch = func(i int) {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			fn(idx)
			<-sem

			var ready []int
			mu.Lock()
			for _, ch := range children[idx] {
				remDeps[ch]--
				if remDeps[ch] == 0 && !launched[ch] {
					launched[ch] = true
					ready = append(ready, ch)
				}
			}
			mu.Unlock()
			for _, r := range ready {
				launch(r)
			}
		}(i)
	}

	var roots []int
	mu.Lock()
	for i := 0; i < n; i++ {
		if remDeps[i] == 0 {
			launched[i] = true
			roots = append(roots, i)
		}
	}
	mu.Unlock()
	for _, r := range roots {
		launch(r)
	}
	wg.Wait()
	_ = ctx
}

// ClaimKeyConflict reports a same-key value conflict across parallel unit
// records, along with whether Verified-Quote Dominance settles it automatically.
type ClaimKeyConflict struct {
	Key            string   `json:"key"`
	VerifiedValues []string `json:"verified_values,omitempty"` // unit:value backed by quote-verified anchors
	UnverifiedVals []string `json:"unverified_values,omitempty"`
	DominantUnit   string   `json:"dominant_unit,omitempty"` // set when exactly one side has quote-verified proof
	Summary        string   `json:"summary"`
}

// AuditClaimKeyConflicts inspects checked claims across all unit records in a
// plan. When two units assert different values for the same structured key
// (`pool_size`, `port`, `timeout`, etc.), it enforces Verified-Quote Dominance:
// an unverified/inferred claim may never override a byte-verified quoted claim.
func AuditClaimKeyConflicts(recs []unitRecord) []ClaimKeyConflict {
	type entry struct {
		unit     string
		claimID  string
		val      string
		verified bool
	}
	byKey := map[string][]entry{}
	for _, rec := range recs {
		if rec.Error != "" {
			continue
		}
		for _, cl := range rec.Claims {
			assertions := store.ExtractKeyAssertions(0, 0, cl.Claim+" "+cl.Quote)
			isVerified := cl.Check == checkVerified && cl.Support == packet.SupportQuoted
			for _, a := range assertions {
				byKey[a.Key] = append(byKey[a.Key], entry{
					unit:     rec.Unit,
					claimID:  cl.ID,
					val:      a.RawValue,
					verified: isVerified,
				})
			}
		}
	}

	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out []ClaimKeyConflict
	for _, k := range keys {
		list := byKey[k]
		distinctVals := map[string]bool{}
		for _, e := range list {
			distinctVals[strings.ToLower(e.val)] = true
		}
		if len(distinctVals) < 2 {
			continue
		}
		var ver, unver []string
		verVals := map[string]string{} // normVal -> unit
		for _, e := range list {
			tag := fmt.Sprintf("%s=%s (%s)", e.claimID, e.val, e.unit)
			if e.verified {
				ver = append(ver, tag)
				verVals[strings.ToLower(e.val)] = e.unit
			} else {
				unver = append(unver, tag)
			}
		}
		c := ClaimKeyConflict{
			Key:            k,
			VerifiedValues: ver,
			UnverifiedVals: unver,
		}
		if len(verVals) == 1 && len(unver) > 0 {
			for val, u := range verVals {
				c.DominantUnit = u
				c.Summary = fmt.Sprintf("key %q: verified-quote dominance favors %s=%s (from %s) over unverified %s",
					k, k, val, u, strings.Join(unver, ", "))
			}
		} else {
			c.Summary = fmt.Sprintf("key %q has conflicting assertions across units: verified=[%s] unverified=[%s]",
				k, strings.Join(ver, ", "), strings.Join(unver, ", "))
		}
		out = append(out, c)
	}
	return out
}
