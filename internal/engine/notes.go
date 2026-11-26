package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// Idea notes are the experiment behind "notes as keys": a model reads the
// conversation once and writes small notes, each one idea in plain words
// with the passage that states it. The engine trusts none of it. A note
// counts only through the quotes code finds in the turn they cite, and each
// quote is pinned to the paragraphs that hold it, so a note is a checked
// pointer into the verbatim text and never a replacement for it. Searching
// the notes finds paragraphs the question's own words would not.
//
// This file is the offline experiment (`cx zoom -notes`). The live form,
// notes stored in the vault and shown in packets as a map, is atlas.go.

// IdeaNote is a note as a model wrote it.
type IdeaNote struct {
	Title    string         `json:"title"`
	Kind     string         `json:"kind"`
	Topics   []string       `json:"topics"`
	Evidence []NoteEvidence `json:"evidence"`
}

// NoteEvidence is where a note says its idea is stated.
type NoteEvidence struct {
	Turn  string `json:"turn"`
	Quote string `json:"quote"`
}

// LoadIdeaNotes reads every JSON-lines file of dir whose name starts with
// the workspace's name and a dash.
func LoadIdeaNotes(dir, workspace string) ([]IdeaNote, error) {
	files, err := filepath.Glob(filepath.Join(dir, filepath.Base(workspace)+"-*.jsonl"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return LoadIdeaNoteFiles(files)
}

// LoadIdeaNoteFiles reads notes from JSON-lines files: one note per line,
// with title, kind, topics and evidence (turn and quote).
func LoadIdeaNoteFiles(files []string) ([]IdeaNote, error) {
	var out []IdeaNote
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if line = strings.TrimSpace(line); line == "" {
				continue
			}
			var n IdeaNote
			if err := json.Unmarshal([]byte(line), &n); err != nil {
				return nil, fmt.Errorf("%s line %d: %w", f, i+1, err)
			}
			out = append(out, n)
		}
	}
	return out, nil
}

// TopicMap folds the topic names of independent note writers into one
// vocabulary, and files each topic under a broader area. In the engine the
// indexer would be shown the existing names and reuse them, as it is shown
// existing threads; writers that ran side by side need this step instead.
type TopicMap struct {
	Map   map[string]string `json:"map"`   // a topic as written to its canonical name
	Areas map[string]string `json:"areas"` // a canonical topic to its area
}

// LoadTopicMap reads dir/topics-<workspace>.json; ok is false when there
// is none.
func LoadTopicMap(dir, workspace string) (TopicMap, bool) {
	return LoadTopicMapFile(filepath.Join(dir, "topics-"+filepath.Base(workspace)+".json"))
}

// LoadTopicMapFile reads a topic map; ok is false when it cannot be read.
func LoadTopicMapFile(path string) (TopicMap, bool) {
	var tm TopicMap
	data, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(data, &tm) != nil {
		return TopicMap{}, false
	}
	return tm, true
}

// NoteCheck is what checking a set of notes found.
type NoteCheck struct {
	Notes, Kept        int // notes read; notes with at least one quote found
	Quotes, Found      int // quotes read; quotes found in the turn they cite
	Atoms, AllAtoms    int // distinct paragraphs pinned; paragraphs in the conversation
	Turns, AllTurns    int // turns with a pinned paragraph; turns
	TitleTokens        int // what all kept titles cost
	ConversationTokens int
}

func (c NoteCheck) String() string {
	pct := func(a, b int) int { return 100 * a / max(b, 1) }
	return fmt.Sprintf("%d notes, %d kept; %d of %d quotes found in the turn they cite (%d%%); they pin %d of %d paragraphs (%d%%) in %d of %d turns; titles cost %d tokens against %d for the conversation (%d%%)",
		c.Notes, c.Kept, c.Found, c.Quotes, pct(c.Found, c.Quotes), c.Atoms, c.AllAtoms, pct(c.Atoms, c.AllAtoms), c.Turns, c.AllTurns,
		c.TitleTokens, c.ConversationTokens, pct(c.TitleTokens, c.ConversationTokens))
}

// atomSpan is the first and last paragraph of t holding bytes start to end.
func atomSpan(t vault.Turn, start, end int) (a, b int) {
	return atomSpanCached(packet.AtomsOf(t), start, end)
}

