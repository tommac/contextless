package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// The atlas is the live form of the idea-note experiment (notes.go): notes
// are stored in the vault as checked records, the packet shows them as a
// map of the conversation, and the model asks for the paragraphs it needs
// by id. Nothing a model writes is trusted: a note exists only through
// quotes that code found in the turn they cite.

// maxPullIDs caps the ids of one recall request while the map is shown.
// The request is bounded by tokens (Config.AtlasPullTokens); the count
// only stops a runaway list.
const maxPullIDs = 80

// atlasShown reports whether packets of this workspace carry the map.
func (e *Engine) atlasShown() bool {
	return e.Cfg.Atlas != config.AtlasOff && e.Cfg.AtlasMapTokens > 0 && e.V.HasNotes()
}

// atlasMapTokens is the map's budget for a call that may run on cands: the
// configured budget, held to a quarter of the largest window's plan cap so
// a map can never crowd the records out of a small window.
func (e *Engine) atlasMapTokens(cands []providers.Profile) int {
	if !e.atlasShown() {
		return 0
	}
	budget := e.Cfg.AtlasMapTokens
	largest := 0
	for _, p := range cands {
		largest = max(largest, p.WithDefaults().PlanCap())
	}
	if largest > 0 {
		budget = min(budget, largest/4)
	}
	return budget
}

// pullTokens is how many tokens of paragraphs one recall request brings
// back while the map is shown; 0 means the request is bounded by count
// and by what fits the packet, as without a map.
func (e *Engine) pullTokens() int {
	if !e.atlasShown() || e.Cfg.AtlasPullTokens <= 0 {
		return 0
	}
	return e.Cfg.AtlasPullTokens
}

// pinQuote finds quote in t (whitespace aside) and pins it to the
// paragraphs that hold it.
func pinQuote(t vault.Turn, quote string) (vault.NoteEvidence, bool) {
	q := strings.Join(strings.Fields(quote), " ")
	if len(q) < 3 {
		return vault.NoteEvidence{}, false
	}
	text, at := normIndex(t.Body, 0)
	i := strings.Index(text, q)
	if i < 0 {
		return vault.NoteEvidence{}, false
	}
	start, end := at[i], at[i+len(q)-1]+1
	a, b := atomSpan(t, start, end)
	if a <= 0 {
		return vault.NoteEvidence{}, false
	}
	return vault.NoteEvidence{Turn: t.ID, BodySHA: t.BodySHA, Quote: t.Body[start:end], Start: start, End: end, A: a, B: b}, true
}

// noteID names a note by what it says and where: the same note written or
// imported twice is one record.
func noteID(title string, ev []vault.NoteEvidence) string {
	h := sha256.New()
	h.Write([]byte(strings.ToLower(strings.Join(strings.Fields(title), " "))))
	for _, e := range ev {
		fmt.Fprintf(h, "\n%s:%d:%d", e.Turn, e.Start, e.End)
	}
	return "no-" + hex.EncodeToString(h.Sum(nil))[:12]
}

// NotePut is what storing a set of proposed notes did.
type NotePut struct {
	Read    int // notes proposed
	Kept    int // notes with a title and at least one quote found
	New     int // of those, notes the vault did not already hold
	Quotes  int // quotes proposed
	Found   int // quotes found in the turn they cite
	Dropped int // notes discarded: no title, or no quote found
}

func (p NotePut) String() string {
	return fmt.Sprintf("%d notes read, %d kept (%d new), %d discarded; %d of %d quotes found in the turn they cite",
		p.Read, p.Kept, p.New, p.Dropped, p.Found, p.Quotes)
}

// putNotes checks proposed notes against the turns they cite and stores
// the ones with at least one quote found. turnOf resolves a cited turn id
// (the indexer's batch, or any turn for an import); tm, when given, folds
// topic names into one vocabulary.
func (e *Engine) putNotes(in []IdeaNote, turnOf func(id string) (vault.Turn, bool), tm *TopicMap, runID string) (NotePut, []vault.Note) {
	res := NotePut{Read: len(in)}
	have := map[string]bool{}
	for _, n := range e.V.Notes() {
		have[n.ID] = true
	}
	var stored []vault.Note
	for _, n := range in {
		note := vault.Note{Title: strings.Join(strings.Fields(n.Title), " "), Kind: strings.ToLower(strings.TrimSpace(n.Kind)), Run: runID}
		seenTopic := map[string]bool{}
		for _, tp := range n.Topics {
			tp = strings.ToLower(strings.Join(strings.Fields(tp), " "))
			if tm != nil {
				if c, ok := tm.Map[tp]; ok {
					tp = c
				}
			}
			if tp != "" && !seenTopic[tp] {
				seenTopic[tp] = true
				note.Topics = append(note.Topics, tp)
			}
		}
		seenSpan := map[string]bool{}
		for _, ev := range n.Evidence {
			res.Quotes++
			t, ok := turnOf(strings.TrimSpace(ev.Turn))
			if !ok {
				continue
			}
			pinned, ok := pinQuote(t, ev.Quote)
			if !ok {
				continue
			}
			res.Found++
			key := fmt.Sprintf("%s:%d:%d", pinned.Turn, pinned.Start, pinned.End)
			if seenSpan[key] {
				continue
			}
			seenSpan[key] = true
			note.Evidence = append(note.Evidence, pinned)
			if note.Seq == 0 || t.Seq < note.Seq {
				note.Seq = t.Seq
			}
			note.Last = max(note.Last, t.Seq)
		}
		if note.Title == "" || len(note.Evidence) == 0 {
			res.Dropped++
			continue
		}
		sort.SliceStable(note.Evidence, func(i, j int) bool {
			if note.Evidence[i].Turn != note.Evidence[j].Turn {
				return note.Evidence[i].Turn < note.Evidence[j].Turn
			}
			return note.Evidence[i].Start < note.Evidence[j].Start
		})
		note.ID = noteID(note.Title, note.Evidence)
		res.Kept++
		if have[note.ID] {
			continue
		}
		if err := e.V.PutNote(note); err != nil {
			e.logf("notes: %v", err)
			res.Kept--
			res.Dropped++
			continue
		}
		have[note.ID] = true
		res.New++
		stored = append(stored, note)
	}
	return res, stored
}

