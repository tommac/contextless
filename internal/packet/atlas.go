package packet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// The map of a conversation ("atlas") is the overview the answer test of
// 2026-11-25 favoured: every checked idea note as one line that names the
// paragraphs holding its quotes, read by the model, which then asks for the
// paragraphs it needs by id. In that test (30 held-out questions, two blind
// judges) a model that read the note lines and asked once for paragraphs
// stated 0.73 of the needed passages against 0.55 for search filled to the
// same size, and 0.84 for the whole conversation at eight times the tokens.
// A code-built line per turn in place of the notes did no better than
// search. A note line alone was a poor substitute for the paragraph (about
// a third of pointed-at passages were stated without being read), which is
// why the map is a pointer and the section says so.
//
// The map is what changes least in a packet after the rules: notes are
// appended as turns are indexed and never rewritten, so while the map fits
// its budget each new batch only extends it, and a provider's prompt cache
// can serve everything before the new lines.

// KeyMap is the selector key of map lines.
const KeyMap = "map"

// MapIntroID is the id of the item that opens the map section.
const MapIntroID = "map-intro"

// mapFoldPrefix starts the id of a line that stands for the notes of one
// topic that did not fit.
const mapFoldPrefix = "map-topic:"

// mapFoldSeq orders the lines of folded topics after every note line, by
// the first turn of each topic.
const mapFoldSeq = 1 << 29

// mapReserve is the share of an overfull map's budget kept for the lines
// that stand for folded topics.
const mapReserve = 0.15

// MapStats says what the map of a packet holds.
type MapStats struct {
	Notes  int `json:"notes"`  // notes with checked evidence before the focus
	Shown  int `json:"shown"`  // notes shown as a line
	Folded int `json:"folded"` // notes that did not fit and are named by topic only
	Stale  int `json:"stale"`  // notes left out because their turn's text is not what was checked
	Tokens int `json:"tokens"` // what the section costs
	Budget int `json:"budget"`
}

// MapLine renders a note as its map line: the paragraphs it points at, its
// kind and its title, the format the answer test put in front of a model.
func MapLine(n vault.Note) string {
	kind := strings.ToLower(strings.TrimSpace(n.Kind))
	if kind == "" {
		kind = "note"
	}
	return fmt.Sprintf("[%s] (%s) %s", strings.Join(n.Atoms(), ", "), kind, strings.Join(strings.Fields(n.Title), " "))
}

func mapTokens(s string) int { return Tokens(providers.Profile{}, len(s)+3) }

// mapIntro is the text that opens the section. pull is how many tokens of
// paragraphs one recall request brings back; folded says whether some
// notes are named by topic only.
func mapIntro(pull int, folded bool) string {
	var b strings.Builder
	b.WriteString("Each line is a note about one thing said in the conversation, written by an indexing model and checked by a program: the words the note quotes are in the paragraphs named in its brackets (t000012#a3 is paragraph 3 of turn t000012, t000012#a3-5 its paragraphs 3 to 5). A note points into the conversation; it is not the conversation's own words. Oldest first.\n\n")
	b.WriteString("When the task depends on what was said and the paragraphs a note points at are not below word for word, do not answer from the note: return status `needs_recall` with those paragraph ids in `recall_ids`, the most important first, everything the answer rests on in one request.")
	if pull > 0 {
		fmt.Fprintf(&b, " One request brings back about %d tokens of paragraphs, in the order you ask (often 30 to 60 paragraphs; long ones cost more), so ask for what matters and use the room.", pull)
	}
	if folded {
		b.WriteString("\n\nA line that starts with a topic in quotes stands for notes that did not fit here. To read them, add `{\"op\": \"notes\", \"target\": \"<the topic, or search words>\"}` to `recall_ops`.")
	}
	return b.String()
}

