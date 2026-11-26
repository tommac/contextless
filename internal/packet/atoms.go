package packet

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// RankedAtom is one atom (a paragraph-sized piece) of an earlier turn,
// ranked for a query.
type RankedAtom struct {
	Turn  vault.Turn
	Atom  vault.Atom
	Score float64
	// CardOnly marks an atom that stands for its whole turn: the turn
	// matched the query through its card text only, so no paragraph of
	// it is a better pick than another.
	CardOnly bool
}

// atomCacheMax caps cached turn splits so long-lived daemons never leak RAM.
const atomCacheMax = 8192

type boundedAtomCache struct {
	mu    sync.RWMutex
	items map[string][]vault.Atom
	ring  []string
	next  int
}

func (c *boundedAtomCache) get(key string) ([]vault.Atom, bool) {
	c.mu.RLock()
	v, ok := c.items[key]
	c.mu.RUnlock()
	return v, ok
}

func (c *boundedAtomCache) put(key string, val []vault.Atom) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.items == nil {
		c.items = make(map[string][]vault.Atom, 256)
	}
	if _, exists := c.items[key]; exists {
		return
	}
	if len(c.ring) < atomCacheMax {
		c.ring = append(c.ring, key)
	} else {
		delete(c.items, c.ring[c.next])
		c.ring[c.next] = key
		c.next = (c.next + 1) % atomCacheMax
	}
	c.items[key] = val
}

// atomCache keeps each turn's atoms up to atomCacheMax entries: bodies are
// immutable, so the split of (turn id, body sha) never changes.
var atomCache boundedAtomCache

// AtomsOf returns the cached deterministic atom split of turn t.
func AtomsOf(t vault.Turn) []vault.Atom { return atomsOf(t) }

func atomsOf(t vault.Turn) []vault.Atom {
	if t.BodySHA == "" {
		return vault.Atoms(t.ID, t.Body)
	}
	key := t.ID + "\x00" + t.BodySHA
	if v, ok := atomCache.get(key); ok {
		return v
	}
	a := vault.Atoms(t.ID, t.Body)
	atomCache.put(key, a)
	return a
}

// AtomLaneDiag captures per-lane rankings for cx explain and diagnostics.
type AtomLaneDiag struct {
	Exact      []string
	FullText   []string
	Value      []string
	WorkingSet []string
	Episode    []string
	Fused      []RankedAtom
	LaneErrors []string
}

// CandidateTurnIDs uses SQLite FTS5, phrase prefilter, value postings, and
// working-set locality to identify candidate turn IDs before beforeSeq without
// reading every turn file from disk. Any SQLite lane errors are returned
// explicitly so callers never mistake a degraded index for absence.
func CandidateTurnIDs(s *store.Store, query string, beforeSeq, depth int, opt CompileOptions) (turnIDs []string, laneErrs []string) {
	if s == nil || strings.TrimSpace(query) == "" {
		return nil, nil
	}
	if depth <= 0 {
		depth = 48
	}
	seen := map[string]bool{}
	addTurn := func(tid string) {
		if tid != "" && !seen[tid] {
			seen[tid] = true
			turnIDs = append(turnIDs, tid)
		}
	}
	if opt.Lanes.FullText {
		hits, err := s.SearchAtomsE(query, depth, beforeSeq)
		if err != nil {
			laneErrs = append(laneErrs, "full_text: "+err.Error())
		} else {
			for _, h := range hits {
				addTurn(h.Turn)
			}
		}
	}
	if opt.Lanes.Exact {
		if q := normSpace(query); len(q) >= 3 && len(q) <= 200 && len(queryWords(query)) > 0 {
			if s.HasValuePostings() {
				phits, err := s.SearchAtomsPhraseE(q, max(depth*4, 64), beforeSeq)
				if err != nil {
					laneErrs = append(laneErrs, "exact_phrase: "+err.Error())
				} else {
					for _, h := range phits {
						addTurn(h.Turn)
					}
				}
			}
		}
	}
	if opt.Lanes.Value && opt.AtomValueTop > 0 && s.HasValuePostings() {
		if qw := queryWords(query); len(qw) > 0 {
			vhits, err := s.SearchValuePostingsE(qw, beforeSeq)
			if err != nil {
				laneErrs = append(laneErrs, "value_postings: "+err.Error())
			} else {
				for _, vh := range vhits {
					addTurn(vault.TurnID(vh.TurnSeq))
				}
			}
		}
	}
	for _, wid := range opt.WorkingSet {
		if ref, ok := vault.ParseRef(wid); ok {
			addTurn(ref.Turn)
		}
	}
	sort.Strings(turnIDs)
	return turnIDs, laneErrs
}

