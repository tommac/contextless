package packet

import (
	"math"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// Experimental readers for questions whose answer is spread over several
// far-apart passages. Each returns verbatim paragraphs within a budget, as
// SearchRead does, so `cx zoom` can compare them within the same token
// total. None is used in live packets.

// spreadIndex holds what the readers share for one conversation.
type spreadIndex struct {
	turns []vault.Turn
	idf   map[string]float64 // by stem, over turns
	df    map[string]int     // turns holding a stem
	in    map[string]map[string]bool
	words map[string]map[string]int
	n     int
}

func newSpreadIndex(turns []vault.Turn) *spreadIndex {
	x := &spreadIndex{turns: turns, idf: map[string]float64{}, df: map[string]int{}, in: map[string]map[string]bool{}, words: map[string]map[string]int{}}
	for _, t := range turns {
		x.in[t.ID] = ContentWords(t.Body)
		for w := range x.in[t.ID] {
			x.df[w]++
		}
		x.n += len(atomsOf(t))
	}
	for w, c := range x.df {
		x.idf[w] = math.Log(1 + float64(len(turns))/float64(c))
	}
	return x
}

// stems counts the content-word stems of an atom's text.
func (x *spreadIndex) stems(id, text string) map[string]int {
	if m, ok := x.words[id]; ok {
		return m
	}
	m := map[string]int{}
	for _, tk := range tokens(text) {
		w := strings.ToLower(strings.Trim(tk.text, ".,:;"))
		if len(w) < 3 || valueStop[w] {
			continue
		}
		m[stem(w)]++
	}
	x.words[id] = m
	return m
}

func readInOrder(order []RankedAtom, budget int) []string { return Texts(pickInOrder(order, budget)) }

// pickInOrder takes paragraphs in the order given while they fit budget:
// one that does not fit is passed over for a smaller one further down, and
// one whose exact text was already taken is passed over too.
func pickInOrder(order []RankedAtom, budget int) []RankedAtom {
	spent := 0
	var picked []RankedAtom
	seen := map[string]bool{}
	for _, r := range order {
		text := r.Atom.Text(r.Turn.Body)
		if repeated(seen, text) {
			continue
		}
		if c := zoomTokens(text) + zoomReadOverhead; spent+c <= budget {
			spent += c
			picked = append(picked, r)
			seen[text] = true
		}
	}
	return picked
}

// PickInOrder is pickInOrder for callers that bring their own order.
func PickInOrder(order []RankedAtom, budget int) []RankedAtom { return pickInOrder(order, budget) }

// Texts are the verbatim texts of picked paragraphs.
func Texts(picked []RankedAtom) []string {
	out := make([]string, len(picked))
	for i, r := range picked {
		out[i] = r.Atom.Text(r.Turn.Body)
	}
	return out
}

// FeedbackMax fuses the question's ranking and the expanded ranking by a
// paragraph's better place in either, in place of the sum of both. With
// the sum, a passage only the expansion finds can never come before one
// that is about sixtieth in both lists.
var FeedbackMax = false

// FeedbackRead is pseudo-relevance feedback (RM3's idea): the distinctive
// words of the first hits are added to the question, the conversation is
// searched again, and the two rankings are fused. A question in vague
// words borrows the conversation's own words for the topic.
func FeedbackRead(turns []vault.Turn, s *store.Store, x *spreadIndex, query string, beforeSeq, budget, seeds, terms int) []string {
	return Texts(FeedbackPick(turns, s, x, query, beforeSeq, budget, seeds, terms))
}

// FeedbackPick is FeedbackRead, returning the paragraphs picked. A word is
// borrowed only if some turn other than the first hits' own holds it (a
// word found nowhere else can lead nowhere else), and counts once per
// paragraph however often the paragraph repeats it.
func FeedbackPick(turns []vault.Turn, s *store.Store, x *spreadIndex, query string, beforeSeq, budget, seeds, terms int) []RankedAtom {
	base := RankAtoms(turns, s, query, beforeSeq, x.n+1)
	q := ContentWords(query)
	first := base[:max(0, min(seeds, len(base)))]
	seedTurns := map[string]bool{}
	for _, r := range first {
		seedTurns[r.Turn.ID] = true
	}
	elsewhere := func(st string) bool {
		inSeeds := 0
		for id := range seedTurns {
			if x.in[id][st] {
				inSeeds++
			}
		}
		return x.df[st] > inSeeds
	}
	weight := map[string]float64{}
	surface := map[string]string{}
	for i, r := range first {
		once := map[string]bool{}
		for _, tk := range tokens(r.Atom.Text(r.Turn.Body)) {
			w := strings.ToLower(strings.Trim(tk.text, ".,:;"))
			st := stem(w)
			if len(w) < 4 || valueStop[w] || termStop[w] || q[st] || once[st] || !elsewhere(st) {
				continue
			}
			once[st] = true
			weight[st] += x.idf[st] / float64(i+1)
			if _, ok := surface[st]; !ok {
				surface[st] = w
			}
		}
	}
	top := make([]string, 0, len(weight))
	for st := range weight {
		top = append(top, st)
	}
	sort.Slice(top, func(i, j int) bool {
		if weight[top[i]] != weight[top[j]] {
			return weight[top[i]] > weight[top[j]]
		}
		return top[i] < top[j]
	})
	expanded := query
	for _, st := range top[:max(0, min(terms, len(top)))] {
		expanded += " " + surface[st]
	}
	again := RankAtoms(turns, s, expanded, beforeSeq, x.n+1)
	score := map[string]float64{}
	byID := map[string]RankedAtom{}
	for _, list := range [][]RankedAtom{base, again} {
		for i, r := range list {
			if v := 1 / float64(rrfK+i+1); FeedbackMax {
				score[r.Atom.ID()] = math.Max(score[r.Atom.ID()], v)
			} else {
				score[r.Atom.ID()] += v
			}
			byID[r.Atom.ID()] = r
		}
	}
	return pickInOrder(sortByScore(byID, score), budget)
}

func sortByScore(byID map[string]RankedAtom, score map[string]float64) []RankedAtom {
	out := make([]RankedAtom, 0, len(byID))
	for _, r := range byID {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := score[out[i].Atom.ID()], score[out[j].Atom.ID()]
		if a != b {
			return a > b
		}
		if out[i].Turn.Seq != out[j].Turn.Seq {
			return out[i].Turn.Seq > out[j].Turn.Seq
		}
		return out[i].Atom.K < out[j].Atom.K
	})
	return out
}

