package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// ZoomSource is one passage a question needs.
type ZoomSource struct {
	Turn   string `json:"turn"`
	Answer string `json:"answer"`
}

// ZoomProbe is a question with the passages that answer it. A file of
// ordinary probes (source and answer) also reads as one-source questions.
type ZoomProbe struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Question string       `json:"question"`
	Sources  []ZoomSource `json:"sources"`
	Source   string       `json:"source"`
	Answer   string       `json:"answer"`
}

// ReadZoomProbes reads one probe per JSON line.
func ReadZoomProbes(path string) ([]ZoomProbe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []ZoomProbe
	for i, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var p ZoomProbe
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, i+1, err)
		}
		if len(p.Sources) == 0 && p.Source != "" {
			p.Sources = []ZoomSource{{p.Source, p.Answer}}
		}
		out = append(out, p)
	}
	return out, nil
}

// ZoomSet is one probe file on one workspace, with the idea notes written
// for that workspace, if any.
type ZoomSet struct {
	Workspace, Name string
	Probes          []ZoomProbe
	Notes           []IdeaNote
	Topics          *TopicMap
}

// The arms of the zoom test. All are given the same total tokens, but not
// the same reader. The catalogs and ladders get up to a quarter for the
// overview (a ladder always shows every root, which can be more) and are
// followed by ZoomWalk, a keyword reader that opens only lines sharing a
// word with the question, ranks paragraphs by summed rarity, and stops when
// no such line is left, so these arms often spend less than the total.
// Search and the note arms read paragraphs in the search ranking; note
// lines are shown, not followed. The two control arms separate these
// effects, and each cell reports the tokens actually used.
const (
	ZoomCatalog = "catalog"       // today's previews, older turns in digest blocks
	ZoomTerms   = "catalog+terms" // the same lines the ladder's turns carry, flat: the control
	ZoomLadder  = "ladder"        // blurred blocks, zoomed toward the question
	ZoomStatic  = "ladder-static" // blurred blocks, zoomed toward the newest
	ZoomSearch  = "search"        // no overview: paragraphs in ranked order
	// controls: the walker's ranking with every paragraph open to it, and
	// the ladder's lines read the way note lines are
	ZoomWalkFree     = "control: walker ranking, no lines"
	ZoomLadderSearch = "ladder lines+search"
	// experimental readers for answers spread over the conversation
	ZoomFeedback = "search+feedback"   // the first hits' distinctive words are searched too
	ZoomLinks    = "search+links"      // the first hits pass score to paragraphs sharing rare words
	ZoomCover    = "search+cover"      // picks for coverage of the question and of the timeline
	ZoomSearchNb = "search+neighbours" // the paragraphs next to the best hits are read too
	// model-written idea notes, checked by code, as keys to paragraphs
	ZoomNotes    = "search+notes"                   // notes that match pass their rank to the paragraphs they quote
	ZoomTopics   = "search+notes+topics"            // and the topics of the best notes bring in their other notes
	ZoomTopicsNb = "search+notes+topics+neighbours" // and the paragraphs next to the best hits
	ZoomView     = "note lines+search"              // part of the budget shows matching note lines, the rest reads paragraphs
	// ZoomAllNotes shows every note line and spends the whole budget on
	// paragraphs besides: the conversation at note resolution. It costs
	// more than the other arms (the note lines come on top), so it is
	// reported apart and is in no verdict.
	ZoomAllNotes = "all note lines, plus search"
)

var zoomArms = []string{ZoomCatalog, ZoomTerms, ZoomLadder, ZoomStatic, ZoomSearch, ZoomWalkFree, ZoomLadderSearch, ZoomFeedback, ZoomLinks, ZoomCover, ZoomSearchNb}

// ZoomOnly, when set, limits a run to these arms, plus search and the flat
// lines every arm is compared with.
var ZoomOnly []string
var noteArms = []string{ZoomNotes, ZoomTopics, ZoomTopicsNb, ZoomView, ZoomAllNotes}

// Settings of the experimental readers, for `cx zoom` to vary.
var (
	FeedbackSeeds, FeedbackTerms = 10, 16 // the terms were chosen on the development conversation
	LinkSeeds                    = 20
	// LinkLambda was 1 at first: a first hit then handed out several times
	// its own score, which pushed search's best paragraphs out of small
	// budgets (the one-source loss at 4,000 tokens). A quarter, chosen on
	// the development conversation, removes that loss.
	LinkLambda  = 0.25
	CoverPool   = 300
	CoverSpread = 0.15
	NoteWeight  = 1.0
	TopicWeight = 1.0
	ViewShare   = 0.5 // share of the budget the view arm spends on note lines
)

