package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

func hasText(read []string, want string) bool {
	for _, r := range read {
		if strings.Contains(r, want) {
			return true
		}
	}
	return false
}

// A note counts only through the quotes code finds in the turn it cites.
// Each found quote is pinned to its paragraph; a note with no such quote is
// dropped; topic names take their canonical form.
func TestIdeaNotesAreCheckedAndPinned(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	turns, _ := h.e.V.Turns()
	notes := []IdeaNote{
		{Title: "The movers come at seven on 9 June", Kind: "Fact", Topics: []string{"Moving Day", "moving day"},
			Evidence: []NoteEvidence{{"t000003", "arrive on the ninth of  June"}, {"t000003", "a sentence nobody wrote in that turn"}}},
		{Title: "The van is booked for Friday", Kind: "fact", Evidence: []NoteEvidence{{"t000005", "the van was booked for Friday"}}},
		{Title: "Storage is unit 214", Kind: "fact", Topics: []string{"storage"},
			Evidence: []NoteEvidence{{"t000004", "storage unit is number 214"}, {"t000099", "no such turn at all"}}},
	}
	tm := &TopicMap{Map: map[string]string{"moving day": "move day", "storage": "storage unit"},
		Areas: map[string]string{"move day": "logistics", "storage unit": "logistics"}}
	checked, chk := CheckIdeaNotes(turns, notes, tm)
	if chk.Notes != 3 || chk.Kept != 2 || chk.Quotes != 5 || chk.Found != 2 || len(checked) != 2 {
		t.Fatalf("check %+v, kept %d", chk, len(checked))
	}
	n := checked[0]
	if strings.Join(n.Atoms, ",") != "t000003#a3" || n.Kind != "fact" || strings.Join(n.Topics, ",") != "move day" || strings.Join(n.Areas, ",") != "logistics" || n.Seq != 3 {
		t.Fatalf("first note %+v", n)
	}
	q := n.Quotes[0]
	if body := turns[2].Body; flat(body[q.Start:q.End]) != "arrive on the ninth of June" {
		t.Fatalf("the quote span is %q", body[q.Start:q.End])
	}
	if strings.Join(checked[1].Atoms, ",") != "t000004#a1" {
		t.Fatalf("second note pinned to %v", checked[1].Atoms)
	}
	if !strings.Contains(chk.String(), "2 of 5 quotes found") {
		t.Fatal(chk.String())
	}
}

// A note in plain words leads to a paragraph the question's own words do
// not match, and the topic of a matching note brings in the other notes
// filed under it. What is read is the verbatim paragraph.
func TestNotesReachParagraphsTheQuestionDoesNotMatch(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	turns, _ := h.e.V.Turns()
	next := h.e.nextTurn("").Seq
	checked, _ := CheckIdeaNotes(turns, []IdeaNote{
		{Title: "The removal crew gets here at seven on 9 June", Kind: "fact", Topics: []string{"relocation"}, Evidence: []NoteEvidence{{"t000003", "arrive on the ninth of June"}}},
		{Title: "The lock-up is number 214 at the Elm Street depot", Kind: "fact", Topics: []string{"relocation"}, Evidence: []NoteEvidence{{"t000004", "storage unit is number 214"}}},
	}, nil)
	ni := packet.NewNoteIndex(checked)
	x := packet.NewSpreadIndex(turns)
	const q = "When does the removal crew get here?"
	if hasText(packet.SearchRead(turns, h.e.S, q, next, 4000), movePlanted) {
		t.Fatal("the fixture is wrong: plain search already finds the paragraph")
	}
	read := packet.NotesRead(turns, h.e.S, x, ni, q, next, 4000, 1, 0)
	if !hasText(read, movePlanted) {
		t.Fatal("the note did not lead to its paragraph")
	}
	if hasText(read, "storage unit is number 214") {
		t.Fatal("without the topic lane, a note that does not match should bring nothing")
	}
	if read = packet.NotesRead(turns, h.e.S, x, ni, q, next, 4000, 1, 1); !hasText(read, "storage unit is number 214") || !hasText(read, movePlanted) {
		t.Fatal("the topic of the matching note should bring in the other note filed under it")
	}
	// a view shows the lines themselves, within its budget, in conversation order
	shown, spent := packet.NotesView(ni, q, 1000, 1)
	if len(shown) != 2 || shown[0] != 0 || spent > 1000 || spent != packet.NoteLineTokens(checked[0])+packet.NoteLineTokens(checked[1]) {
		t.Fatalf("view %v, %d tokens", shown, spent)
	}
	if shown, spent = packet.NotesView(ni, q, 5, 1); len(shown) != 0 || spent != 0 {
		t.Fatalf("a budget too small for a line showed %v", shown)
	}
}