// LinkRead follows links between paragraphs, built by code: two paragraphs
// are linked when they share rare words (the wiki-link of a note system,
// without anyone writing it). The first hits spread their score to the
// paragraphs they are linked to, one step, so a passage that never uses
// the question's words is reached through a passage that does.
func LinkRead(turns []vault.Turn, s *store.Store, x *spreadIndex, query string, beforeSeq, budget, seeds int, lambda float64) []string {
	base := RankAtoms(turns, s, query, beforeSeq, x.n+1)
	rareMax := math.Log(1 + float64(len(turns))/math.Max(2, float64(len(turns))/20)) // in at most 5% of turns
	post := map[string][]RankedAtom{}
	byID := map[string]RankedAtom{}
	for _, t := range turns {
		if beforeSeq > 0 && t.Seq >= beforeSeq {
			continue
		}
		for _, a := range atomsOf(t) {
			r := RankedAtom{Turn: t, Atom: a}
			byID[a.ID()] = r
			for st := range x.stems(a.ID(), a.Text(t.Body)) {
				if x.idf[st] >= rareMax {
					post[st] = append(post[st], r)
				}
			}
		}
	}
	score := map[string]float64{}
	for i, r := range base {
		score[r.Atom.ID()] = 1 / float64(rrfK+i+1)
	}
	q := ContentWords(query)
	for i, r := range base[:min(seeds, len(base))] {
		seed := 1 / float64(rrfK+i+1)
		link := map[string]float64{}
		total := 0.0
		for _, st := range sortedWords(x.stems(r.Atom.ID(), r.Atom.Text(r.Turn.Body))) {
			if x.idf[st] < rareMax || q[st] {
				continue // the question's own words are already searched
			}
			total += x.idf[st]
			for _, o := range post[st] {
				if o.Atom.ID() != r.Atom.ID() {
					link[o.Atom.ID()] += x.idf[st]
				}
			}
		}
		for id, w := range link {
			score[id] += lambda * seed * w / math.Max(total, 1)
		}
	}
	keep := map[string]RankedAtom{}
	for id := range score {
		keep[id] = byID[id]
	}
	return readInOrder(sortByScore(keep, score), budget)
}

