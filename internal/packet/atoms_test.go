package packet

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// para is a paragraph long enough to be an atom of its own.
func para(i int, extra string) string {
	return fmt.Sprintf("Paragraph %d covers ordinary notes about the week, written out at enough length that it stands as one atom of its turn and is not joined to the paragraph after it, with a few more words here to be quite sure that it passes the minimum size. %s", i, extra)
}

func atomTurn(seq int, paras ...string) vault.Turn {
	body := strings.Join(paras, "\n\n")
	return vault.Turn{ID: vault.TurnID(seq), Seq: seq, Role: vault.RoleUser, Body: body, BodySHA: vault.SHA(body)}
}

// PickAtoms fills its budget best first, joins adjacent atoms of a turn
// into one run, names a turn picked whole by its turn id, and leaves out
// what the packet already has.
func TestPickAtomsJoinsRunsAndKeepsToBudget(t *testing.T) {
	const phrase = "the blue heron survey"
	hit := "We agreed on " + phrase + " for the marsh."
	turns := []vault.Turn{
		atomTurn(1, para(1, ""), para(2, hit), para(3, hit), para(4, ""), para(5, ""), para(6, "")),
		atomTurn(2, para(1, hit)),
		atomTurn(3, para(1, ""), para(2, ""), para(3, ""), para(4, hit)),
	}
	for _, tu := range turns {
		if n := len(vault.Atoms(tu.ID, tu.Body)); n != strings.Count(tu.Body, "\n\n")+1 {
			t.Fatalf("%s: %d atoms, want one per paragraph", tu.ID, n)
		}
	}
	ids := func(runs []AtomRun) string {
		var out []string
		for _, r := range runs {
			out = append(out, r.ID())
		}
		return strings.Join(out, " ")
	}
	runs, _ := PickAtoms(turns, nil, phrase, 0, 10_000, nil)
	if got := ids(runs); got != "t000001#a2-3 t000002 t000003#a4" {
		t.Fatalf("runs = %s", got)
	}
	// a budget for one atom takes the newest statement of the phrase
	runs, _ = PickAtoms(turns, nil, phrase, 0, 130, nil)
	if got := ids(runs); got != "t000003#a4" {
		t.Fatalf("with a small budget, runs = %s", got)
	}
	// nothing at or after beforeSeq
	runs, _ = PickAtoms(turns, nil, phrase, 3, 10_000, nil)
	if got := ids(runs); got != "t000001#a2-3 t000002" {
		t.Fatalf("before t000003, runs = %s", got)
	}
	// what the packet already has is not picked again, and is reported
	have := func(turn string, k int) bool { return turn == "t000002" || (turn == "t000001" && k == 2) }
	runs, seen := PickAtoms(turns, nil, phrase, 0, 10_000, have)
	if got := ids(runs); got != "t000001#a3 t000003#a4" {
		t.Fatalf("with turn 2 and atom 2 of turn 1 already in, runs = %s", got)
	}
	if seen["t000002"] <= 0 {
		t.Fatalf("a turn already in whole should be reported as hit: %v", seen)
	}
}

// sentTokens is what a set of runs costs in a brief: each run's text and
// its header, at the default density.
func sentTokens(runs []AtomRun) int {
	n := 0
	for _, r := range runs {
		atoms := vault.Atoms(r.Turn.ID, r.Turn.Body)
		text := vault.SpanText(r.Turn.Body[atoms[r.A-1].Start:atoms[r.B-1].End])
		n += Tokens(providers.Profile{}, len(text)+len(AtomHeader(r.Turn, r.A, r.B, r.N, nil))+8)
	}
	return n
}

