package packet

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// episodeGap groups turns into conversational sessions when thread cards are
// absent (30-minute idle boundary, matching natural multi-session histories).
const episodeGap = 30 * time.Minute

// RouteEpisodeTurns implements the empirically validated two-stage session /
// thread candidate router from reports/2026-11-20-large-memory/fixed-window-results.md:
//
//  1. Group earlier turns into episodes by primary thread (when indexed) or
//     by 30-minute conversational session boundaries (fallback: 20-turn windows).
//  2. Build each episode's lexical bag over text/previews, card keywords,
//     author/speaker names, and ISO dates (YYYY-MM-DD).
//  3. Score episodes by IDF-weighted query-term overlap and pick the top
//     topEpisodes (default 3) episodes.
//  4. Rank turns inside those episodes and return up to maxTurns turn IDs so
//     the caller can interleave them with primary atom candidates before packing.
//
// Works in both full-turn mode and metadata-only lazy mode (when
// only metas + cards are in memory).
func RouteEpisodeTurns(turns []vault.Turn, metas []store.TurnMeta, cards map[string]vault.Card, query string, beforeSeq, topEpisodes, maxTurns int) []string {
	qw := ContentWords(query)
	// Also include raw lowercased short tokens (speaker names, YYYY-MM-DD dates)
	// so speaker and date queries match episodes cleanly.
	extraToks := map[string]bool{}
	for _, tok := range tokens(query) {
		low := strings.ToLower(tok.text)
		if len(low) >= 2 && !valueStop[low] {
			extraToks[low] = true
		}
	}
	if len(qw) == 0 && len(extraToks) == 0 {
		return nil
	}
	if topEpisodes <= 0 {
		topEpisodes = 3
	}
	if maxTurns <= 0 {
		maxTurns = 12
	}

	type turnStub struct {
		id      string
		seq     int
		author  string
		created time.Time
		text    string
	}
	var stubs []turnStub
	if len(metas) > 0 {
		for _, m := range metas {
			if beforeSeq > 0 && m.Seq >= beforeSeq {
				continue
			}
			txt := m.Preview + " " + strings.Join(m.TopTerms, " ")
			if c, ok := cards[m.ID]; ok {
				txt += " " + c.Summary + " " + strings.Join(c.Keywords, " ")
			}
			stubs = append(stubs, turnStub{id: m.ID, seq: m.Seq, author: m.Author, created: m.Created, text: txt})
		}
	} else {
		for _, t := range turns {
			if beforeSeq > 0 && t.Seq >= beforeSeq {
				continue
			}
			txt := t.Body
			if c, ok := cards[t.ID]; ok {
				txt += " " + c.Summary + " " + strings.Join(c.Keywords, " ")
			}
			stubs = append(stubs, turnStub{id: t.ID, seq: t.Seq, author: t.Author, created: t.Created, text: txt})
		}
	}
	if len(stubs) == 0 {
		return nil
	}
	sort.Slice(stubs, func(i, j int) bool { return stubs[i].seq < stubs[j].seq })

	// Group turns into episodes: primary thread if present, else temporal session.
	type episode struct {
		key   string
		turns []turnStub
		terms map[string]bool
		score float64
	}
	epIndex := map[string]int{}
	var episodes []*episode
	sessionIdx := 0
	var lastTime time.Time
	turnsInSess := 0

	addToEp := func(key string, st turnStub) {
		idx, ok := epIndex[key]
		if !ok {
			idx = len(episodes)
			epIndex[key] = idx
			episodes = append(episodes, &episode{key: key, terms: map[string]bool{}})
		}
		ep := episodes[idx]
		ep.turns = append(ep.turns, st)
		for w := range ContentWords(st.text) {
			ep.terms[w] = true
		}
		for _, tok := range tokens(st.text) {
			low := strings.ToLower(tok.text)
			if extraToks[low] {
				ep.terms[low] = true
			}
		}
		if st.author != "" {
			ep.terms[strings.ToLower(st.author)] = true
		}
		if !st.created.IsZero() {
			ep.terms[st.created.Format("2006-01-02")] = true
		}
	}

	for _, st := range stubs {
		if c, ok := cards[st.id]; ok && len(c.Threads) > 0 {
			addToEp("th:"+c.Threads[0], st)
			continue
		}
		newSess := false
		if turnsInSess == 0 {
			newSess = true
		} else if !lastTime.IsZero() && !st.created.IsZero() && st.created.Sub(lastTime) >= episodeGap {
			newSess = true
		} else if turnsInSess >= 20 {
			newSess = true
		}
		if newSess {
			sessionIdx++
			turnsInSess = 0
		}
		turnsInSess++
		if !st.created.IsZero() {
			lastTime = st.created
		}
		addToEp(fmt.Sprintf("sess:%d", sessionIdx), st)
	}
	if len(episodes) == 0 {
		return nil
	}

	// IDF over episodes.
	allQuery := map[string]bool{}
	for w := range qw {
		allQuery[w] = true
	}
	for w := range extraToks {
		allQuery[w] = true
	}
	df := map[string]int{}
	for _, ep := range episodes {
		for w := range allQuery {
			if ep.terms[w] {
				df[w]++
			}
		}
	}
	idf := map[string]float64{}
	for w := range allQuery {
		idf[w] = 1 + math.Log(float64(1+len(episodes))/float64(1+df[w]))
	}
	for _, ep := range episodes {
		sc := 0.0
		for w, wt := range idf {
			if ep.terms[w] {
				sc += wt
			}
		}
		ep.score = sc
	}
	sort.SliceStable(episodes, func(i, j int) bool {
		if episodes[i].score != episodes[j].score {
			return episodes[i].score > episodes[j].score
		}
		return episodes[i].turns[len(episodes[i].turns)-1].seq > episodes[j].turns[len(episodes[j].turns)-1].seq
	})

	// Pull highest-scoring turns from the top episodes in round-robin order.
	type scoredTurn struct {
		id    string
		seq   int
		score float64
	}
	var perEp [][]scoredTurn
	for i := 0; i < len(episodes) && i < topEpisodes; i++ {
		if episodes[i].score <= 0 {
			break
		}
		var list []scoredTurn
		for _, st := range episodes[i].turns {
			tw := ContentWords(st.text)
			sc := 0.0
			for w, wt := range idf {
				if tw[w] || strings.Contains(strings.ToLower(st.text), w) {
					sc += wt
				}
			}
			if sc > 0 {
				list = append(list, scoredTurn{id: st.id, seq: st.seq, score: sc})
			}
		}
		sort.SliceStable(list, func(a, b int) bool {
			if list[a].score != list[b].score {
				return list[a].score > list[b].score
			}
			return list[a].seq > list[b].seq
		})
		if len(list) > 0 {
			perEp = append(perEp, list)
		}
	}

	var out []string
	seen := map[string]bool{}
	for pass := 0; len(out) < maxTurns; pass++ {
		added := false
		for _, list := range perEp {
			if pass < len(list) {
				id := list[pass].id
				if !seen[id] {
					seen[id] = true
					out = append(out, id)
					if len(out) >= maxTurns {
						break
					}
				}
				added = true
			}
		}
		if !added {
			break
		}
	}
	return out
}

// NeedsReplyWindow reports whether turn t looks like a question or split-Q&A
// anchor whose actual answer likely lives in the immediately following turn
// (seq+1). Pairing such anchors with their immediate reply closed 4/4 speaker-
// attribution and split-Q&A failures on LoCoMo while rejecting 4/4 speaker-
// swapped lures.
func NeedsReplyWindow(t vault.Turn) bool {
	body := strings.TrimSpace(t.Body)
	if body == "" || len(body) > 1200 {
		return false
	}
	// Find the last non-empty line.
	lines := strings.Split(body, "\n")
	last := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s != "" {
			last = s
			break
		}
	}
	if strings.HasSuffix(last, "?") {
		return true
	}
	// Short user prompts asking a question followed by a short parenthetical or quote.
	if t.Role == vault.RoleUser && strings.Contains(body, "?") && len(body) <= 320 {
		return true
	}
	return false
}
