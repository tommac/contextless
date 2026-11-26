package vault

import (
	"fmt"
	"strings"
)

// PartBytes is the size of the addressable parts of a long turn. A turn
// longer than this can be recalled one part at a time as tNNNNNN#pK
// instead of whole (ARC's chunked recall: exact, non-overlapping pieces,
// never a truncation). The split is deterministic, so a part id always
// names the same text, and the parts joined are the turn byte for byte.
const PartBytes = 24 * 1024

// SplitParagraphs cuts text into pieces of at most size bytes, at blank
// lines where it can, else at line ends, else at spaces, else at a rune
// boundary. Joined, the pieces are the text exactly.
func SplitParagraphs(text string, size int) []string {
	var out []string
	for len(text) > size {
		cut := -1
		for _, sep := range []string{"\n\n", "\n", " "} {
			if i := strings.LastIndex(text[:size], sep); i > size/3 {
				cut = i + len(sep)
				break
			}
		}
		if cut <= 0 {
			cut = size
			for cut > 0 && !utf8Start(text[cut]) {
				cut--
			}
			if cut == 0 {
				cut = size // not UTF-8: cut at the limit so the split always advances
			}
		}
		out = append(out, text[:cut])
		text = text[cut:]
	}
	if text != "" {
		out = append(out, text)
	}
	return out
}

func utf8Start(b byte) bool { return b&0xC0 != 0x80 }

// Parts splits a turn body into its addressable parts: one part for a
// turn of at most PartBytes.
func Parts(body string) []string {
	if len(body) <= PartBytes {
		return []string{body}
	}
	return SplitParagraphs(body, PartBytes)
}

// PartID names part k (1-based) of a turn.
func PartID(turn string, k int) string { return fmt.Sprintf("%s#p%d", turn, k) }

// ParseRecallID splits a recall id into the turn id and the part number
// (0 for the whole turn or for atoms; ParseRef tells those apart).
func ParseRecallID(id string) (turn string, part int, ok bool) {
	r, ok := ParseRef(id)
	return r.Turn, r.Part, ok
}
