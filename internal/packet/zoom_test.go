package packet

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/vault"
)

func zoomTurns(n int) []vault.Turn {
	var out []vault.Turn
	for i := 1; i <= n; i++ {
		body := para(1, fmt.Sprintf("Entry %d about the orchard fences.", i))
		if i == 17 {
			body = para(1, "") + "\n\n" + para(2, "The heron platform was moved to the east marsh.")
		}
		out = append(out, vault.Turn{ID: vault.TurnID(i), Seq: i, Role: vault.RoleUser, Body: body, BodySHA: vault.SHA(body)})
	}
	return out
}

// Finished blocks of ten climb the ladder; every turn is under exactly one
// root, and a block's line names the turns it covers.
func TestZoomLadderCoversEveryTurnOnce(t *testing.T) {
	z := BuildZoom(zoomTurns(237), true)
	seen := map[int]int{}
	var walk func(n *ZoomNode)
	walk = func(n *ZoomNode) {
		if n.Level == 1 {
			seen[n.First]++
			return
		}
		if n.Level >= 2 && !strings.HasPrefix(n.Text, vault.TurnID(n.First)+" to "+vault.TurnID(n.Last)) {
			t.Errorf("block line %q does not name its span", n.Text)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	levels := map[int]int{}
	for _, r := range z.Roots {
		levels[r.Level]++
		walk(r)
	}
	for i := 1; i <= 237; i++ {
		if seen[i] != 1 {
			t.Fatalf("turn %d is under %d roots", i, seen[i])
		}
	}
	if levels[3] != 2 || levels[2] != 3 || levels[1] != 7 {
		t.Fatalf("roots by level %v, want 2 hundreds, 3 tens, 7 turns", levels)
	}
}

// Views keep to their budget, and a ladder view zooms toward the question.
func TestZoomViewsKeepToBudget(t *testing.T) {
	z := BuildZoom(zoomTurns(237), true)
	roots := viewTokens(z.Roots)
	for _, budget := range []int{roots, 2 * roots, 4000, 100000} {
		for name, v := range map[string][]*ZoomNode{"catalog": z.CatalogView(budget), "ladder": z.LadderView("heron platform marsh", budget), "static": z.LadderView("", budget)} {
			if got := viewTokens(v); got > budget {
				t.Errorf("%s at %d spends %d", name, budget, got)
			}
		}
	}
	open := false
	for _, n := range z.LadderView("heron platform marsh", 4000) {
		open = open || n.ID == "t000017"
	}
	if !open {
		t.Error("the ladder did not zoom into the block holding the match")
	}
	if all := z.CatalogView(100000); len(all) != 237 {
		t.Errorf("with room, the catalog lists every turn, got %d lines", len(all))
	}
}

// The reader reaches a paragraph through blurred lines, reads it verbatim,
// and never spends more than its budget.
func TestZoomWalkReadsVerbatimWithinBudget(t *testing.T) {
	turns := zoomTurns(237)
	z := BuildZoom(turns, true)
	read := z.ZoomWalk(z.LadderView("", 0), "Where was the heron platform moved?", 6000)
	found := false
	for _, r := range read {
		found = found || strings.Contains(r, "The heron platform was moved to the east marsh.")
	}
	if !found {
		t.Fatalf("the paragraph was not reached; read %d", len(read))
	}
	if got := z.ZoomWalk(z.LadderView("", 0), "Where was the heron platform moved?", 50); len(got) != 0 {
		t.Errorf("a budget too small for any paragraph read %d", len(got))
	}
	if got := SearchRead(turns, nil, "the heron platform was moved", 0, 400); len(got) == 0 || !strings.Contains(got[0], "heron platform") {
		t.Errorf("search read %v", got)
	}
}

// repeatTurns is a conversation whose paragraphs share a question's words in
// many combinations, several paragraphs to a combination, the words differing
// in rarity: equal scores are common, and a sum added in another order comes
// out different in its last bit.
func repeatTurns(n int) ([]vault.Turn, string) {
	words := []string{"harbour", "lantern", "granite", "orchard", "winter", "copper", "meadow", "ferry", "saddle", "thistle", "quarry", "willow"}
	var out []vault.Turn
	for i := 1; i <= n; i++ {
		var w []string
		for k, x := range words {
			// each word in its own share of the turns, so that the words differ in rarity
			if i%(k+2) == 0 || (i+k)%7 == 0 {
				w = append(w, x)
			}
		}
		body := fmt.Sprintf("Entry %d of the log mentions the %s and little else worth keeping in mind.", i, strings.Join(w, ", the "))
		out = append(out, vault.Turn{ID: vault.TurnID(i), Seq: i, Role: vault.RoleUser, Body: body, BodySHA: vault.SHA(body)})
	}
	return out, "What about the " + strings.Join(words, " and the ") + "?"
}

// The same walk gives the same reading, every time. A score summed over a
// map adds in a new order on each call and differs in its last bit, so equal
// matches compared unequal and the order changed between identical calls.
func TestZoomWalkIsRepeatable(t *testing.T) {
	turns, question := repeatTurns(90)
	z := BuildZoom(turns, true)
	first := ""
	for i := 0; i < 60; i++ {
		got := strings.Join(z.ZoomWalk(z.LadderView(question, 600), question, 1000), "\n")
		if i == 0 {
			first = got
			if got == "" {
				t.Fatal("the fixture is wrong: the walk read nothing")
			}
		} else if got != first {
			t.Fatalf("call %d read differently from the first call", i+1)
		}
	}
}

// Notes that tie on every score keep one order, and so do the notes a topic
// brings in.
func TestNoteLanesAreRepeatable(t *testing.T) {
	var notes []Note
	for i := 0; i < 40; i++ {
		topic := []string{"harbour works", "ferry timetable", "orchard fences"}[i%3]
		notes = append(notes, Note{Title: "The harbour lantern and the ferry were discussed again", Kind: "fact", Topics: []string{topic, "winter plans"},
			Atoms: []string{vault.AtomID(vault.TurnID(1+i/4), 1)}, Seq: 1 + i/4})
	}
	ni := NewNoteIndex(notes)
	const q = "When were the harbour lantern and the ferry discussed, and what about winter?"
	own, byTopic := ni.lanes(q, true)
	if len(own) != 40 || len(byTopic) != 40 {
		t.Fatalf("the fixture is wrong: %d own, %d by topic", len(own), len(byTopic))
	}
	for i := 0; i < 60; i++ {
		o, b := ni.lanes(q, true)
		if fmt.Sprint(o) != fmt.Sprint(own) || fmt.Sprint(b) != fmt.Sprint(byTopic) {
			t.Fatalf("call %d ranked the notes differently", i+2)
		}
		if shown, _ := NotesView(ni, q, 300, 1); fmt.Sprint(shown) != fmt.Sprint(firstView(ni, q)) {
			t.Fatalf("call %d showed different lines", i+2)
		}
	}
}

func firstView(ni *NoteIndex, q string) []int {
	shown, _ := NotesView(ni, q, 300, 1)
	return shown
}

// A paragraph whose exact text was already taken is passed over, as live
// packets do: the second copy would cost tokens and add nothing.
func TestPickInOrderSkipsRepeatedParagraphs(t *testing.T) {
	long := "The heron platform was moved to the east marsh after the spring floods, and the old posts were left standing where they were."
	var order []RankedAtom
	for i, body := range []string{long, "Short one.", long, "Short one.", "Another paragraph of the record, long enough to count as text of its own in this fixture, about the orchard."} {
		tn := vault.Turn{ID: vault.TurnID(i + 1), Seq: i + 1, Role: vault.RoleUser, Body: body, BodySHA: vault.SHA(body)}
		order = append(order, RankedAtom{Turn: tn, Atom: atomsOf(tn)[0]})
	}
	var ids []string
	for _, r := range pickInOrder(order, 100000) {
		ids = append(ids, r.Turn.ID)
	}
	// the long paragraph once; the short one twice (too short to be worth the rule); the rest
	if got := strings.Join(ids, " "); got != "t000001 t000002 t000004 t000005" {
		t.Fatalf("picked %s", got)
	}
}

// The walk pays for the lines it opens and the paragraphs it reads, reports
// what it spent, and a budget one token short of a block's lines opens nothing.
func TestZoomWalkChargesWhatItOpens(t *testing.T) {
	z := BuildZoom(zoomTurns(37), true)
	const q = "heron platform marsh"
	var block *ZoomNode
	for _, r := range z.Roots {
		if r.Level == 2 && r.First <= 17 && 17 <= r.Last {
			block = r
		}
	}
	if block == nil {
		t.Fatal("the fixture is wrong: no ten-block holds turn 17")
	}
	lines := ViewTokens(block.Children)
	if read, spent := z.ZoomWalkSpent([]*ZoomNode{block}, q, lines-1); len(read) != 0 || spent != 0 {
		t.Fatalf("one token short of the lines: read %d, spent %d", len(read), spent)
	}
	read, spent := z.ZoomWalkSpent([]*ZoomNode{block}, q, 100000)
	want := 0
	for _, n := range z.Paragraphs() {
		if slices.Contains(read, n.Text) {
			want += n.Tokens
		}
	}
	if len(read) == 0 || spent != lines+want || !strings.Contains(strings.Join(read, "\n"), "heron platform") {
		t.Fatalf("read %d paragraphs, spent %d, want %d lines + %d paragraphs", len(read), spent, lines, want)
	}
	for _, r := range read {
		if !strings.Contains(r, "heron") && !strings.Contains(r, "platform") && !strings.Contains(r, "marsh") {
			t.Fatalf("read a paragraph with none of the question's words: %q", r)
		}
	}
}