// CoverCostPower is how strongly CoverRead prefers a cheap paragraph: its
// gain is divided by cost to this power. At 1 (gain per token, the usual
// rule for a budgeted pick) it skipped search's best paragraphs for short
// ones, because nothing in the gain grows with a paragraph's length: on the
// development conversation the first-ranked paragraph went unread in a third
// of the questions at 4,000 tokens. At 0 cost only decides what fits. 0.3
// did best of 0, 0.3 and 1 on that conversation.
var CoverCostPower = 0.3

// CoverRead picks paragraphs for coverage, not just rank: each new one is
// worth the question words it adds, with diminishing returns for words
// already covered, plus a bonus for a stretch of the conversation not yet
// represented, all per token (budgeted submodular selection). An inventory
// or a timeline needs many different passages, not the best one five times.
func CoverRead(turns []vault.Turn, s *store.Store, x *spreadIndex, query string, beforeSeq, budget, pool int, spread float64) []string {
	base := RankAtoms(turns, s, query, beforeSeq, pool)
	q := ContentWords(query)
	qs := sortedWords(q) // sums run over the words in a fixed order, so equal gains stay equal
	qsum := 0.0
	for _, w := range qs {
		qsum += x.idf[w]
	}
	if qsum == 0 {
		return readInOrder(base, budget) // none of the question's words is a known stem: nothing to cover, so ranked order
	}
	covered := map[string]float64{}
	blocks := map[int]bool{}
	chosen := map[int]bool{}
	seen := map[string]bool{}
	spent := 0
	var read []string
	for {
		best, bestGain := -1, 0.0
		for i, r := range base {
			if chosen[i] {
				continue
			}
			text := r.Atom.Text(r.Turn.Body)
			if repeated(seen, text) {
				continue
			}
			cost := zoomTokens(text) + zoomReadOverhead
			if spent+cost > budget {
				continue
			}
			gain := 0.0
			st := x.stems(r.Atom.ID(), text)
			for _, w := range qs {
				if c := float64(st[w]); c > 0 {
					gain += x.idf[w] * (math.Sqrt(covered[w]+c) - math.Sqrt(covered[w]))
				}
			}
			if !blocks[r.Turn.Seq/ZoomFanout] {
				gain += spread * qsum
			}
			gain += qsum * 0.5 / float64(rrfK+i+1) * float64(rrfK) // keep the ranking's own order as a prior
			if g := gain / math.Pow(float64(cost), CoverCostPower); g > bestGain {
				best, bestGain = i, g
			}
		}
		if best < 0 {
			return read
		}
		r := base[best]
		text := r.Atom.Text(r.Turn.Body)
		chosen[best] = true
		spent += zoomTokens(text) + zoomReadOverhead
		for w, c := range x.stems(r.Atom.ID(), text) {
			if q[w] {
				covered[w] += float64(c)
			}
		}
		blocks[r.Turn.Seq/ZoomFanout] = true
		read = append(read, text)
		seen[text] = true
	}
}

// NewSpreadIndex prepares the experimental readers for a conversation.
func NewSpreadIndex(turns []vault.Turn) *spreadIndex { return newSpreadIndex(turns) }

// Note is a model-written note the engine has checked: one idea in plain
// words, and the paragraphs its quotes were found in.
type Note struct {
	Title  string
	Kind   string
	Topics []string
	Areas  []string    // the broader areas its topics belong to, if known
	Atoms  []string    // ids of the paragraphs holding its checked quotes
	Quotes []QuoteSpan // where each checked quote is
	Seq    int         // the first turn it points at
}

