package route

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
)

func TestPenalizeHonorsNamedResetsAndWindows(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	h, err := store.OpenFile(filepath.Join(cfg.Root, "machine.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	r, err := New(cfg, h)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 14, 17, 0, 0, 0, time.UTC)
	r.now = func() time.Time { return now }
	cases := []struct {
		provider, text string
		min, max       time.Duration
	}{
		{"muse", "Subscription quota exhausted. Your usage window resets at 2026-11-16T00:00:00Z", 31 * time.Hour, 31*time.Hour + 2*time.Minute},
		{"kimi", "Error: EPERM: operation not permitted, watch '/Users/x'\nprovider.api_error: You've reached your weekly (7-day) usage limit.", 6 * time.Hour, 6 * time.Hour},
		{"codex", "429 Too Many Requests", 15 * time.Minute, 15 * time.Minute},
		{"grok", "rate limit, retry after 90 seconds", 90 * time.Second, 90 * time.Second},
	}
	for _, c := range cases {
		d := r.Penalize(c.provider, providers.Quota, 0, c.text)
		if d < c.min || d > c.max {
			t.Errorf("%s: cooldown %s, want %s to %s", c.provider, d, c.min, c.max)
		}
	}
	if ex := providers.QuotaExcerpt(cases[1].text); ex == "" || ex[:5] == "Error" {
		t.Errorf("the excerpt should name the quota line, got %q", ex)
	}
}

func TestRankByStatsReordersByReliabilityAndSpeed(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.StatsRouting = true
	h, err := store.OpenFile(filepath.Join(cfg.Root, "machine.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	r, err := New(cfg, h)
	if err != nil {
		t.Fatal(err)
	}
	cands := []providers.Profile{
		{ID: "slow-flaky", Provider: "a"},
		{ID: "fast-solid", Provider: "b"},
		{ID: "unseen", Provider: "c"},
	}
	stats := []store.Stat{
		{Profile: "slow-flaky", Kind: "respond", Runs: 6, OK: 2, MedianWall: 40000},
		{Profile: "fast-solid", Kind: "respond", Runs: 6, OK: 6, MedianWall: 1800},
	}
	ranked := r.RankByStats(cands, "respond", stats)
	if ranked[0].ID != "fast-solid" {
		t.Fatalf("expected fast-solid ranked first, got %+v", ranked)
	}
	if ranked[len(ranked)-1].ID != "slow-flaky" {
		t.Fatalf("expected slow-flaky ranked last, got %+v", ranked)
	}
}
