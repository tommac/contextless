package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// A long turn is recalled one part at a time: a search lands on the part
// that holds the phrase, and the catalog line says how many parts exist.
func TestRecallOfOnePartOfALongTurnWithAtomsOff(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.e.Cfg.RelevantTokens = -1 // relevant_tokens: -1, whole turns as the ranked unit
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
		if !strings.Contains(c.Brief, fmt.Sprintf("part %d of %d (%s)", k, len(parts), partID)) || !strings.Contains(c.Brief, parts[k-1]) {
			t.Errorf("round 2 lacks part %d verbatim", k)
		}
		other := 1
		if k == 1 {
			other = 2
		}
		if strings.Contains(c.Brief, parts[other-1]) {
			t.Errorf("round 2 holds part %d, which nobody asked for", other)
		}
		return done(c, "ESC-88-1907.")
	}
	if _, err := h.e.Ask(context.Background(), "Can you find it for me?"); err != nil {
		t.Fatal(err)
	}
	if round != 2 {
		t.Fatalf("round %d", round)
	}
}

// With relevant_tokens off, the ranked section holds whole turns, as
// before atoms: the turn that matches the message is in the first packet
// verbatim, under its turn header, not as an excerpt.
func TestRankedSectionIsWholeTurnsWithAtomsOff(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.e.Cfg.RelevantTokens = -1
	planted := longTurn("house", 9, 5, "The warranty on the heat pump runs until March 2031 and the installer is Calloway Mechanical.")
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
		if !strings.Contains(c.Brief, planted) {
			t.Errorf("the ranked turn is not in the first packet whole")
		}
		if strings.Contains(c.Brief, "excerpt: paragraph") {
			t.Errorf("an excerpt was sent with relevant_tokens off")
		}
		return done(c, "March 2031.")
	}
	if _, err := h.e.Ask(context.Background(), "Until when does the warranty on the heat pump run, and who is the installer?"); err != nil {
		t.Fatal(err)
	}
	if round != 1 {
		t.Fatalf("round %d", round)
	}
}