// ImportNotes stores notes written outside the engine (the JSON-lines
// format of LoadIdeaNotes), after the same check as the indexer's: a note
// is kept only through quotes found in the turns they cite.
func (e *Engine) ImportNotes(notes []IdeaNote, tm *TopicMap) (NotePut, error) {
	turns, err := e.V.Turns()
	if err != nil {
		return NotePut{}, err
	}
	byID := make(map[string]vault.Turn, len(turns))
	for _, t := range turns {
		byID[t.ID] = t
	}
	res, stored := e.putNotes(notes, func(id string) (vault.Turn, bool) { t, ok := byID[id]; return t, ok }, tm, "")
	if len(stored) > 0 {
		e.V.Commit(fmt.Sprintf("Import %d idea notes", len(stored)))
	}
	return res, nil
}

// AtlasReport says what the workspace's map holds and costs.
type AtlasReport struct {
	Notes, Stale           int
	Topics                 int
	Paragraphs, AllParas   int // distinct paragraphs a note points at; paragraphs in the conversation
	Turns, AllTurns        int
	MapTokens, AllTokens   int // every note line; the whole conversation
	Budget                 int
	Kinds                  map[string]int
	Shown, Folded          int // under Budget, for the newest message
	Mode                   string
	UncoveredTurns         []string // turns no note points at, oldest first (at most 20)
	UncoveredTurnsTotal    int
	UncoveredTokens        int
	LargestUncoveredTokens int
}

// Atlas reports on the workspace's notes: how much of the conversation
// they point at, what the map costs, and which turns no note covers.
func (e *Engine) Atlas() (AtlasReport, error) {
	rep := AtlasReport{Kinds: map[string]int{}, Mode: e.Cfg.Atlas, Budget: e.Cfg.AtlasMapTokens}
	turns, err := e.V.Turns()
	if err != nil {
		return rep, err
	}
	sha := make(map[string]string, len(turns))
	for _, t := range turns {
		sha[t.ID] = t.BodySHA
		rep.AllParas += len(packet.AtomsOf(t))
		rep.AllTokens += packet.Tokens(providers.Profile{}, len(t.Body))
	}
	rep.AllTurns = len(turns)
	usable, stale := packet.UsableNotes(e.V.Notes(), 0, func(turn, s string) bool { return sha[turn] == s })
	rep.Notes, rep.Stale = len(usable), stale
	paras, covered, topics := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, n := range usable {
		kind := n.Kind
		if kind == "" {
			kind = "note"
		}
		rep.Kinds[kind]++
		for _, tp := range n.Topics {
			topics[tp] = true
		}
		for _, ev := range n.Evidence {
			covered[ev.Turn] = true
			for k := ev.A; k <= ev.B; k++ {
				paras[vault.AtomID(ev.Turn, k)] = true
			}
		}
		rep.MapTokens += packet.Tokens(providers.Profile{}, len(packet.MapLine(n))+3)
	}
	rep.Topics, rep.Paragraphs, rep.Turns = len(topics), len(paras), len(covered)
	for _, t := range turns {
		if covered[t.ID] {
			continue
		}
		rep.UncoveredTurnsTotal++
		tok := packet.Tokens(providers.Profile{}, len(t.Body))
		rep.UncoveredTokens += tok
		rep.LargestUncoveredTokens = max(rep.LargestUncoveredTokens, tok)
		if len(rep.UncoveredTurns) < 20 {
			rep.UncoveredTurns = append(rep.UncoveredTurns, t.ID)
		}
	}
	last := len(turns) + 1
	_, st := packet.BuildMap(usable, nil, last, "", e.Cfg.AtlasMapTokens, e.Cfg.AtlasPullTokens)
	rep.Shown, rep.Folded = st.Shown, st.Folded
	return rep, nil
}

