package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// longTurn is a turn of n paragraphs about one topic; paragraph at holds
// the planted sentence.
func longTurn(topic string, n, at int, planted string) string {
	var b strings.Builder
	for p := 1; p <= n; p++ {
		fmt.Fprintf(&b, "Paragraph %d of the %s notes goes over routine matters that came up during the week, "+
			"who followed up with whom, and which items stay open until the next review of the %s work. ", p, topic, topic)
		if p == at {
			b.WriteString(planted + " ")
		}
		b.WriteString("Nothing else in this paragraph needs to be remembered later.\n\n")
	}
	return b.String()
}

// The ranked section of a packet is made of excerpts: the paragraphs that
// match the message, verbatim, each headed by its turn, its place in the
// turn and the message it answers. The rest of the turn stays on disk,
// named in the catalog, and the paragraphs around an excerpt come back by
// id.
func TestRankedSectionIsExcerpts(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	planted := "The warranty on the heat pump runs until March 2031 and the installer is Calloway Mechanical."
	for i := 1; i <= 40; i++ {
		role, reply, body := vault.RoleUser, "", fmt.Sprintf("What is next on the %s list for week %d?", []string{"garden", "garage", "attic"}[i%3], i)
		if i%2 == 0 {
			role, reply = vault.RoleAssistant, vault.TurnID(i-1)
			body = longTurn([]string{"garden", "garage", "attic"}[i%3], 9, 0, "")
		}
		if i == 12 {
			body = longTurn("house", 9, 5, planted)
		}
		if _, err := h.e.V.AppendTurn(role, "", "", reply, body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	src, _ := h.e.V.Turn("t000012")
	atoms := vault.Atoms(src.ID, src.Body)
	k := 0
	for _, a := range atoms {
		if strings.Contains(a.Text(src.Body), planted) {
			k = a.K
		}
	}
	if k < 2 || k >= len(atoms) {
		t.Fatalf("the planted paragraph should sit inside the turn, got atom %d of %d", k, len(atoms))
	}
	first := atoms[0].Text(src.Body)
	before := atoms[k-2].Text(src.Body)
	round := 0
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		round++
		switch round {
		case 1:
			head := fmt.Sprintf("t000012 · assistant · %s · excerpt: paragraph %d of %d (t000012#a%d) · in reply to t000011: \"What is next on",
				src.Created.Format("2006-01-02 15:04"), k, len(atoms), k)
			if !strings.Contains(c.Brief, head) {
				t.Errorf("round 1 lacks the excerpt header %q", head)
			}
			if !strings.Contains(c.Brief, planted) {
				t.Errorf("round 1 lacks the planted paragraph")
			}
			if strings.Contains(c.Brief, first) {
				t.Errorf("round 1 holds the first paragraph of the turn, which does not match the question")
			}
			if !strings.Contains(c.Brief, "- t000012 assistant (not yet indexed, opens with)") {
				t.Errorf("the catalog should still name the turn an excerpt comes from")
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall,
				RecallIDs: []string{vault.AtomRangeID("t000012", k-1, k+1), "t000012#a99"}}
		case 2:
			want := fmt.Sprintf("excerpt: paragraphs %d to %d of %d (t000012#a%d-%d)", k-1, k+1, len(atoms), k-1, k+1)
			if !strings.Contains(c.Brief, want) || !strings.Contains(c.Brief, before) {
				t.Errorf("round 2 lacks the paragraphs around the excerpt (%s)", want)
			}
			if strings.Contains(c.Brief, fmt.Sprintf("excerpt: paragraph %d of %d (t000012#a%d)", k, len(atoms), k)) {
				t.Errorf("round 2 repeats the single paragraph beside the run that holds it")
			}
			note := fmt.Sprintf("t000012 has %d paragraph(s), so t000012#a99 does not exist", len(atoms))
			if !strings.Contains(c.Brief, note) {
				t.Errorf("round 2 lacks the note for the bad id: %q", note)
			}
			return done(c, "March 2031, Calloway Mechanical.")
		}
		return fault{"unexpected round"}
	}
	if _, err := h.e.Ask(context.Background(), "When does the heat pump warranty run out, and who was the installer?"); err != nil {
		t.Fatal(err)
	}
	if round != 2 {
		t.Fatalf("round %d", round)
	}
}

// A history short enough is sent whole and in order, with no catalog and
// nothing ranked. A history that would only fit a larger provider's
// window is not: old turns never move a call off the first profile.
func TestShortHistoryIsSentWhole(t *testing.T) {
	build := func(alphaWindow int) (*harness, []string) {
		h := newHarness(t, alphaWindow, nil)
		h.e.Cfg.WholeHistoryTokens = 60_000
		var bodies []string
		for i := 1; i <= 30; i++ {
			body := fmt.Sprintf("Note %d. ", i) + strings.Repeat(fmt.Sprintf("Line about item %d of the inventory. ", i), 160)
			bodies = append(bodies, strings.TrimSpace(body))
			if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
				t.Fatal(err)
			}
		}
		if err := h.e.Reindex(); err != nil {
			t.Fatal(err)
		}
		return h, bodies
	}

	h, bodies := build(200_000)
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		for i, b := range bodies {
			if !strings.Contains(c.Brief, b) {
				t.Errorf("turn %d is not in the packet whole", i+1)
			}
		}
		if strings.Contains(c.Brief, "## Catalog") || strings.Contains(c.Brief, "excerpt:") {
			t.Errorf("a whole history needs no catalog and no excerpts")
		}
		return done(c, "ok")
	}
	if _, err := h.e.Ask(context.Background(), "Which item was on line 7?"); err != nil {
		t.Fatal(err)
	}

	// the same history (about 54k tokens) against a first profile that
	// holds a selected packet but not all of it
	h, _ = build(90_000)
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		if c.Provider != "alpha" {
			t.Errorf("the call moved to %s to fit old turns", c.Provider)
		}
		if !strings.Contains(c.Brief, "## Catalog") {
			t.Errorf("a history that does not fit the first profile is selected, with a catalog")
		}
		return done(c, "ok")
	}
	if _, err := h.e.Ask(context.Background(), "Which item was on line 7?"); err != nil {
		t.Fatal(err)
	}
}