// noteBefore is n with only the evidence that lies before turn beforeSeq
// and whose turn still has the body the quote was checked against.
func noteBefore(n vault.Note, beforeSeq int, current func(turn, sha string) bool) (vault.Note, bool) {
	out := n
	out.Evidence = nil
	out.Seq, out.Last = 0, 0
	for _, ev := range n.Evidence {
		seq, ok := vault.TurnSeq(ev.Turn)
		if !ok || (beforeSeq > 0 && seq >= beforeSeq) {
			continue
		}
		if current != nil && !current(ev.Turn, ev.BodySHA) {
			continue
		}
		out.Evidence = append(out.Evidence, ev)
		if out.Seq == 0 || seq < out.Seq {
			out.Seq = seq
		}
		if seq > out.Last {
			out.Last = seq
		}
	}
	return out, len(out.Evidence) > 0 && strings.TrimSpace(out.Title) != ""
}

// UsableNotes keeps the notes a packet for the turn at beforeSeq may show:
// evidence before that turn, on turns whose text is what was checked.
// stale counts the notes dropped only because a turn's text differs.
func UsableNotes(notes []vault.Note, beforeSeq int, current func(turn, sha string) bool) (usable []vault.Note, stale int) {
	for _, n := range notes {
		if x, ok := noteBefore(n, beforeSeq, current); ok {
			usable = append(usable, x)
		} else if _, early := noteBefore(n, beforeSeq, nil); early {
			stale++
		}
	}
	sort.SliceStable(usable, func(i, j int) bool {
		if usable[i].Seq != usable[j].Seq {
			return usable[i].Seq < usable[j].Seq
		}
		return usable[i].ID < usable[j].ID
	})
	return usable, stale
}

// RankNotes orders notes for a query: those whose title or topics match,
// best first (BM25), then the rest, newest first.
func RankNotes(notes []vault.Note, query string) []int {
	pn := make([]Note, len(notes))
	for i, n := range notes {
		pn[i] = Note{Title: n.Title, Kind: n.Kind, Topics: n.Topics, Seq: n.Seq}
	}
	order := make([]int, 0, len(notes))
	seen := make([]bool, len(notes))
	for _, s := range NewNoteIndex(pn).Search(query) {
		order = append(order, s.i)
		seen[s.i] = true
	}
	for i := len(notes) - 1; i >= 0; i-- {
		if !seen[i] {
			order = append(order, i)
		}
	}
	return order
}

// noteTopic is the topic a note is folded under.
func noteTopic(n vault.Note) string {
	for _, t := range n.Topics {
		if t = strings.ToLower(strings.Join(strings.Fields(t), " ")); t != "" {
			return t
		}
	}
	return "other"
}

