package vault

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Atoms are the paragraph-sized pieces of a turn: the unit a packet's
// ranked section is built from (docs/CONTEXT-COMPILER.md). An atom is an
// address into the immutable body, never a rewrite: atom k of a turn
// always names the same bytes, and the atoms of a turn tile its body, so
// joined they are the body exactly.
//
// The split is deterministic and Markdown-aware: blocks end at blank
// lines, except inside a fenced code block; blocks are joined until they
// reach AtomMinBytes, and a short heading, bold label or line ending in a
// colon always opens the next atom, never closes one; a block longer than
// AtomMaxBytes (a long code block or table included) is cut at a blank
// line, else a line end, else a sentence end, else a space. Changing
// these rules changes atom ids, so AtomSeg names the rules in force.
const (
	AtomMinBytes = 200  // about 60 tokens
	AtomMaxBytes = 2400 // about 720 tokens
	AtomSeg      = 2
)

// Atom is one paragraph-sized span of a turn's body.
type Atom struct {
	Turn  string
	K     int // 1-based
	Start int // byte offsets: atom k ends where atom k+1 starts
	End   int
	Kind  string // prose, code, list, table or heading
}

// ID is the atom's recall id, tNNNNNN#aK.
func (a Atom) ID() string { return AtomID(a.Turn, a.K) }

// Text is the atom without the blank lines around it.
func (a Atom) Text(body string) string { return strings.TrimSpace(body[a.Start:a.End]) }

// AtomID names atom k (1-based) of a turn.
func AtomID(turn string, k int) string { return fmt.Sprintf("%s#a%d", turn, k) }

// AtomRangeID names atoms a to b of a turn: tNNNNNN#aK, or tNNNNNN#aK-M
// for a run of adjacent atoms.
func AtomRangeID(turn string, a, b int) string {
	if b <= a {
		return AtomID(turn, a)
	}
	return fmt.Sprintf("%s#a%d-%d", turn, a, b)
}

type span struct{ start, end int }

// Atoms splits a turn body into its atoms. An empty body has none.
func Atoms(turn, body string) []Atom {
	if strings.TrimSpace(body) == "" {
		return nil
	}
	// blocks: runs of non-blank lines; a fence keeps its blank lines
	var blocks []span
	fence := false
	cur := span{-1, -1}
	for off := 0; off < len(body); {
		nl := strings.IndexByte(body[off:], '\n')
		end := len(body)
		if nl >= 0 {
			end = off + nl + 1
		}
		line := strings.TrimSpace(body[off:end])
		if isFence(line) {
			fence = !fence
		}
		if line == "" && !fence {
			if cur.start >= 0 {
				blocks = append(blocks, cur)
				cur = span{-1, -1}
			}
		} else {
			if cur.start < 0 {
				cur.start = off
			}
			cur.end = end
		}
		off = end
	}
	if cur.start >= 0 {
		blocks = append(blocks, cur)
	}

	// short blocks join the next one; a heading or label that follows a
	// group long enough to stand alone opens the next group; a short
	// tail joins the last group
	var groups []span
	g := span{-1, -1}
	for _, b := range blocks {
		if g.start >= 0 && b.start-g.start >= AtomMinBytes && leadsNext(body[b.start:b.end]) {
			groups = append(groups, span{g.start, b.start})
			g = span{-1, -1}
		}
		if g.start < 0 {
			g.start = b.start
		}
		g.end = b.end
		if g.end-g.start >= AtomMinBytes && !leadsNext(body[b.start:b.end]) {
			groups = append(groups, g)
			g = span{-1, -1}
		}
	}
	if g.start >= 0 {
		if len(groups) > 0 {
			groups[len(groups)-1].end = g.end
		} else {
			groups = append(groups, g)
		}
	}

	// long groups are cut
	var pieces []span
	for _, g := range groups {
		for g.end-g.start > AtomMaxBytes {
			cut := cutPoint(body, g.start)
			pieces = append(pieces, span{g.start, cut})
			g.start = cut
		}
		pieces = append(pieces, g)
	}

	// tile the body: the whitespace between pieces belongs to the piece
	// before it, the leading whitespace to the first
	out := make([]Atom, len(pieces))
	for i, p := range pieces {
		start, end := p.start, len(body)
		if i == 0 {
			start = 0
		}
		if i+1 < len(pieces) {
			end = pieces[i+1].start
		}
		out[i] = Atom{Turn: turn, K: i + 1, Start: start, End: end, Kind: atomKind(body[p.start:p.end])}
	}
	return out
}

