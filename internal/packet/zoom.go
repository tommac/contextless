package packet

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// A zoom ladder is a pyramid over the immutable turns: paragraphs at the
// bottom, a line per turn, then a line per ten turns, per hundred, and so
// on. Upper levels are blurred: here they are built by code alone, from
// each span's opening words, frequent terms and values. They only say where
// to look; what is read in the end is always a verbatim paragraph.
//
// This is the experiment behind `cx zoom`, not part of live packets. A
// deterministic reader stands in for a model: it follows the lines whose
// words best match a question, paying tokens for every line and paragraph.

// ZoomFanout is how many spans one level above covers.
const ZoomFanout = 10

// zoomDigestTurns is the size of the block a flat catalog falls back to.
const zoomDigestTurns = DigestBlock

// zoomReadOverhead is the header a verbatim paragraph comes with.
const zoomReadOverhead = 40

// ZoomNode is one line of a ladder, or a verbatim paragraph (level 0).
type ZoomNode struct {
	ID          string
	Level       int // 0 paragraph, 1 turn, 2 and up blocks
	First, Last int // turn sequence numbers covered
	Text        string
	Tokens      int
	Children    []*ZoomNode
	words       map[string]bool
	terms       []string
	weight      map[string]float64 // each term's weight: how often, times how rare
	values      []string
	opening     string
}

// ZoomTree is a ladder over a conversation.
type ZoomTree struct {
	Turns   []*ZoomNode // level 1, in order
	Roots   []*ZoomNode // the highest complete blocks, then the unfinished tail
	Digests []*ZoomNode // fixed blocks of DigestBlock turns over Turns
	idf     map[string]float64
}

func zoomTokens(s string) int { return Tokens(providers.Profile{}, len(s)+3) }

// ReadTokens is what reading one paragraph verbatim costs a reader of the
// zoom test: its text and the header it comes with.
func ReadTokens(text string) int { return zoomTokens(text) + zoomReadOverhead }

func (n *ZoomNode) finish(text string) {
	n.Text, n.Tokens = text, zoomTokens(text)
	n.words = ContentWords(text)
}

func take(xs []string, n int) []string { return xs[:min(n, len(xs))] }

func detail(terms, values []string, nt, nv int) string {
	s := ""
	if t := take(terms, nt); len(t) > 0 {
		s += " · terms: " + strings.Join(t, ", ")
	}
	if v := take(values, nv); len(v) > 0 {
		s += " · values: " + strings.Join(v, ", ")
	}
	return s
}

// turnTerms are a turn's most distinctive words: frequent in it, rare in
// the conversation. Plain frequency would fill the lines above with the
// words every turn shares, which tell a reader nothing about where to look.
func turnTerms(body string, idf map[string]float64, n int) ([]string, map[string]float64) {
	weight, first := map[string]float64{}, map[string]int{}
	for i, tk := range tokens(body) {
		w := strings.ToLower(strings.Trim(tk.text, ".,:;"))
		if len(w) < 4 || valueStop[w] || termStop[w] {
			continue
		}
		if _, ok := first[w]; !ok {
			first[w] = i
		}
		weight[w] += idf[stem(w)]
	}
	return topWeighted(weight, first, n)
}

func topWeighted(weight map[string]float64, first map[string]int, n int) ([]string, map[string]float64) {
	words := make([]string, 0, len(weight))
	for w := range weight {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		if weight[words[i]] != weight[words[j]] {
			return weight[words[i]] > weight[words[j]]
		}
		if first[words[i]] != first[words[j]] {
			return first[words[i]] < first[words[j]]
		}
		return words[i] < words[j]
	})
	words = take(words, n)
	kept := map[string]float64{}
	for _, w := range words {
		kept[w] = weight[w]
	}
	return words, kept
}

// mergeRanked orders the items of child lists by how many children hold
// them, then by first appearance.
func mergeRanked(lists [][]string) []string {
	count, first := map[string]int{}, map[string]int{}
	for _, l := range lists {
		for _, w := range l {
			if _, ok := first[w]; !ok {
				first[w] = len(first)
			}
			count[w]++
		}
	}
	out := make([]string, 0, len(count))
	for w := range count {
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool {
		if count[out[i]] != count[out[j]] {
			return count[out[i]] > count[out[j]]
		}
		return first[out[i]] < first[out[j]]
	})
	return out
}

