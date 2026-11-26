package packet

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/vault"
)

func mapNote(id, title, topic string, turn, a, b int) vault.Note {
	return vault.Note{ID: id, Title: title, Kind: "fact", Topics: []string{topic}, Seq: turn, Last: turn,
		Evidence: []vault.NoteEvidence{{Turn: vault.TurnID(turn), BodySHA: "sha-" + vault.TurnID(turn), A: a, B: b}}}
}

func TestMapLineNamesParagraphRanges(t *testing.T) {
	n := vault.Note{Title: "  The port\n moved to 8443 ", Kind: "Change", Evidence: []vault.NoteEvidence{
		{Turn: "t000040", A: 1, B: 2}, {Turn: "t000012", A: 3, B: 3}, {Turn: "t000012", A: 3, B: 3}}}
	if got, want := MapLine(n), "[t000012#a3, t000040#a1-2] (change) The port moved to 8443"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got := MapLine(vault.Note{Title: "x", Evidence: []vault.NoteEvidence{{Turn: "t000001", A: 1, B: 1}}}); got != "[t000001#a1] (note) x" {
		t.Fatalf("a note with no kind: %q", got)
	}
}

// While every line fits, the map is the same for every message and only
// grows at its end: the property a prompt cache needs.
func TestMapIsStableAndAppendOnlyWhileItFits(t *testing.T) {
	var notes []vault.Note
	for i := 1; i <= 30; i++ {
		notes = append(notes, mapNote(fmt.Sprintf("no-%03d", i), fmt.Sprintf("Fact number %d about the harbour", i), "harbour", i, 1, 1))
	}
	render := func(items []Item) string {
		return Render(Task{RunID: "r", Kind: KindRespond, Role: "responder"}, items, nil)
	}
	section := func(brief string) string {
		i := strings.Index(brief, "## Map of the conversation")
		j := strings.Index(brief[i:], "\n## Task")
		return brief[i : i+j]
	}
	a, _ := BuildMap(notes, nil, 31, "what about the harbour wall", 40000, 8000)
	b, _ := BuildMap(notes, nil, 31, "something else entirely, number 7", 40000, 8000)
	if section(render(a)) != section(render(b)) {
		t.Fatal("the map changed with the message although every line fits")
	}
	more := append(append([]vault.Note{}, notes...), mapNote("no-031", "Fact number 31 about the harbour", "harbour", 31, 1, 1))
	c, _ := BuildMap(more, nil, 32, "q", 40000, 8000)
	if before, after := section(render(a)), section(render(c)); !strings.HasPrefix(after, before) {
		t.Fatalf("a new note did not just extend the map:\n%s\n---\n%s", before, after)
	}
}

func TestMapLeavesOutStaleAndLaterNotes(t *testing.T) {
	notes := []vault.Note{
		mapNote("no-a", "Early fact", "x", 3, 1, 1),
		mapNote("no-b", "Fact whose turn changed", "x", 4, 1, 1),
		mapNote("no-c", "Later fact", "x", 9, 1, 1),
	}
	current := func(turn, sha string) bool { return turn != "t000004" && sha == "sha-"+turn }
	items, st := BuildMap(notes, current, 9, "fact", 40000, 0)
	if st.Notes != 1 || st.Shown != 1 || st.Stale != 1 || len(items) != 2 {
		t.Fatalf("stats %+v, %d items", st, len(items))
	}
	for _, it := range items {
		if strings.Contains(it.Text, "changed") || strings.Contains(it.Text, "Later") {
			t.Errorf("a stale or later note is shown: %q", it.Text)
		}
	}
	// a note with evidence on both sides of the focus keeps the earlier part
	both := vault.Note{ID: "no-d", Title: "Stated twice", Evidence: []vault.NoteEvidence{{Turn: "t000002", A: 1, B: 1}, {Turn: "t000020", A: 2, B: 2}}}
	items, _ = BuildMap([]vault.Note{both}, nil, 9, "", 40000, 0)
	if len(items) != 2 || !strings.Contains(items[0].Text+items[1].Text, "[t000002#a1] (note) Stated twice") {
		t.Fatalf("items %+v", items)
	}
	if items, st := BuildMap(notes, nil, 9, "", 0, 0); items != nil || st.Shown != 0 {
		t.Errorf("a zero budget should send no map")
	}
}

// Past the budget every note is still accounted for: shown as a line, or
// counted under its topic.
func TestMapOverBudgetAccountsForEveryNote(t *testing.T) {
	var notes []vault.Note
	for i := 1; i <= 200; i++ {
		topic := []string{"alpha", "beta", "gamma", "delta"}[i%4]
		notes = append(notes, mapNote(fmt.Sprintf("no-%03d", i), fmt.Sprintf("Reading %d of the %s series was within tolerance", i, topic), topic, i, 1, 1))
	}
	budget := 1500
	items, st := BuildMap(notes, nil, 201, "the gamma series reading 150", budget, 8000)
	if st.Shown == 0 || st.Folded == 0 || st.Shown+st.Folded != 200 || st.Tokens > budget {
		t.Fatalf("stats %+v", st)
	}
	counted, sawBest := 0, false
	for _, it := range items {
		if strings.HasPrefix(it.ID, mapFoldPrefix) {
			var n int
			var topic string
			if _, err := fmt.Sscanf(it.Text, "%q: %d more", &topic, &n); err != nil {
				t.Fatalf("fold line %q: %v", it.Text, err)
			}
			counted += n
		}
		if strings.Contains(it.Text, "Reading 150 of the gamma series") {
			sawBest = true
		}
	}
	if counted != st.Folded {
		t.Errorf("fold lines count %d notes, %d were folded", counted, st.Folded)
	}
	if !sawBest {
		t.Errorf("the best match for the message is not among the lines shown")
	}
	if got := MatchNotes(notes, "Gamma", 500); len(got) != 50 {
		t.Errorf("topic lookup found %d notes", len(got))
	}
	if got := MatchNotes(notes, "reading 150", 3); len(got) != 3 || !strings.Contains(got[0].Title, "Reading 150") {
		t.Errorf("search lookup: %+v", got)
	}
}
