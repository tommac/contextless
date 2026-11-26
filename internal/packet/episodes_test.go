package packet

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

func TestRouteEpisodeTurnsFindsCrossTurnSessionContext(t *testing.T) {
	base := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	turns := []vault.Turn{
		// Episode 1 (09:00): discusses "Kyoto pottery trip" and mentions "blue celadon bowl" in next turn.
		{ID: "t000001", Seq: 1, Role: vault.RoleUser, Author: "Alice", Created: base, Body: "Planning our Kyoto pottery trip in autumn."},
		{ID: "t000002", Seq: 2, Role: vault.RoleAssistant, Author: "Bob", Created: base.Add(2 * time.Minute), Body: "We should pick up the blue celadon bowl from the riverside kiln."},
		// Episode 2 (3 hours later): unrelated CI discussion.
		{ID: "t000003", Seq: 3, Role: vault.RoleUser, Author: "Alice", Created: base.Add(3 * time.Hour), Body: "Fixing Go module caching in CI pipelines."},
		{ID: "t000004", Seq: 4, Role: vault.RoleAssistant, Author: "Bob", Created: base.Add(3*time.Hour + 2*time.Minute), Body: "Cache key now hashes go.sum properly."},
	}
	// Query names Episode 1's topic ("Kyoto pottery") + "kiln" so Episode 1 wins routing
	// and surfaces t000002 cleanly.
	routed := RouteEpisodeTurns(turns, nil, nil, "Which kiln for the Kyoto pottery trip?", 5, 2, 6)
	found := false
	for _, id := range routed {
		if id == "t000002" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected episode routing to surface t000002, got %v", routed)
	}
}

func TestAutoReplyWindowPairsQuestionWithAnswer(t *testing.T) {
	v, err := vault.Create(filepath.Join(t.TempDir(), "ws"), "Q&A Test", "Verify auto reply window")
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.OpenFile(filepath.Join(v.Dir, ".cx", "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Reproduces the LoCoMo split-Q&A pattern: t1 asks the question containing the
	// lexical keywords ("running shoes"), while t2 gives the decisive answer without
	// repeating the word "shoes".
	t1, _ := v.AppendTurn(vault.RoleUser, "Alice", "", "", "Hey Bob, are those neon trail shoes for walking or running?")
	t2, _ := v.AppendTurn(vault.RoleAssistant, "Bob", "", t1.ID, "Strictly for sprinting uphill, they keep my knees fresh.")
	// Add 6 unrelated filler turns so t1/t2 fall outside Recent=2.
	for i := 0; i < 6; i++ {
		_, _ = v.AppendTurn(vault.RoleUser, "Alice", "", "", "Unrelated weekly grocery reminder item.")
	}
	turns, _ := v.Turns()
	var at []store.AtomTurn
	for _, tr := range turns {
		var recs []store.AtomRec
		for _, a := range AtomsOf(tr) {
			recs = append(recs, store.AtomRec{ID: a.ID(), Text: a.Text(tr.Body)})
		}
		at = append(at, store.AtomTurn{Turn: tr.ID, Seq: tr.Seq, Atoms: recs})
	}
	_ = s.IndexAtomTurns(at)

	focus := vault.Turn{ID: "t000009", Seq: 9, Role: vault.RoleUser, Body: "What does Bob use the neon trail shoes for?"}

	coptOff := DefaultCompileOptions()
	coptOff.AutoReplyWindow = false
	itemsOff, err := Select(v, s, focus, SelectOptions{Recent: 2, RelevantTokens: 2000, Compile: &coptOff})
	if err != nil {
		t.Fatal(err)
	}
	hasT2Off := false
	for _, it := range itemsOff {
		if it.Key == KeyRelevant && strings.Contains(it.Text, "Strictly for sprinting uphill") {
			hasT2Off = true
		}
	}

	coptOn := DefaultCompileOptions()
	coptOn.AutoReplyWindow = true
	itemsOn, err := Select(v, s, focus, SelectOptions{Recent: 2, RelevantTokens: 2000, Compile: &coptOn})
	if err != nil {
		t.Fatal(err)
	}
	hasT2On := false
	for _, it := range itemsOn {
		if it.Key == KeyRelevant && strings.Contains(it.Text, "Strictly for sprinting uphill") {
			hasT2On = true
		}
	}
	if hasT2Off {
		t.Fatalf("expected t2 absent when AutoReplyWindow=false")
	}
	if !hasT2On {
		t.Fatalf("expected AutoReplyWindow=true to co-select reply turn %s verbatim", t2.ID)
	}
}
