package engine

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// The offline zoom test says which passages a packet holds. Whether a model
// then answers better is a different question, and a packet has to be put in
// front of a model to settle it. DumpPackets writes, for every probe, the
// packet each arm would give at one budget, in one fixed plain format, and
// the passages a complete answer draws on, so that the same model can answer
// from each packet and judges who do not know the arm can score the
// answers. It runs no model itself.
//
// Three things differ from the offline test, on purpose. A packet is filled
// by its size as written, so arms that are said to cost the same do (the
// offline test charges every paragraph a flat header instead). Packets and
// gold passages are in folders with meaningless names, one file each, so a
// model that is given one file finds nothing else beside it. And the index
// records, for every arm and needed passage, whether the packet held the
// passage, only pointed at it with a note line, or neither, so answers can
// be read against what was in front of the model.

// DumpArm names a packet arm and its short name in the index. Total, when
// set, is the arm's own budget in place of the dump's.
type DumpArm struct {
	Name, Dir string
	Total     int
}

// Arms that exist only in the answer test.
const (
	dumpRecent    = "recent only"                    // no ranking: the newest paragraphs
	dumpNone      = "the question alone"             // control below: no records at all, what can be guessed
	dumpOracle    = "the passages themselves"        // control above: exactly the paragraphs a complete answer draws on
	dumpSearchBig = "search at the all-notes total"  // plain search filled to the size of the all-notes packet
	dumpTitles    = "every note line, no paragraphs" // the conversation at note resolution only
	dumpCatalog   = "a code-built line per turn"     // what can be shown with no notes written: every turn's opening, terms and values
	dumpFull      = "the whole conversation"
)

// DumpArms are the arms of the answer test. The first six hold at most the
// budget: three ways to select, the newest paragraphs with no ranking, and
// a control on each side. The next two are larger: every note line plus
// the budget's worth of paragraphs, and plain search filled to that same
// size. The last two hold no paragraph: the note lines alone, and a
// code-built line per turn. A model can also be shown either of those and
// then ask for paragraphs by id (PullExcerpts).
var DumpArms = []DumpArm{
	{Name: ZoomSearch, Dir: "search"},
	{Name: ZoomTopics, Dir: "notes"},
	{Name: ZoomView, Dir: "lines"},
	{Name: dumpRecent, Dir: "recent"},
	{Name: dumpNone, Dir: "none"},
	{Name: dumpOracle, Dir: "oracle"},
	{Name: ZoomAllNotes, Dir: "allnotes"},
	{Name: dumpSearchBig, Dir: "searchbig"},
	{Name: dumpTitles, Dir: "titles"},
	{Name: dumpCatalog, Dir: "catalog"},
}

// DumpSmall are further budgets at which the two arms that only select
// paragraphs, plain search and notes, are also written (arms search-2000,
// notes-2000 and so on). An answer is often as good from 8,000 tokens as
// from the whole conversation, whatever chose them; smaller packets show
// where the ways of choosing part.
var DumpSmall []int

// DumpWhole also writes, for every probe, a packet holding the whole
// conversation (arm full): what a model answers with everything in front of
// it. It is far larger than every other packet, so it is written only when
// asked for.
var DumpWhole = false

// DumpSalt, when set, replaces the random salt that folder names are made
// from. Tests set it.
var DumpSalt = ""

// DumpMeta describes a dump as a whole (dump.json).
type DumpMeta struct {
	Total int    `json:"total"` // the budget, in tokens as written
	Salt  string `json:"salt"`  // folder names are OpaqueID(salt, question id, arm)
	Tag   string `json:"tag"`   // the tag excerpts are wrapped in
}

// DumpIndex describes one probe's packets (index.json).
type DumpIndex struct {
	QID       string `json:"qid"`
	Workspace string `json:"workspace"`
	Kind      string `json:"kind"`
	Type      string `json:"type,omitempty"` // the probe's own label for what it asks: evolution, inventory and so on
	Sources   int    `json:"sources"`
	// SourceAtoms are, for each needed passage, the paragraphs holding it.
	SourceAtoms [][]string `json:"source_atoms"`
	Gold        string     `json:"gold"` // folder under g/ holding gold.md
	// By arm: the folder under p/ holding packet.md, its tokens as written,
	// how many paragraphs it holds, and one letter per needed passage: h
	// when the packet holds the passage word for word, q when it does not
	// but a note line shown quotes the passage's own words, p when a note
	// line shown is pinned to its paragraph for other words, n otherwise.
	Packets    map[string]string `json:"packets"`
	Tokens     map[string]int    `json:"tokens"`
	Paragraphs map[string]int    `json:"paragraphs"`
	Status     map[string]string `json:"status"`
}

