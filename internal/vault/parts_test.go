package vault

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPartsAreLosslessAndStable(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 40; i++ {
		b.WriteString(strings.Repeat("A paragraph of ordinary words for the split. ", 60) + "\n\n")
	}
	b.WriteString(strings.Repeat("one very long paragraph without breaks ", 1500))
	b.WriteString(strings.Repeat("é漢", 9000)) // no spaces at all, multibyte
	body := b.String()
	parts := Parts(body)
	if len(parts) < 4 {
		t.Fatalf("want several parts, got %d", len(parts))
	}
	if strings.Join(parts, "") != body {
		t.Fatal("parts do not join back to the turn")
	}
	for i, p := range parts {
		if len(p) > PartBytes {
			t.Errorf("part %d is %d bytes, over %d", i+1, len(p), PartBytes)
		}
		if !utf8.ValidString(p) {
			t.Errorf("part %d cuts a character", i+1)
		}
	}
	again := Parts(body)
	for i := range parts {
		if again[i] != parts[i] {
			t.Fatalf("part %d differs between two splits of the same turn", i+1)
		}
	}
	if got := Parts("short"); len(got) != 1 || got[0] != "short" {
		t.Fatalf("a short turn is one part: %q", got)
	}
}

func TestParseRecallID(t *testing.T) {
	for _, c := range []struct {
		in   string
		turn string
		part int
		ok   bool
	}{
		{"t000123", "t000123", 0, true},
		{" t000123#p4 ", "t000123", 4, true},
		{"t000123#p0", "", 0, false},
		{"t12", "", 0, false},
		{"t000123#4", "", 0, false},
		{"fa-abc", "", 0, false},
	} {
		turn, part, ok := ParseRecallID(c.in)
		if turn != c.turn || part != c.part || ok != c.ok {
			t.Errorf("%q: got %q %d %v", c.in, turn, part, ok)
		}
	}
}