func (r AtlasReport) String() string {
	pct := func(a, b int) int { return 100 * a / max(b, 1) }
	var b strings.Builder
	fmt.Fprintf(&b, "atlas %s: %d notes under %d topics", r.Mode, r.Notes, r.Topics)
	if r.Stale > 0 {
		fmt.Fprintf(&b, " (%d more left out: their turn's text is not what was checked)", r.Stale)
	}
	fmt.Fprintf(&b, "\nthey point at %d of %d paragraphs (%d%%) in %d of %d turns (%d%%)", r.Paragraphs, r.AllParas, pct(r.Paragraphs, r.AllParas), r.Turns, r.AllTurns, pct(r.Turns, r.AllTurns))
	fmt.Fprintf(&b, "\nevery note line costs %d tokens against %d for the conversation (%d%%); the map budget is %d", r.MapTokens, r.AllTokens, pct(r.MapTokens, r.AllTokens), r.Budget)
	if r.Folded > 0 {
		fmt.Fprintf(&b, ": %d lines shown, %d notes named by topic only", r.Shown, r.Folded)
	}
	if len(r.Kinds) > 0 {
		kinds := make([]string, 0, len(r.Kinds))
		for k := range r.Kinds {
			kinds = append(kinds, k)
		}
		sort.Slice(kinds, func(i, j int) bool {
			if r.Kinds[kinds[i]] != r.Kinds[kinds[j]] {
				return r.Kinds[kinds[i]] > r.Kinds[kinds[j]]
			}
			return kinds[i] < kinds[j]
		})
		b.WriteString("\nkinds:")
		for _, k := range kinds {
			fmt.Fprintf(&b, " %s %d;", k, r.Kinds[k])
		}
	}
	if r.UncoveredTurnsTotal > 0 {
		fmt.Fprintf(&b, "\n%d turns have no note (%d tokens, the largest %d): %s", r.UncoveredTurnsTotal, r.UncoveredTokens, r.LargestUncoveredTokens, strings.Join(r.UncoveredTurns, ", "))
		if r.UncoveredTurnsTotal > len(r.UncoveredTurns) {
			b.WriteString(", ...")
		}
		b.WriteString("\nthey are still reached by search and by their catalog lines")
	}
	return strings.TrimSuffix(b.String(), ";")
}

// noteTopicsListing gives the indexer the topic names already in use, the
// commonest first, so it files new notes under them instead of coining a
// second name for the same subject.
func (e *Engine) noteTopicsListing(limit int) string {
	count := map[string]int{}
	for _, n := range e.V.Notes() {
		for _, tp := range n.Topics {
			count[tp]++
		}
	}
	names := make([]string, 0, len(count))
	for tp := range count {
		names = append(names, tp)
	}
	sort.Slice(names, func(i, j int) bool {
		if count[names[i]] != count[names[j]] {
			return count[names[i]] > count[names[j]]
		}
		return names[i] < names[j]
	})
	if len(names) > limit {
		names = names[:limit]
	}
	var b strings.Builder
	for _, tp := range names {
		fmt.Fprintf(&b, "- %s (%d)\n", tp, count[tp])
	}
	if b.Len() == 0 {
		return "(none yet)"
	}
	return strings.TrimSpace(b.String())
}

// opNotes answers a notes recall op: the lines of the notes filed under a
// topic, or matching search words, as a supplied record.
func (e *Engine) opNotes(target string, limit, maxSeq, idx int) (*packet.Item, string) {
	if e.Cfg.Atlas == config.AtlasOff {
		return nil, "notes: this workspace shows no map of the conversation."
	}
	if limit <= 0 || limit > 80 {
		limit = 40
	}
	usable, _ := packet.UsableNotes(e.V.Notes(), maxSeq+1, func(turn, sha string) bool {
		t, err := e.V.Turn(turn)
		return err == nil && t.BodySHA == sha
	})
	if len(usable) == 0 {
		return nil, "notes: this conversation has no notes yet; use recall_queries to search it."
	}
	hits := packet.MatchNotes(usable, target, limit+1)
	if len(hits) == 0 {
		return nil, fmt.Sprintf("notes(%q) matched no topic and no note; use recall_queries to search the conversation itself.", target)
	}
	more := len(hits) > limit
	if more {
		hits = hits[:limit]
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Seq < hits[j].Seq })
	var b strings.Builder
	for _, n := range hits {
		b.WriteString("- " + packet.MapLine(n) + "\n")
	}
	item := &packet.Item{
		Key: packet.KeyExtra, ID: fmt.Sprintf("recall-notes-%d", idx+1), Seq: 94 + idx,
		Header: fmt.Sprintf("Notes for %q (pointers into the conversation, oldest first; recall their paragraphs by id)", target),
		Text:   strings.TrimSpace(b.String()),
	}
	note := fmt.Sprintf("notes(%q) listed %d note(s).", target, len(hits))
	if more {
		note += " More match; narrow the words to see others."
	}
	return item, note
}