// cutPoint is where to cut a block starting at start that is longer than
// AtomMaxBytes: after a blank line, else a line end, else a sentence end,
// else a space, in the last two thirds of the allowed length; else at a
// rune boundary.
func cutPoint(body string, start int) int {
	s := body[start : start+AtomMaxBytes]
	for _, sep := range []string{"\n\n", "\n", ". ", "? ", "! ", "; ", " "} {
		if i := strings.LastIndex(s, sep); i > len(s)/3 {
			return start + i + len(sep)
		}
	}
	cut := start + AtomMaxBytes
	for cut > start && !utf8Start(body[cut]) {
		cut--
	}
	if cut == start {
		// not UTF-8: cut at the limit so the split always advances
		return start + AtomMaxBytes
	}
	return cut
}

func isFence(trimmedLine string) bool {
	return strings.HasPrefix(trimmedLine, "```") || strings.HasPrefix(trimmedLine, "~~~")
}

// leadsNext reports a short block that introduces what follows it: one
// heading line, one bold-only line, or a block ending in a colon.
func leadsNext(block string) bool {
	b := strings.TrimSpace(block)
	if b == "" || len(b) >= AtomMinBytes {
		return false
	}
	if strings.HasSuffix(b, ":") {
		return true
	}
	if strings.Contains(b, "\n") {
		return false
	}
	return strings.HasPrefix(b, "#") || (len(b) > 4 && strings.HasPrefix(b, "**") && strings.HasSuffix(b, "**"))
}

// InFence reports whether byte offset off of body lies inside a fenced
// code block: after an opening fence line and before its closing one.
func InFence(body string, off int) bool {
	fence := false
	for pos := 0; pos < off && pos < len(body); {
		end := len(body)
		if nl := strings.IndexByte(body[pos:], '\n'); nl >= 0 {
			end = pos + nl + 1
		}
		if end > off {
			break // the line holding off has not ended before it
		}
		if isFence(strings.TrimSpace(body[pos:end])) {
			fence = !fence
		}
		pos = end
	}
	return fence
}

func atomKind(text string) string {
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		l := strings.TrimSpace(line)
		switch {
		case l == "" || strings.HasPrefix(l, "#"):
			continue
		case strings.HasPrefix(l, "```") || strings.HasPrefix(l, "~~~"):
			return "code"
		case strings.HasPrefix(l, "|"):
			return "table"
		case strings.HasPrefix(l, "- ") || strings.HasPrefix(l, "* ") || listNum.MatchString(l):
			return "list"
		default:
			return "prose"
		}
	}
	return "heading"
}

var listNum = regexp.MustCompile(`^\d+[.)] `)

// Ref is a parsed recall id: a whole turn, one part of a long turn
// (#pK), or a run of atoms (#aK or #aK-M).
type Ref struct {
	Turn    string
	Part    int // 1-based; 0 when the ref is not a part
	Atom    int // first atom, 1-based; 0 when the ref is not atoms
	AtomEnd int // last atom (Atom for a single one)
}

// Whole reports whether the ref names a whole turn.
func (r Ref) Whole() bool { return r.Part == 0 && r.Atom == 0 }

var refRe = regexp.MustCompile(`^(t\d{6,})(?:#p(\d+)|#a(\d+)(?:-(\d+))?)?$`)

// ParseRef parses a recall id.
func ParseRef(id string) (Ref, bool) {
	m := refRe.FindStringSubmatch(strings.TrimSpace(id))
	if m == nil {
		return Ref{}, false
	}
	r := Ref{Turn: m[1]}
	num := func(s string) int {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return -1
		}
		return n
	}
	switch {
	case m[2] != "":
		if r.Part = num(m[2]); r.Part < 1 {
			return Ref{}, false
		}
	case m[3] != "":
		if r.Atom = num(m[3]); r.Atom < 1 {
			return Ref{}, false
		}
		r.AtomEnd = r.Atom
		if m[4] != "" {
			if r.AtomEnd = num(m[4]); r.AtomEnd < r.Atom {
				return Ref{}, false
			}
		}
	}
	return r, true
}

// AtomsText is the verbatim text of atoms a to b (1-based, inclusive) of
// body, or false when the turn has fewer atoms.
func AtomsText(turn, body string, a, b int) (string, bool) {
	atoms := Atoms(turn, body)
	if a < 1 || b < a || b > len(atoms) {
		return "", false
	}
	return SpanText(body[atoms[a-1].Start:atoms[b-1].End]), true
}

