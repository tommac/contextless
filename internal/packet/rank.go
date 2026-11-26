package packet

import (
	"sort"
	"strings"
	"unicode"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// Ranked is one earlier turn ranked for a query.
type Ranked struct {
	Turn  vault.Turn
	Score float64
}

// Lanes switches the ranking lanes; all are on. It exists so the probe
// measurement can compare lanes on the same conversation.
type Lanes struct{ Exact, FullText, Value bool }

// LaneWeights tunes convex/RRF fusion weights across lanes. Zero values fall
// back to equal weight (1.0) for backward compatibility.
type LaneWeights struct {
	Exact      float64 `json:"exact" yaml:"exact"`
	FullText   float64 `json:"full_text" yaml:"full_text"`
	Value      float64 `json:"value" yaml:"value"`
	WorkingSet float64 `json:"working_set" yaml:"working_set"`
}

// CompileOptions captures packet ranking and layout knobs as an immutable
// per-call value instead of mutating package globals.
type CompileOptions struct {
	Lanes           Lanes
	Weights         LaneWeights
	AtomValueTop    int
	CatalogSpine    bool
	WorkingSet      []string // recently recalled or verified-cited atom/turn ids
	EpisodeTurns    []string // episode/session-routed turn ids interleaved when SessionRoute is on
	AdaptivePack    bool     // source-diverse anchor-outward excerpt growth on overflow
	SessionRoute    bool     // two-stage episode/thread candidate routing interleaved with atoms
	AutoReplyWindow bool     // co-select immediate reply turn (seq+1) for question-ending anchors
}

// DefaultCompileOptions snapshots the current package defaults into a safe
// per-call struct.
func DefaultCompileOptions() CompileOptions {
	return CompileOptions{
		Lanes:        RankLanes,
		Weights:      LaneWeights{Exact: 1, FullText: 1, Value: 1, WorkingSet: 0.5},
		AtomValueTop: AtomValueTop,
		CatalogSpine: CatalogSpine,
	}
}

// RankLanes are the default lanes RankTurns fuses. Prefer passing
// CompileOptions on new callers to avoid global state.
var RankLanes = Lanes{Exact: true, FullText: true, Value: true}

// rrfK is the Reciprocal Rank Fusion constant. LANTERN reports recovery
// flat across k from 10 to 200, so the usual 60 is kept.
const rrfK = 60

// RankTurns ranks the turns before beforeSeq for a query, best first, by
// fusing three ranked lanes with Reciprocal Rank Fusion, with each lane's
// top hit placed first. No model is called and nothing is embedded; every
// lane reads the verbatim turns.
//
//   - exact: the whole query occurs in the turn (case and whitespace
//     aside), newest first, so a value that changed is found before the
//     value it replaced.
//   - full text: bm25 over every word (the store's FTS index).
//   - value: the turn states a value (a number, a date or time, an
//     amount, a code identifier or path) within a few words of a word of
//     the query. Config values, amounts and dates are the facts most often
//     lost in LANTERN's evaluation: their questions share only topic words
//     with the turn that answers them, and so does every other turn on
//     the topic. A topic word right next to a value is the signal that
//     separates the answer from the talk around it.
func RankTurns(turns []vault.Turn, s *store.Store, query string, beforeSeq, limit int) []Ranked {
	if limit <= 0 || strings.TrimSpace(query) == "" {
		return nil
	}
	depth := max(4*limit, 24)
	byID, exact, fts, value := rankLanes(turns, s, query, beforeSeq, depth)

	score := fuseLanes(exact, fts, value)
	out := make([]Ranked, 0, len(score))
	for id, sc := range score {
		out = append(out, Ranked{Turn: byID[id], Score: sc})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Turn.Seq > out[j].Turn.Seq
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// fuseLanes scores ids by Reciprocal Rank Fusion over the lanes that are
// switched on.
func fuseLanes(exact, fts, value []string) map[string]float64 {
	return fuseLanesWith(exact, fts, value, nil, DefaultCompileOptions())
}

func fuseLanesWith(exact, fts, value, working []string, opt CompileOptions) map[string]float64 {
	if !opt.Lanes.Exact {
		exact = nil
	}
	if !opt.Lanes.FullText {
		fts = nil
	}
	if !opt.Lanes.Value {
		value = nil
	}
	wExact, wFTS, wVal, wWork := opt.Weights.Exact, opt.Weights.FullText, opt.Weights.Value, opt.Weights.WorkingSet
	if wExact <= 0 {
		wExact = 1
	}
	if wFTS <= 0 {
		wFTS = 1
	}
	if wVal <= 0 {
		wVal = 1
	}
	if wWork <= 0 {
		wWork = 0.5
	}
	score := map[string]float64{}
	type weightedLane struct {
		ids []string
		w   float64
		pin bool
	}
	for _, wl := range []weightedLane{
		{exact, wExact, true},
		{fts, wFTS, true},
		{value, wVal, true},
		{working, wWork, false},
	} {
		for rank, id := range wl.ids {
			score[id] += wl.w / float64(rrfK+rank+1)
		}
	}
	// each primary lane's best hit always makes the cut: fusion alone favors a
	// record that is middling in two lanes over the clear winner of one
	for _, lane := range [][]string{exact, value, fts} {
		if len(lane) > 0 && score[lane[0]] < 1 {
			score[lane[0]] += 1
		}
	}
	return score
}

// rankLanes returns the earlier turns by id and the three lanes, each
// ranked best first and cut at depth.
func rankLanes(turns []vault.Turn, s *store.Store, query string, beforeSeq, depth int) (byID map[string]vault.Turn, exact, fts, value []string) {
	byID = map[string]vault.Turn{}
	var earlier []vault.Turn
	for _, t := range turns {
		if beforeSeq > 0 && t.Seq >= beforeSeq {
			continue
		}
		byID[t.ID] = t
		earlier = append(earlier, t)
	}
	if q := normSpace(query); len(q) >= 3 && len(q) <= 200 {
		for i := len(earlier) - 1; i >= 0; i-- {
			if strings.Contains(normSpace(earlier[i].Body), q) {
				exact = append(exact, earlier[i].ID)
				if len(exact) == depth {
					break
				}
			}
		}
	}

	if s != nil {
		for _, h := range s.Search(query, depth, vault.TypeTurn) {
			if _, ok := byID[h.ID]; ok {
				fts = append(fts, h.ID)
			}
		}
	}

	if qw := queryWords(query); len(qw) > 0 {
		type vc struct {
			id   string
			seq  int
			best int
		}
		var hits []vc
		for _, t := range earlier {
			if b := valueScore(t.Body, qw); b > 0 {
				hits = append(hits, vc{t.ID, t.Seq, b})
			}
		}
		ftsRank := map[string]int{}
		for i, id := range fts {
			ftsRank[id] = i + 1
		}
		rankOf := func(id string) int {
			if r, ok := ftsRank[id]; ok {
				return r
			}
			return len(fts) + 1
		}
		// among equally direct statements, the full-text lane decides,
		// then the newer turn
		sort.SliceStable(hits, func(i, j int) bool {
			a, b := hits[i], hits[j]
			if a.best != b.best {
				return a.best > b.best
			}
			if ra, rb := rankOf(a.id), rankOf(b.id); ra != rb {
				return ra < rb
			}
			return a.seq > b.seq
		})
		for i := 0; i < len(hits) && i < depth; i++ {
			value = append(value, hits[i].id)
		}
	}

	return byID, exact, fts, value
}

// Positions is where one turn sits (1 is first, 0 is absent) in each lane
// and in the fused ranking of every earlier turn.
type Positions struct{ Fused, Exact, FullText, Value int }

// RankPositions reports where turn id sits for query, to explain a probe
// whose source did not make the packet.
func RankPositions(turns []vault.Turn, s *store.Store, query string, beforeSeq int, id string) Positions {
	var p Positions
	at := func(list []string) int {
		for i, x := range list {
			if x == id {
				return i + 1
			}
		}
		return 0
	}
	n := len(turns) + 1
	_, exact, fts, value := rankLanes(turns, s, query, beforeSeq, n)
	p.Exact, p.FullText, p.Value = at(exact), at(fts), at(value)
	var fused []string
	for _, r := range RankTurns(turns, s, query, beforeSeq, n) {
		fused = append(fused, r.Turn.ID)
	}
	p.Fused = at(fused)
	return p
}

func normSpace(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

// valueStop are words too common in questions to tie a value to a topic.
var valueStop = map[string]bool{
	"the": true, "and": true, "for": true, "that": true, "this": true, "with": true, "you": true, "are": true,
	"was": true, "were": true, "did": true, "does": true, "what": true, "which": true, "when": true, "where": true,
	"who": true, "whom": true, "how": true, "why": true, "now": true, "our": true, "use": true, "uses": true,
	"has": true, "have": true, "had": true, "not": true, "from": true, "into": true, "about": true, "there": true,
	"they": true, "them": true, "set": true, "get": true, "got": true, "still": true, "much": true, "many": true,
	"originally": true, "again": true, "should": true, "would": true, "could": true, "will": true,
}

// stem folds a few English inflections so "leaves" meets "leave" and
// "cabinets" meets "cabinet": a plural or tense ending comes off, then a
// final "e".
func stem(w string) string {
	for _, suf := range []string{"ies", "es", "s", "ed", "ing"} {
		if len(w) > len(suf)+3 && strings.HasSuffix(w, suf) {
			if suf == "ies" {
				w = w[:len(w)-3] + "y"
			} else {
				w = w[:len(w)-len(suf)]
			}
			break
		}
	}
	if len(w) > 4 && strings.HasSuffix(w, "e") {
		w = w[:len(w)-1]
	}
	return w
}

func queryWords(q string) map[string]bool {
	out := map[string]bool{}
	for _, tok := range tokens(q) {
		w := strings.ToLower(tok.text)
		if len(w) < 3 || valueStop[w] {
			continue
		}
		out[stem(w)] = true
	}
	return out
}

// token is one word of a turn, with its sentence number and whether it
// opens the sentence.
type token struct {
	text      string
	sent      int
	sentStart bool
}

// tokens splits text at spaces and at punctuation that cannot be inside
// a value (keeping $ , . : _ / - and # inside a token), then trims
// punctuation from both ends.
func tokens(text string) []token {
	raw := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("()[]{}<>\"'`;!?|*=", r)
	})
	out := make([]token, 0, len(raw))
	start, sent := true, 0
	for _, t := range raw {
		end := strings.HasSuffix(t, ".") || strings.HasSuffix(t, ":") || strings.HasSuffix(t, "\n")
		t = strings.Trim(t, ".,:;-_/#")
		if t != "" {
			out = append(out, token{t, sent, start})
			start = false
		}
		if end {
			start = true
			sent++
		}
	}
	return out
}

// isValue reports whether a token states a value: it has a digit, it is
// a code identifier (camelCase, snake_case) or a path, or it is a
// capitalized word inside a sentence (a name).
func isValue(tok token) bool {
	hasLetter, hasDigit, lowerThenUpper := false, false, false
	var prev rune
	for i, r := range tok.text {
		switch {
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsLetter(r):
			hasLetter = true
			if i > 0 && unicode.IsUpper(r) && unicode.IsLower(prev) {
				lowerThenUpper = true
			}
		}
		prev = r
	}
	if hasDigit {
		return true
	}
	if !hasLetter {
		return false
	}
	if lowerThenUpper || strings.ContainsAny(tok.text, "_/") {
		return true
	}
	first, _ := utf8DecodeFirst(tok.text)
	return !tok.sentStart && unicode.IsUpper(first) && len(tok.text) > 1
}

func utf8DecodeFirst(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}

// valueScore rates how directly body states a value about the query: the
// most distinct query words within four words of any one value in the
// same sentence. One statement decides, not the turn's length.
func valueScore(body string, qw map[string]bool) (best int) {
	toks := tokens(body)
	for i, t := range toks {
		// a word the question already holds is its topic, not the
		// value being asked for
		if !isValue(t) || qw[stem(strings.ToLower(t.text))] {
			continue
		}
		seen := map[string]bool{}
		for j := max(0, i-4); j <= min(len(toks)-1, i+4); j++ {
			if toks[j].sent != t.sent || j == i {
				continue
			}
			if w := stem(strings.ToLower(toks[j].text)); qw[w] {
				seen[w] = true
			}
		}
		best = max(best, len(seen))
	}
	return best
}

// Values are the values a text states (numbers, dates, times, amounts,
// identifiers, paths, names), as the value lane reads them, lower case and
// each once, in order.
func Values(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tokens(text) {
		if !isValue(t) {
			continue
		}
		v := strings.ToLower(strings.Trim(t.text, ".,:;"))
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// ContentWords are the stemmed words of a text that carry its topic, the
// words a query is matched on.
func ContentWords(text string) map[string]bool { return queryWords(text) }

// ExtractValuePostings builds the inverted value-occurrence postings for one
// atom, matching valueScore's exact 4-token same-sentence neighborhood rule.
// Persisting these postings in SQLite removes full-corpus tokenization on
// every query while preserving 100% score equivalence.
func ExtractValuePostings(turnSeq, atomK int, text string) []store.ValuePosting {
	toks := tokens(text)
	var out []store.ValuePosting
	occID := 0
	for i, t := range toks {
		if !isValue(t) {
			continue
		}
		occID++
		valStem := stem(strings.ToLower(t.text))
		seen := map[string]bool{}
		for j := max(0, i-4); j <= min(len(toks)-1, i+4); j++ {
			if toks[j].sent != t.sent || j == i {
				continue
			}
			w := strings.ToLower(toks[j].text)
			if len(w) < 3 || valueStop[w] {
				continue
			}
			st := stem(w)
			if st == "" || seen[st] {
				continue
			}
			seen[st] = true
			out = append(out, store.ValuePosting{
				Term:    st,
				TurnSeq: turnSeq,
				AtomK:   atomK,
				OccID:   occID,
				ValStem: valStem,
			})
		}
	}
	return out
}