// atomSpanCached is the first and last paragraph of t holding bytes start to end,
// using sorted boundaries over a pre-split atom slice via binary search.
func atomSpanCached(atoms []vault.Atom, start, end int) (a, b int) {
	i := sort.Search(len(atoms), func(k int) bool { return atoms[k].End > start })
	if i >= len(atoms) {
		return 0, 0
	}
	j := sort.Search(len(atoms)-i, func(k int) bool { return atoms[i+k].End >= end }) + i
	if j >= len(atoms) {
		j = len(atoms) - 1
	}
	return atoms[i].K, atoms[j].K
}

// CheckIdeaNotes keeps what code can confirm: for each note, the quotes
// that occur in the turn they cite, pinned to their paragraphs. Caches atom
// splits and normalized bodies per immutable turn across all quotes.
func CheckIdeaNotes(turns []vault.Turn, notes []IdeaNote, tm *TopicMap) ([]packet.Note, NoteCheck) {
	byID := map[string]vault.Turn{}
	atomsByTurn := make(map[string][]vault.Atom, len(turns))
	type normBody struct {
		text string
		at   []int
	}
	normByTurn := map[string]normBody{}
	chk := NoteCheck{Notes: len(notes), AllTurns: len(turns)}
	for _, t := range turns {
		byID[t.ID] = t
		as := packet.AtomsOf(t)
		atomsByTurn[t.ID] = as
		chk.AllAtoms += len(as)
		chk.ConversationTokens += packet.Tokens(providers.Profile{}, len(t.Body))
	}
	findInTurn := func(t vault.Turn, quote string) (int, int, bool) {
		q := strings.Join(strings.Fields(quote), " ")
		if len(q) < 3 {
			return 0, 0, false
		}
		nb, ok := normByTurn[t.ID]
		if !ok {
			nb.text, nb.at = normIndex(t.Body, 0)
			normByTurn[t.ID] = nb
		}
		i := strings.Index(nb.text, q)
		if i < 0 {
			return 0, 0, false
		}
		return nb.at[i], nb.at[i+len(q)-1] + 1, true
	}
	pinned, pinnedTurns := map[string]bool{}, map[string]bool{}
	var out []packet.Note
	for _, n := range notes {
		pn := packet.Note{Title: strings.TrimSpace(n.Title), Kind: strings.ToLower(strings.TrimSpace(n.Kind))}
		have := map[string]bool{}
		for _, tp := range n.Topics {
			tp = strings.ToLower(strings.TrimSpace(tp))
			if tm != nil {
				if c, ok := tm.Map[tp]; ok {
					tp = c
				}
			}
			if tp == "" || have[tp] {
				continue
			}
			have[tp] = true
			pn.Topics = append(pn.Topics, tp)
			if tm != nil {
				if a := tm.Areas[tp]; a != "" && !have["area:"+a] {
					have["area:"+a] = true
					pn.Areas = append(pn.Areas, a)
				}
			}
		}
		seen := map[string]bool{}
		for _, ev := range n.Evidence {
			chk.Quotes++
			t, ok := byID[strings.TrimSpace(ev.Turn)]
			if !ok {
				continue
			}
			start, end, ok := findInTurn(t, ev.Quote)
			if !ok {
				continue
			}
			chk.Found++
			pn.Quotes = append(pn.Quotes, packet.QuoteSpan{Turn: t.ID, BodySHA: t.BodySHA, Start: start, End: end, Seg: vault.AtomSeg})
			a, b := atomSpanCached(atomsByTurn[t.ID], start, end)
			for k := a; k <= b && k > 0; k++ {
				id := vault.AtomID(t.ID, k)
				if !seen[id] {
					seen[id] = true
					pn.Atoms = append(pn.Atoms, id)
				}
				pinned[id], pinnedTurns[t.ID] = true, true
			}
			if pn.Seq == 0 || t.Seq < pn.Seq {
				pn.Seq = t.Seq
			}
		}
		if pn.Title == "" || len(pn.Atoms) == 0 {
			continue
		}
		chk.Kept++
		chk.TitleTokens += packet.Tokens(providers.Profile{}, len(pn.Title)+24)
		out = append(out, pn)
	}
	chk.Atoms, chk.Turns = len(pinned), len(pinnedTurns)
	// oldest first, whatever order the writers' files came in: a reader is
	// shown notes in conversation order, and ties between notes are broken
	// by position
	sort.SliceStable(out, func(a, b int) bool { return out[a].Seq < out[b].Seq })
	return out, chk
}