// The atom index is stemmed and searches each atom with its turn's card
// text; a card that arrives later re-indexes the turn's atoms.
func TestAtomIndexStemsAndFollowsTheCard(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	for i := 1; i <= 12; i++ {
		body := fmt.Sprintf("Filler turn %d about the compost and the garden beds, long enough to be a paragraph of its own in the index.", i)
		if i == 4 {
			body = "The committee rejected both proposals for the east fence after the survey came back."
		}
		if i == 9 {
			body = "Let's go with the walnut ones, not the white oak."
		}
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	top := func(q string) string {
		hits := h.e.S.SearchAtoms(q, 3, 0)
		if len(hits) == 0 {
			return ""
		}
		return hits[0].Turn
	}
	if got := top("rejections of the proposal"); got != "t000004" {
		t.Errorf("stemmed search for a different word form found %q, want t000004", got)
	}
	if got := top("which wood for the cabinets"); got == "t000009" {
		t.Errorf("before its card, the turn shares no word with the question")
	}
	tu, _ := h.e.V.Turn("t000009")
	card := vault.Card{Turn: tu.ID, BodySHA: tu.BodySHA, Summary: "The user picks the wood for the kitchen cabinets.", Keywords: []string{"cabinets", "wood"}}
	if err := h.e.V.PutCard(card); err != nil {
		t.Fatal(err)
	}
	h.e.indexTurn(tu, &card)
	if got := top("which wood for the cabinets"); got != "t000009" {
		t.Errorf("with its card, the turn's atom should be found by the card's words, got %q", got)
	}
	if n := len(h.e.S.SearchAtoms("walnut", 10, 0)); n != 1 {
		t.Errorf("re-indexing a turn should replace its atoms, found %d", n)
	}
	if hits := h.e.S.SearchAtoms("walnut", 10, 9); len(hits) != 0 {
		t.Errorf("a search before turn 9 should not see turn 9")
	}
}
