package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

func TestRecallQueriesRespectNonemptyLimit(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	focus := vault.Turn{Seq: 8}
	queries := []string{"quartz", "saffron", "indigo", "malachite", "topaz", "  ", "cobalt"}
	for _, word := range queries {
		if strings.TrimSpace(word) != "" {
			if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", word); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	for _, budget := range []int{0, 8000} {
		h.e.Cfg.RelevantTokens = budget
		add, notes := h.e.recallRequest(nil, queries, focus, nil)
		if !slices.Equal(add, []string{"t000001", "t000002", "t000003", "t000004", "t000005"}) ||
			len(notes) != 6 || !strings.Contains(notes[5], `"cobalt" and later ones did not run`) {
			t.Fatalf("budget %d: added %v, notes %v", budget, add, notes)
		}
	}
}

func TestRecallQueryReportsUnreadableTurnHistory(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	turn, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", "original text")
	if err != nil {
		t.Fatal(err)
	}
	focus := vault.Turn{Seq: 2}
	queries := []string{"", " quartz ", "indigo"}
	_, notes := h.e.recallRequest(nil, queries, focus, nil)
	if len(notes) != 2 || !strings.Contains(notes[0], "no ranked passages") {
		t.Fatalf("ordinary misses: %v", notes)
	}
	path := filepath.Join(h.e.V.Dir, vault.TurnPath(turn.ID))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(data), "original text", "edited text", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	assertFailed := func() {
		t.Helper()
		add, notes := h.e.recallRequest(nil, queries, focus, nil)
		if len(add) != 0 || len(notes) != 1 || !strings.Contains(notes[0], `Search "quartz" and any remaining searches could not run`) {
			t.Fatalf("failed search: %v, %v", add, notes)
		}
	}
	assertFailed()
	dir := filepath.Dir(path)
	if err := os.Rename(dir, dir+"-unavailable"); err != nil {
		t.Fatal(err)
	}
	assertFailed()
	if _, notes := h.e.recallRequest(nil, []string{" ", ""}, focus, nil); len(notes) != 0 {
		t.Fatalf("blank queries: %v", notes)
	}
}

// A model that cannot tell from any catalog line where a detail is can
// search for it (recall_queries); a bad id or part gets a note back, not
// silence.
func TestRecallSearchFindsTurnsNoCardPointsTo(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	planted := "Long notes from the walk around the property with the surveyor, the fence line, the drainage ditch and the old well cover. The gate code for the north shed is 4471, set by Dana last spring."
	for i := 1; i <= 40; i++ {
		body := fmt.Sprintf("Filler turn %d about the garden beds and the compost.", i)
		if i == 5 {
			body = planted
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
		switch round {
		case 1:
			if strings.Contains(c.Brief, planted) {
				t.Errorf("the planted turn should not be in the first packet")
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall,
				RecallIDs: []string{"t000099", "t000005#p2"}, RecallQueries: []string{"north shed", "zebra crossing"}}
		default:
			for _, want := range []string{
				"### t000005 · user", planted,
				`Search "north shed" found t000005`,
				`Search "zebra crossing" returned no ranked passages`,
				"t000099 is not before the message being answered",
				"t000005 has 1 part(s), so t000005#p2 does not exist",
			} {
				if !strings.Contains(c.Brief, want) {
					t.Errorf("round 2 packet lacks %q", want)
				}
			}
			return done(c, "4471.")
		}
	}
	r, err := h.e.Ask(context.Background(), "What was that number again?")
	if err != nil {
		t.Fatal(err)
	}
	if round != 2 || r.Assistant.Body != "4471." {
		t.Fatalf("round %d answer %q", round, r.Assistant.Body)
	}
}

// A long turn is never recalled whole by accident: a search lands on the
// paragraph that holds the phrase and returns it as an excerpt, the
// catalog line says how many parts exist, and one part can be asked for
// by its id.
func TestRecallOfOnePartOfALongTurn(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	var doc strings.Builder
	for s := 1; s <= 60; s++ {
		fmt.Fprintf(&doc, "Section %d.\n\n", s)
		line := fmt.Sprintf("Clause %d of the agreement covers routine obligations of both parties. ", s)
		if s == 41 {
			line += "The escrow reference is ESC-88-1907 and it is held by the title company. "
		}
		doc.WriteString(strings.Repeat(line, 22) + "\n\n")
	}
	long, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", doc.String())
	if err != nil {
		t.Fatal(err)
	}
	parts := vault.Parts(long.Body)
	if len(parts) < 3 || strings.Join(parts, "") != long.Body {
		t.Fatalf("expected a lossless split into several parts, got %d", len(parts))
	}
	k := 0
	for i, p := range parts {
		if strings.Contains(p, "ESC-88-1907") {
			k = i + 1
		}
	}
	for i := 2; i <= 20; i++ {
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", fmt.Sprintf("Short note %d about the move.", i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	partID := vault.PartID(long.ID, k)
	round := 0
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		round++
		if round == 1 {
			want := fmt.Sprintf("in %d parts (%s to %s)", len(parts), vault.PartID(long.ID, 1), vault.PartID(long.ID, len(parts)))
			if !strings.Contains(c.Brief, want) {
				t.Errorf("catalog line of the long turn lacks %q", want)
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall, RecallQueries: []string{"ESC-88-1907"}}
		}
		other := 1
		if k == 1 {
			other = 2
		}
		if round == 2 {
			// the search found the paragraph, not the 24 KB part around it
			if !strings.Contains(c.Brief, "excerpt: paragraph") || !strings.Contains(c.Brief, "The escrow reference is ESC-88-1907") {
				t.Errorf("round 2 lacks the paragraph with the reference as an excerpt")
			}
			if !strings.Contains(c.Brief, "Search \"ESC-88-1907\" found "+long.ID+"#a") {
				t.Errorf("round 2 is not told which paragraph the search found")
			}
			if strings.Contains(c.Brief, parts[k-1]) || strings.Contains(c.Brief, parts[other-1]) {
				t.Errorf("round 2 holds a whole part, which nobody asked for")
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall, RecallIDs: []string{partID}}
		}
		if !strings.Contains(c.Brief, fmt.Sprintf("part %d of %d (%s)", k, len(parts), partID)) || !strings.Contains(c.Brief, parts[k-1]) {
			t.Errorf("round 3 lacks part %d verbatim", k)
		}
		if strings.Contains(c.Brief, parts[other-1]) {
			t.Errorf("round 3 holds part %d, which nobody asked for", other)
		}
		return done(c, "ESC-88-1907.")
	}
	if _, err := h.e.Ask(context.Background(), "Can you find it for me?"); err != nil {
		t.Fatal(err)
	}
	if round != 3 {
		t.Fatalf("round %d", round)
	}
}

// Recalled turns that do not fit with the rest of the packet are refused
// with a note, and the run is asked again. Before, the overflow came back
// as ErrTooBig and the engine split the user's short message into parts.
func TestRecallThatDoesNotFitIsToldNotSplit(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	huge := strings.Repeat("Line of a very large pasted log with request ids and timings.\n", 20_000)
	if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", huge); err != nil {
		t.Fatal(err)
	}
	for i := 2; i <= 12; i++ {
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", fmt.Sprintf("Short turn %d.", i)); err != nil {
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
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall, RecallIDs: []string{"t000001"}}
		}
		if !strings.Contains(c.Brief, "could not all be put back together with the rest of this packet: t000001") {
			t.Errorf("round 2 lacks the overflow note")
		}
		return done(c, "Too large to read whole; asked for parts instead.")
	}
	r, err := h.e.Ask(context.Background(), "What was in the log?")
	if err != nil {
		t.Fatal(err)
	}
	if len(h.callsOf(packet.KindUnit)) > 0 {
		t.Fatalf("the short message was split into parts")
	}
	if round != 2 || !strings.HasPrefix(r.Assistant.Body, "Too large") {
		t.Fatalf("round %d answer %q", round, r.Assistant.Body)
	}
}