// QuoteSpan is a checked quote: durable byte span Start to End of a turn's body.
type QuoteSpan struct {
	Turn       string
	BodySHA    string
	Start, End int
	Seg        int
}

// NoteIndex searches notes by their titles and topics.
type NoteIndex struct {
	Notes   []Note
	stems   []map[string]int
	lens    []float64
	avg     float64
	idf     map[string]float64
	byTopic map[string][]int
}

// topicKey folds a topic's wording: its stemmed words, sorted.
func topicKey(topic string) string {
	var ws []string
	for _, tk := range tokens(topic) {
		w := strings.ToLower(strings.Trim(tk.text, ".,:;"))
		if len(w) >= 3 {
			ws = append(ws, stem(w))
		}
	}
	sort.Strings(ws)
	return strings.Join(ws, " ")
}

// NewNoteIndex indexes checked notes.
func NewNoteIndex(notes []Note) *NoteIndex {
	ni := &NoteIndex{Notes: notes, idf: map[string]float64{}, byTopic: map[string][]int{}}
	df := map[string]int{}
	total := 0.0
	for i, n := range notes {
		m := map[string]int{}
		count := 0
		for _, tk := range tokens(n.Title + " " + strings.Join(n.Topics, " ")) {
			w := strings.ToLower(strings.Trim(tk.text, ".,:;"))
			if len(w) < 3 || valueStop[w] {
				continue
			}
			m[stem(w)]++
			count++
		}
		for w := range m {
			df[w]++
		}
		ni.stems, ni.lens = append(ni.stems, m), append(ni.lens, float64(count))
		total += float64(count)
		seen := map[string]bool{}
		for _, tp := range n.Topics {
			if k := topicKey(tp); k != "" && !seen[k] {
				seen[k] = true
				ni.byTopic[k] = append(ni.byTopic[k], i)
			}
		}
	}
	ni.avg = total / math.Max(1, float64(len(notes)))
	for w, c := range df {
		ni.idf[w] = math.Log(1 + (float64(len(notes))-float64(c)+0.5)/(float64(c)+0.5))
	}
	return ni
}

type scoredNote struct {
	i     int
	score float64
}

// Search ranks notes for a question (BM25 over title and topics).
func (ni *NoteIndex) Search(query string) []scoredNote {
	const k1, b = 1.2, 0.75
	q := sortedWords(ContentWords(query))
	var out []scoredNote
	for i, m := range ni.stems {
		s := 0.0
		for _, w := range q {
			if tf := float64(m[w]); tf > 0 {
				s += ni.idf[w] * tf * (k1 + 1) / (tf + k1*(1-b+b*ni.lens[i]/math.Max(ni.avg, 1)))
			}
		}
		if s > 0 {
			out = append(out, scoredNote{i, s})
		}
	}
	sort.SliceStable(out, func(a, c int) bool {
		if out[a].score != out[c].score {
			return out[a].score > out[c].score
		}
		if ni.Notes[out[a].i].Seq != ni.Notes[out[c].i].Seq {
			return ni.Notes[out[a].i].Seq > ni.Notes[out[c].i].Seq
		}
		return out[a].i < out[c].i
	})
	return out
}

// Settings of the topic lane of NotesRead, for `cx zoom` to vary.
var (
	TopicSeeds = 8    // best notes whose topics are followed
	TopicShare = 0.34 // a topic is followed when it has this share of the best topic's weight
)

// NotesRead searches the notes as well as the paragraphs. A note that
// matches the question passes its rank to the paragraphs it is pinned to,
// so a passage whose own words do not match is reached through the plain
// words of a note about it; and, with TopicWeight, the topics of the best
// notes bring in the other notes filed under them, the way a tag page
// gathers a subject's notes. What is read is always the verbatim
// paragraph. The note lanes are fused with ordinary search by rank:
// noteWeight and topicWeight weigh them against the search lane (a topic
// weight of 0 turns that lane off).
func NotesRead(turns []vault.Turn, s *store.Store, x *spreadIndex, ni *NoteIndex, query string, beforeSeq, budget int, noteWeight, topicWeight float64) []string {
	return Texts(NotesPick(turns, s, x, ni, query, beforeSeq, budget, noteWeight, topicWeight, false))
}

