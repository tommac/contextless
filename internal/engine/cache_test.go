package engine

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// conversationHistory is n exchanges, each user turn asking about a topic
// and each answer long enough to rank.
func conversationHistory(t *testing.T, h *harness, n int) {
	t.Helper()
	topics := []string{"garden", "garage", "attic", "kitchen"}
	for i := 1; i <= n; i++ {
		topic := topics[i%len(topics)]
		u, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", fmt.Sprintf("What is left to do in the %s for week %d?", topic, i))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := h.e.V.AppendTurn(vault.RoleAssistant, "", "", u.ID, longTurn(topic, 4, 0, "")); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
}

func replayBrief(t *testing.T, h *harness, id string) string {
	t.Helper()
	u, err := h.e.V.Turn(id)
	if err != nil {
		t.Fatal(err)
	}
	cands, _ := h.e.R.Candidates(config.RoleResponder, nil)
	run := "run-" + id
	task := packet.Task{RunID: run, Kind: packet.KindRespond, Role: config.RoleResponder, Instructions: respondInstructions(run), Schema: packet.RespondSchema(true)}
	s, err := h.e.buildPacket(task, u, packet.SelectOptions{}, cands, packet.OutputTokens[packet.KindRespond])
	if err != nil {
		t.Fatal(err)
	}
	return s.Brief
}

// The brief runs from what every call shares to what is specific to this
// one: no run id before the task, and the message to answer last.
func TestBriefRunsFromStableToVolatile(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	conversationHistory(t, h, 30)
	b := replayBrief(t, h, "t000059")
	last := -1
	for _, section := range []string{"## Rules", "## Session", "## Catalog", "## Conversation records (verbatim", "## Conversation records chosen", "## Task", "## The message to answer"} {
		at := strings.Index(b, section)
		if at <= last {
			t.Fatalf("%q is at %d, after %d", section, at, last)
		}
		last = at
	}
	if at := strings.Index(b, "run-t000059"); at < strings.Index(b, "## Task") {
		t.Fatal("the run id comes before the task")
	}
}

// Consecutive messages share everything up to the end of the recent
// turns while the window keeps its start, whatever is ranked for each.
func TestConsecutiveBriefsShareTheirPrefix(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	conversationHistory(t, h, 30)
	packet.CatalogSpine = true
	defer func() { packet.CatalogSpine = false }()
	h.e.Cfg.RecentEpoch = 4
	h.e.Cfg.RecentTurns = 8
	a, b := replayBrief(t, h, "t000053"), replayBrief(t, h, "t000055")
	shared := commonPrefix(a, b)
	end := strings.Index(a, "## Conversation records chosen")
	if end < 0 || shared < end {
		t.Fatalf("the briefs part at byte %d, before the chosen records at %d: ...%q", shared, end, a[max(0, shared-60):min(len(a), shared+60)])
	}
	// the window grows from an aligned start instead of sliding
	h.e.Cfg.RecentEpoch = -1
	a, b = replayBrief(t, h, "t000053"), replayBrief(t, h, "t000055")
	if commonPrefix(a, b) >= strings.Index(a, "## Conversation records chosen") {
		t.Fatal("a sliding window should change the prefix")
	}
}

// A turn sent whole for one message keeps its catalog line, so that
// message does not change the catalog the next one shares.
func TestCatalogListsTurnsSentWhole(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	conversationHistory(t, h, 30)
	packet.CatalogSpine = true
	defer func() { packet.CatalogSpine = false }()
	u, _ := h.e.V.Turn("t000059")
	s, err := packet.Select(h.e.V, h.e.S, u, packet.SelectOptions{Recent: 8, RelevantTokens: 8000, Pinned: []string{"t000003"}})
	if err != nil {
		t.Fatal(err)
	}
	b := packet.Render(packet.Task{RunID: "r", Kind: packet.KindRespond}, s, nil)
	cat := b[strings.Index(b, "## Catalog"):strings.Index(b, "## Conversation records (verbatim")]
	if !strings.Contains(cat, "- t000003 user") || !strings.Contains(b, "### t000003 ·") {
		t.Fatal("a recalled turn should be both in the catalog and whole in the records")
	}
}

func TestCacheableTokensFollowTheProviderRule(t *testing.T) {
	for in, want := range map[int]int{0: 0, 1023: 0, 1024: 1024, 1151: 1024, 1152: 1152, 5000: 4992} {
		if got := cacheableTokens(in); got != want {
			t.Errorf("cacheableTokens(%d) = %d, want %d", in, got, want)
		}
	}
	r := CacheReport([]ReplayCall{{Brief: 1000, Cached: 0}, {Brief: 1000, Cached: 800}}, 0.1, 1)
	if !strings.Contains(r, "640 if every call hits the cache, 820 if half do, 1000 if none do") {
		t.Fatalf("report:\n%s", r)
	}
}
