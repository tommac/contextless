package engine

import (
	"fmt"
	"testing"

	"github.com/tommac/contextless/internal/vault"
)

// oldWorkspace builds 12 turns, a card on turn 9, a full index, and then
// drops the atom tables: the state the code before atoms left behind.
func oldWorkspace(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t, 200_000, nil)
	for i := 1; i <= 12; i++ {
		body := fmt.Sprintf("Filler turn %d about the compost and the garden beds, long enough to be a paragraph of its own in the index.", i)
		if i == 9 {
			body = "Let's go with the walnut ones, not the white oak."
		}
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
			t.Fatal(err)
		}
	}
	tu, _ := h.e.V.Turn("t000009")
	if err := h.e.V.PutCard(vault.Card{Turn: tu.ID, BodySHA: tu.BodySHA, Summary: "The user picks the wood for the kitchen cabinets.", Keywords: []string{"cabinets", "wood"}}); err != nil {
		t.Fatal(err)
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	if _, err := h.e.S.DB.Exec(`DROP TABLE atoms_fts; DROP TABLE atom_turns; DROP TABLE meta;`); err != nil {
		t.Fatal(err)
	}
	return h
}

// A workspace the old code left has no atom tables. A read-only open,
// while a writer holds the workspace, builds the atom index with the
// cards; a change of the splitting rules rebuilds it.
func TestOldWorkspaceGetsItsAtomIndexOnReadOnlyOpen(t *testing.T) {
	h := oldWorkspace(t)
	ro, err := OpenReadOnly(h.e.Cfg, "test")
	if err != nil {
		t.Fatalf("read-only open of an old workspace: %v", err)
	}
	if n := len(ro.S.AtomTurns()); n != 12 {
		t.Errorf("read-only open indexed the atoms of %d of 12 turns", n)
	}
	if hits := ro.S.SearchAtoms("which wood for the cabinets", 3, 0); len(hits) == 0 || hits[0].Turn != "t000009" {
		t.Errorf("after the upgrade the card's words do not reach the turn's atom: %+v", hits)
	}
	ro.Close()

	if err := h.e.S.SetMeta("atom_seg", "0"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.e.S.DB.Exec(`INSERT INTO atoms_fts(id,turn,seq,text,card) VALUES('t000009#a7','t000009',9,'a stale walnut atom from older rules','')`); err != nil {
		t.Fatal(err)
	}
	ro, err = OpenReadOnly(h.e.Cfg, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	if n := len(ro.S.SearchAtoms("walnut", 10, 0)); n != 1 {
		t.Errorf("after a change of the splitting rules %d atoms hold walnut, want 1", n)
	}
	if n := len(ro.S.AtomTurns()); n != 12 {
		t.Errorf("after a change of the splitting rules %d of 12 turns are indexed", n)
	}
}

// The same upgrade through the writing open, the one cx chat and cx say use.
func TestOldWorkspaceGetsItsAtomIndexOnWritingOpen(t *testing.T) {
	h := oldWorkspace(t)
	cfg := h.e.Cfg
	h.e.Close()
	w, err := Open(cfg, "test")
	if err != nil {
		t.Fatalf("writing open of an old workspace: %v", err)
	}
	defer w.Close()
	if n := len(w.S.AtomTurns()); n != 12 {
		t.Errorf("writing open indexed the atoms of %d of 12 turns", n)
	}
	if hits := w.S.SearchAtoms("which wood for the cabinets", 3, 0); len(hits) == 0 || hits[0].Turn != "t000009" {
		t.Errorf("after the upgrade the card's words do not reach the turn's atom: %+v", hits)
	}
}