// A fact that changes a value supersedes the old fact: the packet shows
// the new value as current, carries the old one as what it replaced, and
// dates both by their source turns.
func TestFactThatChangesAValueSupersedesTheOldOne(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "", "", "", "The standup is at 9:15 every weekday.")
	for i := 0; i < 5; i++ {
		_, _ = h.e.V.AppendTurn(vault.RoleUser, "", "", "", fmt.Sprintf("Unrelated note %d.", i))
	}
	t3, _ := h.e.V.AppendTurn(vault.RoleUser, "", "", "", "Move the standup to 10:30 from now on.")
	h.e.applyIndex([]vault.Turn{t1}, packet.IndexHandoff{Status: "completed", Facts: []packet.FactOp{
		{Statement: "Standup is at 9:15 on weekdays.", Source: t1.ID, Quote: "The standup is at 9:15"}}}, "run-a")
	facts := h.e.V.Facts()
	if len(facts) != 1 {
		t.Fatalf("want one fact, got %d", len(facts))
	}
	old := facts[0].ID
	var listed bool
	for _, it := range h.e.ledgerListings([]vault.Turn{t3}) {
		listed = listed || strings.Contains(it.Text, old)
	}
	if !listed {
		t.Fatalf("the indexer of %s is not shown fact %s to supersede", t3.ID, old)
	}
	h.e.applyIndex([]vault.Turn{t3}, packet.IndexHandoff{Status: "completed", Facts: []packet.FactOp{
		{Statement: "Standup is at 10:30.", Source: t3.ID, Quote: "Move the standup to 10:30", Supersedes: []string{old}}}}, "run-b")
	for _, f := range h.e.V.Facts() {
		if f.ID == old && f.Status != "superseded" {
			t.Fatalf("old fact is %q", f.Status)
		}
	}
	sized, err := h.e.DryPacket("What time is the standup?")
	if err != nil {
		t.Fatal(err)
	}
	day := t3.Created.Format("2006-01-02")
	for _, want := range []string{
		"Standup is at 10:30.",
		fmt.Sprintf(`(source %s · %s: "Move the standup to 10:30")`, t3.ID, day),
		fmt.Sprintf("Replaces %s: Standup is at 9:15 on weekdays.", old),
	} {
		if !strings.Contains(sized.Brief, want) {
			t.Errorf("packet lacks %q", want)
		}
	}
	if strings.Contains(sized.Brief, "### "+old) {
		t.Errorf("the superseded fact is still presented as its own record")
	}
}
