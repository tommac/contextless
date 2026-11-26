package engine

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

var recordRe = regexp.MustCompile(`(?m)^### (t\d{6}\S*) ·.*\((t\d{6}#a[0-9-]+)\)|^### (t\d{6}) ·`)

func chosenRecords(brief string) []string {
	at := strings.Index(brief, "## Conversation records chosen")
	if at < 0 {
		return nil
	}
	section := brief[at:]
	if end := strings.Index(section, "\n## Task"); end >= 0 {
		section = section[:end]
	}
	var out []string
	for _, m := range recordRe.FindAllStringSubmatch(section, -1) {
		out = append(out, m[2]+m[3])
	}
	sort.Strings(out)
	return out
}

// A dry run ranks exactly as the real call for the same message does: the
// message is not in the index while its own packet is built.
func TestDryPacketRanksLikeTheRealCall(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	conversationHistory(t, h, 30)
	const msg = "What is left to do in the garage and the attic for week twelve, and which items stay open?"
	dry, err := h.e.DryPacket(msg)
	if err != nil {
		t.Fatal(err)
	}
	var live string
	h.reply = func(c call) any {
		if c.Kind == packet.KindRespond {
			live = c.Brief
			_, atoms := h.e.S.AtomTurns()["t000061"]
			if h.e.S.HasRecord("t000061") || atoms {
				t.Error("the message is in the index while its own packet is built")
			}
			return done(c, "ok")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	a, b := chosenRecords(dry.Brief), chosenRecords(live)
	if len(a) == 0 || strings.Join(a, " ") != strings.Join(b, " ") {
		t.Fatalf("dry %v, live %v", a, b)
	}
	if !h.e.S.HasRecord("t000061") {
		t.Fatal("the message should be indexed once it is answered")
	}
}

// An also turn counts through any accepted wording, and one that holds
// none is reported, since it can never count.
func TestProbesAcceptOtherWordings(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	p := Probe{ID: "p", Type: "t", Question: "Which depot holds the storage unit, number 214?", Source: "t000003", Answer: "at seven in the morning",
		Also: []string{"t000004", "t000005"}}
	res, _ := h.e.RunProbes([]Probe{p})
	if res[0].Any && res[0].Where != WhereVerbatim {
		t.Fatalf("no also turn holds the answer: %+v", res[0])
	}
	if strings.Join(res[0].AlsoBad, ",") != "t000004,t000005" {
		t.Fatalf("also_bad %v", res[0].AlsoBad)
	}
	p.Answers = []string{"number 214 at the Elm Street depot"}
	res, _ = h.e.RunProbes([]Probe{p})
	if !res[0].Any || strings.Join(res[0].AlsoBad, ",") != "t000005" {
		t.Fatalf("with the other wording: any %v, also_bad %v", res[0].Any, res[0].AlsoBad)
	}
	if !strings.Contains(ProbeReport(res), "1 also entries name a turn holding no accepted answer") {
		t.Fatal("the report should name the unusable entry")
	}
}

// Each context arm builds the packet it says it does.
func TestContextArms(t *testing.T) {
	defer func() { c := config.Default(); _ = ApplyContextArm(&c, ArmCurrent) }()
	const q = "When do the movers arrive on the ninth of June?"
	build := func(arm string) (*harness, packet.Sized) {
		h := newHarness(t, 200_000, nil)
		h.e.Cfg.WholeHistoryTokens = 60_000
		if err := ApplyContextArm(&h.e.Cfg, arm); err != nil {
			t.Fatal(err)
		}
		moveHistory(t, h)
		for i := 0; i < 20; i++ {
			if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", "Another short note about labelling the boxes for the flat."); err != nil {
				t.Fatal(err)
			}
		}
		if err := h.e.Reindex(); err != nil {
			t.Fatal(err)
		}
		s, err := h.e.DryPacket(q)
		if err != nil {
			t.Fatal(err)
		}
		return h, s
	}
	has := func(s packet.Sized) bool { return strings.Contains(s.Brief, movePlanted) }
	if _, s := build(ArmFull); !s.WholeHistory || !has(s) || strings.Contains(s.Brief, "## Catalog") {
		t.Errorf("full: whole %v", s.WholeHistory)
	}
	if _, s := build(ArmRecent); has(s) || s.WholeHistory || strings.Contains(s.Brief, "chosen for this message") || !strings.Contains(s.Brief, "## Catalog") {
		t.Errorf("recent should rank nothing")
	}
	if h, s := build(ArmBM25); !has(s) || !strings.Contains(s.Brief, "excerpt: paragraph") || packet.RankLanes.Value || h.e.Cfg.WholeHistoryTokens != -1 {
		t.Errorf("bm25: lanes %+v", packet.RankLanes)
	}
	if _, s := build(ArmTurns); !has(s) || strings.Contains(s.Brief, "excerpt: paragraph") {
		t.Errorf("turns should send whole turns")
	}
	if err := ApplyContextArm(&config.Config{}, "nonsense"); err == nil {
		t.Error("an unknown arm should be refused")
	}
	h, _ := build(ArmCurrent)
	turns, _ := h.e.V.Turns()
	next := h.e.nextTurn("").Seq
	if n := packet.TokensToAnswer(turns, h.e.S, q, next, "t000003", "at seven in the morning", true); n <= 0 {
		t.Errorf("tokens to answer %d", n)
	}
	if n := packet.TokensToAnswer(turns, h.e.S, q, next, "t000003", "a sentence nobody wrote", true); n != 0 {
		t.Errorf("an unreachable answer gave %d", n)
	}
}

func TestExactSignTest(t *testing.T) {
	for _, c := range []struct {
		a, b int
		want string
	}{{0, 0, "1.0000"}, {1, 1, "1.0000"}, {0, 5, "0.0625"}, {1, 7, "0.0703"}, {11, 7, "0.4807"}} {
		if got := strings.TrimSpace(sprintf4(exactSignP(c.a, c.b))); got != c.want {
			t.Errorf("exactSignP(%d,%d) = %s, want %s", c.a, c.b, got, c.want)
		}
	}
}

func TestGapReport(t *testing.T) {
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	r := GapReport([][]time.Time{{at, at.Add(4 * time.Minute), at.Add(24 * time.Minute), at.Add(3 * time.Hour)}, {at.Add(100 * time.Hour)}})
	for _, want := range []string{"3 gaps between user messages in 2 session(s); median 20m0s", "within 5m0s: 1 (33%)", "within 30m0s: 2 (66%)", "within 1h0m0s: 2 (66%)", "within 12h0m0s: 3 (100%)",
		// gaps of 4, 20 and 156 minutes: a 5-minute cache is read once and written twice, a 12-hour one is read three times
		"  0.87  cache kept 5 minutes, write 1.25x", "  0.48  cache kept 30 minutes, write 1.25x", "  0.73  cache kept 1 hour, write 2x", "  0.10  cache kept 12 hours, write 2.5x",
		// four pings carry the 20-minute gap: (0.1) + (0.4 + 0.1) + (0.4 + 1.25) over three messages
		"keep-alive pings on the 5m0s cache for up to 20m0s of silence bring it to 0.75"} {
		if !strings.Contains(r, want) {
			t.Errorf("report lacks %q:\n%s", want, r)
		}
	}
	// a fast pace with one short pause: a few pings are cheaper than writing the prefix again
	var fast []time.Time
	for i := 0; i < 20; i++ {
		fast = append(fast, at.Add(time.Duration(i)*6*time.Minute))
	}
	if r := GapReport([][]time.Time{fast}); !strings.Contains(r, "keep-alive pings on the 5m0s cache for up to 10m0s of silence bring it to 0.20") {
		t.Errorf("a steady six-minute pace should favour keep-alives:\n%s", r)
	}
	// long pauses only: every ping is wasted
	slow := []time.Time{at, at.Add(3 * time.Hour), at.Add(6 * time.Hour), at.Add(9 * time.Hour)}
	if r := GapReport([][]time.Time{slow}); !strings.Contains(r, "keep-alive pings on the 5m0s cache do not pay at this pace") {
		t.Errorf("three-hour pauses should not favour keep-alives:\n%s", r)
	}
	if !strings.Contains(GapReport(nil), "no message times") {
		t.Error("no sessions should say so")
	}
}

func sprintf4(f float64) string { return fmt.Sprintf("%.4f", f) }