// NotesPick is NotesRead, returning the paragraphs picked. With
// neighbours, the paragraphs next to the best hits are read too.
func NotesPick(turns []vault.Turn, s *store.Store, x *spreadIndex, ni *NoteIndex, query string, beforeSeq, budget int, noteWeight, topicWeight float64, neighbours bool) []RankedAtom {
	base := RankAtoms(turns, s, query, beforeSeq, x.n+1)
	byID := map[string]RankedAtom{}
	for _, t := range turns {
		if beforeSeq > 0 && t.Seq >= beforeSeq {
			continue
		}
		for _, a := range atomsOf(t) {
			byID[a.ID()] = RankedAtom{Turn: t, Atom: a}
		}
	}
	score := map[string]float64{}
	keep := map[string]RankedAtom{}
	for i, r := range base {
		score[r.Atom.ID()] = 1 / float64(rrfK+i+1)
		keep[r.Atom.ID()] = r
	}
	lane := func(order []int, weight float64) {
		rank := 0
		seen := map[string]bool{}
		for _, i := range order {
			for _, id := range ni.Notes[i].Atoms {
				r, ok := byID[id]
				if !ok || seen[id] {
					continue
				}
				seen[id] = true
				score[id] += weight / float64(rrfK+rank+1)
				keep[id] = r
				rank++
			}
		}
	}
	order, torder := ni.lanes(query, topicWeight > 0)
	lane(order, noteWeight)
	if topicWeight > 0 {
		lane(torder, topicWeight)
	}
	ranked := sortByScore(keep, score)
	if neighbours {
		ranked = withNeighbours(ranked, NeighbourRadius, NeighbourTop)
	}
	return pickInOrder(ranked, budget)
}

// lanes ranks the notes for a question two ways: by their own match, and
// (with topics) by membership of the topics the best matches are filed
// under, most followed topic first, then own match, then conversation
// order.
func (ni *NoteIndex) lanes(query string, topics bool) (own, byTopic []int) {
	hits := ni.Search(query)
	own = make([]int, len(hits))
	for i, h := range hits {
		own[i] = h.i
	}
	if topics && len(hits) > 0 {
		weight := map[string]float64{}
		for rank, h := range hits[:min(TopicSeeds, len(hits))] {
			seen := map[string]bool{}
			for _, tp := range ni.Notes[h.i].Topics {
				if k := topicKey(tp); k != "" && !seen[k] {
					seen[k] = true
					weight[k] += 1 / float64(rrfK+rank+1)
				}
			}
		}
		best := 0.0
		for _, w := range weight {
			best = math.Max(best, w)
		}
		noteScore := map[int]float64{}
		for _, h := range hits {
			noteScore[h.i] = h.score
		}
		type member struct {
			i     int
			topic float64
			own   float64
		}
		var members []member
		in := map[int]bool{}
		for _, k := range sortedWords(weight) {
			w := weight[k]
			if w < TopicShare*best {
				continue
			}
			for _, i := range ni.byTopic[k] {
				if !in[i] {
					in[i] = true
					members = append(members, member{i, w, noteScore[i]})
				} else {
					for j := range members {
						if members[j].i == i && w > members[j].topic {
							members[j].topic = w
						}
					}
				}
			}
		}
		// Tried and dropped on the development conversation, where neither
		// changed a result: spreading a topic's notes evenly over the
		// timeline, and preferring the kinds of note a question's wording
		// asks for ("how did it change", "what did we try").
		sort.SliceStable(members, func(a, c int) bool {
			if members[a].topic != members[c].topic {
				return members[a].topic > members[c].topic
			}
			if members[a].own != members[c].own {
				return members[a].own > members[c].own
			}
			if ni.Notes[members[a].i].Seq != ni.Notes[members[c].i].Seq {
				return ni.Notes[members[a].i].Seq < ni.Notes[members[c].i].Seq
			}
			return members[a].i < members[c].i
		})
		byTopic = make([]int, len(members))
		for j, m := range members {
			byTopic[j] = m.i
		}
	}
	return own, byTopic
}