// OpaqueID is the folder name of a question's packet for an arm, or of its
// gold passages (arm "gold"): sixteen hex digits that say nothing without
// the salt. The answer-test script computes the same value.
func OpaqueID(salt, qid, arm string) string {
	h := func(tag string) string {
		f := fnv.New32a()
		f.Write([]byte(salt + "|" + tag + "|" + qid + "|" + arm))
		return fmt.Sprintf("%08x", f.Sum32())
	}
	return h("a") + h("b")
}

func writtenTokens(text string) int { return packet.Tokens(providers.Profile{}, len(text)) }

const dumpPreamble = `# Records of a conversation

You are answering a question about a long conversation between a user and an assistant. You cannot see the conversation itself, only what is below, which a program put together from it. Turns are numbered in order of time, t000001 first. The question is the last section.
`

const dumpNotesIntro = `
## Notes

Each line is a note about one thing said in the conversation, written by an indexing model and checked by a program: the words the note quotes are in the paragraphs named in its brackets (t000012#a3 is paragraph 3 of turn t000012). A note points into the conversation; it is not the conversation's own words. Oldest first.

`

const dumpCatalogIntro = `
## Turns

One line per turn of the conversation, made by a program: the turn's id and author, how it opens, its most distinctive terms and values, and its size. Oldest first.

`

const dumpExcerptsIntro = `
## Excerpts

Each excerpt is one paragraph copied word for word from the conversation, oldest first. Its tag gives the turn, the author and which paragraph of the turn it is.
`

const dumpWholeIntro = `
## Conversation

Every turn of the conversation, word for word, oldest first.
`

// DumpAtom is one paragraph of a dumped conversation.
type DumpAtom struct {
	ID   string `json:"id"`
	Turn string `json:"turn"`
	Role string `json:"role"`
	K    int    `json:"k"`
	Of   int    `json:"of"`
	Text string `json:"text"`
}

func dumpAtom(r packet.RankedAtom) DumpAtom {
	return DumpAtom{ID: r.Atom.ID(), Turn: r.Turn.ID, Role: r.Turn.Role, K: r.Atom.K, Of: len(vault.Atoms(r.Turn.ID, r.Turn.Body)), Text: r.Atom.Text(r.Turn.Body)}
}

// renderExcerpt wraps a paragraph in the dump's tag. Conversation text can
// hold headings and fences of its own; a tag that occurs nowhere in the
// conversation cannot be mistaken for them.
func renderExcerpt(a DumpAtom, tag string) string {
	return fmt.Sprintf("\n<%s turn=\"%s\" by=\"%s\" paragraph=\"%d of %d\" id=\"%s\">\n%s\n</%s>\n", tag, a.Turn, a.Role, a.K, a.Of, a.ID, a.Text, tag)
}

// dumpTag is a tag name that occurs in no turn.
func dumpTag(turns []vault.Turn) string {
	for _, tag := range []string{"excerpt", "excerpt-r", "record-x", "cx-excerpt-7f3a"} {
		used := false
		for _, t := range turns {
			if strings.Contains(t.Body, "<"+tag) || strings.Contains(t.Body, "</"+tag) {
				used = true
				break
			}
		}
		if !used {
			return tag
		}
	}
	return "cx-excerpt-b91e04d2"
}

func noteLine(n packet.Note) string {
	return fmt.Sprintf("- [%s] (%s) %s\n", strings.Join(n.Atoms, ", "), n.Kind, n.Title)
}

// dumpPacket is what one packet holds.
type dumpPacket struct {
	notes   []packet.Note       // note lines, shown in the order given (oldest first)
	catalog []string            // code-built turn lines
	picked  []packet.RankedAtom // paragraphs, shown oldest first
	whole   []vault.Turn        // every turn, for the full arm
}