// ZoomCell is one arm at one total budget for one kind of question.
type ZoomCell struct {
	Kind      string  `json:"kind"`
	Arm       string  `json:"arm"`
	Total     int     `json:"total_tokens"`
	Questions int     `json:"questions"`
	Recall    float64 `json:"mean_source_recall"`
	Used      float64 `json:"mean_tokens_used"`
	// Ceiling is, for an arm the keyword reader follows, the share of
	// passages it reaches with no budget limit: what its rule for opening
	// lines allows at all.
	Ceiling  float64 `json:"mean_recall_without_budget,omitempty"`
	Walked   bool    `json:"walked,omitempty"`
	AllFound int     `json:"questions_with_every_source"`
	// Pointed and Paragraph count a needed passage as reached when it was
	// read, or when a note line shown has a checked quote overlapping the
	// passage's own words (Pointed), or is pinned to its paragraph
	// (Paragraph). They say a checked pointer was delivered, not the text.
	Pointed   float64 `json:"mean_recall_or_note_on_same_words"`
	Paragraph float64 `json:"mean_recall_or_note_on_same_paragraph"`
	WinsC     int     `json:"wins_vs_control"`
	LossesC   int     `json:"losses_vs_control"`
	PC        float64 `json:"p_vs_control"`
	WinsS     int     `json:"wins_vs_search"`
	LossesS   int     `json:"losses_vs_search"`
	PS        float64 `json:"p_vs_search"`
	perQ      map[string]float64
}

// srcLoc is where a needed passage is: its bytes in its turn, and the
// paragraphs holding them.
type srcLoc struct {
	turn       string
	start, end int
	atoms      map[string]bool
	ok         bool
}

// sourceRecall is the share of a probe's passages in what was read, and
// the shares when a shown note on the same words, or on the same
// paragraph, also counts.
func sourceRecall(p ZoomProbe, read []string, locs []srcLoc, shown []packet.Note) (strict float64, all bool, pointed, paragraph float64) {
	var text []string
	for _, r := range read {
		text = append(text, norm(r))
	}
	found, nPointed, nPara := 0, 0, 0
	for i, s := range p.Sources {
		a := norm(s.Answer)
		hit := false
		for _, t := range text {
			if strings.Contains(t, a) {
				hit = true
				break
			}
		}
		words, para := hit, hit
		if !hit && i < len(locs) && locs[i].ok {
			for _, n := range shown {
				for _, q := range n.Quotes {
					if q.Turn == locs[i].turn && q.Start < locs[i].end && locs[i].start < q.End {
						words = true
					}
				}
				for _, id := range n.Atoms {
					if locs[i].atoms[id] {
						para = true
					}
				}
			}
		}
		if hit {
			found++
		}
		if words {
			nPointed++
		}
		if para {
			nPara++
		}
	}
	n := float64(len(p.Sources))
	return float64(found) / n, found == len(p.Sources), float64(nPointed) / n, float64(nPara) / n
}