// The budget bounds what is sent, headers included: every run carries its
// own header, so many separate picks cost more than one run of the same
// paragraphs, and an atom that does not fit is skipped for a smaller one.
func TestPickAtomsBudgetCountsHeadersAndIsNotExceeded(t *testing.T) {
	const phrase = "the grey seal count"
	hit := "We logged " + phrase + " at the point."
	var turns []vault.Turn
	for i := 1; i <= 30; i++ {
		// the match is the middle paragraph, so every pick is its own run
		turns = append(turns, atomTurn(i, para(1, ""), para(2, hit+fmt.Sprintf(" Entry %d.", i)), para(3, "")))
	}
	for _, budget := range []int{150, 400, 1000, 2500} {
		runs, _ := PickAtoms(turns, nil, phrase, 0, budget, nil)
		if len(runs) == 0 {
			t.Fatalf("budget %d picked nothing", budget)
		}
		if got := sentTokens(runs); got > budget {
			t.Errorf("budget %d: the runs picked cost %d tokens as sent", budget, got)
		}
		one := sentTokens(runs[:1])
		// the charge is a few tokens above the cost, never below it
		if (len(runs)+2)*one <= budget {
			t.Errorf("budget %d: %d runs of %d tokens leave room for two more", budget, len(runs), one)
		}
	}
	// a long atom that does not fit is passed over for a short one
	long := atomTurn(40, strings.Repeat("A sentence about "+phrase+" that goes on. ", 40))
	short := atomTurn(41, para(1, hit))
	runs, _ := PickAtoms([]vault.Turn{short, long}, nil, phrase, 0, 200, nil)
	if len(runs) != 1 || runs[0].Turn.ID != "t000041" {
		t.Fatalf("want the short turn only, got %+v", runs)
	}
}

// The same paragraph pasted into several turns is sent once.
func TestPickAtomsSkipsRepeatedParagraphs(t *testing.T) {
	notice := "Reminder: the osprey nest platform inspection is due before the first of the month, and whoever goes should bring the long ladder and the spare bolts from the shed."
	var turns []vault.Turn
	for i := 1; i <= 9; i++ {
		turns = append(turns, atomTurn(i, para(1, fmt.Sprintf("Week %d.", i)), notice, para(3, "")))
	}
	runs, _ := PickAtoms(turns, nil, "osprey nest platform inspection", 0, 10_000, nil)
	n := 0
	for _, r := range runs {
		atoms := vault.Atoms(r.Turn.ID, r.Turn.Body)
		n += strings.Count(r.Turn.Body[atoms[r.A-1].Start:atoms[r.B-1].End], notice)
	}
	if n != 1 {
		t.Fatalf("the repeated paragraph was picked %d times, want once (the newest)", n)
	}
}

// The exact lane matches the message as a phrase at word boundaries, and
// a message with no content word draws nothing.
func TestExactLaneNeedsAWordBoundaryAndAContentWord(t *testing.T) {
	if !containsPhrase("we saw the heron today", "heron") || containsPhrase("the herons left", "heron") || containsPhrase("there", "the") {
		t.Fatal("containsPhrase is wrong about word boundaries")
	}
	turns := []vault.Turn{atomTurn(1, para(1, "Then there was the other theme of the weather."))}
	for _, msg := range []string{"the", "yes", "and"} {
		if runs, _ := PickAtoms(turns, nil, msg, 0, 10_000, nil); len(runs) != 0 {
			t.Errorf("message %q drew %d excerpts", msg, len(runs))
		}
	}
}

// Only the value lane's best AtomValueTop atoms are fused, and a negative
// setting is the lane switched off, not a crash.
func TestRankAtomsFusesOnlyTheBestOfTheValueLane(t *testing.T) {
	var turns []vault.Turn
	for i := 1; i <= 8; i++ {
		turns = append(turns, atomTurn(i, para(1, fmt.Sprintf("The budget for the dock repair is $%d,400 this year.", i))))
	}
	defer func(old int) { AtomValueTop = old }(AtomValueTop)
	for _, c := range []struct{ top, want int }{{2, 2}, {5, 5}, {0, 0}, {-1, 0}} {
		AtomValueTop = c.top
		if got := len(RankAtoms(turns, nil, "What is the dock repair budget?", 0, 50)); got != c.want {
			t.Errorf("AtomValueTop %d: %d atoms ranked, want %d", c.top, got, c.want)
		}
	}
}

func TestJoinAtoms(t *testing.T) {
	got := fmt.Sprint(JoinAtoms(map[int]bool{7: true, 2: true, 3: true, 9: true, 8: true, 4: true}))
	if got != "[[2 4] [7 9]]" {
		t.Fatalf("JoinAtoms = %s", got)
	}
}