// zoomBlock is the line over children. Its terms are those the block says
// most about: each child's weight for a term, again times the term's
// rarity, so a word one turn stresses outranks a word every turn shares.
// The rarity compounds at each level, so about half of a ten-block's shown
// terms occur in one turn of it. This is the only recipe the zoom test ran
// as its result; a review's sensitivity run that ranked terms by how many
// children list them read about 5 points more at 16,000 tokens, with the
// same verdict.
func (z *ZoomTree) zoomBlock(children []*ZoomNode, level int) *ZoomNode {
	n := &ZoomNode{Level: level, First: children[0].First, Last: children[len(children)-1].Last, Children: children}
	n.ID = fmt.Sprintf("z%d:%s-%s", level, vault.TurnID(n.First), vault.TurnID(n.Last))
	var values [][]string
	weight, first := map[string]float64{}, map[string]int{}
	for _, c := range children {
		values = append(values, c.values)
		for _, w := range c.terms {
			if _, ok := first[w]; !ok {
				first[w] = len(first)
			}
			weight[w] += c.weight[w] * z.idf[stem(w)]
		}
		if n.opening == "" {
			n.opening = c.opening
		}
	}
	n.terms, n.weight = topWeighted(weight, first, 24)
	n.values = take(mergeRanked(values), 10)
	n.finish(fmt.Sprintf("%s to %s, %d turns: %s%s", vault.TurnID(n.First), vault.TurnID(n.Last), n.Last-n.First+1, n.opening, detail(n.terms, n.values, 12, 5)))
	return n
}

// BuildZoom builds the ladder. With detailed lines every turn line carries
// its terms and values; otherwise it is today's catalog preview.
func BuildZoom(turns []vault.Turn, detailed bool) *ZoomTree {
	z := &ZoomTree{idf: map[string]float64{}}
	df := map[string]int{}
	for _, t := range turns {
		for w := range ContentWords(t.Body) {
			df[w]++
		}
	}
	for w, c := range df {
		z.idf[w] = math.Log(1 + float64(len(turns))/float64(c))
	}
	for _, t := range turns {
		n := &ZoomNode{ID: t.ID, Level: 1, First: t.Seq, Last: t.Seq, opening: oneLine(t.Body, 90), values: take(Values(t.Body), 8)}
		n.terms, n.weight = turnTerms(t.Body, z.idf, 12)
		if detailed {
			n.finish(fmt.Sprintf("%s %s: %s%s", t.ID, t.Role, n.opening, detail(n.terms, n.values, 6, 4)))
		} else {
			n.finish(PreviewLine(t))
		}
		for _, a := range atomsOf(t) {
			text := a.Text(t.Body)
			p := &ZoomNode{ID: a.ID(), First: t.Seq, Last: t.Seq, Text: text, Tokens: zoomTokens(text) + zoomReadOverhead, words: ContentWords(text)}
			n.Children = append(n.Children, p)
		}
		z.Turns = append(z.Turns, n)
	}
	for i := 0; i < len(z.Turns); i += zoomDigestTurns {
		z.Digests = append(z.Digests, z.zoomBlock(z.Turns[i:min(i+zoomDigestTurns, len(z.Turns))], 2))
	}
	// finished blocks climb; what does not fill a block stays a root
	level, cur := 2, z.Turns
	var tails [][]*ZoomNode
	for len(cur) >= ZoomFanout {
		var up []*ZoomNode
		n := len(cur) / ZoomFanout * ZoomFanout
		for i := 0; i < n; i += ZoomFanout {
			up = append(up, z.zoomBlock(cur[i:i+ZoomFanout], level))
		}
		tails = append(tails, cur[n:])
		cur, level = up, level+1
	}
	z.Roots = append(z.Roots, cur...)
	for i := len(tails) - 1; i >= 0; i-- {
		z.Roots = append(z.Roots, tails[i]...)
	}
	return z
}

// score adds up the rarity of the question's words a node holds. The words
// come sorted: a sum over a map adds in a different order on every call,
// which changes its last bit, and two equal scores would then compare
// unequal and the walk would not be repeatable.
func (z *ZoomTree) score(n *ZoomNode, q []string) float64 {
	s := 0.0
	for _, w := range q {
		if n.words[w] {
			s += z.idf[w]
		}
	}
	return s
}

// repeated says whether a paragraph's exact text was read already. A
// conversation repeats itself (pasted text, quoted earlier answers), and a
// second copy costs tokens and adds nothing; live packets skip it by the
// same rule (PickAtoms).
func repeated(seen map[string]bool, text string) bool {
	return len(text) >= atomRepeatBytes && seen[text]
}

