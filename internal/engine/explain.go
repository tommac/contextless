package engine

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// Explain compiles a dry-run packet for msg and returns a complete database
// query-planner breakdown: query class, lane hits, included/skipped atoms,
// section token costs, cache-prefix sharing against the previous message,
// prefix-closed knapsack recommendation, and estimated cost/latency per
// routed profile. dropIDs removes the named turn/atom IDs from the built
// packet. Nothing is selected again, so it does not show cause.
func (e *Engine) Explain(msg string, dropIDs []string) (packet.ExplainPlan, error) {
	focus := e.nextTurn(msg)
	cands, _ := e.R.Candidates(config.RoleResponder, nil)
	if len(cands) == 0 {
		return packet.ExplainPlan{}, fmt.Errorf("no responder profile available")
	}
	task := packet.Task{
		RunID: "run-explain", Kind: packet.KindRespond, Role: config.RoleResponder,
		Instructions: respondTextMax("run-explain", true, e.Cfg.MaxUnits),
		Schema:       packet.RespondSchema(true),
	}
	copt := e.compileOptions()
	sized, err := e.buildPacket(task, focus, packet.SelectOptions{Compile: &copt}, cands, packet.OutputTokens[packet.KindRespond])
	if err != nil {
		return packet.ExplainPlan{}, err
	}
	if len(dropIDs) > 0 {
		dropSet := map[string]bool{}
		for _, id := range dropIDs {
			dropSet[strings.TrimSpace(id)] = true
		}
		var kept []packet.Item
		for _, it := range sized.Items {
			ref, ok := vault.ParseRef(it.ID)
			if dropSet[it.ID] || (ok && dropSet[ref.Turn]) {
				sized.Dropped = append(sized.Dropped, it)
				continue
			}
			kept = append(kept, it)
		}
		sized.Items = kept
		task.NativeSchema = sized.Profile.NativeSchema
		sized.Brief = packet.Render(task, sized.Items, sized.Dropped)
		sized.Budget.Brief = packet.Tokens(sized.Profile, len(sized.Brief))
		sized.Budget.Planned = sized.Profile.SysOverhead + sized.Budget.Brief + sized.Budget.Output
	}

	// Section costs
	secMap := map[string]*packet.SectionCost{}
	var secKeys []string
	for _, it := range sized.Items {
		sc := secMap[it.Key]
		if sc == nil {
			sc = &packet.SectionCost{Key: it.Key}
			secMap[it.Key] = sc
			secKeys = append(secKeys, it.Key)
		}
		sc.Count++
		sc.Tokens += it.Tokens
	}
	sort.Strings(secKeys)
	sections := make([]packet.SectionCost, 0, len(secKeys))
	for _, k := range secKeys {
		sections = append(sections, *secMap[k])
	}

	// Lane diagnostics + top candidate explanations
	turns, _ := e.V.Turns()
	if copt.SessionRoute && len(copt.EpisodeTurns) == 0 {
		copt.EpisodeTurns = packet.RouteEpisodeTurns(turns, nil, e.V.Cards(), msg, focus.Seq, 3, 12)
	}
	diag := packet.RankAtomsDiag(turns, e.S, msg, focus.Seq, 20, copt)
	includedIDs := map[string]string{}
	for _, it := range sized.Items {
		if it.Key == packet.KeyCatalog {
			continue
		}
		includedIDs[it.ID] = it.Key
		if ref, ok := vault.ParseRef(it.ID); ok {
			if ref.Whole() {
				includedIDs[ref.Turn] = it.Key
			} else if ref.Atom > 0 {
				for k := ref.Atom; k <= ref.AtomEnd; k++ {
					includedIDs[vault.AtomID(ref.Turn, k)] = it.Key
				}
			}
		}
	}
	laneMem := func(list []string, id string) int {
		for i, x := range list {
			if x == id {
				return i + 1
			}
		}
		return 0
	}
	topCands := make([]packet.LaneHitEntry, 0, min(12, len(diag.Fused)))
	for _, r := range diag.Fused[:min(12, len(diag.Fused))] {
		aid := r.Atom.ID()
		var why []string
		var poly packet.ProvenancePoly
		if p := laneMem(diag.Exact, aid); p > 0 {
			why = append(why, fmt.Sprintf("exact#%d", p))
			poly.Add("exact", fmt.Sprintf("#%d", p), 0.90)
		}
		if p := laneMem(diag.FullText, aid); p > 0 {
			why = append(why, fmt.Sprintf("fts#%d", p))
			poly.Add("fts", fmt.Sprintf("#%d", p), max(0.35, 0.80-0.04*float64(p-1)))
		}
		if p := laneMem(diag.Value, aid); p > 0 {
			why = append(why, fmt.Sprintf("val#%d", p))
			poly.Add("value", fmt.Sprintf("#%d", p), 0.78)
		}
		if p := laneMem(diag.WorkingSet, aid); p > 0 {
			why = append(why, fmt.Sprintf("ws#%d", p))
			poly.Add("working_set", fmt.Sprintf("#%d", p), 0.65)
		}
		if p := laneMem(diag.Episode, aid); p > 0 {
			why = append(why, fmt.Sprintf("ep#%d", p))
			poly.Add("episode", fmt.Sprintf("#%d", p), 0.60)
		}
		incKey := includedIDs[aid]
		if incKey == "" {
			incKey = includedIDs[r.Turn.ID]
		}
		inc := incKey != ""
		reason := strings.Join(why, ",")
		if inc {
			reason += " -> " + incKey
		} else {
			reason += " (budget full)"
		}
		topCands = append(topCands, packet.LaneHitEntry{
			ID: aid, Turn: r.Turn.ID, Score: r.Score, Included: inc, Reason: reason, Provenance: poly.Format(),
		})
	}

	// Cache prefix against previous user message + cadence advice
	var userTimes []time.Time
	var prevUser *vault.Turn
	for i := range turns {
		if turns[i].Role == vault.RoleUser {
			userTimes = append(userTimes, turns[i].Created)
			prevUser = &turns[i]
		}
	}
	var gaps []time.Duration
	for i := 1; i < len(userTimes); i++ {
		if d := userTimes[i].Sub(userTimes[i-1]); d >= 0 {
			gaps = append(gaps, d)
		}
	}
	advice := packet.AdviseCachePolicy(gaps, 0.10, 1.25, 5*time.Minute)

	cachePrefixToks := 0
	if prevUser != nil {
		ptask := packet.Task{RunID: "run-prev", Kind: packet.KindRespond, Role: config.RoleResponder,
			Instructions: respondTextMax("run-prev", true, e.Cfg.MaxUnits), Schema: packet.RespondSchema(true)}
		if ps, perr := e.buildPacket(ptask, *prevUser, packet.SelectOptions{Compile: &copt}, cands, packet.OutputTokens[packet.KindRespond]); perr == nil {
			cachePrefixToks = cacheableTokens(packet.Tokens(sized.Profile, commonPrefix(ps.Brief, sized.Brief)))
		}
	}
	cacheShare := 0
	if sized.Budget.Brief > 0 {
		cacheShare = 100 * cachePrefixToks / sized.Budget.Brief
	}

	// Prefix-closed knapsack simulation across P0 (rules+session), P1 (+catalog), P2 (+recent)
	var prefixOpt *packet.PrefixOptResult
	if len(diag.Fused) > 0 {
		qw := packet.ContentWords(msg)
		qIDF := map[string]float64{}
		for w := range qw {
			qIDF[w] = 1.0
		}
		var suffixes []packet.SuffixCandidate
		for rank, r := range diag.Fused {
			txt := r.Atom.Text(r.Turn.Body)
			terms := map[string]float64{}
			for w := range packet.ContentWords(txt) {
				if qw[w] {
					terms[w] = 1.0
				}
			}
			suffixes = append(suffixes, packet.SuffixCandidate{
				ID:        r.Atom.ID(),
				Tokens:    packet.Tokens(sized.Profile, len(txt)+80),
				RankPrior: 0.5 / float64(60+rank+1),
				Terms:     terms,
			})
		}
		catToks, recToks := 0, 0
		var recIDs []string
		recTerms := map[string]float64{}
		for _, it := range sized.Items {
			switch it.Key {
			case packet.KeyCatalog:
				catToks += it.Tokens
			case packet.KeyRecent:
				recToks += it.Tokens
				recIDs = append(recIDs, it.ID)
				for w := range packet.ContentWords(it.Text) {
					if qw[w] {
						recTerms[w]++
					}
				}
			}
		}
		res := packet.OptimizePrefixClosed(packet.PrefixOptInput{
			QueryIDF: qIDF,
			Prefixes: []packet.PrefixCandidate{
				{Name: "P0:rules+session", Tokens: 800},
				{Name: "P1:+catalog", Tokens: 800 + catToks},
				{Name: "P2:+catalog+recent", Tokens: 800 + catToks + recToks, AtomIDs: recIDs, TermCount: recTerms},
			},
			Suffixes:     suffixes,
			BilledBudget: float64(max(e.Cfg.RelevantTokens, 4000)),
			ReadPrice:    0.10,
			FreshPrice:   1.0,
		})
		prefixOpt = &res
	}

	// Per-profile cost & latency estimates fitted from store.Stats()
	statsByProfile := map[string]store.Stat{}
	if e.S != nil {
		for _, st := range e.S.Stats() {
			if st.Kind == packet.KindRespond || statsByProfile[st.Profile].Runs == 0 {
				statsByProfile[st.Profile] = st
			}
		}
	}
	estimates := make([]packet.ProfileEstimate, 0, len(cands))
	for _, p := range cands {
		p = p.WithDefaults()
		briefToks := packet.Tokens(p, len(sized.Brief))
		outToks := packet.OutputTokens[packet.KindRespond] + packet.HandoffTokens
		planned := p.SysOverhead + briefToks + outToks
		estOut := 1500 // realistic median response tokens for USD comparison
		uncachedUSD := providers.EstimateUSD(p, providers.Usage{InputTokens: p.SysOverhead + briefToks, OutputTokens: estOut})
		cachedUSD := providers.EstimateUSD(p, providers.Usage{InputTokens: p.SysOverhead + briefToks, CachedTokens: cachePrefixToks, OutputTokens: estOut})
		pe := packet.ProfileEstimate{
			Profile:       p.ID,
			Fits:          planned <= p.PlanCap(),
			PlannedTokens: planned,
			CapTokens:     p.PlanCap(),
			EstUSD:        uncachedUSD,
			EstCachedUSD:  cachedUSD,
		}
		if st, ok := statsByProfile[p.ID]; ok && st.Runs > 0 {
			pe.ObservedRuns = st.Runs
			pe.MedianWallMs = st.MedianWall
			pe.SuccessRate = math.Round(float64(st.OK)/float64(st.Runs)*1000) / 1000
		}
		estimates = append(estimates, pe)
	}

	// Compute the coverage score against a stand-in objective, plus active
	// materialized key-state views + 2-hop rare-key PPR bridges.
	var subCert *packet.SubmodularCert
	if len(diag.Fused) > 0 {
		qw := packet.ContentWords(msg)
		qIDF := map[string]float64{}
		for w := range qw {
			qIDF[w] = 1.0
		}
		subItems := make([]packet.SubmodularItem, 0, len(diag.Fused))
		for rank, r := range diag.Fused {
			txt := r.Atom.Text(r.Turn.Body)
			terms := map[string]float64{}
			for w := range packet.ContentWords(txt) {
				if qw[w] {
					terms[w] = 1.0
				}
			}
			subItems = append(subItems, packet.SubmodularItem{
				ID:       r.Atom.ID(),
				Tokens:   packet.Tokens(sized.Profile, len(txt)+80),
				Prior:    0.5 / float64(60+rank+1),
				Terms:    terms,
				Shingles: packet.WordShingles(txt),
			})
		}
		cert := packet.CertifySubmodularSelection(qIDF, subItems, max(e.Cfg.RelevantTokens, 2000))
		subCert = &cert
	}
	var activeKeys []store.KeyStateRow
	var bridgeHits []store.BridgeHit
	if e.S != nil {
		if allKVs, kerr := e.S.ListActiveKeys("", focus.Seq-1, 50); kerr == nil {
			lowMsg := strings.ToLower(msg)
			for _, kv := range allKVs {
				if strings.Contains(lowMsg, kv.Key) || len(allKVs) <= 8 {
					activeKeys = append(activeKeys, kv)
				}
			}
		}
		var seedSeqs []int
		for i := 0; i < len(diag.Fused) && i < 6; i++ {
			seedSeqs = append(seedSeqs, diag.Fused[i].Turn.Seq)
		}
		if len(seedSeqs) > 0 {
			bridgeHits, _ = e.S.BridgeRareKeysPPR(seedSeqs, focus.Seq, 5)
		}
	}

	idxStatus := sized.Index
	if len(diag.LaneErrors) > 0 {
		idxStatus.Complete = false
		idxStatus.LaneErrors = append(append([]string{}, idxStatus.LaneErrors...), diag.LaneErrors...)
	}
	omittedHandle := sized.OmittedHandle
	if omittedHandle == "" {
		omittedHandle = task.OmittedHandle
	}
	return packet.ExplainPlan{
		Query:             msg,
		Class:             packet.ClassifyQuery(msg),
		SelectedProfile:   sized.Profile.ID,
		WholeHistory:      sized.WholeHistory,
		AdaptivePack:      copt.AdaptivePack,
		Budget:            sized.Budget,
		Sections:          sections,
		OmittedCount:      len(sized.Dropped),
		OmittedHandle:     omittedHandle,
		CachePrefixTokens: cachePrefixToks,
		CacheSharePct:     cacheShare,
		LaneCounts: map[string]int{
			"exact":       len(diag.Exact),
			"full_text":   len(diag.FullText),
			"value":       len(diag.Value),
			"working_set": len(diag.WorkingSet),
			"episode":     len(diag.Episode),
		},
		TopCandidates:    topCands,
		ProfileEstimates: estimates,
		CacheAdvice:      advice,
		PrefixOpt:        prefixOpt,
		SubmodularCert:   subCert,
		ActiveKeyViews:   activeKeys,
		BridgeHits:       bridgeHits,
		DroppedForTest:   dropIDs,
		Index:            idxStatus,
	}, nil
}
