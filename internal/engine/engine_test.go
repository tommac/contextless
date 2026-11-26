package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// scripted is a provider adapter whose "CLI" is /bin/cp of a handoff that a
// Go function wrote from the brief, so tests drive every engine path with
// real process runs and no model.
type scripted struct {
	name  string
	dir   string
	reply func(c call) any // a Handoff, an IndexHandoff, or a fault
	mu    *sync.Mutex
	calls *[]call
}

type call struct {
	Provider, Kind, Role, RunID, Brief string
}

type fault struct{ stderr string }

// withEdits writes files into the workspace before the handoff, as an
// agent editing the checkout would.
type withEdits struct {
	H     any
	Edits map[string]string
}

var hdrRe = regexp.MustCompile("run `([^`]+)` · kind `([^`]+)` · role `([^`]+)`")

func (s scripted) Provider() string { return s.name }
func (s scripted) BuildCommand(inv providers.Invocation) ([]string, []string, []byte, error) {
	b, err := os.ReadFile(inv.BriefPath)
	if err != nil {
		return nil, nil, nil, err
	}
	m := hdrRe.FindStringSubmatch(string(b))
	c := call{Provider: s.name, RunID: m[1], Kind: m[2], Role: m[3], Brief: string(b)}
	s.mu.Lock()
	*s.calls = append(*s.calls, c)
	s.mu.Unlock()
	out := s.reply(c)
	if we, ok := out.(withEdits); ok {
		for rel, body := range we.Edits {
			full := filepath.Join(inv.Workspace, rel)
			_ = os.MkdirAll(filepath.Dir(full), 0o755)
			if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
				return nil, nil, nil, err
			}
		}
		out = we.H
	}
	if f, ok := out.(fault); ok {
		return []string{"/bin/sh", "-c", fmt.Sprintf("echo %q >&2; exit 1", f.stderr)}, os.Environ(), nil, nil
	}
	data, _ := json.Marshal(out)
	tmp := filepath.Join(s.dir, c.RunID+".json")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return nil, nil, nil, err
	}
	return []string{"/bin/cp", tmp, inv.HandoffPath}, os.Environ(), nil, nil
}
func (scripted) ParseEvent(line []byte) (providers.Event, bool) { return providers.Event{}, false }
func (scripted) FinalUsage(stdout []byte, ev []providers.Event) (providers.Usage, error) {
	return providers.Usage{}, providers.ErrNoUsage
}
func (scripted) ExitClass(code int, stderr []byte, ev []providers.Event) providers.ExitClass {
	if code == 0 {
		return providers.Ok
	}
	if c := providers.SignatureClass(string(stderr)); c != "" {
		return c
	}
	return providers.Crash
}
func (scripted) InstructionFile() string                                    { return "AGENTS.md" }
func (scripted) ProjectSettings(inv providers.Invocation) map[string][]byte { return nil }

type harness struct {
	t     *testing.T
	e     *Engine
	mu    sync.Mutex
	calls []call
	reply func(c call) any
}

// newHarness builds a workspace with three scripted providers: alpha (a
// small window), beta and gamma (large windows). window sets alpha's.
func newHarness(t *testing.T, alphaWindow int, reply func(c call) any) *harness {
	t.Helper()
	root := t.TempDir()
	cfg := config.Default()
	cfg.Root = root
	cfg.Sandbox = false
	cfg.Enable = nil
	// the tests are about selection, recall and sizing, so the rule that
	// sends a short history whole is off except where a test turns it on
	cfg.WholeHistoryTokens = 0
	prof := func(provider string, window int) providers.Profile {
		return providers.Profile{Provider: provider, Model: "m1", Window: window, EffectiveWindow: window,
			SysOverhead: 1000, LiveUsage: providers.LivePerTurn, NativeSchema: true, Enabled: true, MaxConcurrent: 2}
	}
	cfg.Profiles = []providers.Profile{prof("alpha", alphaWindow), prof("beta", 400_000), prof("gamma", 400_000)}
	cfg.Lanes = map[string]int{"alpha": 2, "beta": 2, "gamma": 2}
	cfg.Routing = map[string][]string{
		config.RoleResponder:  {"alpha/m1", "beta/m1"},
		config.RoleCoder:      {"beta/m1", "alpha/m1"},
		config.RoleResearcher: {"gamma/m1"},
		config.RoleLongReader: {"beta/m1", "gamma/m1"},
		config.RoleAssembler:  {"alpha/m1", "gamma/m1"},
		config.RoleIndexer:    {"gamma/m1"},
	}
	h := &harness{t: t, reply: reply}
	e, err := Create(cfg, "test", "Test chat", "testing")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { e.Close() })
	for _, p := range []string{"alpha", "beta", "gamma"} {
		e.R.SetAdapter(p, scripted{name: p, dir: root, mu: &h.mu, calls: &h.calls, reply: func(c call) any { return h.reply(c) }})
	}
	h.e = e
	return h
}