// RunZoom runs every arm. Questions with several sources and with one are
// reported apart. The lines it also returns say what checking the notes
// found, and how far the needed passages are from plain search and from
// the notes. Use scratch copies: opening may update the index.
func RunZoom(cfg config.Config, sets []ZoomSet, totals []int, logf func(string)) ([]ZoomCell, []string, error) {
	arms := zoomArms
	withNotes := len(sets) > 0
	for _, set := range sets {
		withNotes = withNotes && len(set.Notes) > 0
	}
	if withNotes {
		arms = append(append([]string{}, zoomArms...), noteArms...)
	}
	if len(ZoomOnly) > 0 {
		want := map[string]bool{ZoomSearch: true, ZoomTerms: true}
		for _, a := range ZoomOnly {
			want[a] = true
		}
		var kept []string
		for _, a := range arms {
			if want[a] {
				kept = append(kept, a)
				delete(want, a)
			}
		}
		for a := range want {
			return nil, nil, fmt.Errorf("zoom: no arm %q", a)
		}
		arms = kept
	}
	var info []string
	type diag struct{ sources, ranked, pinned, noteTurn int }
	diags := map[string]*diag{}
	checkedWS := map[string]bool{}
	cells := map[string]*ZoomCell{}
	cell := func(kind, arm string, total int) *ZoomCell {
		k := fmt.Sprint(kind, "|", arm, "|", total)
		if cells[k] == nil {
			cells[k] = &ZoomCell{Kind: kind, Arm: arm, Total: total, perQ: map[string]float64{}}
		}
		return cells[k]
	}
	for _, set := range sets {
		logf("zoom: " + set.Name)
		e, err := OpenReadOnly(cfg, set.Workspace)
		if err != nil {
			return nil, nil, err
		}
		turns, err := e.V.Turns()
		if err != nil {
			e.Close()
			return nil, nil, err
		}
		next := e.nextTurn("").Seq
		preview, detailed := packet.BuildZoom(turns, false), packet.BuildZoom(turns, true)
		for _, total := range totals {
			if roots := packet.ViewTokens(detailed.Roots); total <= roots {
				info = append(info, fmt.Sprintf("%s: at %d total tokens the ladder's roots alone cost %d, so its arms can read nothing", set.Name, total, roots))
			}
		}
		spread := packet.NewSpreadIndex(turns)
		byTurn := map[string]vault.Turn{}
		for _, t := range turns {
			byTurn[t.ID] = t
		}
		var notes *packet.NoteIndex
		pinnedAtom, pinnedTurn := map[string]bool{}, map[string]bool{}
		if withNotes {
			checked, chk := CheckIdeaNotes(turns, set.Notes, set.Topics)
			notes = packet.NewNoteIndex(checked)
			for _, n := range checked {
				for _, id := range n.Atoms {
					pinnedAtom[id] = true
					if ref, ok := vault.ParseRef(id); ok {
						pinnedTurn[ref.Turn] = true
					}
				}
			}
			if !checkedWS[set.Workspace] {
				checkedWS[set.Workspace] = true
				lines := 0
				for _, n := range checked {
					lines += packet.NoteLineTokens(n)
				}
				info = append(info, fmt.Sprintf("notes for %s: %s; shown as lines with their ids, all of them cost %d tokens", set.Workspace, chk, lines))
			}
		}
		for _, p := range set.Probes {
			kind := "one source"
			if len(p.Sources) > 1 {
				kind = "whole conversation"
			}
			d := diags[kind]
			if d == nil {
				d = &diag{}
				diags[kind] = d
			}
			locs := make([]srcLoc, len(p.Sources))
			for i, src := range p.Sources {
				d.sources++
				if rank, _ := packet.AnswerRank(turns, e.S, spread, p.Question, next, src.Turn, norm(src.Answer)); rank > 0 {
					d.ranked++
				}
				if t, ok := byTurn[src.Turn]; ok {
					if start, end, ok := findQuote(t.Body, 0, len(t.Body), src.Answer); ok {
						locs[i] = srcLoc{turn: t.ID, start: start, end: end, atoms: map[string]bool{}, ok: true}
						a, b := atomSpan(t, start, end)
						pin := false
						for k := a; k <= b && k > 0; k++ {
							locs[i].atoms[vault.AtomID(t.ID, k)] = true
							pin = pin || pinnedAtom[vault.AtomID(t.ID, k)]
						}
						if pin {
							d.pinned++
						}
					}
					if pinnedTurn[t.ID] {
						d.noteTurn++
					}
				}
			}
			for _, total := range totals {
				over := total / 4
				for _, arm := range arms {
					var read []string
					var shown []packet.Note
					used, walked, ceiling := -1, false, 0.0
					walk := func(z *packet.ZoomTree, v []*packet.ZoomNode) []string {
						view := packet.ViewTokens(v)
						r, spent := z.ZoomWalkSpent(v, p.Question, total-view)
						all, _ := z.ZoomWalkSpent(v, p.Question, 1<<40)
						ceiling, _, _, _ = sourceRecall(p, all, locs, nil)
						used, walked = view+spent, true
						return r
					}
					overview := 0 // tokens of lines shown beside what is read
					switch arm {
					case ZoomCatalog:
						read = walk(preview, preview.CatalogView(over))
					case ZoomTerms:
						read = walk(detailed, detailed.CatalogView(over))
					case ZoomLadder:
						read = walk(detailed, detailed.LadderView(p.Question, over))
					case ZoomStatic:
						read = walk(detailed, detailed.LadderView("", over))
					case ZoomSearch:
						read = packet.SearchRead(turns, e.S, p.Question, next, total)
					case ZoomWalkFree:
						// paragraphs are paid for when read, so there is no view to pay for
						var all []string
						read, used = detailed.ZoomWalkSpent(detailed.Paragraphs(), p.Question, total)
						all, _ = detailed.ZoomWalkSpent(detailed.Paragraphs(), p.Question, 1<<40)
						ceiling, _, _, _ = sourceRecall(p, all, locs, nil)
						walked = true
					case ZoomLadderSearch:
						overview = packet.ViewTokens(detailed.LadderView(p.Question, over))
						read = packet.SearchRead(turns, e.S, p.Question, next, total-overview)
					case ZoomFeedback:
						read = packet.FeedbackRead(turns, e.S, spread, p.Question, next, total, FeedbackSeeds, FeedbackTerms)
					case ZoomLinks:
						read = packet.LinkRead(turns, e.S, spread, p.Question, next, total, LinkSeeds, LinkLambda)
					case ZoomCover:
						read = packet.CoverRead(turns, e.S, spread, p.Question, next, total, CoverPool, CoverSpread)
					case ZoomSearchNb:
						read = packet.Texts(packet.SearchNeighboursPick(turns, e.S, p.Question, next, total))
					case ZoomTopicsNb:
						read = packet.Texts(packet.NotesPick(turns, e.S, spread, notes, p.Question, next, total, NoteWeight, TopicWeight, true))
					case ZoomNotes:
						read = packet.NotesRead(turns, e.S, spread, notes, p.Question, next, total, NoteWeight, 0)
					case ZoomTopics:
						read = packet.NotesRead(turns, e.S, spread, notes, p.Question, next, total, NoteWeight, TopicWeight)
					case ZoomView:
						idx, spent := packet.NotesView(notes, p.Question, int(float64(total)*ViewShare), TopicWeight)
						for _, i := range idx {
							shown = append(shown, notes.Notes[i])
						}
						overview = spent
						read = packet.NotesRead(turns, e.S, spread, notes, p.Question, next, total-spent, NoteWeight, TopicWeight)
					case ZoomAllNotes:
						shown = notes.Notes
						for _, n := range shown {
							overview += packet.NoteLineTokens(n)
						}
						read = packet.NotesRead(turns, e.S, spread, notes, p.Question, next, total, NoteWeight, TopicWeight)
					}
					if used < 0 {
						used = overview
						for _, r := range read {
							used += packet.ReadTokens(r)
						}
					}
					rec, all, pointed, para := sourceRecall(p, read, locs, shown)
					c := cell(kind, arm, total)
					c.Questions++
					c.Recall += rec
					c.Used += float64(used)
					c.Ceiling += ceiling
					c.Walked = walked
					c.Pointed += pointed
					c.Paragraph += para
					if all {
						c.AllFound++
					}
					c.perQ[set.Name+"|"+p.ID] = rec
				}
			}
		}
		e.Close()
	}
	var out []ZoomCell
	seen := map[string]bool{}
	for _, set := range sets {
		for _, p := range set.Probes {
			kind := "one source"
			if len(p.Sources) > 1 {
				kind = "whole conversation"
			}
			if seen[kind] {
				continue
			}
			seen[kind] = true
			for _, total := range totals {
				for _, arm := range arms {
					c := *cell(kind, arm, total)
					ctl, srch := cell(kind, ZoomTerms, total), cell(kind, ZoomSearch, total)
					c.WinsC, c.LossesC, c.PC = pairedSign(c.perQ, ctl.perQ)
					c.WinsS, c.LossesS, c.PS = pairedSign(c.perQ, srch.perQ)
					if c.Questions > 0 {
						c.Recall /= float64(c.Questions)
						c.Used /= float64(c.Questions)
						c.Ceiling /= float64(c.Questions)
						c.Pointed /= float64(c.Questions)
						c.Paragraph /= float64(c.Questions)
					}
					out = append(out, c)
				}
			}
		}
	}
	for _, kind := range []string{"whole conversation", "one source"} {
		if d := diags[kind]; d != nil && d.sources > 0 {
			line := fmt.Sprintf("%s: %d needed passages; plain search ranks %d (%d%%) at any depth", kind, d.sources, d.ranked, 100*d.ranked/d.sources)
			if withNotes {
				line += fmt.Sprintf("; a checked note points at the paragraph of %d (%d%%), and at the turn of %d (%d%%)", d.pinned, 100*d.pinned/d.sources, d.noteTurn, 100*d.noteTurn/d.sources)
			}
			info = append(info, line)
		}
	}
	return out, info, nil
}