// render writes a packet: the sections it has, then the question.
func (d dumpPacket) render(question, tag string) string {
	var b strings.Builder
	b.WriteString(dumpPreamble)
	if len(d.notes) > 0 {
		b.WriteString(dumpNotesIntro)
		for _, n := range d.notes {
			b.WriteString(noteLine(n))
		}
	}
	if len(d.catalog) > 0 {
		b.WriteString(dumpCatalogIntro)
		for _, l := range d.catalog {
			b.WriteString(l)
		}
	}
	if len(d.picked) > 0 {
		picked := append([]packet.RankedAtom{}, d.picked...)
		sort.SliceStable(picked, func(i, j int) bool {
			if picked[i].Turn.Seq != picked[j].Turn.Seq {
				return picked[i].Turn.Seq < picked[j].Turn.Seq
			}
			return picked[i].Atom.K < picked[j].Atom.K
		})
		b.WriteString(dumpExcerptsIntro)
		for _, r := range picked {
			b.WriteString(renderExcerpt(dumpAtom(r), tag))
		}
	}
	if len(d.whole) > 0 {
		b.WriteString(dumpWholeIntro)
		for _, t := range d.whole {
			fmt.Fprintf(&b, "\n<turn id=\"%s\" by=\"%s\">\n%s\n</turn>\n", t.ID, t.Role, strings.TrimSpace(t.Body))
		}
	}
	fmt.Fprintf(&b, "\n## Question\n\n%s\n", strings.TrimSpace(question))
	return b.String()
}

// fitWritten takes paragraphs in the order given while their excerpts, as
// written, fit room tokens: one that does not fit is passed over for a
// smaller one further down, and so is one whose exact text was taken.
func fitWritten(order []packet.RankedAtom, room int, tag string) []packet.RankedAtom {
	var picked []packet.RankedAtom
	seen := map[string]bool{}
	spent := 0
	for _, r := range order {
		a := dumpAtom(r)
		if len(a.Text) >= 80 && seen[a.Text] {
			continue
		}
		if c := writtenTokens(renderExcerpt(a, tag)); spent+c <= room {
			spent += c
			picked = append(picked, r)
			seen[a.Text] = true
		}
	}
	return picked
}

// sourceSpans are, for each needed passage of a probe, where its words are
// and the paragraphs holding them. A passage that is not in its turn is an
// error: the answer test would otherwise score against a passage no packet
// could hold.
func sourceSpans(p ZoomProbe, byTurn map[string]vault.Turn) ([]srcLoc, [][]packet.RankedAtom, error) {
	locs := make([]srcLoc, len(p.Sources))
	atoms := make([][]packet.RankedAtom, len(p.Sources))
	for i, src := range p.Sources {
		t, ok := byTurn[src.Turn]
		if !ok {
			return nil, nil, fmt.Errorf("probe %s: source %d names turn %s, which is not in the conversation", p.ID, i+1, src.Turn)
		}
		start, end, ok := findQuote(t.Body, 0, len(t.Body), src.Answer)
		a, b := 0, 0
		if ok {
			a, b = atomSpan(t, start, end)
		}
		all := vault.Atoms(t.ID, t.Body)
		if a < 1 || b > len(all) {
			return nil, nil, fmt.Errorf("probe %s: the words of source %d are not in turn %s", p.ID, i+1, src.Turn)
		}
		locs[i] = srcLoc{turn: t.ID, start: start, end: end, atoms: map[string]bool{}, ok: true}
		for k := a; k <= b; k++ {
			locs[i].atoms[vault.AtomID(t.ID, k)] = true
			atoms[i] = append(atoms[i], packet.RankedAtom{Turn: t, Atom: all[k-1]})
		}
	}
	return locs, atoms, nil
}

