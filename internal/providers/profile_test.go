package providers

import (
	"math"
	"testing"
)

func TestSeedProfiles(t *testing.T) {
	profiles := SeedProfiles()
	ids := map[string]Profile{}
	for _, p := range profiles {
		if err := p.Validate(); err != nil {
			t.Error(err)
		}
		if _, dup := ids[p.ID]; dup {
			t.Errorf("duplicate id %s", p.ID)
		}
		ids[p.ID] = p
	}
	want := map[string]struct {
		window, effective int
		live              string
		kill, plan        float64
		enabled           bool
	}{
		"claude/claude-opus-4-7/high":         {200_000, 200_000, LivePerTurn, 0.85, 0.70, true},
		"claude/claude-sonnet-4-6/medium":     {200_000, 200_000, LivePerTurn, 0.85, 0.70, true},
		"claude/claude-opus-4-7[1m]/high":     {1_000_000, 1_000_000, LivePerTurn, 0.85, 0.70, true},
		"codex/gpt-6.1-sol/high":              {500_000, 500_000, LiveNone, 0.85, 0.60, true},
		"codex/gpt-6-sol/high":                {272_000, 272_000, LiveNone, 0.85, 0.60, true},
		"codex/gpt-5.6-sol/high":              {272_000, 272_000, LiveNone, 0.85, 0.60, true},
		"gemini/gemini-3-pro-preview":         {1_000_000, 500_000, LiveNone, 0.85, 0.60, true},
		"gemini/gemini-3-flash-preview":       {1_000_000, 500_000, LiveNone, 0.85, 0.60, true},
		"kimi/kimi-code/k3/high":              {1_048_576, 1_048_576, LiveNone, 0.80, 0.60, true},
		"kimi/kimi-code/kimi-for-coding/high": {1_048_576, 1_048_576, LiveNone, 0.80, 0.60, true},
		"grok/grok-4.5/high":                  {128_000, 128_000, LiveNone, 0.85, 0.60, false},
		"muse/muse-spark-1.3/high":            {128_000, 128_000, LivePerTurn, 0.85, 0.70, false},
	}
	for id, w := range want {
		p, ok := ids[id]
		if !ok {
			t.Errorf("missing profile %s", id)
			continue
		}
		if p.Window != w.window || p.EffectiveWindow != w.effective || p.LiveUsage != w.live ||
			p.KillFraction != w.kill || p.PlanFraction != w.plan || p.Enabled != w.enabled {
			t.Errorf("%s: got %+v", id, p)
		}
		if p.TaintFraction != 0.90 || p.TokensPerWord != 1.5 {
			t.Errorf("%s: taint %v tpw %v", id, p.TaintFraction, p.TokensPerWord)
		}
	}
	claude := ids["claude/claude-opus-4-7/high"]
	if claude.Ceiling() != 180_000 || claude.PlanCap() != 140_000 || claude.KillLine() != 170_000 || claude.TaintLine() != 180_000 {
		t.Errorf("claude lines: %d %d %d", claude.Ceiling(), claude.PlanCap(), claude.KillLine())
	}
	if ids["codex/gpt-6.1-sol/high"].Ceiling() != 450_000 {
		t.Error("codex ceiling")
	}
	if ids["gemini/gemini-3-pro-preview"].PlanCap() != 300_000 {
		t.Error("gemini plan cap must use the 0.60 rule on 500k")
	}
	if ids["claude/claude-opus-4-7/high"].RateIn != 0 || ids["gemini/gemini-3-pro-preview"].RateIn == 0 {
		t.Error("rates: subscription 0, api billed non-zero")
	}
}

func TestProfileID(t *testing.T) {
	cases := []struct{ id, provider, model, effort string }{
		{"claude/claude-opus-4-7/high", "claude", "claude-opus-4-7", "high"},
		{"kimi/kimi-code/k3/high", "kimi", "kimi-code/k3", "high"},
		{"kimi/kimi-code/k3", "kimi", "kimi-code/k3", ""},
		{"gemini/gemini-3-pro-preview", "gemini", "gemini-3-pro-preview", ""},
	}
	for _, c := range cases {
		p, m, e, ok := ParseProfileID(c.id)
		if !ok || p != c.provider || m != c.model || e != c.effort {
			t.Errorf("%s: %q %q %q %v", c.id, p, m, e, ok)
		}
		if ProfileID(p, m, e) != c.id {
			t.Errorf("roundtrip %s", c.id)
		}
	}
	if _, _, _, ok := ParseProfileID("claude"); ok {
		t.Error("bare provider must fail")
	}
}

func TestEstimateUSD(t *testing.T) {
	p := Profile{RateIn: 2, RateOut: 12}
	u := Usage{InputTokens: 1_000_000, CachedTokens: 500_000, OutputTokens: 100_000, ReasoningTokens: 50_000}
	got := EstimateUSD(p, u)
	want := 500_000*2/1e6 + 500_000*2*0.1/1e6 + 100_000*12/1e6
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("got %v want %v", got, want)
	}
	if EstimateUSD(Profile{}, u) != 0 {
		t.Error("subscription estimate must be 0")
	}
}

func TestValidate(t *testing.T) {
	p := testProfile("claude")
	p.KillFraction = 0.95
	if p.Validate() == nil {
		t.Error("kill above taint must fail")
	}
	p = testProfile("claude")
	p.PlanFraction = 0.9
	if p.Validate() == nil {
		t.Error("plan above 0.8 must fail")
	}
}
