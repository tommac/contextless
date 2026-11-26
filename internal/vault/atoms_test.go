package vault

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestAtomsTileTheBodyAndKeepBlocksWhole(t *testing.T) {
	long := strings.Repeat("The packet is sized before the model starts, and nothing is compressed. ", 60) // about 4.3 KB, one paragraph
	body := "## Plan\n\n" +
		"Short remark.\n\n" +
		"The first real paragraph explains the plan in enough words to stand on its own as one atom of the turn, " +
		"with the port 5433 and the decision to keep every turn verbatim, never rewritten and never summarized to make it fit.\n\n" +
		"```go\nfunc main() {\n\t// a code block long enough to be an atom of its own, so the test can see it stay whole\n\tconfig := load(\"cx.yaml\")\n\n\tfmt.Println(\"a blank line inside a fence does not end the block\")\n}\n```\n\n" +
		long + "\n\n" +
		"- one item in a list that is long enough to be an atom of its own\n- two, with a second line of words so the list passes the minimum size\n- three, and the closing words of the list arrive here at the end of the turn.\n"
	atoms := Atoms("t000007", body)
	if len(atoms) < 4 {
		t.Fatalf("want at least 4 atoms, got %d", len(atoms))
	}
	var joined strings.Builder
	for i, a := range atoms {
		if a.K != i+1 || a.Turn != "t000007" {
			t.Fatalf("atom %d numbered %d of %s", i, a.K, a.Turn)
		}
		if i > 0 && a.Start != atoms[i-1].End {
			t.Fatalf("atom %d starts at %d, previous ends at %d", a.K, a.Start, atoms[i-1].End)
		}
		if n := len(a.Text(body)); n > AtomMaxBytes {
			t.Fatalf("atom %d is %d bytes, over %d", a.K, n, AtomMaxBytes)
		}
		joined.WriteString(body[a.Start:a.End])
	}
	if joined.String() != body {
		t.Fatal("atoms joined are not the body")
	}
	// the heading and the short remark join the paragraph after them
	if first := atoms[0].Text(body); !strings.HasPrefix(first, "## Plan") || !strings.Contains(first, "port 5433") {
		t.Fatalf("heading and remark should lead the first paragraph's atom: %q", first)
	}
	// the fenced block stays whole despite its blank line
	var code *Atom
	for i := range atoms {
		if strings.Contains(atoms[i].Text(body), "func main()") {
			code = &atoms[i]
		}
	}
	if code == nil || !strings.Contains(code.Text(body), "does not end the block\")\n}\n```") || code.Kind != "code" {
		t.Fatalf("the code block was split or misclassified: %+v", code)
	}
	// the long paragraph is cut at sentence ends
	for _, a := range atoms {
		if txt := a.Text(body); strings.HasPrefix(txt, "The packet is sized") && !strings.HasSuffix(txt, "compressed.") {
			t.Fatalf("long paragraph cut mid-sentence: ...%q", txt[len(txt)-30:])
		}
	}
	// stable: the same body gives the same atoms
	again := Atoms("t000007", body)
	for i := range atoms {
		if atoms[i] != again[i] {
			t.Fatal("atoms are not deterministic")
		}
	}
	if Atoms("t000008", "  \n\n ") != nil {
		t.Fatal("an empty body has no atoms")
	}
	if one := Atoms("t000009", "Just a short line."); len(one) != 1 || one[0].Text("Just a short line.") != "Just a short line." {
		t.Fatalf("a short turn is one atom: %+v", one)
	}
}

func tiles(t *testing.T, body string) []Atom {
	t.Helper()
	atoms := Atoms("t000001", body)
	var joined strings.Builder
	for i, a := range atoms {
		if i > 0 && a.Start != atoms[i-1].End {
			t.Fatalf("atom %d starts at %d, previous ends at %d", a.K, a.Start, atoms[i-1].End)
		}
		if a.End <= a.Start {
			t.Fatalf("atom %d is empty", a.K)
		}
		joined.WriteString(body[a.Start:a.End])
	}
	if joined.String() != body {
		t.Fatal("atoms joined are not the body")
	}
	return atoms
}

// No separators at all: the cut falls back to a rune boundary, also when
// the limit lands inside a rune, and the leading whitespace of a body
// belongs to its first atom.
func TestAtomsNeverCutARune(t *testing.T) {
	for _, body := range []string{
		strings.Repeat("é", AtomMaxBytes),
		"a" + strings.Repeat("é", AtomMaxBytes),
		"ab" + strings.Repeat("漢", AtomMaxBytes),
		"x" + strings.Repeat("😀", AtomMaxBytes),
		"\n\n  \nA body that starts with blank lines.\n\nSecond paragraph.\n",
	} {
		atoms := tiles(t, body)
		if atoms[0].Start != 0 {
			t.Fatalf("the first atom starts at %d", atoms[0].Start)
		}
		for _, a := range atoms {
			if !utf8.ValidString(body[a.Start:a.End]) {
				t.Fatalf("atom %d cuts a rune", a.K)
			}
		}
	}
}