func (h *harness) callsOf(kind string) []call {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []call
	for _, c := range h.calls {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

func done(c call, answer string) packet.Handoff {
	return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: answer}
}

func TestAskStoresVerbatimAndIndexes(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return done(c, "Use **Postgres 17**; the billing cutover can happen Friday.")
		case packet.KindIndex:
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed",
				Cards: []packet.IndexCard{
					{Turn: "t000001", Summary: "User asks which database to use for billing.", Threads: []string{"new:billing"}, Mentions: []string{"new:pg"}},
					{Turn: "t000002", Summary: "Assistant recommends Postgres 17 and a Friday cutover.", Threads: []string{"new:billing"}, Mentions: []string{"new:pg"}},
				},
				NewThreads:  []packet.NewThread{{Key: "new:billing", Title: "Billing database migration"}},
				NewEntities: []packet.NewEntity{{Key: "new:pg", Name: "Postgres", Kind: "product"}},
				Facts: []packet.FactOp{
					{Statement: "Billing runs on MySQL today", Source: "t000001", Quote: "billing   runs on MySQL", Entities: []string{"new:pg"}},
					{Statement: "invented", Source: "t000001", Quote: "this quote is not in the turn"},
				},
				Decisions: []packet.DecisionOp{{Statement: "Cut over on Friday", Source: "t000002", Quote: "cutover can happen Friday"}},
			}
		}
		return fault{"unexpected"}
	}
	msg := "Our billing runs on MySQL today.\n\n  Which database should we move it to?  "
	r, err := h.e.Ask(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}
	h.e.Wait()
	if r.User.Body != strings.TrimRight(msg, " \n") {
		t.Fatalf("user turn not verbatim: %q", r.User.Body)
	}
	got, err := h.e.V.Turn(r.Assistant.ID)
	if err != nil || got.Body != "Use **Postgres 17**; the billing cutover can happen Friday." || got.Author != "alpha/m1" {
		t.Fatalf("assistant turn = %+v, %v", got, err)
	}
	if len(h.e.V.Cards()) != 2 || len(h.e.V.Threads()) != 1 || len(h.e.V.Entities()) != 1 {
		t.Fatalf("index not applied: cards %d threads %d entities %d", len(h.e.V.Cards()), len(h.e.V.Threads()), len(h.e.V.Entities()))
	}
	facts := h.e.V.Facts()
	if len(facts) != 1 || facts[0].Statement != "Billing runs on MySQL today" {
		t.Fatalf("want the one verbatim-anchored fact, got %+v", facts)
	}
	if ds := h.e.V.Decisions(); len(ds) != 1 {
		t.Fatalf("decisions = %+v", ds)
	}
	out, err := exec.Command("git", "-C", h.e.V.Dir, "log", "--oneline").CombinedOutput()
	if err != nil || strings.Count(string(out), "\n") < 3 {
		t.Fatalf("git log: %s %v", out, err)
	}

	// the next message's packet carries the decision (tier 0) and the
	// thread's turns verbatim
	h.reply = func(c call) any {
		if c.Kind == packet.KindRespond {
			if !strings.Contains(c.Brief, "Cut over on Friday") {
				t.Errorf("decision missing from the next packet")
			}
			if !strings.Contains(c.Brief, "Use **Postgres 17**; the billing cutover can happen Friday.") {
				t.Errorf("earlier answer not verbatim in the next packet")
			}
			return done(c, "ok")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), "Remind me what we decided about the billing cutover?"); err != nil {
		t.Fatal(err)
	}
}