// RankAtoms ranks the atoms of the turns before beforeSeq for a query,
// best first: the exact-phrase and full-text lanes of RankTurns, read per
// atom and fused the same way, plus the value lane's AtomValueTop best.
// The full-text lane uses the stemmed atom index, which also searches
// each atom with its turn's card text.
func RankAtoms(turns []vault.Turn, s *store.Store, query string, beforeSeq, limit int) []RankedAtom {
	diag := RankAtomsDiag(turns, s, query, beforeSeq, limit, DefaultCompileOptions())
	return diag.Fused
}

// RankAtomsDiag runs atom ranking under explicit CompileOptions and returns
// both the fused results and each contributing lane.
func RankAtomsDiag(turns []vault.Turn, s *store.Store, query string, beforeSeq, limit int, opt CompileOptions) AtomLaneDiag {
	if limit <= 0 || strings.TrimSpace(query) == "" {
		return AtomLaneDiag{}
	}
	byID, exact, fts, value, laneErrs := atomLanes(turns, s, query, beforeSeq, max(2*limit, 48))
	if top := max(opt.AtomValueTop, 0); len(value) > top {
		value = value[:top]
	}
	var working []string
	for _, wid := range opt.WorkingSet {
		if _, ok := byID[wid]; ok {
			working = append(working, wid)
		}
	}
	score := fuseLanesWith(exact, fts, value, working, opt)
	// Interleave episode/session-routed turns as a complementary discovery lane
	// (the +37 LongMemEval / +68 LoCoMo gains are a Python prototype's, not yet
	// reproduced by this code): pick each routed turn's best
	// query-overlapping atom (or first atom) and fuse with RRF weight 0.75.
	var epAtoms []string
	if len(opt.EpisodeTurns) > 0 {
		qw := ContentWords(query)
		byTurnAtoms := map[string][]RankedAtom{}
		for _, r := range byID {
			byTurnAtoms[r.Turn.ID] = append(byTurnAtoms[r.Turn.ID], r)
		}
		for rank, tid := range opt.EpisodeTurns {
			list := byTurnAtoms[tid]
			if len(list) == 0 {
				continue
			}
			best := list[0]
			bestHits := -1
			for _, cand := range list {
				tw := ContentWords(cand.Atom.Text(cand.Turn.Body))
				h := 0
				for w := range qw {
					if tw[w] {
						h++
					}
				}
				if h > bestHits || (h == bestHits && cand.Atom.K < best.Atom.K) {
					bestHits = h
					best = cand
				}
			}
			aid := best.Atom.ID()
			epAtoms = append(epAtoms, aid)
			score[aid] += 0.75 / float64(rrfK+rank+1)
		}
	}
	out := make([]RankedAtom, 0, len(score))
	for id, sc := range score {
		r, ok := byID[id]
		if !ok {
			continue
		}
		r.Score = sc
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Turn.Seq != b.Turn.Seq {
			return a.Turn.Seq > b.Turn.Seq
		}
		return a.Atom.K < b.Atom.K
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return AtomLaneDiag{Exact: exact, FullText: fts, Value: value, WorkingSet: working, Episode: epAtoms, Fused: out, LaneErrors: laneErrs}
}

// AtomValueTop is how many of the value lane's best atoms are fused.
// Among turns the value lane is a full lane. Among thousands of short
// atoms a number near a topic word is common: fused in full, the lane cost
// up to 5 of 30 answers on probes written fresh for this test, because
// each atom it brings in takes budget from the full-text lane's better
// ones. Its first hits are still often right, so only those are kept
// (docs/CONTEXT-COMPILER.md, section 2b). `cx probe -value-top` measures
// other settings.
var AtomValueTop = 2

// atomLanes returns the earlier atoms by id, the three lanes ranked best first
// and cut at depth, and any SQLite lane errors encountered.
func atomLanes(turns []vault.Turn, s *store.Store, query string, beforeSeq, depth int) (byID map[string]RankedAtom, exact, fts, value, laneErrs []string) {
	byID = map[string]RankedAtom{}
	var earlier []RankedAtom
	for _, t := range turns {
		if beforeSeq > 0 && t.Seq >= beforeSeq {
			continue
		}
		for _, a := range atomsOf(t) {
			r := RankedAtom{Turn: t, Atom: a}
			byID[a.ID()] = r
			earlier = append(earlier, r)
		}
	}

	// the exact lane: the message as a phrase, at word boundaries. A
	// message with no content word ("yes", "the") matches nothing: among
	// thousands of atoms it would match inside other words everywhere.
	// When an FTS5 atom index is available, use phrase prefiltering so we
	// avoid scanning every atom in the corpus.
	if q := normSpace(query); len(q) >= 3 && len(q) <= 200 && len(queryWords(query)) > 0 {
		usedIndex := false
		if s != nil && s.HasValuePostings() {
			phits, err := s.SearchAtomsPhraseE(q, max(depth*4, 64), beforeSeq)
			if err != nil {
				laneErrs = append(laneErrs, "exact_phrase: "+err.Error())
			} else {
				usedIndex = true
				for _, h := range phits {
					if r, ok := byID[h.ID]; ok && containsPhrase(normSpace(r.Atom.Text(r.Turn.Body)), q) {
						exact = append(exact, h.ID)
						if len(exact) >= depth {
							break
						}
					}
				}
			}
		}
		if !usedIndex {
			for i := len(earlier) - 1; i >= 0 && len(exact) < depth; i-- {
				if containsPhrase(normSpace(earlier[i].Atom.Text(earlier[i].Turn.Body)), q) {
					exact = append(exact, earlier[i].Atom.ID())
				}
			}
		}
	}

	// the full-text lane. An atom that matched only through its turn's
	// card text says the turn is relevant, not which paragraph: every
	// atom of the turn matches alike. Where the turn also has a paragraph
	// that matches in its own words, those paragraphs stand; otherwise
	// the turn's first hit stands for the whole turn.
	if s != nil {
		hits, err := s.SearchAtomsE(query, depth, beforeSeq)
		if err != nil {
			laneErrs = append(laneErrs, "full_text: "+err.Error())
		}
		own := map[string]bool{}
		for _, h := range hits {
			if !h.CardOnly {
				own[h.Turn] = true
			}
		}
		stood := map[string]bool{}
		for _, h := range hits {
			r, ok := byID[h.ID]
			if !ok {
				continue
			}
			if h.CardOnly {
				if own[h.Turn] || stood[h.Turn] {
					continue
				}
				stood[h.Turn] = true
				r.CardOnly = true
				byID[h.ID] = r
			}
			fts = append(fts, h.ID)
		}
	}

	if qw := queryWords(query); len(qw) > 0 {
		type vc struct {
			id       string
			seq, k   int
			best, at int
		}
		ftsRank := map[string]int{}
		for i, id := range fts {
			ftsRank[id] = i + 1
		}
		var hits []vc
		usedPostings := false
		if s != nil && s.HasValuePostings() {
			phits, err := s.SearchValuePostingsE(qw, beforeSeq)
			if err != nil {
				laneErrs = append(laneErrs, "value_postings: "+err.Error())
			} else {
				usedPostings = true
				for _, ph := range phits {
					id := vault.AtomID(vault.TurnID(ph.TurnSeq), ph.AtomK)
					if _, ok := byID[id]; !ok {
						continue
					}
					at, ok := ftsRank[id]
					if !ok {
						at = len(fts) + 1
					}
					hits = append(hits, vc{id, ph.TurnSeq, ph.AtomK, ph.Score, at})
				}
			}
		}
		if !usedPostings {
			for _, r := range earlier {
				if b := valueScore(r.Atom.Text(r.Turn.Body), qw); b > 0 {
					id := r.Atom.ID()
					at, ok := ftsRank[id]
					if !ok {
						at = len(fts) + 1
					}
					hits = append(hits, vc{id, r.Turn.Seq, r.Atom.K, b, at})
				}
			}
		}
		// among equally direct statements, the full-text lane decides,
		// then the newer turn
		sort.SliceStable(hits, func(i, j int) bool {
			a, b := hits[i], hits[j]
			if a.best != b.best {
				return a.best > b.best
			}
			if a.at != b.at {
				return a.at < b.at
			}
			if a.seq != b.seq {
				return a.seq > b.seq
			}
			return a.k < b.k
		})
		for i := 0; i < len(hits) && i < depth; i++ {
			value = append(value, hits[i].id)
		}
	}
	return byID, exact, fts, value, laneErrs
}

// containsPhrase reports whether phrase occurs in text with no letter or
// digit directly before or after it.
func containsPhrase(text, phrase string) bool {
	for from := 0; ; {
		i := strings.Index(text[from:], phrase)
		if i < 0 {
			return false
		}
		i += from
		before, _ := utf8.DecodeLastRuneInString(text[:i])
		after, _ := utf8.DecodeRuneInString(text[i+len(phrase):])
		if !wordRune(before) && !wordRune(after) {
			return true
		}
		from = i + 1
	}
}

func wordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// RankAtomPositions reports where the first atom of turn id that holds
// answer (lower case, single spaces; "" for any atom) sits for query in
// each lane and in the fused ranking of every earlier atom, to explain a
// probe whose source did not make the packet. A card-only match stands
// for its whole turn. The value position is the lane's own; only its
// AtomValueTop best are fused.
func RankAtomPositions(turns []vault.Turn, s *store.Store, query string, beforeSeq int, id, answer string) Positions {
	var p Positions
	n := 1
	for _, t := range turns {
		n += len(atomsOf(t))
	}
	byID, exact, fts, value, _ := atomLanes(turns, s, query, beforeSeq, n)
	holds := func(r RankedAtom) bool {
		return r.Turn.ID == id && (answer == "" || r.CardOnly || strings.Contains(normSpace(r.Atom.Text(r.Turn.Body)), answer))
	}
	at := func(list []string) int {
		for i, x := range list {
			if holds(byID[x]) {
				return i + 1
			}
		}
		return 0
	}
	p.Exact, p.FullText, p.Value = at(exact), at(fts), at(value)
	for i, r := range RankAtoms(turns, s, query, beforeSeq, n) {
		if holds(r) {
			p.Fused = i + 1
			break
		}
	}
	return p
}

// TokensToAnswer is how many tokens of ranked text a reader goes through,
// in rank order, to reach the source turn's text holding answer (lower
// case, single spaces); 0 when the ranking never reaches it. It says how
// large a budget this ranking needs for this question, whatever the budget
// in force. With atoms is false it walks whole turns.
func TokensToAnswer(turns []vault.Turn, s *store.Store, query string, beforeSeq int, source, answer string, atoms bool) int {
	sum := 0
	if !atoms {
		for _, r := range RankTurns(turns, s, query, beforeSeq, len(turns)) {
			sum += Tokens(providers.Profile{}, len(r.Turn.Body))
			if r.Turn.ID == source && strings.Contains(normSpace(r.Turn.Body), answer) {
				return sum
			}
		}
		return 0
	}
	n := 1
	for _, t := range turns {
		n += len(atomsOf(t))
	}
	for _, r := range RankAtoms(turns, s, query, beforeSeq, n) {
		text := r.Atom.Text(r.Turn.Body)
		sum += Tokens(providers.Profile{}, len(text))
		if r.Turn.ID == source && strings.Contains(normSpace(text), answer) {
			return sum
		}
	}
	return 0
}

// atomCandidates is the least number of ranked atoms selection looks at
// to fill its budget; a larger budget looks at more.
const atomCandidates = 200

// atomRepeatBytes is the size from which a paragraph already chosen is
// not chosen again from another turn (re-pasted text, a recurring notice).
const atomRepeatBytes = 80

// AtomRun is a run of adjacent atoms of one turn chosen for a packet.
type AtomRun struct {
	Turn  vault.Turn
	A, B  int // first and last atom, 1-based
	N     int // atoms in the turn
	Score float64
}

// ID is the run's recall id: the turn id when the run is the whole turn.
func (r AtomRun) ID() string {
	if r.A == 1 && r.B == r.N {
		return r.Turn.ID
	}
	return vault.AtomRangeID(r.Turn.ID, r.A, r.B)
}

// PickAtoms fills a token budget with the best-ranked atoms for query and
// joins adjacent picks of a turn into runs. The budget counts what is
// sent: each atom's text and the header of each run. have reports an atom
// that is already in the packet (its turn is in whole, or a recalled
// excerpt or part holds it): it costs nothing and is not picked again,
// and its turn is reported in hit with its best score while the atoms
// ranked above it would themselves fit the budget. A turn that matched
// only through its card goes in whole when it fits what is left of the
// budget, and is left to its catalog line when it does not. A paragraph
// whose exact text was already picked from another turn is skipped.
func PickAtoms(turns []vault.Turn, s *store.Store, query string, beforeSeq, budget int, have func(turn string, k int) bool) (runs []AtomRun, hit map[string]float64) {
	return pickAtomsWith(turns, s, query, beforeSeq, budget, have, DefaultCompileOptions())
}

func pickAtomsWith(turns []vault.Turn, s *store.Store, query string, beforeSeq, budget int, have func(turn string, k int) bool, opt CompileOptions) (runs []AtomRun, hit map[string]float64) {
	hit = map[string]float64{}
	chosen := map[string]map[int]float64{}
	byID := map[string]vault.Turn{}
	for _, t := range turns {
		byID[t.ID] = t
	}
	tokens := func(n int) int { return Tokens(providers.Profile{}, n) }
	headerCost := func(t vault.Turn) int {
		var replyTo *vault.Turn
		if q, ok := byID[t.ReplyTo]; ok {
			replyTo = &q
		}
		// the header line, its markup and a range in place of one number
		return tokens(len(AtomHeader(t, 1, 1, len(atomsOf(t)), replyTo)) + 24)
	}
	// take adds atom k of t and returns what it adds to the packet: its
	// text, plus a header when it starts a run, less one when it joins two
	cost := func(t vault.Turn, a vault.Atom) int {
		c := tokens(len(a.Text(t.Body)) + 2)
		_, left := chosen[t.ID][a.K-1]
		_, right := chosen[t.ID][a.K+1]
		switch {
		case !left && !right:
			c += headerCost(t)
		case left && right:
			c -= headerCost(t)
		}
		return c
	}
	take := func(t vault.Turn, a vault.Atom, score float64) {
		if chosen[t.ID] == nil {
			chosen[t.ID] = map[int]float64{}
		}
		chosen[t.ID][a.K] = score
	}
	seen := map[string]bool{}
	used, offered := 0, 0
	for _, r := range RankAtomsDiag(turns, s, query, beforeSeq, max(atomCandidates, budget/40), opt).Fused {
		t := r.Turn
		if r.CardOnly {
			// the whole turn or nothing
			var rest []vault.Atom
			for _, a := range atomsOf(t) {
				if _, ok := chosen[t.ID][a.K]; !ok && (have == nil || !have(t.ID, a.K)) {
					rest = append(rest, a)
				}
			}
			if len(rest) == 0 {
				if offered <= budget && r.Score > hit[t.ID] {
					hit[t.ID] = r.Score
				}
				continue
			}
			whole := tokens(len(t.Body)) + headerCost(t)
			offered += whole
			if used+whole > budget {
				continue
			}
			for _, a := range rest {
				used += cost(t, a)
				take(t, a, r.Score)
			}
			continue
		}
		text := r.Atom.Text(t.Body)
		offered += tokens(len(text))
		if have != nil && have(t.ID, r.Atom.K) {
			if offered <= budget && r.Score > hit[t.ID] {
				hit[t.ID] = r.Score
			}
			continue
		}
		if _, ok := chosen[t.ID][r.Atom.K]; ok {
			continue
		}
		if len(text) >= atomRepeatBytes && seen[text] {
			continue
		}
		c := cost(t, r.Atom)
		if used+c > budget {
			continue // a smaller atom further down may still fit
		}
		used += c
		take(t, r.Atom, r.Score)
		seen[text] = true
	}
	for id, ks := range chosen {
		t := byID[id]
		n := len(atomsOf(t))
		for _, x := range JoinAtoms(ks) {
			run := AtomRun{Turn: t, A: x[0], B: x[1], N: n}
			for k := x[0]; k <= x[1]; k++ {
				run.Score = max(run.Score, ks[k])
			}
			runs = append(runs, run)
		}
	}
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Turn.Seq != runs[j].Turn.Seq {
			return runs[i].Turn.Seq < runs[j].Turn.Seq
		}
		return runs[i].A < runs[j].A
	})
	return runs, hit
}