// Bytes that are not UTF-8 must not hang the split: a body of
// continuation bytes has no rune start to walk back to.
func TestAtomsAndPartsAdvanceOnBytesThatAreNotText(t *testing.T) {
	junk := strings.Repeat("\x80", 3*AtomMaxBytes)
	if n := len(tiles(t, junk)); n < 3 {
		t.Fatalf("%d atoms", n)
	}
	big := strings.Repeat("\x80", PartBytes+1)
	if parts := Parts(big); len(parts) != 2 || strings.Join(parts, "") != big {
		t.Fatalf("%d parts", len(parts))
	}
}

// The splitting rules are pinned: changing them without a new AtomSeg
// would silently move every atom id. A fenced block with a blank line
// inside stays one atom though each half could stand alone, and a heading
// after a paragraph long enough to stand alone opens the next atom.
func TestAtomSplitIsPinnedToAtomSeg(t *testing.T) {
	para := strings.Repeat("Words of a paragraph that is long enough to stand alone. ", 5)
	code := "```go\n" + strings.Repeat("\tstep(items) // first half of the block\n", 8) + "\n" +
		strings.Repeat("\tstep(items) // second half of the block\n", 8) + "```"
	body := para + "\n\n## Results\n\n" + para + "\n\n" + code + "\n\n**Next steps**\n\n" + para + "\n"
	atoms := tiles(t, body)
	var got []string
	for _, a := range atoms {
		txt := a.Text(body)
		got = append(got, fmt.Sprintf("%s:%d", a.Kind, len(txt)))
		if strings.HasSuffix(txt, "## Results") || strings.HasSuffix(txt, "**Next steps**") {
			t.Errorf("atom %d ends with the heading written for the next one", a.K)
		}
	}
	want := "prose:284 prose:296 code:658 prose:300"
	if s := strings.Join(got, " "); AtomSeg != 2 || s != want {
		t.Fatalf("AtomSeg %d splits the body as %q, want %q: a change of the rules needs a new AtomSeg", AtomSeg, s, want)
	}
	if !strings.HasPrefix(atoms[1].Text(body), "## Results") || !strings.HasPrefix(atoms[3].Text(body), "**Next steps**") {
		t.Fatal("a heading should open the atom of its section")
	}
}

func TestInFenceAndSpanText(t *testing.T) {
	body := "Intro.\n\n```\nline one\nline two\n```\n\nAfter.\n"
	in := strings.Index(body, "line two")
	if !InFence(body, in) || InFence(body, 0) || InFence(body, strings.Index(body, "After")) {
		t.Fatal("InFence is wrong about the block")
	}
	if got := SpanText("\n  \n    indented first line\nsecond\n\n"); got != "    indented first line\nsecond" {
		t.Fatalf("SpanText = %q", got)
	}
}

func TestParseRef(t *testing.T) {
	cases := []struct {
		in   string
		want Ref
		ok   bool
	}{
		{"t000123", Ref{Turn: "t000123"}, true},
		{" t000123#p2 ", Ref{Turn: "t000123", Part: 2}, true},
		{"t000123#a4", Ref{Turn: "t000123", Atom: 4, AtomEnd: 4}, true},
		{"t000123#a4-6", Ref{Turn: "t000123", Atom: 4, AtomEnd: 6}, true},
		{"t000123#a0", Ref{}, false},
		{"t000123#a6-4", Ref{}, false},
		{"t000123#p0", Ref{}, false},
		{"t123", Ref{}, false},
		{"t000123#x1", Ref{}, false},
	}
	for _, c := range cases {
		got, ok := ParseRef(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("ParseRef(%q) = %+v, %v; want %+v, %v", c.in, got, ok, c.want, c.ok)
		}
	}
	if turn, part, ok := ParseRecallID("t000123#a4"); !ok || turn != "t000123" || part != 0 {
		t.Errorf("ParseRecallID of an atom: %s %d %v", turn, part, ok)
	}
	if id := AtomRangeID("t000123", 4, 6); id != "t000123#a4-6" {
		t.Errorf("AtomRangeID = %s", id)
	}
	para := strings.Repeat("words that make a paragraph long enough to stand alone ", 5)
	body := "First " + para + "\n\nSecond " + para
	if txt, ok := AtomsText("t000001", body, 2, 2); !ok || !strings.HasPrefix(txt, "Second") {
		t.Errorf("AtomsText(2) = %q, %v", txt, ok)
	}
	if _, ok := AtomsText("t000001", body, 1, 3); ok {
		t.Error("AtomsText past the last atom should fail")
	}
}