// SpanText is a span of a body as it is sent: without the blank lines
// before it and the whitespace after it. The indentation of its first
// line is kept, so an indented code block or list item stays what it was.
func SpanText(span string) string {
	span = strings.TrimRight(span, " \t\r\n")
	for {
		i := strings.IndexByte(span, '\n')
		if i < 0 || strings.TrimSpace(span[:i]) != "" {
			return span
		}
		span = span[i+1:]
	}
}

// EvidenceAnchor is a durable, content-addressed byte span into an immutable
// turn. Storing (Turn, BodySHA, Start, End, Seg) lets citations and exported
// notes survive physical repacking or future segmentation changes while
// keeping a human-readable AtomRef alongside them.
type EvidenceAnchor struct {
	Turn    string `json:"turn"`
	BodySHA string `json:"body_sha"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Seg     int    `json:"seg"`
	AtomRef string `json:"atom_ref,omitempty"`
}

// VerifyAnchor verifies a durable anchor against a turn's body and digest and
// returns its exact substring.
func VerifyAnchor(body, bodySHA string, a EvidenceAnchor) (string, bool) {
	if a.BodySHA != "" && bodySHA != "" && a.BodySHA != bodySHA {
		return "", false
	}
	if a.Start < 0 || a.End <= a.Start || a.End > len(body) {
		return "", false
	}
	return body[a.Start:a.End], true
}

// UnclosedFence reports whether text leaves a Markdown code fence open at its
// end, returning the fence marker (e.g. "```" or "~~~") needed to close it.
func UnclosedFence(text string) (marker string, open bool) {
	var fenceChar byte
	fenceLen := 0
	for off := 0; off < len(text); {
		nl := strings.IndexByte(text[off:], '\n')
		end := len(text)
		if nl >= 0 {
			end = off + nl + 1
		}
		line := strings.TrimSpace(text[off:end])
		if len(line) >= 3 && (line[0] == '`' || line[0] == '~') {
			ch := line[0]
			n := 0
			for n < len(line) && line[n] == ch {
				n++
			}
			if n >= 3 {
				if !open {
					open = true
					fenceChar = ch
					fenceLen = n
				} else if ch == fenceChar && n >= fenceLen && strings.TrimSpace(line[n:]) == "" {
					open = false
					fenceChar = 0
					fenceLen = 0
				}
			}
		}
		off = end
	}
	if !open {
		return "", false
	}
	return strings.Repeat(string(fenceChar), fenceLen), true
}

// EnclosingBlock expands atom k (1-based) to the complete enclosing code
// block, table, or contiguous list (including an immediately preceding heading
// or colon label that introduces it). Returns (startAtom, endAtom, kind).
func EnclosingBlock(turn, body string, k int) (int, int, string) {
	atoms := Atoms(turn, body)
	if k < 1 || k > len(atoms) {
		return 0, 0, ""
	}
	idx := k - 1
	lo, hi := idx, idx
	a := atoms[idx]
	// Case 1: inside a split fenced code block.
	if a.Kind == "code" || InFence(body, a.Start) || InFence(body, a.End) {
		for lo > 0 && InFence(body, atoms[lo].Start) {
			lo--
		}
		for hi+1 < len(atoms) && InFence(body, atoms[hi].End) {
			hi++
		}
		return lo + 1, hi + 1, "code"
	}
	// Case 2: contiguous table or list across atom splits.
	if a.Kind == "table" || a.Kind == "list" {
		kind := a.Kind
		for lo > 0 && atoms[lo-1].Kind == kind {
			lo--
		}
		for hi+1 < len(atoms) && atoms[hi+1].Kind == kind {
			hi++
		}
		if lo > 0 && (atoms[lo-1].Kind == "heading" || leadsNext(atoms[lo-1].Text(body))) {
			lo--
		}
		return lo + 1, hi + 1, kind
	}
	// Case 3: heading or label introducing the next block.
	if a.Kind == "heading" || leadsNext(a.Text(body)) {
		if hi+1 < len(atoms) {
			_, nextHi, nextKind := EnclosingBlock(turn, body, k+1)
			if nextHi > 0 {
				return lo + 1, nextHi, nextKind
			}
		}
	}
	return lo + 1, hi + 1, a.Kind
}