func TestRecallPutsRequestedTurnsBackVerbatim(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	secret := "Notes from the long planning call about the shed, the fence, the compost bins and the drainage along the north wall of the garden. The vault combination is 31-7-22, and only Dana may use it."
	for i := 1; i <= 30; i++ {
		body := fmt.Sprintf("Filler turn %d about the garden.", i)
		if i == 3 {
			body = secret
		}
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	round := 0
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		round++
		if round == 1 {
			if strings.Contains(c.Brief, "### t000003 · user") || strings.Contains(c.Brief, "31-7-22") {
				t.Errorf("turn 3 should only be in the catalog on round 1")
			}
			if !strings.Contains(c.Brief, "t000003 user (not yet indexed, opens with)") {
				t.Errorf("catalog line for t000003 missing")
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall, RecallIDs: []string{"t000003", "t999999", "t000031"}}
		}
		if !strings.Contains(c.Brief, "### t000003 · user") || !strings.Contains(c.Brief, secret) {
			t.Errorf("recalled turn not verbatim in round 2")
		}
		return done(c, "It is 31-7-22 (Dana only).")
	}
	r, err := h.e.Ask(context.Background(), "What is the safe code?")
	if err != nil {
		t.Fatal(err)
	}
	if round != 2 || r.Assistant.Body != "It is 31-7-22 (Dana only)." {
		t.Fatalf("round %d answer %q", round, r.Assistant.Body)
	}
}

func TestPartialIsContinued(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPartial, Answer: "Chapter 1 text.",
				UnitsDone: []string{"ch1"}, UnitsRemaining: []string{"ch2"}, NextStep: "write chapter 2"}
		case packet.KindContinue:
			if !strings.Contains(c.Brief, "Chapter 1 text.") || !strings.Contains(c.Brief, "write chapter 2") {
				t.Errorf("continuation lacks the work so far or next_step")
			}
			return done(c, "Chapter 2 text.")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	r, err := h.e.Ask(context.Background(), "Write two chapters.")
	if err != nil {
		t.Fatal(err)
	}
	if r.Assistant.Body != "Chapter 1 text.\n\nChapter 2 text." {
		t.Fatalf("answer %q", r.Assistant.Body)
	}
}

func TestPlanFansOutAcrossProvidersAndAssembles(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
				{ID: "u1", Role: "coder", Title: "Write the parser", Instructions: "Write a CSV parser in Go."},
				{ID: "u2", Role: "researcher", Title: "Find the RFC", Instructions: "Find the CSV RFC number."},
			}}
		case packet.KindUnit:
			if c.Role == "coder" {
				return done(c, "func Parse() {}")
			}
			return done(c, "RFC 4180")
		case packet.KindAssemble:
			if !strings.Contains(c.Brief, "func Parse() {}") || !strings.Contains(c.Brief, "RFC 4180") {
				t.Errorf("assembler missing unit outputs")
			}
			return done(c, "Here is the parser, following RFC 4180: func Parse() {}")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	r, err := h.e.Ask(context.Background(), "Write a CSV parser and tell me which RFC it follows.")
	if err != nil {
		t.Fatal(err)
	}
	units := h.callsOf(packet.KindUnit)
	byRole := map[string]string{}
	for _, u := range units {
		byRole[u.Role] = u.Provider
	}
	if byRole["coder"] != "beta" || byRole["researcher"] != "gamma" {
		t.Fatalf("units routed to %v", byRole)
	}
	if !strings.Contains(r.Assistant.Body, "RFC 4180") || !strings.Contains(r.Assistant.Author, "assembling") {
		t.Fatalf("assembled answer %q by %q", r.Assistant.Body, r.Assistant.Author)
	}
	arts, _ := filepath.Glob(filepath.Join(h.e.V.Dir, "artifacts", "*", "u*.md"))
	if len(arts) != 2 {
		t.Fatalf("unit artifacts = %v", arts)
	}
}

