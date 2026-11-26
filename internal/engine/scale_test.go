package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// TestLongConversationNeverLosesATurn builds a conversation about four
// times the largest window, all in one thread so selection pulls every
// turn, and checks the packets: each fits its cap, the focus and recent
// turns are verbatim, every other turn is either verbatim or named (catalog,
// omitted list or id range), and a recalled old turn returns byte for byte.
func TestLongConversationNeverLosesATurn(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	const n = 1200
	th := vault.Thread{ID: "th-bigthread", Title: "The long project", Status: "open", Created: time.Now()}
	if err := h.e.V.PutThread(th); err != nil {
		t.Fatal(err)
	}
	bodies := map[string]string{}
	for i := 1; i <= n; i++ {
		role := vault.RoleUser
		if i%2 == 0 {
			role = vault.RoleAssistant
		}
		body := fmt.Sprintf("Turn %d of the long project. Marker M%05d. ", i, i) + strings.Repeat(fmt.Sprintf("detail-%d ", i%97), 400)
		tu, err := h.e.V.AppendTurn(role, "", "", "", body)
		if err != nil {
			t.Fatal(err)
		}
		bodies[tu.ID] = tu.Body
		if err := h.e.V.PutCard(vault.Card{Turn: tu.ID, BodySHA: tu.BodySHA, Summary: fmt.Sprintf("Step %d of the long project.", i), Threads: []string{th.ID}}); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, b := range bodies {
		total += len(b)
	}
	if tok := packet.Tokens(providers.Profile{}, total); tok < 4*280_000 {
		t.Fatalf("history is only %d tokens; the test needs it past every window", tok)
	}

	target := "t000007"
	round := 0
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		round++
		checkCoverage(t, c.Brief, n)
		if round == 1 {
			if strings.Contains(c.Brief, "### "+target+" ·") {
				t.Errorf("%s unexpectedly verbatim in round 1; pick an older target", target)
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall, RecallIDs: []string{target}}
		}
		if !strings.Contains(c.Brief, "### "+target+" ·") || !strings.Contains(c.Brief, bodies[target]) {
			t.Errorf("recalled %s is not verbatim", target)
		}
		return done(c, "Marker M00007.")
	}

	start := time.Now()
	r, err := h.e.Ask(context.Background(), "What was the marker in turn 7 of the long project?")
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if r.Assistant.Body != "Marker M00007." || round != 2 {
		t.Fatalf("answer %q after %d rounds", r.Assistant.Body, round)
	}
	// every respond run's manifest fits its cap, and the recent turns are in
	for _, ri := range r.Runs {
		m := readManifest(t, filepath.Join(h.e.RunsDir, ri.ID, "manifest.json"))
		if m.Budget.Planned > m.Budget.Cap {
			t.Fatalf("run %s planned %d over cap %d", ri.ID, m.Budget.Planned, m.Budget.Cap)
		}
		if len(m.Omitted) == 0 {
			t.Fatalf("run %s omitted nothing; history should exceed the window", ri.ID)
		}
		in := map[string]bool{}
		for _, it := range m.Records {
			in[it.ID] = true
		}
		for i := n - 7; i <= n; i++ {
			if !in[vault.TurnID(i)] {
				t.Fatalf("recent turn %s missing from run %s", vault.TurnID(i), ri.ID)
			}
		}
		t.Logf("run %s on %s: %d records verbatim, %d omitted, %d of %d tokens", ri.ID, ri.Profile, len(m.Records), len(m.Omitted), m.Budget.Planned, m.Budget.Cap)
	}
	t.Logf("%d turns, %d MB of history, two packets built and run in %s", n, total>>20, elapsed.Round(time.Millisecond))
	if elapsed > 20*time.Second {
		t.Fatalf("selection and sizing too slow: %s", elapsed)
	}
}

var (
	verbatimRe = regexp.MustCompile(`(?m)^### (t\d{6}) ·(.*)$`)
	listedRe   = regexp.MustCompile(`(?m)^- (t\d{6}) `)
	rangeRe    = regexp.MustCompile(`t(\d{6})–t(\d{6})`)
	singleRe   = regexp.MustCompile(`\bt(\d{6})\b`)
)

// checkCoverage fails unless every earlier turn is verbatim or named.
func checkCoverage(t *testing.T, brief string, n int) {
	t.Helper()
	seen := map[int]bool{}
	mark := func(id string) {
		if s, err := strconv.Atoi(strings.TrimPrefix(id, "t")); err == nil {
			seen[s] = true
		}
	}
	for _, m := range verbatimRe.FindAllStringSubmatch(brief, -1) {
		// an excerpt is not the turn: an excerpted turn must still be
		// named by a catalog line, a range or the omitted list
		if !strings.Contains(m[2], " excerpt: ") && !strings.Contains(m[2], " part ") {
			mark(m[1])
		}
	}
	for _, m := range listedRe.FindAllStringSubmatch(brief, -1) {
		mark(m[1])
	}
	if i := strings.Index(brief, "## Omitted for budget"); i >= 0 {
		om := brief[i:]
		for _, m := range rangeRe.FindAllStringSubmatch(om, -1) {
			a, _ := strconv.Atoi(m[1])
			b, _ := strconv.Atoi(m[2])
			for s := a; s <= b; s++ {
				seen[s] = true
			}
		}
		for _, m := range singleRe.FindAllStringSubmatch(om, -1) {
			mark("t" + m[1])
		}
	}
	for s := 1; s <= n; s++ {
		if !seen[s] {
			t.Fatalf("turn %s is neither in the packet nor named in it", vault.TurnID(s))
		}
	}
}

func readManifest(t *testing.T, path string) manifest {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}
