package engine

import (
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
)

func TestSourceRecallAndVerdict(t *testing.T) {
	p := ZoomProbe{Sources: []ZoomSource{{"t1", "the ninth of June"}, {"t2", "Unit  214"}, {"t3", "never said"}}}
	read := []string{"They arrive on the ninth of  June.", "It is unit 214 at the depot."}
	rec, all, pointed, para := sourceRecall(p, read, nil, nil)
	if all || rec < 0.66 || rec > 0.67 || pointed != rec || para != rec {
		t.Fatalf("recall %v, all %v, pointed %v, paragraph %v", rec, all, pointed, para)
	}
	// a shown note counts for a passage that was not read: on the same
	// words when its checked quote overlaps them, on the same paragraph
	// when it is pinned there
	locs := []srcLoc{{}, {}, {turn: "t3", start: 100, end: 140, atoms: map[string]bool{"t3#a2": true}, ok: true}}
	onWords := packet.Note{Quotes: []packet.QuoteSpan{{Turn: "t3", Start: 130, End: 160}}, Atoms: []string{"t3#a2"}}
	onParagraph := packet.Note{Quotes: []packet.QuoteSpan{{Turn: "t3", Start: 300, End: 340}}, Atoms: []string{"t3#a2"}}
	elsewhere := packet.Note{Quotes: []packet.QuoteSpan{{Turn: "t4", Start: 100, End: 140}}, Atoms: []string{"t4#a1"}}
	if _, _, pointed, para = sourceRecall(p, read, locs, []packet.Note{onWords}); pointed != 1 || para != 1 {
		t.Fatalf("a note on the same words: pointed %v, paragraph %v", pointed, para)
	}
	if rec, _, pointed, para = sourceRecall(p, read, locs, []packet.Note{onParagraph, elsewhere}); pointed != rec || para != 1 {
		t.Fatalf("a note on the same paragraph only: pointed %v, paragraph %v", pointed, para)
	}
	cell := func(kind string, total, w, l int) ZoomCell {
		return ZoomCell{Kind: kind, Arm: ZoomLadder, Total: total, WinsC: w, LossesC: l, PC: exactSignP(w, l)}
	}
	win := []ZoomCell{cell("whole conversation", 4, 20, 3), cell("whole conversation", 8, 18, 4), cell("whole conversation", 16, 15, 3), cell("whole conversation", 32, 6, 5), cell("one source", 4, 3, 5)}
	if !strings.Contains(ZoomVerdict(win), "Verdict: supported") {
		t.Fatal(ZoomVerdict(win))
	}
	lose := append([]ZoomCell{}, win...)
	lose[4] = cell("one source", 4, 2, 20)
	if !strings.Contains(ZoomVerdict(lose), "Verdict: not supported") {
		t.Fatal("a significant one-source loss should fail the rule")
	}
	if !strings.Contains(ZoomVerdict(win[2:]), "Verdict: not supported") {
		t.Fatal("too few winning budgets should fail the rule")
	}
}
