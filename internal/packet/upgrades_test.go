package packet

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

func TestRenderClosesUnclosedCodeFence(t *testing.T) {
	items := []Item{
		{
			Key: KeyRelevant, ID: "t000001#a1", Seq: 1,
			Header: "t000001 · user (excerpt)",
			Text:   "Here is part of the code:\n```go\nfunc broken() {\n\treturn",
		},
		{
			Key: KeyRecent, ID: "t000002", Seq: 2,
			Header: "t000002 · assistant",
			Text:   "Next turn text must not be swallowed inside a code block.",
		},
	}
	brief := Render(Task{RunID: "run-1", Kind: KindRespond, Role: "responder", Instructions: "Reply"}, items, nil)
	if strings.Count(brief, "```")%2 != 0 {
		t.Fatalf("expected balanced code fences in rendered brief:\n%s", brief)
	}
	if !strings.Contains(brief, "Next turn text must not be swallowed") {
		t.Fatalf("missing subsequent turn text")
	}
	// Original verbatim text must still appear intact before the closing fence.
	if !strings.Contains(brief, "func broken() {\n\treturn") {
		t.Fatalf("verbatim snippet was altered")
	}
}

func TestPickAtomsAdaptiveDiversityAndBudget(t *testing.T) {
	s, err := store.OpenFile(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Three turns, each with a >220-byte lead paragraph plus two large follow-up
	// paragraphs mentioning "deployment" so a 450-token budget forces overflow
	// and round-robin source diversity picks across distinct turns.
	leadPad := strings.Repeat("canary service deployment specification details ", 6)
	pad := strings.Repeat("background operational notes for deployment verification ", 12)
	turns := []vault.Turn{
		{ID: "t000001", Seq: 1, Role: vault.RoleUser, Body: "Canary deployment region is us-east-1. " + leadPad + "\n\n" + pad + "first-extra\n\n" + pad + "first-tail"},
		{ID: "t000002", Seq: 2, Role: vault.RoleUser, Body: "Canary deployment timeout is 45s. " + leadPad + "\n\n" + pad + "second-extra\n\n" + pad + "second-tail"},
		{ID: "t000003", Seq: 3, Role: vault.RoleUser, Body: "Canary deployment rollback policy uses blue-green. " + leadPad + "\n\n" + pad + "third-extra\n\n" + pad + "third-tail"},
	}
	var at []store.AtomTurn
	for _, tr := range turns {
		var recs []store.AtomRec
		for _, a := range AtomsOf(tr) {
			recs = append(recs, store.AtomRec{ID: a.ID(), Text: a.Text(tr.Body)})
		}
		at = append(at, store.AtomTurn{Turn: tr.ID, Seq: tr.Seq, Atoms: recs})
	}
	if err := s.IndexAtomTurns(at); err != nil {
		t.Fatal(err)
	}

	opt := DefaultCompileOptions()
	opt.AdaptivePack = true
	runs, _ := PickAtomsAdaptive(turns, s, "deployment canary region timeout rollback", 10, 450, nil, opt)
	distinct := map[string]bool{}
	for _, r := range runs {
		distinct[r.Turn.ID] = true
	}
	if len(distinct) < 2 {
		t.Fatalf("expected adaptive packer to draw from >=2 distinct source turns under tight budget, got %d (%+v)", len(distinct), runs)
	}
}

func TestOmittedCatalogBuildOpenVerifyAndTamper(t *testing.T) {
	dir := t.TempDir()
	// 6,000 turns: turns 1..5500 omitted except multiples of 500; turns 5501..6000 included.
	omittedFn := func(seq int) bool {
		return seq <= 5500 && seq%500 != 0
	}
	h, err := BuildOmittedCatalog(dir, "deadbeef", 6000, omittedFn)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.JSON()) > 260 {
		t.Fatalf("expected compact omitted handle (<260 bytes), got %d: %s", len(h.JSON()), h.JSON())
	}
	if h.Omitted != 5500-11 {
		t.Fatalf("omitted count = %d, want %d", h.Omitted, 5500-11)
	}
	cat, err := OpenOmittedCatalog(dir, h, "deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	defer cat.Close()

	if ok, err := cat.Contains(42); err != nil || !ok {
		t.Fatalf("turn 42 should be omitted: ok=%v err=%v", ok, err)
	}
	if ok, err := cat.Contains(500); err != nil || ok {
		t.Fatalf("turn 500 should not be omitted: ok=%v err=%v", ok, err)
	}
	if ok, err := cat.Contains(5999); err != nil || ok {
		t.Fatalf("turn 5999 should not be omitted: ok=%v err=%v", ok, err)
	}
	page, next, err := cat.Enumerate(&OmittedCursor{Index: h.Index, After: 498}, 4)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{499, 501, 502, 503}
	if len(page) != len(want) || page[0] != want[0] || page[1] != want[1] || page[3] != want[3] || next.After != 503 {
		t.Fatalf("Enumerate = %v next=%+v, want %v", page, next, want)
	}
	// Tampering with the .cxi file must be detected on open or read.
	cxiPath := filepath.Join(dir, h.Index+".cxi")
	raw, _ := os.ReadFile(cxiPath)
	raw[0] ^= 0xff
	_ = os.WriteFile(cxiPath, raw, 0o644)
	if _, err := OpenOmittedCatalog(dir, h, "deadbeef"); err == nil {
		t.Fatal("expected tamper detection error when .cxi bytes are corrupted")
	}
}

func TestOptimizePrefixClosedAndCacheAdvice(t *testing.T) {
	// Best-single-item fallback check: tiny cheap low-value items vs one big high-value item
	// whose gain-per-token is slightly lower so greedy picks small items and misses the big one.
	res := OptimizePrefixClosed(PrefixOptInput{
		QueryIDF: map[string]float64{"alpha": 10, "beta": 0.4},
		Prefixes: []PrefixCandidate{
			{Name: "P0", Tokens: 100},
			{Name: "P1", Tokens: 2000, TermCount: map[string]float64{"alpha": 1}},
		},
		Suffixes: []SuffixCandidate{
			{ID: "small", Tokens: 10, Terms: map[string]float64{"beta": 1}},
			{ID: "big", Tokens: 950, Terms: map[string]float64{"alpha": 1}},
		},
		BilledBudget: 1000,
		ReadPrice:    0.10,
		FreshPrice:   1.0,
	})
	if res.PrefixName != "P1" {
		t.Fatalf("expected cached prefix P1 (cost 200 billed for alpha coverage) to win, got %+v", res)
	}
	// Single-item fallback specifically on empty prefix:
	fb := OptimizePrefixClosed(PrefixOptInput{
		QueryIDF: map[string]float64{"rare": 10, "common": 0.6},
		Prefixes: []PrefixCandidate{{Name: "P0", Tokens: 0}},
		Suffixes: []SuffixCandidate{
			{ID: "tiny", Tokens: 10, Terms: map[string]float64{"common": 1}},
			{ID: "huge", Tokens: 995, Terms: map[string]float64{"rare": 1}},
		},
		BilledBudget: 1000,
		ReadPrice:    0.10,
		FreshPrice:   1.0,
	})
	if !fb.UsedFallback || len(fb.SuffixIDs) != 1 || fb.SuffixIDs[0] != "huge" {
		t.Fatalf("expected 0.405 single-item fallback to pick 'huge', got %+v", fb)
	}

	tight := AdviseCachePolicy([]time.Duration{30 * time.Second, 45 * time.Second, 60 * time.Second}, 0.10, 1.25, 5*time.Minute)
	if !tight.RecommendStable {
		t.Fatalf("expected RecommendStable=true for rapid cadence, got %+v", tight)
	}
	sparse := AdviseCachePolicy([]time.Duration{20 * time.Minute, 30 * time.Minute, 45 * time.Minute}, 0.10, 1.25, 5*time.Minute)
	if sparse.RecommendStable {
		t.Fatalf("expected RecommendStable=false for sparse cadence, got %+v", sparse)
	}
}

func TestClassifyQueryAndFormatExplain(t *testing.T) {
	if got := ClassifyQuery("What was the pool_size value on 2026-11-10?"); got != ClassLookup {
		t.Fatalf("ClassifyQuery lookup = %q", got)
	}
	if got := ClassifyQuery("How did our rollout decision change over time?"); got != ClassTimeline {
		t.Fatalf("ClassifyQuery timeline = %q", got)
	}
	if got := ClassifyQuery("Refactor the parser and fix the bug"); got != ClassCode {
		t.Fatalf("ClassifyQuery code = %q", got)
	}
	if got := ClassifyQuery("Compare all options across the conversation"); got != ClassSynthesis {
		t.Fatalf("ClassifyQuery synthesis = %q", got)
	}
	txt := FormatExplain(ExplainPlan{
		Query: "test", Class: ClassGeneral, SelectedProfile: "claude/opus",
		Budget:     Budget{Planned: 1200, Cap: 120000, Brief: 400},
		Sections:   []SectionCost{{Key: KeyRecent, Count: 2, Tokens: 300}},
		LaneCounts: map[string]int{"exact": 1, "full_text": 2, "value": 0, "working_set": 1},
	})
	if !strings.Contains(txt, "CONTEXT PLAN") || !strings.Contains(txt, "claude/opus") {
		t.Fatalf("unexpected FormatExplain output:\n%s", txt)
	}
}