// sourceStatus is one letter per needed passage for a packet: h, q, p or n
// (see DumpIndex).
func sourceStatus(locs []srcLoc, d dumpPacket) string {
	held := map[string]bool{}
	for _, r := range d.picked {
		held[r.Atom.ID()] = true
	}
	var b strings.Builder
	for _, loc := range locs {
		status := byte('h')
		for id := range loc.atoms {
			if !held[id] && len(d.whole) == 0 {
				status = 'n'
			}
		}
		if status == 'n' {
			for _, n := range d.notes {
				for _, q := range n.Quotes {
					if q.Turn == loc.turn && q.Start < loc.end && loc.start < q.End {
						status = 'q'
					}
				}
				if status != 'q' {
					for _, id := range n.Atoms {
						if loc.atoms[id] {
							status = 'p'
						}
					}
				}
			}
		}
		b.WriteByte(status)
	}
	return b.String()
}

// DumpPackets writes the answer test's files into dir: p/<id>/packet.md for
// every probe and arm at the budget total (tokens as written), g/<id>/gold.md
// with the passages a complete answer draws on, atoms/<workspace>.json for
// PullExcerpts, dump.json and index.json. Every set needs notes. Use
// scratch copies: opening may update the index.
func DumpPackets(cfg config.Config, sets []ZoomSet, total int, dir string, logf func(string)) ([]DumpIndex, error) {
	meta := DumpMeta{Total: total, Salt: DumpSalt}
	if meta.Salt == "" {
		raw := make([]byte, 12)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		meta.Salt = hex.EncodeToString(raw)
	}
	for _, sub := range []string{"p", "g", "atoms"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return nil, err
		}
	}
	arms := dumpArms()
	var index []DumpIndex
	wroteAtoms, seenQ := map[string]bool{}, map[string]bool{}
	for _, set := range sets {
		if len(set.Notes) == 0 {
			return nil, fmt.Errorf("dump: no notes for %s", set.Workspace)
		}
		logf("dump: " + set.Name)
		e, err := OpenReadOnly(cfg, set.Workspace)
		if err != nil {
			return nil, err
		}
		one, err := dumpSet(e, set, &meta, arms, dir, wroteAtoms, seenQ)
		e.Close()
		if err != nil {
			return nil, err
		}
		index = append(index, one...)
	}
	js, _ := json.MarshalIndent(meta, "", " ")
	if err := os.WriteFile(filepath.Join(dir, "dump.json"), append(js, '\n'), 0o644); err != nil {
		return nil, err
	}
	js, _ = json.MarshalIndent(index, "", " ")
	return index, os.WriteFile(filepath.Join(dir, "index.json"), append(js, '\n'), 0o644)
}