// JoinAtoms turns a set of atom numbers of one turn into runs of adjacent
// atoms, in order: the one rule for joining, shared by selection, recall
// and recall search.
func JoinAtoms[V any](ks map[int]V) [][2]int {
	order := make([]int, 0, len(ks))
	for k := range ks {
		order = append(order, k)
	}
	sort.Ints(order)
	var out [][2]int
	for _, k := range order {
		if n := len(out); n > 0 && k == out[n-1][1]+1 {
			out[n-1][1] = k
			continue
		}
		out = append(out, [2]int{k, k})
	}
	return out
}

// AtomHeader is the header of an excerpt: the turn, which of its
// paragraphs these are, their recall id and, for a reply, the opening of
// the message it answers, so a paragraph is not read without its
// question.
func AtomHeader(t vault.Turn, a, b, n int, replyTo *vault.Turn) string {
	which := fmt.Sprintf("paragraph %d", a)
	if b > a {
		which = fmt.Sprintf("paragraphs %d to %d", a, b)
	}
	h := fmt.Sprintf("%s · excerpt: %s of %d (%s)", TurnHeader(t), which, n, vault.AtomRangeID(t.ID, a, b))
	if replyTo != nil {
		h += fmt.Sprintf(" · in reply to %s: \"%s\"", replyTo.ID, oneLine(replyTo.Body, 100))
	}
	return h
}