func TestQuotaFailsOverAndCools(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.reply = func(c call) any {
		if c.Provider == "alpha" {
			return fault{"Error: 429 Too Many Requests: rate limit reached"}
		}
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		return done(c, "answered by beta")
	}
	r, err := h.e.Ask(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if r.Assistant.Author != "beta/m1" {
		t.Fatalf("author %q", r.Assistant.Author)
	}
	if until, _ := h.e.R.Cooling("alpha"); until.IsZero() {
		t.Fatalf("alpha should be cooling after a 429")
	}
	cands, _ := h.e.R.Candidates(config.RoleResponder, nil)
	if len(cands) != 1 || cands[0].Provider != "beta" {
		t.Fatalf("candidates while alpha cools: %v", cands)
	}
}

func TestOversizedMessageIsSplitNotSummarized(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	// shrink every window so a 400 KB message fits none
	for _, p := range []string{"alpha", "beta", "gamma"} {
		prof, _ := h.e.R.Profile(p + "/m1")
		prof.Window, prof.EffectiveWindow = 80_000, 80_000
		h.e.R.SetProfile(prof)
	}
	var para []string
	for i := 0; i < 4000; i++ {
		para = append(para, fmt.Sprintf("Paragraph %04d: the quick brown fox jumps over the lazy dog, line item %d.", i, i))
	}
	big := "Count the paragraphs that mention a fox.\n\n" + strings.Join(para, "\n\n")
	var mu sync.Mutex
	seen := map[string]bool{}
	var asmBrief string
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindUnit:
			mu.Lock()
			last := ""
			for _, m := range regexp.MustCompile(`Paragraph (\d{4})`).FindAllStringSubmatch(c.Brief, -1) {
				seen[m[1]] = true
				last = m[1]
			}
			mu.Unlock()
			// each part cites the message, quoting the last paragraph it holds
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "some foxes",
				Claims: []packet.UnitClaim{{Claim: "This part mentions a fox.", Evidence: []string{"t000001"},
					Quote: "Paragraph " + last + ": the quick brown fox", Support: "quoted"}}}
		case packet.KindAssemble:
			asmBrief = c.Brief
			return done(c, "4000 paragraphs mention a fox.")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	r, err := h.e.Ask(context.Background(), big)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.callsOf(packet.KindUnit)) < 2 {
		t.Fatalf("expected several part runs, got %d", len(h.callsOf(packet.KindUnit)))
	}
	if len(seen) != 4000 {
		t.Fatalf("every paragraph must reach some part verbatim; saw %d of 4000", len(seen))
	}
	if r.Assistant.Body != "4000 paragraphs mention a fox." {
		t.Fatalf("answer %q", r.Assistant.Body)
	}
	n := len(h.callsOf(packet.KindUnit))
	if got := strings.Count(asmBrief, "quoted, quote found in t000001#a"); got != n {
		t.Fatalf("%d of %d parts' quotes were found in the message", got, n)
	}
	if _, err := os.Stat(filepath.Join(h.e.V.Dir, "artifacts", "t000001-parts", "part1.json")); err != nil {
		t.Fatalf("the parts' records are not stored: %v", err)
	}
}

func TestSplitParagraphsIsLossless(t *testing.T) {
	text := strings.Repeat("alpha beta gamma.\n\nδέλτα ε ζ η θ ι κ λ\n", 500)
	parts := splitParagraphs(text, 700)
	if strings.Join(parts, "") != text {
		t.Fatal("parts do not reassemble the text")
	}
	for _, p := range parts {
		if len(p) > 700 {
			t.Fatalf("part of %d bytes", len(p))
		}
	}
}