func dumpSet(e *Engine, set ZoomSet, meta *DumpMeta, arms []DumpArm, dir string, wroteAtoms, seenQ map[string]bool) ([]DumpIndex, error) {
	turns, err := e.V.Turns()
	if err != nil {
		return nil, err
	}
	tag := dumpTag(turns)
	if meta.Tag != "" && meta.Tag != tag {
		return nil, fmt.Errorf("dump: the conversations need different excerpt tags (%s, %s); dump them apart", meta.Tag, tag)
	}
	meta.Tag = tag
	next := e.nextTurn("").Seq
	byTurn := map[string]vault.Turn{}
	for _, t := range turns {
		byTurn[t.ID] = t
	}
	checked, _ := CheckIdeaNotes(turns, set.Notes, set.Topics) // oldest first
	notes := packet.NewNoteIndex(checked)
	spread := packet.NewSpreadIndex(turns)
	ws := filepath.Base(set.Workspace)
	var newest []packet.RankedAtom // every paragraph, newest first
	for i := len(turns) - 1; i >= 0; i-- {
		atoms := vault.Atoms(turns[i].ID, turns[i].Body)
		for k := len(atoms) - 1; k >= 0; k-- {
			newest = append(newest, packet.RankedAtom{Turn: turns[i], Atom: atoms[k]})
		}
	}
	if !wroteAtoms[ws] {
		wroteAtoms[ws] = true
		all := make([]DumpAtom, 0, len(newest))
		for i := len(newest) - 1; i >= 0; i-- {
			all = append(all, dumpAtom(newest[i]))
		}
		js, _ := json.Marshal(all)
		if err := os.WriteFile(filepath.Join(dir, "atoms", ws+".json"), js, 0o644); err != nil {
			return nil, err
		}
	}
	var catalog []string
	for _, n := range packet.BuildZoom(turns, true).TurnLines() {
		size := 0
		for _, c := range n.Children {
			size += writtenTokens(c.Text)
		}
		catalog = append(catalog, fmt.Sprintf("- %s (%d paragraphs, about %d tokens)\n", n.Text, len(n.Children), size))
	}
	const everything = 1 << 30
	var index []DumpIndex
	for _, p := range set.Probes {
		qid := ws + "-" + p.ID
		if seenQ[qid] {
			return nil, fmt.Errorf("dump: two probes are both %s; their packets would overwrite each other", qid)
		}
		seenQ[qid] = true
		locs, srcAtoms, err := sourceSpans(p, byTurn)
		if err != nil {
			return nil, err
		}
		ix := DumpIndex{QID: qid, Workspace: set.Workspace, Kind: "one source", Type: p.Type, Sources: len(p.Sources), Gold: OpaqueID(meta.Salt, qid, "gold"),
			Packets: map[string]string{}, Tokens: map[string]int{}, Paragraphs: map[string]int{}, Status: map[string]string{}}
		if len(p.Sources) > 1 {
			ix.Kind = "whole conversation"
		}
		for _, as := range srcAtoms {
			var ids []string
			for _, r := range as {
				ids = append(ids, r.Atom.ID())
			}
			ix.SourceAtoms = append(ix.SourceAtoms, ids)
		}
		// what a packet costs before its first paragraph
		frame := writtenTokens(dumpPacket{}.render(p.Question, tag)) + writtenTokens(dumpExcerptsIntro)
		searchOrder := packet.SearchPick(turns, e.S, p.Question, next, everything)
		notesOrder := packet.NotesPick(turns, e.S, spread, notes, p.Question, next, everything, NoteWeight, TopicWeight, false)
		linesOf := func(ns []packet.Note) int {
			sum := writtenTokens(dumpNotesIntro)
			for _, n := range ns {
				sum += writtenTokens(noteLine(n))
			}
			return sum
		}
		allNotesSize := 0
		for _, arm := range arms {
			var d dumpPacket
			total := meta.Total
			if arm.Total > 0 {
				total = arm.Total
			}
			switch arm.Name {
			case ZoomSearch:
				d.picked = fitWritten(searchOrder, total-frame, tag)
			case ZoomTopics:
				d.picked = fitWritten(notesOrder, total-frame, tag)
			case ZoomView:
				// note lines in ranked order up to their share of the budget, shown oldest first
				share, spent := int(float64(total)*ViewShare)-writtenTokens(dumpNotesIntro), 0
				var idx []int
				for _, i := range packet.NoteOrder(notes, p.Question, TopicWeight) {
					if c := writtenTokens(noteLine(notes.Notes[i])); spent+c <= share {
						spent += c
						idx = append(idx, i)
					}
				}
				sort.Ints(idx)
				for _, i := range idx {
					d.notes = append(d.notes, notes.Notes[i])
				}
				room := total - frame
				if len(d.notes) > 0 {
					room -= linesOf(d.notes)
				}
				d.picked = fitWritten(notesOrder, room, tag)
			case dumpRecent:
				d.picked = fitWritten(newest, total-frame, tag)
			case dumpNone:
			case dumpOracle:
				seen := map[string]bool{}
				for _, as := range srcAtoms {
					for _, r := range as {
						if !seen[r.Atom.ID()] {
							seen[r.Atom.ID()] = true
							d.picked = append(d.picked, r)
						}
					}
				}
			case ZoomAllNotes:
				d.notes = notes.Notes
				d.picked = fitWritten(notesOrder, total-frame, tag)
				allNotesSize = writtenTokens(d.render(p.Question, tag))
			case dumpSearchBig:
				if allNotesSize == 0 {
					return nil, fmt.Errorf("dump: the %s arm must come after the all-notes arm", arm.Dir)
				}
				d.picked = fitWritten(searchOrder, allNotesSize-frame, tag)
			case dumpTitles:
				d.notes = notes.Notes
			case dumpCatalog:
				d.catalog = catalog
			case dumpFull:
				d.whole = turns
			default:
				return nil, fmt.Errorf("dump: no packet rule for arm %q", arm.Name)
			}
			text := d.render(p.Question, tag)
			id := OpaqueID(meta.Salt, qid, arm.Dir)
			ix.Packets[arm.Dir], ix.Tokens[arm.Dir], ix.Paragraphs[arm.Dir], ix.Status[arm.Dir] = id, writtenTokens(text), len(d.picked), sourceStatus(locs, d)
			if err := writeIn(filepath.Join(dir, "p", id), "packet.md", text); err != nil {
				return nil, err
			}
		}
		// the gold passages, for the judges: no turn ids, so an answer earns nothing by citing one
		var g strings.Builder
		fmt.Fprintf(&g, "# Question\n\n%s\n\n# Passages a complete answer draws on\n", strings.TrimSpace(p.Question))
		for i, src := range p.Sources {
			as := srcAtoms[i]
			t := as[0].Turn
			fmt.Fprintf(&g, "\n## Source %d\n\nNeeded words: \"%s\"\n\nThe paragraph they are in:\n\n%s\n", i+1, flat(src.Answer), vault.SpanText(t.Body[as[0].Atom.Start:as[len(as)-1].Atom.End]))
		}
		if err := writeIn(filepath.Join(dir, "g", ix.Gold), "gold.md", g.String()); err != nil {
			return nil, err
		}
		index = append(index, ix)
	}
	return index, nil
}