// NotesVerdict applies the rule fixed before the held-out run: notes as
// keys are worth building into the engine only if, on whole-conversation
// questions of a conversation not used to choose their settings, they find
// more needed passages than plain search, significantly (p below 0.05), at
// three or more of the four budgets, with no significant loss on
// one-source questions at any budget.
func NotesVerdict(cells []ZoomCell, arm string) string {
	wins, sigLoss, budgets := 0, 0, 0
	for _, c := range cells {
		if c.Arm != arm {
			continue
		}
		switch c.Kind {
		case "whole conversation":
			budgets++
			if c.WinsS > c.LossesS && c.PS < 0.05 {
				wins++
			}
		case "one source":
			if c.LossesS > c.WinsS && c.PS < 0.05 {
				sigLoss++
			}
		}
	}
	if budgets == 0 {
		return ""
	}
	verdict := "not supported"
	if wins >= 3 && sigLoss == 0 {
		verdict = "supported"
	}
	return fmt.Sprintf("%s against plain search: significantly ahead at %d of %d budgets on whole-conversation questions; significant one-source losses at %d. Verdict: %s.\n", arm, wins, budgets, sigLoss, verdict)
}

func pairedSign(a, b map[string]float64) (wins, losses int, p float64) {
	for id, x := range a {
		if y, ok := b[id]; ok {
			if x > y {
				wins++
			} else if x < y {
				losses++
			}
		}
	}
	return wins, losses, exactSignP(wins, losses)
}