func TestLockRecoveryRetryAndCritiques(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	// a second writer on the same workspace is refused; a reader is fine
	if _, err := Open(h.e.Cfg, "test"); err == nil || !strings.Contains(err.Error(), "in use") {
		t.Fatalf("second Open: %v", err)
	}
	ro, err := OpenReadOnly(h.e.Cfg, "test")
	if err != nil {
		t.Fatal(err)
	}
	ro.Close()

	// every provider fails: the user turn stays, unanswered
	h.reply = func(c call) any { return fault{"segmentation fault"} }
	if _, err := h.e.Ask(context.Background(), "first try"); err == nil {
		t.Fatal("want an error when every provider fails")
	}
	p, ok := h.e.Pending()
	if !ok || p.Body != "first try" {
		t.Fatalf("pending = %+v %v", p, ok)
	}
	// the second attempt's brief names the first attempt's failure
	calls := h.callsOf(packet.KindRespond)
	if len(calls) < 2 || !strings.Contains(calls[1].Brief, "Earlier attempts at this same call failed") || !strings.Contains(calls[1].Brief, "alpha/m1") {
		t.Fatalf("critique not carried into attempt 2")
	}

	// a run left "running" by a dead process is abandoned on the next open
	cfg := h.e.Cfg
	_, _ = h.e.S.DB.Exec(`INSERT INTO runs(id,kind,role,profile,status,started) VALUES('run-dead','respond','responder','alpha/m1','running','2026-01-01T00:00:00Z')`)
	h.e.Close()
	e2, err := Open(cfg, "test")
	if err != nil {
		t.Fatal(err)
	}
	h.e = e2
	for _, prov := range []string{"alpha", "beta", "gamma"} {
		e2.R.SetAdapter(prov, scripted{name: prov, dir: t.TempDir(), mu: &h.mu, calls: &h.calls, reply: func(c call) any { return h.reply(c) }})
		_ = e2.health.Cooldown(prov, "reset", time.Now().Add(-time.Minute))
	}
	if e2.Abandoned != 1 {
		t.Fatalf("abandoned = %d", e2.Abandoned)
	}
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		return done(c, "second try worked")
	}
	r, err := e2.Retry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.User.ID != p.ID || r.Assistant.Body != "second try worked" {
		t.Fatalf("retry answered %s with %q", r.User.ID, r.Assistant.Body)
	}
	if _, ok := e2.Pending(); ok {
		t.Fatal("still pending after retry")
	}
}

func TestReviewOnAnotherProviderRevises(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.e.Cfg.Review = "always"
	h.e.Cfg.Routing[config.RoleReviewer] = []string{"alpha/m1", "gamma/m1"}
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return done(c, "Paris is the capital of Germany.")
		case packet.KindReview:
			if c.Provider == "alpha" {
				t.Errorf("the writer's provider reviewed its own answer")
			}
			if !strings.Contains(c.Brief, "Answer under review") || !strings.Contains(c.Brief, "Paris is the capital of Germany.") {
				t.Errorf("reviewer did not get the answer verbatim")
			}
			return packet.ReviewHandoff{RunID: c.RunID, Status: "completed", Verdict: "revise", Problems: []packet.ReviewProblem{
				{Quote: "capital of Germany", Problem: "Paris is the capital of France; Berlin is Germany's.", Fix: "Say France."},
				{Quote: "a quote that is not there", Problem: "invented"},
			}}
		case packet.KindRevise:
			if !strings.Contains(c.Brief, "Berlin is Germany's") || strings.Contains(c.Brief, "invented") {
				t.Errorf("revision brief should carry only the verbatim-anchored problem")
			}
			return done(c, "Paris is the capital of France.")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	r, err := h.e.Ask(context.Background(), "What is Paris the capital of?")
	if err != nil {
		t.Fatal(err)
	}
	if r.Assistant.Body != "Paris is the capital of France." || !strings.Contains(r.Assistant.Author, "revised after review by gamma/m1") {
		t.Fatalf("answer %q by %q", r.Assistant.Body, r.Assistant.Author)
	}
}