// NoteLineTokens is what showing a note costs: its title and its ids.
func NoteLineTokens(n Note) int { return zoomTokens(n.Title) + 10 }

// NoteOrder ranks the notes for a question: the notes that match it and,
// with topicWeight, the other notes of the topics the best matches are filed
// under, fused by rank. Best first; every tie is broken, so the order is the
// same on every call.
func NoteOrder(ni *NoteIndex, query string, topicWeight float64) []int {
	own, byTopic := ni.lanes(query, topicWeight > 0)
	score := map[int]float64{}
	for rank, i := range own {
		score[i] += 1 / float64(rrfK+rank+1)
	}
	for rank, i := range byTopic {
		score[i] += topicWeight / float64(rrfK+rank+1)
	}
	order := make([]int, 0, len(score))
	for i := range score {
		order = append(order, i)
	}
	sort.Slice(order, func(a, b int) bool {
		if score[order[a]] != score[order[b]] {
			return score[order[a]] > score[order[b]]
		}
		return order[a] < order[b]
	})
	return order
}

// NotesView picks the note lines to show for a question within budget, in
// NoteOrder, and returns them in the order the notes were given. A line is
// a plain statement with a checked pointer to the paragraph that says it, at
// a fraction of the paragraph's cost; the paragraph itself comes back by
// recall.
func NotesView(ni *NoteIndex, query string, budget int, topicWeight float64) (shown []int, spent int) {
	for _, i := range NoteOrder(ni, query, topicWeight) {
		if c := NoteLineTokens(ni.Notes[i]); spent+c <= budget {
			spent += c
			shown = append(shown, i)
		}
	}
	sort.Ints(shown)
	return shown, spent
}

// AnswerRank is where plain search ranks the first paragraph of turn that
// holds answer (lower case, single spaces), and the tokens read to reach
// it; rank 0 when the ranking never reaches it.
func AnswerRank(turns []vault.Turn, s *store.Store, x *spreadIndex, query string, beforeSeq int, turn, answer string) (rank, tokens int) {
	sum := 0
	for i, r := range RankAtoms(turns, s, query, beforeSeq, x.n+1) {
		text := r.Atom.Text(r.Turn.Body)
		sum += zoomTokens(text) + zoomReadOverhead
		if r.Turn.ID == turn && strings.Contains(normSpace(text), answer) {
			return i + 1, sum
		}
	}
	return 0, 0
}

// Settings of neighbour expansion, for `cx zoom` to vary.
var (
	NeighbourRadius = 1  // paragraphs read on each side of a hit, within its turn
	NeighbourTop    = 12 // how many of the best hits are expanded
)

// withNeighbours widens a ranked order: after each of the first top hits
// come the paragraphs next to it in its turn, nearest first. The words that
// answer a question are often in the paragraph beside the one that matches
// it (a note points at the right turn far more often than at the right
// paragraph), and reading a neighbour costs what any paragraph costs, so
// the expansion is paid for out of the same budget.
func withNeighbours(order []RankedAtom, radius, top int) []RankedAtom {
	if radius <= 0 || top <= 0 {
		return order
	}
	seen := map[string]bool{}
	var out []RankedAtom
	add := func(r RankedAtom) {
		if id := r.Atom.ID(); !seen[id] {
			seen[id] = true
			out = append(out, r)
		}
	}
	for i, r := range order {
		add(r)
		if i >= top {
			continue
		}
		atoms := atomsOf(r.Turn)
		for d := 1; d <= radius; d++ {
			for _, k := range []int{r.Atom.K - d, r.Atom.K + d} {
				if k >= 1 && k <= len(atoms) {
					add(RankedAtom{Turn: r.Turn, Atom: atoms[k-1], Score: r.Score})
				}
			}
		}
	}
	return out
}

// SearchNeighboursPick is plain search with neighbour expansion.
func SearchNeighboursPick(turns []vault.Turn, s *store.Store, query string, beforeSeq, budget int) []RankedAtom {
	n := 1
	for _, t := range turns {
		n += len(atomsOf(t))
	}
	return pickInOrder(withNeighbours(RankAtoms(turns, s, query, beforeSeq, n), NeighbourRadius, NeighbourTop), budget)
}