// BuildMap makes the map section's items for a focus at beforeSeq. While
// every note line fits budget tokens, all are shown, oldest first, and the
// section is the same for every message until a note is added. Past the
// budget the lines shown are the ones ranked for query, and the notes left
// out are named by topic with their count and turn range, so nothing is
// dropped without a line that says where it is. pull is the size of one
// recall request, told to the reader.
func BuildMap(notes []vault.Note, current func(turn, sha string) bool, beforeSeq int, query string, budget, pull int) ([]Item, MapStats) {
	st := MapStats{Budget: budget}
	if budget <= 0 {
		return nil, st
	}
	usable, stale := UsableNotes(notes, beforeSeq, current)
	st.Notes, st.Stale = len(usable), stale
	if len(usable) == 0 {
		return nil, st
	}
	// the budget covers the whole section, its opening text included
	introCost := mapTokens(mapIntro(pull, true))
	budget -= introCost
	if budget <= 0 {
		return nil, st
	}
	lines := make([]string, len(usable))
	total := 0
	for i, n := range usable {
		lines[i] = MapLine(n)
		total += mapTokens(lines[i])
	}
	show := make([]bool, len(usable))
	if total <= budget {
		for i := range show {
			show[i] = true
		}
	} else {
		room := int(float64(budget) * (1 - mapReserve))
		spent := 0
		for _, i := range RankNotes(usable, query) {
			if c := mapTokens(lines[i]); spent+c <= room {
				spent += c
				show[i] = true
			}
		}
	}
	var items []Item
	type fold struct {
		topic       string
		n           int
		first, last int
	}
	folds := map[string]*fold{}
	for i, n := range usable {
		if show[i] {
			st.Shown++
			st.Tokens += mapTokens(lines[i])
			items = append(items, Item{Key: KeyMap, ID: n.ID, Tier: 0, Seq: n.Seq, Sub: i + 1, Text: lines[i]})
			continue
		}
		st.Folded++
		tp := noteTopic(n)
		f := folds[tp]
		if f == nil {
			f = &fold{topic: tp, first: n.Seq, last: n.Last}
			folds[tp] = f
		}
		f.n++
		f.first, f.last = min(f.first, n.Seq), max(f.last, n.Last)
	}
	if len(folds) > 0 {
		fs := make([]*fold, 0, len(folds))
		for _, f := range folds {
			fs = append(fs, f)
		}
		// the topics with most folded notes are named first when the
		// reserve cannot name them all; the lines themselves go oldest first
		sort.SliceStable(fs, func(i, j int) bool {
			if fs[i].n != fs[j].n {
				return fs[i].n > fs[j].n
			}
			return fs[i].topic < fs[j].topic
		})
		room := budget - st.Tokens
		restNotes, restTopics := 0, 0
		for _, f := range fs {
			noun := "notes"
			if f.n == 1 {
				noun = "note"
			}
			line := fmt.Sprintf("%q: %d more %s, turns %s to %s", f.topic, f.n, noun, vault.TurnID(f.first), vault.TurnID(f.last))
			c := mapTokens(line)
			if c > room-40 { // keep room for the closing line
				restNotes += f.n
				restTopics++
				continue
			}
			room -= c
			st.Tokens += c
			items = append(items, Item{Key: KeyMap, ID: mapFoldPrefix + f.topic, Tier: 0, Seq: mapFoldSeq + f.first, Text: line})
		}
		if restTopics > 0 {
			line := fmt.Sprintf("and %d more notes under %d other topics; search them with the notes op", restNotes, restTopics)
			st.Tokens += mapTokens(line)
			items = append(items, Item{Key: KeyMap, ID: mapFoldPrefix, Tier: 0, Seq: 2 * mapFoldSeq, Text: line})
		}
	}
	intro := mapIntro(pull, st.Folded > 0)
	st.Tokens += introCost
	items = append(items, Item{Key: KeyMap, ID: MapIntroID, Tier: 0, Text: intro})
	return items, st
}

// MapStatsOf counts what the map items of a built packet hold.
func MapStatsOf(items []Item) (shown, folds, tokens int) {
	for _, it := range items {
		if it.Key != KeyMap {
			continue
		}
		tokens += mapTokens(it.Text)
		switch {
		case it.ID == MapIntroID:
		case strings.HasPrefix(it.ID, mapFoldPrefix):
			folds++
		default:
			shown++
		}
	}
	return shown, folds, tokens
}

// MatchNotes finds the notes a notes op asks for: those filed under a
// topic named target, else the best matches of target as search words.
func MatchNotes(notes []vault.Note, target string, limit int) []vault.Note {
	want := strings.ToLower(strings.Join(strings.Fields(target), " "))
	if want == "" || limit <= 0 {
		return nil
	}
	var out []vault.Note
	for _, n := range notes {
		for _, t := range n.Topics {
			if strings.ToLower(strings.Join(strings.Fields(t), " ")) == want {
				out = append(out, n)
				break
			}
		}
	}
	if len(out) == 0 {
		pn := make([]Note, len(notes))
		for i, n := range notes {
			pn[i] = Note{Title: n.Title, Kind: n.Kind, Topics: n.Topics, Seq: n.Seq}
		}
		for _, s := range NewNoteIndex(pn).Search(target) {
			out = append(out, notes[s.i])
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