func TestNotesVerdictNeedsThreeClearBudgets(t *testing.T) {
	cell := func(kind string, total, w, l int) ZoomCell {
		return ZoomCell{Kind: kind, Arm: ZoomTopics, Total: total, WinsS: w, LossesS: l, PS: exactSignP(w, l)}
	}
	win := []ZoomCell{cell("whole conversation", 4, 40, 15), cell("whole conversation", 8, 38, 16), cell("whole conversation", 16, 30, 12),
		cell("whole conversation", 32, 12, 10), cell("one source", 4, 4, 6)}
	if !strings.Contains(NotesVerdict(win, ZoomTopics), "Verdict: supported") {
		t.Fatal(NotesVerdict(win, ZoomTopics))
	}
	if !strings.Contains(NotesVerdict(win[1:], ZoomTopics), "Verdict: not supported") {
		t.Fatal("two clear budgets are not enough")
	}
	hurt := append(append([]ZoomCell{}, win...), cell("one source", 8, 2, 14))
	if !strings.Contains(NotesVerdict(hurt, ZoomTopics), "Verdict: not supported") {
		t.Fatal("a significant one-source loss fails the rule")
	}
	if NotesVerdict(win, ZoomNotes) != "" {
		t.Fatal("an arm with no cells has no verdict")
	}
}