// ZoomVerdict applies the rule fixed before the test was run: the ladder is
// worth building into packets only if, on whole-conversation questions, it
// finds more sources than the same lines laid out flat at three or more of
// the four budgets, significantly (p below 0.05) at those budgets, and does
// not lose significantly on one-source questions at any budget.
func ZoomVerdict(cells []ZoomCell) string {
	wins, sigLoss, budgets := 0, 0, 0
	for _, c := range cells {
		if c.Arm != ZoomLadder {
			continue
		}
		switch c.Kind {
		case "whole conversation":
			budgets++
			if c.WinsC > c.LossesC && c.PC < 0.05 {
				wins++
			}
		case "one source":
			if c.LossesC > c.WinsC && c.PC < 0.05 {
				sigLoss++
			}
		}
	}
	verdict := "not supported for this keyword reader: keep the ladder out of packets. No model has read a ladder, and this test cannot say what one would do with the same lines"
	if budgets > 0 && wins >= 3 && wins*4 >= budgets*3 && sigLoss == 0 {
		verdict = "supported: the ladder earns a live test"
	}
	return fmt.Sprintf("Rule fixed in advance: ladder significantly ahead of flat lines at 3 or more of 4 budgets on whole-conversation questions, with no significant loss on one-source questions.\nLadder significantly ahead at %d of %d budgets; significant one-source losses at %d. Verdict: %s.\n", wins, budgets, sigLoss, verdict)
}

// ZoomReport renders the cells.
func ZoomReport(cells []ZoomCell) string {
	var b strings.Builder
	b.WriteString("| Questions | Total tokens | Arm | n | Tokens used | Passages read | This reader with no budget | Or a shown note quotes the same words | Or a shown note is on the same paragraph | Every passage read | Against flat lines (won, lost, p) | Against search (won, lost, p) |\n|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, c := range cells {
		ceiling := "-"
		if c.Walked {
			ceiling = fmt.Sprintf("%.1f%%", 100*c.Ceiling)
		}
		fmt.Fprintf(&b, "| %s | %d | %s | %d | %.0f | %.1f%% | %s | %.1f%% | %.1f%% | %d | %d, %d, %.3f | %d, %d, %.3f |\n", c.Kind, c.Total, c.Arm, c.Questions, c.Used, 100*c.Recall, ceiling, 100*c.Pointed, 100*c.Paragraph, c.AllFound,
			c.WinsC, c.LossesC, c.PC, c.WinsS, c.LossesS, c.PS)
	}
	b.WriteString("\n" + ZoomVerdict(cells))
	// the all-notes arm costs more than the others, so no verdict applies to it
	for _, arm := range []string{ZoomFeedback, ZoomLinks, ZoomCover, ZoomSearchNb, ZoomNotes, ZoomTopics, ZoomTopicsNb, ZoomView} {
		if v := NotesVerdict(cells, arm); v != "" {
			b.WriteString("Rule fixed in advance, on a conversation not used to choose settings: significantly ahead of plain search at 3 or more of 4 budgets, with no significant one-source loss. " + v)
		}
	}
	return b.String()
}

// WriteZoom writes zoom.md and zoom.json into dir.
func WriteZoom(dir string, cells []ZoomCell) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	js, _ := json.MarshalIndent(cells, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "zoom.json"), append(js, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "zoom.md"), []byte(ZoomReport(cells)), 0o644)
}