func writeIn(dir, name, text string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644)
}

// dumpArms are the arms a dump writes with the current settings.
func dumpArms() []DumpArm {
	arms := append([]DumpArm{}, DumpArms...)
	for _, b := range DumpSmall {
		arms = append(arms, DumpArm{Name: ZoomSearch, Dir: fmt.Sprintf("search-%d", b), Total: b}, DumpArm{Name: ZoomTopics, Dir: fmt.Sprintf("notes-%d", b), Total: b})
	}
	if DumpWhole {
		arms = append(arms, DumpArm{Name: dumpFull, Dir: "full"})
	}
	return arms
}

// DumpSummary says what the packets of each arm cost.
func DumpSummary(index []DumpIndex) string {
	sum, n := map[string]int{}, 0
	for _, ix := range index {
		n++
		for a, t := range ix.Tokens {
			sum[a] += t
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d probes; mean packet tokens as written, by arm:", n)
	var names []string
	for a := range sum {
		names = append(names, a)
	}
	order := map[string]int{}
	for i, a := range DumpArms {
		order[a.Dir] = i + 1
	}
	sort.Slice(names, func(i, j int) bool {
		if order[names[i]] != order[names[j]] {
			return order[names[i]] != 0 && (order[names[j]] == 0 || order[names[i]] < order[names[j]])
		}
		if len(names[i]) != len(names[j]) {
			return len(names[i]) < len(names[j])
		}
		return names[i] < names[j]
	})
	for _, a := range names {
		fmt.Fprintf(&b, " %s %d;", a, sum[a]/max(n, 1))
	}
	return strings.TrimSuffix(b.String(), ";") + "\n"
}

// PullLog records one excerpt request of the answer test.
type PullLog struct {
	Packet    string   `json:"packet"`
	QID       string   `json:"qid"`
	Arm       string   `json:"arm"`
	Requested []string `json:"requested"`
	Given     []string `json:"given"`
	LeftOut   []string `json:"left_out"`
	NotFound  []string `json:"not_found"`
	Tokens    int      `json:"tokens"`
	Budget    int      `json:"budget"`
}

// PullExcerpts is the second step of the answer test's pull arms. A model
// that was shown a packet with no paragraphs (the note lines, or the
// code-built turn lines) names the paragraphs it wants to read, most
// important first, by id: a paragraph (t000012#a3), a run of them
// (t000012#a3-5) or a whole turn (t000012). It is given as many as fit the
// dump's budget as written, in the order asked for, word for word and
// oldest first, exactly as a packet shows them. One request per packet and
// run: the request is logged to dir/pulls/<run>/<packet>.json and the
// excerpts are written beside it as <packet>.md, and a second request is
// refused, so what a pull cost can be audited afterwards. It reads only
// the dump, never a workspace. PullPath is where the excerpts are.
func PullExcerpts(dir, run, pkt string, ids []string) (string, error) {
	if run == "" || strings.ContainsAny(run, "/\\.") || pkt == "" || strings.ContainsAny(pkt, "/\\.") {
		return "", fmt.Errorf("pull: bad run %q or packet %q", run, pkt)
	}
	var index []DumpIndex
	if err := readJSON(filepath.Join(dir, "index.json"), &index); err != nil {
		return "", err
	}
	log := PullLog{Packet: pkt, Requested: ids}
	ws := ""
	for _, ix := range index {
		for arm, id := range ix.Packets {
			if id == pkt {
				ws, log.QID, log.Arm = filepath.Base(ix.Workspace), ix.QID, arm
			}
		}
	}
	if ws == "" {
		return "", fmt.Errorf("pull: no packet %q in %s", pkt, dir)
	}
	var meta DumpMeta
	if err := readJSON(filepath.Join(dir, "dump.json"), &meta); err != nil {
		return "", err
	}
	log.Budget = meta.Total
	var atoms []DumpAtom
	if err := readJSON(filepath.Join(dir, "atoms", ws+".json"), &atoms); err != nil {
		return "", err
	}
	byID, byTurn := map[string]int{}, map[string][]int{}
	for i, a := range atoms {
		byID[a.ID] = i
		byTurn[a.Turn] = append(byTurn[a.Turn], i)
	}
	var order []int
	seen := map[int]bool{}
	for _, id := range ids {
		id = strings.Trim(strings.TrimSpace(id), "[],;")
		ref, ok := vault.ParseRef(id)
		var want []int
		switch {
		case !ok || ref.Part > 0:
		case ref.Atom > 0:
			for k := ref.Atom; k <= ref.AtomEnd; k++ {
				if i, ok := byID[vault.AtomID(ref.Turn, k)]; ok {
					want = append(want, i)
				}
			}
		default:
			want = byTurn[ref.Turn]
		}
		if len(want) == 0 {
			if id != "" {
				log.NotFound = append(log.NotFound, id)
			}
			continue
		}
		for _, i := range want {
			if !seen[i] {
				seen[i] = true
				order = append(order, i)
			}
		}
	}
	if len(order) == 0 {
		return "", fmt.Errorf("pull: none of the ids is a paragraph or turn of this conversation; ids look like t000012#a3 or t000012")
	}
	var given []int
	for _, i := range order {
		if c := writtenTokens(renderExcerpt(atoms[i], meta.Tag)); log.Tokens+c <= meta.Total {
			log.Tokens += c
			given = append(given, i)
			log.Given = append(log.Given, atoms[i].ID)
		} else {
			log.LeftOut = append(log.LeftOut, atoms[i].ID)
		}
	}
	// the log is the lock: O_EXCL makes a second request for the packet fail
	logDir := filepath.Join(dir, "pulls", run)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return "", err
	}
	f, err := os.OpenFile(filepath.Join(logDir, pkt+".json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("pull: the excerpts for this packet were already given; there is one request per question")
		}
		return "", err
	}
	js, _ := json.MarshalIndent(log, "", " ")
	_, werr := f.Write(append(js, '\n'))
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return "", werr
	}
	sort.Ints(given) // atoms are stored in conversation order
	var b strings.Builder
	b.WriteString(strings.TrimPrefix(dumpExcerptsIntro, "\n"))
	for _, i := range given {
		b.WriteString(renderExcerpt(atoms[i], meta.Tag))
	}
	if len(log.LeftOut) > 0 {
		fmt.Fprintf(&b, "\nNot given, because the %d tokens allowed were spent: %s\n", meta.Total, strings.Join(log.LeftOut, ", "))
	}
	if len(log.NotFound) > 0 {
		fmt.Fprintf(&b, "\nNot a paragraph or turn of this conversation: %s\n", strings.Join(log.NotFound, ", "))
	}
	if err := os.WriteFile(PullPath(dir, run, pkt), []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return b.String(), nil
}

// PullPath is the file PullExcerpts writes a packet's excerpts to.
func PullPath(dir, run, pkt string) string { return filepath.Join(dir, "pulls", run, pkt+".md") }

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