// sortedWords lists the keys of a word map in a fixed order.
func sortedWords[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for w := range m {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}

func viewTokens(v []*ZoomNode) int {
	sum := 0
	for _, n := range v {
		sum += n.Tokens
	}
	return sum
}

// TurnLines are the code-built lines of every turn, oldest first: with
// detailed lines, each turn's id, author, opening, most distinctive terms
// and values.
func (z *ZoomTree) TurnLines() []*ZoomNode { return z.Turns }

// CatalogView is a flat catalog within budget: the newest turn lines, with
// older turns folded into digest blocks, as packets do when space runs out.
func (z *ZoomTree) CatalogView(budget int) []*ZoomNode {
	digestsFor := func(keep int) []*ZoomNode { // blocks holding a turn older than the kept ones
		var out []*ZoomNode
		cut := len(z.Turns) - keep
		for i, d := range z.Digests {
			if i*zoomDigestTurns < cut {
				nd := *d
				nd.Children = d.Children[:min(len(d.Children), cut-i*zoomDigestTurns)]
				out = append(out, &nd)
			}
		}
		return out
	}
	best := []*ZoomNode{}
	for keep := 0; keep <= len(z.Turns); keep++ {
		v := append(digestsFor(keep), z.Turns[len(z.Turns)-keep:]...)
		if viewTokens(v) <= budget {
			best = v
		} else if keep > 0 {
			break
		}
	}
	if len(best) == 0 { // even the digests do not fit: the newest that do
		v := digestsFor(0)
		for i := len(v) - 1; i >= 0 && viewTokens(best)+v[i].Tokens <= budget; i-- {
			best = append([]*ZoomNode{v[i]}, best...)
		}
	}
	return best
}

// LadderView starts from every root and zooms into blocks within budget:
// the blocks that best match query first, then the newest. With no query
// the view depends only on the conversation, so it stays stable.
func (z *ZoomTree) LadderView(query string, budget int) []*ZoomNode {
	// every root is always in view, so the whole conversation is covered
	// at some level; a budget below that buys no zoom
	v := append([]*ZoomNode{}, z.Roots...)
	q := sortedWords(ContentWords(query))
	for {
		best, bestScore := -1, -1.0
		for i, n := range v {
			if n.Level < 2 {
				continue
			}
			if viewTokens(v)-n.Tokens+viewTokens(n.Children) > budget {
				continue
			}
			if s := z.score(n, q); s > bestScore || (s == bestScore && n.Last > v[best].Last) {
				best, bestScore = i, s
			}
		}
		if best < 0 {
			return v
		}
		next := append([]*ZoomNode{}, v[:best]...)
		next = append(next, v[best].Children...)
		v = append(next, v[best+1:]...)
	}
}

type zoomItem struct {
	n     *ZoomNode
	score float64
}
type zoomHeap []zoomItem

func (h zoomHeap) Len() int { return len(h) }
func (h zoomHeap) Less(i, j int) bool {
	a, b := h[i], h[j]
	if a.score != b.score {
		return a.score > b.score
	}
	if a.n.Last != b.n.Last {
		return a.n.Last > b.n.Last
	}
	return a.n.ID < b.n.ID
}
func (h zoomHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *zoomHeap) Push(x any)   { *h = append(*h, x.(zoomItem)) }
func (h *zoomHeap) Pop() any     { o := *h; x := o[len(o)-1]; *h = o[:len(o)-1]; return x }

// ZoomWalk is the deterministic reader: from a view it repeatedly opens the
// line or paragraph that best matches query, paying for each line revealed
// and each paragraph read, and returns the verbatim paragraphs read within
// budget. Text with none of the query's words is never opened.
func (z *ZoomTree) ZoomWalk(view []*ZoomNode, query string, budget int) []string {
	read, _ := z.ZoomWalkSpent(view, query, budget)
	return read
}

// ZoomWalkSpent is ZoomWalk, also returning the tokens it spent. The walk
// stops when no line or paragraph sharing a word with the query is left, so
// it often spends less than it was given.
func (z *ZoomTree) ZoomWalkSpent(view []*ZoomNode, query string, budget int) ([]string, int) {
	q := sortedWords(ContentWords(query))
	h := &zoomHeap{}
	push := func(n *ZoomNode) {
		if s := z.score(n, q); s > 0 {
			heap.Push(h, zoomItem{n, s})
		}
	}
	for _, n := range view {
		push(n)
	}
	spent := 0
	var read []string
	seen := map[string]bool{}
	for h.Len() > 0 {
		n := heap.Pop(h).(zoomItem).n
		if n.Level == 0 {
			if repeated(seen, n.Text) {
				continue
			}
			if spent+n.Tokens <= budget {
				spent += n.Tokens
				read = append(read, n.Text)
				seen[n.Text] = true
			}
			continue
		}
		cost := 0
		if n.Level >= 2 {
			cost = viewTokens(n.Children) // a turn's paragraphs are paid for when read
		}
		if spent+cost > budget {
			continue
		}
		spent += cost
		for _, c := range n.Children {
			push(c)
		}
	}
	return read, spent
}

// Paragraphs are every paragraph node, oldest first: a view with no lines,
// for measuring the walk's own ranking apart from its navigation.
func (z *ZoomTree) Paragraphs() []*ZoomNode {
	var out []*ZoomNode
	for _, t := range z.Turns {
		out = append(out, t.Children...)
	}
	return out
}

// ViewTokens is what showing a view costs.
func ViewTokens(v []*ZoomNode) int { return viewTokens(v) }

// SearchRead is the flat baseline: paragraphs in ranked order, within the
// same budget, each paying the same header.
func SearchRead(turns []vault.Turn, s *store.Store, query string, beforeSeq, budget int) []string {
	return Texts(SearchPick(turns, s, query, beforeSeq, budget))
}

// SearchPick is SearchRead, returning the paragraphs picked.
func SearchPick(turns []vault.Turn, s *store.Store, query string, beforeSeq, budget int) []RankedAtom {
	n := 1
	for _, t := range turns {
		n += len(atomsOf(t))
	}
	return pickInOrder(RankAtoms(turns, s, query, beforeSeq, n), budget)
}