// dumpHistory is a conversation long enough for a small budget to bind:
// 40 turns of three paragraphs. Three paragraphs, far apart, say how a ferry
// timetable changed; the third shares no word with the question. One long
// paragraph about the ferry is in the conversation twice, word for word.
func dumpHistory(t *testing.T, h *harness) {
	t.Helper()
	subjects := []string{"orchard", "harbour", "quarry", "meadow"}
	for i := 1; i <= 40; i++ {
		role := vault.RoleUser
		if i%2 == 0 {
			role = vault.RoleAssistant
		}
		var paras []string
		for k := 1; k <= 3; k++ {
			subj := subjects[(i+k)%4]
			paras = append(paras, fmt.Sprintf("Entry %d part %d on the %s: the crew walked the %s boundary again and wrote down what had changed since the last visit, which was not much beyond a broken gate and two missing markers near post %d.", i, k, subj, subj, i*10+k))
		}
		switch i {
		case 5:
			paras[1] = "The ferry timetable was settled in spring, after two long meetings with the people at the pier office: the first ferry leaves the north pier at six in the morning, and the last one comes back at nine in the evening."
		case 20:
			paras[0] = "After the storm the ferry timetable changed, as the pier office told everyone by letter in that same week: the last ferry now comes back at seven, because the pier lights were damaged and could not be mended quickly."
		case 33:
			paras[2] = "The pier lights were repaired in autumn by the same firm that had put them up long ago, and the late crossing at nine was restored for the winter months, to the relief of everyone on the evening shift."
		case 10, 30:
			paras[1] = "A reminder that was sent to everyone twice about the ferry: bring the season ticket, arrive ten minutes early, and keep the ferry timetable card in the glove box of the van, where the old card used to be kept."
		}
		if _, err := h.e.V.AppendTurn(role, "", "", "", strings.Join(paras, "\n\n")); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
}

var dumpProbe = ZoomProbe{ID: "q1", Type: "evolution", Question: "How did the ferry timetable change over the year?", Sources: []ZoomSource{
	{"t000005", "the last one comes back at nine in the evening"}, {"t000020", "the last ferry now comes back at seven"}, {"t000033", "the late crossing at nine was restored"}}}

// dumpNotes: none on the first needed passage; one pinned to the second's
// paragraph for other words; one quoting the third's own words; and fillers.
func dumpNotes() []IdeaNote {
	notes := []IdeaNote{
		{Title: "The ferry timetable lost its late crossing after storm damage to the pier lights", Kind: "change", Topics: []string{"ferry timetable"}, Evidence: []NoteEvidence{{"t000020", "because the pier lights were damaged"}}},
		{Title: "The ferry timetable got its nine o'clock last crossing back in autumn", Kind: "change", Topics: []string{"ferry timetable"}, Evidence: []NoteEvidence{{"t000033", "the late crossing at nine was restored for the winter"}}},
	}
	for i := 1; i <= 40; i += 3 {
		notes = append(notes, IdeaNote{Title: fmt.Sprintf("Boundary walk %d found a broken gate and two missing markers", i), Kind: "result", Topics: []string{"boundary walks"},
			Evidence: []NoteEvidence{{vault.TurnID(i), fmt.Sprintf("Entry %d part 1 on the", i)}}})
	}
	return notes
}

var excerptTag = regexp.MustCompile(`<excerpt turn="(t\d+)" by="\w+" paragraph="(\d+) of \d+" id="(t\d+#a\d+)">`)

// The answer test's packets. Every arm's packet is alone in a folder whose
// name says nothing, is filled to the budget as written, and holds what its
// arm is meant to hold; the index says so, and what each packet did with
// every needed passage; the gold passages are apart, without turn ids.
func TestDumpPacketsHoldWhatEachArmShould(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	dumpHistory(t, h)
	DumpSalt = "testsalt"
	defer func() { DumpSalt = "" }()
	const total = 1500
	set := ZoomSet{Workspace: "test", Name: "test/q", Probes: []ZoomProbe{dumpProbe}, Notes: dumpNotes()}
	dir := t.TempDir()
	index, err := DumpPackets(h.e.Cfg, []ZoomSet{set}, total, dir, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != 1 || index[0].QID != "test-q1" || index[0].Kind != "whole conversation" || index[0].Sources != 3 || len(index[0].Packets) != len(DumpArms) {
		t.Fatalf("index %+v", index)
	}
	ix := index[0]
	// the answer-test script computes folder names itself, by the same rule: these two values pin it
	if OpaqueID("testsalt", "test-q1", "search") != "f1090acd2cca0c6a" || OpaqueID("testsalt", "test-q1", "gold") != "9e28835f90630f10" {
		t.Fatalf("folder names changed: %s %s", OpaqueID("testsalt", "test-q1", "search"), OpaqueID("testsalt", "test-q1", "gold"))
	}
	if fmt.Sprint(ix.SourceAtoms) != "[[t000005#a2] [t000020#a1] [t000033#a3]]" {
		t.Fatalf("source paragraphs %v", ix.SourceAtoms)
	}
	text, ids := map[string]string{}, map[string][]string{}
	for _, arm := range DumpArms {
		id := ix.Packets[arm.Dir]
		if id != OpaqueID("testsalt", "test-q1", arm.Dir) || len(id) != 16 || strings.Contains(id, arm.Dir) {
			t.Fatalf("%s: folder %q", arm.Dir, id)
		}
		files, _ := os.ReadDir(filepath.Join(dir, "p", id))
		if len(files) != 1 || files[0].Name() != "packet.md" {
			t.Fatalf("%s: a packet's folder should hold the packet and nothing else, got %d files", arm.Dir, len(files))
		}
		b, err := os.ReadFile(filepath.Join(dir, "p", id, "packet.md"))
		if err != nil {
			t.Fatal(err)
		}
		p := string(b)
		text[arm.Dir] = p
		if !strings.HasPrefix(p, dumpPreamble) || !strings.HasSuffix(p, "\n## Question\n\nHow did the ferry timetable change over the year?\n") {
			t.Fatalf("%s: every packet opens and ends the same way:\n%.300s", arm.Dir, p)
		}
		for _, m := range excerptTag.FindAllStringSubmatch(p, -1) {
			ids[arm.Dir] = append(ids[arm.Dir], m[3])
		}
		if ix.Tokens[arm.Dir] != writtenTokens(p) || ix.Paragraphs[arm.Dir] != len(ids[arm.Dir]) || strings.Count(p, "</excerpt>") != len(ids[arm.Dir]) {
			t.Errorf("%s: index says %d tokens and %d paragraphs; the file has %d and %d", arm.Dir, ix.Tokens[arm.Dir], ix.Paragraphs[arm.Dir], writtenTokens(p), len(ids[arm.Dir]))
		}
		// oldest first, whatever the ranking was
		for i := 1; i < len(ids[arm.Dir]); i++ {
			a, _ := vault.ParseRef(ids[arm.Dir][i-1])
			c, _ := vault.ParseRef(ids[arm.Dir][i])
			if a.Turn > c.Turn || (a.Turn == c.Turn && a.Atom >= c.Atom) {
				t.Fatalf("%s: excerpts out of conversation order: %v", arm.Dir, ids[arm.Dir])
			}
		}
	}
	has := func(arm, id string) bool { return slices.Contains(ids[arm], id) }
	// nothing in a packet says which arm it is: the packets that hold paragraphs only are the same around them,
	// and the note and turn sections are where they belong and nowhere else
	block := regexp.MustCompile(`(?s)\n<excerpt .*?</excerpt>\n`)
	for _, arm := range []string{"notes", "recent", "oracle", "searchbig"} {
		if block.ReplaceAllString(text[arm], "") != block.ReplaceAllString(text["search"], "") {
			t.Errorf("the %s packet differs from the search packet outside its excerpts", arm)
		}
	}
	for _, arm := range DumpArms {
		notes, turns := arm.Dir == "lines" || arm.Dir == "allnotes" || arm.Dir == "titles", arm.Dir == "catalog"
		if strings.Contains(text[arm.Dir], "\n## Notes\n") != notes || strings.Contains(text[arm.Dir], "\n## Turns\n") != turns || strings.Contains(text[arm.Dir], "\n## Conversation\n") {
			t.Errorf("%s: wrong sections", arm.Dir)
		}
	}
	// a packet that does not hold a needed passage does not hold its words either
	for _, arm := range []string{"none", "recent"} {
		for _, src := range dumpProbe.Sources {
			if strings.Contains(text[arm], src.Answer) {
				t.Errorf("the %s packet holds the needed words %q", arm, src.Answer)
			}
		}
	}
	// the budget, as written: the four that select fill it to within one paragraph and never pass it
	for _, arm := range []string{"search", "notes", "lines", "recent"} {
		if got := ix.Tokens[arm]; got > total || got < total-130 {
			t.Errorf("%s holds %d tokens as written, want at most %d and within a paragraph of it", arm, got, total)
		}
	}
	if ix.Tokens["none"] >= 200 || len(ids["none"]) != 0 || strings.Contains(text["none"], "## Excerpts") || strings.Contains(text["none"], "## Notes") {
		t.Errorf("the question-only packet should hold no records:\n%s", text["none"])
	}
	// plain search: ranked paragraphs only; the repeated paragraph once; the passage with none of the question's words missing
	if strings.Contains(text["search"], "## Notes") || !has("search", "t000005#a2") || !has("search", "t000020#a1") || has("search", "t000033#a3") ||
		has("search", "t000010#a2") == has("search", "t000030#a2") || ix.Status["search"] != "hhn" {
		t.Errorf("search packet: %v, status %s", ids["search"], ix.Status["search"])
	}
	// notes choose paragraphs: the note brings the third passage in, and no note line is shown
	if strings.Contains(text["notes"], "## Notes") || !has("notes", "t000033#a3") || ix.Status["notes"] != "hhh" {
		t.Errorf("notes packet: %v, status %s", ids["notes"], ix.Status["notes"])
	}
	// lines: matching note lines within their share, oldest first, then fewer paragraphs than the notes packet
	lines := text["lines"]
	a, b := strings.Index(lines, "- [t000020#a1] (change)"), strings.Index(lines, "- [t000033#a3] (change)")
	if a < 0 || b < a || len(ids["lines"]) == 0 || len(ids["lines"]) >= len(ids["notes"]) || !has("lines", "t000033#a3") {
		t.Errorf("lines packet should show the two matching notes in order and fewer paragraphs (%d against %d):\n%.900s", len(ids["lines"]), len(ids["notes"]), lines)
	}
	if end := strings.Index(lines, "\n## Excerpts"); end < 0 || writtenTokens(lines[len(dumpPreamble):end]) > int(total*ViewShare) {
		t.Errorf("the note lines should keep to their share of the budget")
	}
	// recent: the newest paragraphs and nothing else
	if n := len(ids["recent"]); n < 8 || ids["recent"][n-1] != "t000040#a3" || ids["recent"][0] < "t000035" || ix.Status["recent"] != "nnn" {
		t.Errorf("recent packet: %v, status %s", ids["recent"], ix.Status["recent"])
	}
	// oracle: exactly the needed paragraphs
	if fmt.Sprint(ids["oracle"]) != "[t000005#a2 t000020#a1 t000033#a3]" || ix.Status["oracle"] != "hhh" || strings.Contains(text["oracle"], "## Notes") {
		t.Errorf("oracle packet: %v, status %s", ids["oracle"], ix.Status["oracle"])
	}
	// every note line, oldest first, with and without paragraphs
	titles, all := text["titles"], text["allnotes"]
	if strings.Count(titles, "\n- [") != 16 || len(ids["titles"]) != 0 || strings.Contains(titles, "## Excerpts") || !strings.Contains(titles, "- [t000001#a1] (result) Boundary walk 1 found") {
		t.Errorf("titles packet should be the 16 note lines and the question:\n%.600s", titles)
	}
	if strings.Index(titles, "Boundary walk 19 ") > strings.Index(titles, "- [t000020#a1]") || strings.Index(titles, "- [t000020#a1]") > strings.Index(titles, "Boundary walk 22 ") {
		t.Error("note lines should be in conversation order")
	}
	// no note on the first passage; one pinned to the second's paragraph for other words; one quoting the third's
	if ix.Status["titles"] != "npq" {
		t.Errorf("titles status %s, want npq", ix.Status["titles"])
	}
	if !strings.HasPrefix(all, titles[:strings.Index(titles, "\n## Question")]) || len(ids["allnotes"]) != len(ids["notes"]) || ix.Status["allnotes"] != "hhh" {
		t.Errorf("the all-notes packet is the titles packet plus the notes packet's paragraphs: %d against %d paragraphs", len(ids["allnotes"]), len(ids["notes"]))
	}
	// plain search at the same written size as all-notes: more than plain search, and no smaller than all-notes by more than a paragraph
	if d := ix.Tokens["allnotes"] - ix.Tokens["searchbig"]; d < 0 || d > 130 || len(ids["searchbig"]) <= len(ids["search"]) || strings.Contains(text["searchbig"], "## Notes") ||
		has("searchbig", "t000033#a3") || ix.Status["searchbig"] != "hhn" {
		t.Errorf("searchbig %d tokens against allnotes %d; %d paragraphs against search's %d", ix.Tokens["searchbig"], ix.Tokens["allnotes"], len(ids["searchbig"]), len(ids["search"]))
	}
	// the code-built catalog: a line per turn with its size, no paragraph
	cat := text["catalog"]
	if strings.Count(cat, "\n- t0000") != 40 || !strings.Contains(cat, "- t000005 user: ") || !strings.Contains(cat, "(3 paragraphs, about ") || len(ids["catalog"]) != 0 || ix.Status["catalog"] != "nnn" {
		t.Errorf("catalog packet:\n%.500s", cat)
	}
	// gold: each source's own words and paragraph, no turn ids, alone in its folder
	gold, err := os.ReadFile(filepath.Join(dir, "g", ix.Gold, "gold.md"))
	if err != nil || ix.Gold != OpaqueID("testsalt", "test-q1", "gold") {
		t.Fatal(err, ix.Gold)
	}
	g := string(gold)
	for i, want := range []string{"The ferry timetable was settled in spring", "After the storm the ferry timetable changed", "The pier lights were repaired in autumn"} {
		at := strings.Index(g, fmt.Sprintf("## Source %d\n\nNeeded words: \"%s\"", i+1, dumpProbe.Sources[i].Answer))
		next := strings.Index(g, fmt.Sprintf("## Source %d\n", i+2))
		if next < 0 {
			next = len(g)
		}
		if at < 0 || !strings.Contains(g[at:next], want) {
			t.Errorf("gold source %d lacks its words or its paragraph", i+1)
		}
	}
	if regexp.MustCompile(`t\d{6}`).MatchString(g) {
		t.Error("gold should name no turn: an answer must earn nothing by citing an id")
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, "p")); len(entries) != len(DumpArms) {
		t.Errorf("%d packet folders for %d arms", len(entries), len(DumpArms))
	}
	var meta DumpMeta
	if err := readJSON(filepath.Join(dir, "dump.json"), &meta); err != nil || meta.Total != total || meta.Salt != "testsalt" || meta.Tag != "excerpt" {
		t.Fatalf("dump.json %+v %v", meta, err)
	}
	if !strings.Contains(DumpSummary(index), "1 probes; mean packet tokens as written, by arm: search ") || strings.Contains(DumpSummary(index), "full") {
		t.Fatal(DumpSummary(index))
	}
	if _, err := os.Stat(filepath.Join(dir, "p", OpaqueID("testsalt", "test-q1", "full"))); err == nil {
		t.Fatal("the whole-conversation packet is written only when asked for")
	}
}

// What a dump refuses, and the whole-conversation packet.
func TestDumpPacketsRefusalsAndTheWholeConversation(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	dumpHistory(t, h)
	DumpSalt = "testsalt"
	defer func() { DumpSalt = "" }()
	set := ZoomSet{Workspace: "test", Name: "test/q", Probes: []ZoomProbe{dumpProbe}, Notes: dumpNotes()}
	if _, err := DumpPackets(h.e.Cfg, []ZoomSet{set, set}, 1500, t.TempDir(), func(string) {}); err == nil || !strings.Contains(err.Error(), "overwrite") {
		t.Fatalf("two probes with one id should be refused: %v", err)
	}
	lost := set
	lost.Probes = []ZoomProbe{{ID: "q2", Question: "What about the ferry?", Sources: []ZoomSource{{"t000005", "words that are in no turn at all"}, {"t000020", "the last ferry now comes back at seven"}}}}
	if _, err := DumpPackets(h.e.Cfg, []ZoomSet{lost}, 1500, t.TempDir(), func(string) {}); err == nil || !strings.Contains(err.Error(), "source 1 are not in turn t000005") {
		t.Fatalf("a needed passage that is not in its turn should stop the dump: %v", err)
	}
	bare := set
	bare.Notes = nil
	if _, err := DumpPackets(h.e.Cfg, []ZoomSet{bare}, 1500, t.TempDir(), func(string) {}); err == nil {
		t.Fatal("a dump needs notes")
	}
	DumpWhole, DumpSmall = true, []int{700}
	defer func() { DumpWhole, DumpSmall = false, nil }()
	dir := t.TempDir()
	index, err := DumpPackets(h.e.Cfg, []ZoomSet{set}, 1500, dir, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	// the two selecting arms again at a smaller budget: the same choice, cut sooner
	for _, arm := range []string{"search", "notes"} {
		small, big := index[0].Tokens[arm+"-700"], index[0].Tokens[arm]
		if small > 700 || small < 700-130 || big <= small || index[0].Paragraphs[arm+"-700"] >= index[0].Paragraphs[arm] || index[0].Paragraphs[arm+"-700"] < 3 {
			t.Errorf("%s at 700: %d tokens and %d paragraphs, against %d and %d at 1500", arm, small, index[0].Paragraphs[arm+"-700"], big, index[0].Paragraphs[arm])
		}
		b, err := os.ReadFile(filepath.Join(dir, "p", index[0].Packets[arm+"-700"], "packet.md"))
		if err != nil {
			t.Fatal(err)
		}
		full, _ := os.ReadFile(filepath.Join(dir, "p", index[0].Packets[arm], "packet.md"))
		for _, m := range excerptTag.FindAllStringSubmatch(string(b), -1) {
			if !strings.Contains(string(full), `id="`+m[3]+`"`) {
				t.Errorf("%s at 700 holds %s, which the larger packet does not", arm, m[3])
			}
		}
	}
	if index[0].Type != "evolution" || !strings.Contains(DumpSummary(index), " search-700 ") {
		t.Errorf("type %q, summary %s", index[0].Type, DumpSummary(index))
	}
	whole, err := os.ReadFile(filepath.Join(dir, "p", index[0].Packets["full"], "packet.md"))
	if err != nil {
		t.Fatal(err)
	}
	w := string(whole)
	if strings.Count(w, "\n<turn id=\"t0000") != 40 || !strings.Contains(w, "<turn id=\"t000020\" by=\"assistant\">\nAfter the storm the ferry timetable changed") || !strings.HasPrefix(w, dumpPreamble) ||
		!strings.HasSuffix(w, "over the year?\n") || index[0].Status["full"] != "hhh" || index[0].Tokens["full"] <= index[0].Tokens["allnotes"] || !strings.Contains(DumpSummary(index), " full ") {
		t.Fatalf("the whole-conversation packet:\n%.400s", w)
	}
	// a conversation that uses the excerpt tag itself gets another tag
	if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", "Someone pasted <excerpt turn=\"x\"> into the chat, and </excerpt> after it."); err != nil {
		t.Fatal(err)
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	dir2 := t.TempDir()
	if _, err := DumpPackets(h.e.Cfg, []ZoomSet{set}, 1500, dir2, func(string) {}); err != nil {
		t.Fatal(err)
	}
	var meta DumpMeta
	if err := readJSON(filepath.Join(dir2, "dump.json"), &meta); err != nil || meta.Tag != "excerpt-r" {
		t.Fatalf("tag %q %v", meta.Tag, err)
	}
}

// The pull arms' second step: the paragraphs a model names by id come back
// word for word, oldest first, within the dump's budget as written, once
// per packet and run, and the request is logged.
func TestPullExcerptsGivesWhatWasAskedOnceWithinTheBudget(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	dumpHistory(t, h)
	DumpSalt = "testsalt"
	defer func() { DumpSalt = "" }()
	set := ZoomSet{Workspace: "test", Name: "test/q", Probes: []ZoomProbe{dumpProbe}, Notes: dumpNotes()}
	dir := t.TempDir()
	index, err := DumpPackets(h.e.Cfg, []ZoomSet{set}, 1500, dir, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	pkt, other := index[0].Packets["titles"], index[0].Packets["catalog"]
	got, err := PullExcerpts(dir, "r1", pkt, []string{"t000033#a3,", "[t000005#a2]", "t000099#a1", "nonsense", "t000005#a2"})
	if err != nil {
		t.Fatal(err)
	}
	at5, at33 := strings.Index(got, `<excerpt turn="t000005" by="user" paragraph="2 of 3" id="t000005#a2">`), strings.Index(got, `id="t000033#a3">`)
	if !strings.HasPrefix(got, "## Excerpts\n") || at5 < 0 || at33 < at5 || !strings.Contains(got, "the last one comes back at nine in the evening.\n</excerpt>") || strings.Count(got, "</excerpt>") != 2 {
		t.Fatalf("pull:\n%s", got)
	}
	if !strings.Contains(got, "Not a paragraph or turn of this conversation: t000099#a1, nonsense") || strings.Contains(got, "Not given") {
		t.Fatalf("pull should say what it could not find, and only that:\n%s", got)
	}
	if onDisk, err := os.ReadFile(PullPath(dir, "r1", pkt)); err != nil || string(onDisk) != got {
		t.Fatalf("the excerpts should also be in the file a reader is pointed at: %v", err)
	}
	var log PullLog
	if err := readJSON(filepath.Join(dir, "pulls", "r1", pkt+".json"), &log); err != nil {
		t.Fatal(err)
	}
	if strings.Join(log.Given, ",") != "t000033#a3,t000005#a2" || len(log.Requested) != 5 || log.Budget != 1500 || log.Tokens < 100 || log.Tokens > 1500 || len(log.LeftOut) != 0 ||
		log.QID != "test-q1" || log.Arm != "titles" || log.Packet != pkt {
		t.Fatalf("log %+v", log)
	}
	if _, err := PullExcerpts(dir, "r1", pkt, []string{"t000003"}); err == nil || !strings.Contains(err.Error(), "already given") {
		t.Fatalf("a second request for the packet in the same run should be refused: %v", err)
	}
	// another packet of the same question has its own request, and so has another run
	if _, err := PullExcerpts(dir, "r1", other, []string{"t000003#a1"}); err != nil {
		t.Fatal(err)
	}
	// a turn id asks for the whole turn; a run of paragraphs for each of them
	whole, err := PullExcerpts(dir, "r2", pkt, []string{"t000003", "t000004#a1-2"})
	if err != nil || strings.Count(whole, `<excerpt turn="t000003" `) != 3 || !strings.Contains(whole, `id="t000004#a1">`) || !strings.Contains(whole, `id="t000004#a2">`) || strings.Count(whole, "</excerpt>") != 5 {
		t.Fatalf("a turn id should bring every paragraph of the turn: %v\n%s", err, whole)
	}
	for name, call := range map[string]func() (string, error){
		"no usable id":         func() (string, error) { return PullExcerpts(dir, "r3", pkt, []string{"nonsense", "t000099"}) },
		"unknown packet":       func() (string, error) { return PullExcerpts(dir, "r3", "0123456789abcdef", []string{"t000003"}) },
		"a question id":        func() (string, error) { return PullExcerpts(dir, "r3", "test-q1", []string{"t000003"}) },
		"a run that is a path": func() (string, error) { return PullExcerpts(dir, "../r3", pkt, []string{"t000003"}) },
	} {
		if _, err := call(); err == nil {
			t.Errorf("%s should be an error", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "pulls", "r3")); err == nil {
		t.Error("a refused request should leave no log")
	}
	// more than the budget: what fits, most wanted first, and the rest named
	var ask []string
	for i := 40; i >= 20; i-- {
		ask = append(ask, vault.TurnID(i))
	}
	part, err := PullExcerpts(dir, "r4", pkt, ask)
	if err != nil {
		t.Fatal(err)
	}
	if err := readJSON(filepath.Join(dir, "pulls", "r4", pkt+".json"), &log); err != nil {
		t.Fatal(err)
	}
	if log.Tokens > 1500 || log.Tokens < 1370 || log.Given[0] != "t000040#a1" || len(log.Given)+len(log.LeftOut) != 63 || len(log.LeftOut) == 0 ||
		!strings.Contains(part, "Not given, because the 1500 tokens allowed were spent: "+log.LeftOut[0]) || strings.Count(part, "</excerpt>") != len(log.Given) {
		t.Fatalf("the budget should cut the request, most wanted first: %d tokens, %d given, %d left out", log.Tokens, len(log.Given), len(log.LeftOut))
	}
	sum := 0
	for _, id := range log.Given {
		ref, _ := vault.ParseRef(id)
		turns, _ := h.e.V.Turns()
		tn := turns[0]
		for _, x := range turns {
			if x.ID == ref.Turn {
				tn = x
			}
		}
		atoms := vault.Atoms(tn.ID, tn.Body)
		sum += writtenTokens(renderExcerpt(DumpAtom{ID: id, Turn: tn.ID, Role: tn.Role, K: ref.Atom, Of: len(atoms), Text: atoms[ref.Atom-1].Text(tn.Body)}, "excerpt"))
	}
	if sum != log.Tokens {
		t.Fatalf("the log says %d tokens; the excerpts as written are %d", log.Tokens, sum)
	}
}

// A packet filled from any order holds a repeated paragraph once, and never
// more than its room as written.
func TestFitWrittenSkipsRepeatsAndKeepsToItsRoom(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	dumpHistory(t, h)
	turns, _ := h.e.V.Turns()
	var order []packet.RankedAtom
	for _, i := range []int{29, 9, 4, 19} { // turns 30 and 10 hold the same paragraph
		atoms := vault.Atoms(turns[i].ID, turns[i].Body)
		for _, a := range atoms {
			order = append(order, packet.RankedAtom{Turn: turns[i], Atom: a})
		}
	}
	var ids []string
	sum := 0
	for _, r := range fitWritten(order, 1<<30, "excerpt") {
		ids = append(ids, r.Atom.ID())
	}
	if got := strings.Join(ids, " "); !strings.Contains(got, "t000030#a2") || strings.Contains(got, "t000010#a2") || len(ids) != 11 {
		t.Fatalf("picked %s", got)
	}
	picked := fitWritten(order, 300, "excerpt")
	for _, r := range picked {
		sum += writtenTokens(renderExcerpt(dumpAtom(r), "excerpt"))
	}
	if len(picked) == 0 || len(picked) > 3 || sum > 300 {
		t.Fatalf("%d paragraphs, %d tokens in a room of 300", len(picked), sum)
	}
}

// The zoom test end to end: every arm has its cell, reports what it spent
// within the total, the walked arms report their ceiling, and -only limits
// the arms and refuses one it does not know.
func TestRunZoomReportsTokensAndControls(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	dumpHistory(t, h)
	set := ZoomSet{Workspace: "test", Name: "test/q", Probes: []ZoomProbe{dumpProbe}, Notes: dumpNotes()}
	cells, _, err := RunZoom(h.e.Cfg, []ZoomSet{set}, []int{1500}, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]ZoomCell{}
	for _, c := range cells {
		by[c.Arm] = c
	}
	if len(by) != len(zoomArms)+len(noteArms) {
		t.Fatalf("%d arms", len(by))
	}
	for arm, c := range by {
		if arm != ZoomAllNotes && (c.Used <= 0 || c.Used > 1500) {
			t.Errorf("%s used %.0f of 1500", arm, c.Used)
		}
		walked := arm == ZoomCatalog || arm == ZoomTerms || arm == ZoomLadder || arm == ZoomStatic || arm == ZoomWalkFree
		if c.Walked != walked || (walked && c.Ceiling < c.Recall) {
			t.Errorf("%s walked %v, ceiling %.2f, recall %.2f", arm, c.Walked, c.Ceiling, c.Recall)
		}
	}
	if by[ZoomAllNotes].Used <= 1500 {
		t.Errorf("the all-notes arm pays for its lines on top: %.0f", by[ZoomAllNotes].Used)
	}
	if !strings.Contains(ZoomReport(cells), "for this keyword reader") {
		t.Error("the ladder verdict should be limited to the keyword reader")
	}
	ZoomOnly = []string{ZoomCover}
	defer func() { ZoomOnly = nil }()
	if cells, _, err = RunZoom(h.e.Cfg, []ZoomSet{set}, []int{1500}, func(string) {}); err != nil || len(cells) != 3 {
		t.Fatalf("%d cells, %v", len(cells), err)
	}
	ZoomOnly = []string{"no such arm"}
	if _, _, err = RunZoom(h.e.Cfg, []ZoomSet{set}, []int{1500}, func(string) {}); err == nil {
		t.Fatal("an unknown arm should be an error")
	}
}

// The search variants add to search and take nothing away by themselves:
// with nothing borrowed or passed on they are search; every one keeps to its
// budget; coverage reads search's first paragraph and, when no question word
// is known, falls back to the ranking.
func TestSearchVariantsStartFromSearch(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	dumpHistory(t, h)
	turns, _ := h.e.V.Turns()
	next := h.e.nextTurn("").Seq
	x := packet.NewSpreadIndex(turns)
	const q, budget = "How did the ferry timetable change over the year?", 1200
	plain := packet.SearchRead(turns, h.e.S, q, next, budget)
	if got := packet.Texts(packet.FeedbackPick(turns, h.e.S, x, q, next, budget, 10, 0)); !slices.Equal(got, plain) {
		t.Error("feedback with no borrowed word should be search")
	}
	if got := packet.LinkRead(turns, h.e.S, x, q, next, budget, 20, 0); !slices.Equal(got, plain) {
		t.Error("links that pass nothing should be search")
	}
	size := func(read []string) (n int) {
		for _, r := range read {
			n += packet.ReadTokens(r)
		}
		return n
	}
	cover := packet.CoverRead(turns, h.e.S, x, q, next, budget, 300, CoverSpread)
	for name, read := range map[string][]string{"feedback": packet.FeedbackRead(turns, h.e.S, x, q, next, budget, 10, 8), "links": packet.LinkRead(turns, h.e.S, x, q, next, budget, 20, LinkLambda), "cover": cover} {
		if len(read) == 0 || size(read) > budget {
			t.Errorf("%s read %d paragraphs, %d tokens", name, len(read), size(read))
		}
	}
	if !slices.Contains(cover, plain[0]) {
		t.Error("coverage should read search's first paragraph")
	}
	if a, b := packet.FeedbackRead(turns, h.e.S, x, q, next, budget, 10, 8), packet.FeedbackRead(turns, h.e.S, x, q, next, budget, 10, 8); !slices.Equal(a, b) {
		t.Error("feedback should be repeatable")
	}
}
