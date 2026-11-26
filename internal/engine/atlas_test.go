package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// atlasConversation appends n filler turns, with the given bodies at the
// given sequence numbers, and rebuilds the index.
func atlasConversation(t *testing.T, h *harness, n int, special map[int]string) {
	t.Helper()
	for i := 1; i <= n; i++ {
		body := fmt.Sprintf("Filler turn %d about the garden beds and the compost.", i)
		if s, ok := special[i]; ok {
			body = s
		}
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
}

const (
	atlasPara1 = "We walked the whole fence line with the surveyor on Tuesday and marked every post that leans. There are fourteen of them between the orchard corner and the lane, most on the low side where the ground stays wet into June."
	atlasPara2 = "The decision is to replace the north gate with a steel one, because the oak frame rotted through at the hinge side. Dana asked for the same width as before so the trailer still clears it, and for the latch to sit on the lane side."
	atlasPara3 = "The quote from Halvorsen came to 1,840 for the gate, fitted, with a ten year warranty on the galvanizing. Delivery is three weeks from the deposit, and the old posts stay unless they fail the load test on the day."
)

// A model reads the map, asks for the paragraphs a note points at, and
// gets them word for word: the path the answer test of 2026-11-25
// favoured, in the live engine.
func TestAtlasMapPointsAndPullReturnsParagraphs(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	atlasConversation(t, h, 40, map[int]string{5: atlasPara1 + "\n\n" + atlasPara2 + "\n\n" + atlasPara3})
	res, err := h.e.ImportNotes([]IdeaNote{
		{Title: "North gate is replaced with a steel one because the oak frame rotted at the hinge", Kind: "Decision", Topics: []string{"North Gate"},
			Evidence: []NoteEvidence{{Turn: "t000005", Quote: "replace the north gate with a\nsteel one"}}},
		{Title: "Halvorsen quoted 1,840 for the gate fitted", Kind: "fact", Topics: []string{"north gate", "costs"},
			Evidence: []NoteEvidence{{Turn: "t000005", Quote: "came to 1,840 for the gate"}}},
		{Title: "The surveyor recommended a concrete footing", Kind: "fact",
			Evidence: []NoteEvidence{{Turn: "t000005", Quote: "a concrete footing is needed"}}}, // not in the turn
		{Title: "A note about a turn that does not exist", Kind: "fact",
			Evidence: []NoteEvidence{{Turn: "t000999", Quote: "anything at all"}}},
		{Title: "", Kind: "fact", Evidence: []NoteEvidence{{Turn: "t000005", Quote: "marked every post"}}}, // no title
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Read != 5 || res.Kept != 2 || res.New != 2 || res.Dropped != 3 || res.Quotes != 5 || res.Found != 3 {
		t.Fatalf("import: %+v", res)
	}
	// importing the same notes again changes nothing
	if again, _ := h.e.ImportNotes([]IdeaNote{{Title: "Halvorsen quoted 1,840 for the gate fitted", Kind: "fact",
		Evidence: []NoteEvidence{{Turn: "t000005", Quote: "came to 1,840 for the gate"}}}}, nil); again.Kept != 1 || again.New != 0 {
		t.Fatalf("second import: %+v", again)
	}
	notes := h.e.V.Notes()
	if len(notes) != 2 {
		t.Fatalf("%d notes stored", len(notes))
	}
	for _, n := range notes {
		ev := n.Evidence[0]
		turn, _ := h.e.V.Turn(ev.Turn)
		if turn.Body[ev.Start:ev.End] != ev.Quote || ev.BodySHA != turn.BodySHA {
			t.Errorf("note %s: the stored quote is not the bytes of the turn it cites", n.ID)
		}
	}

	round := 0
	h.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		round++
		if round == 1 {
			for _, want := range []string{
				"## Map of the conversation",
				"- [t000005#a2] (decision) North gate is replaced with a steel one because the oak frame rotted at the hinge",
				"- [t000005#a3] (fact) Halvorsen quoted 1,840 for the gate fitted",
				"One request brings back about 8000 tokens",
			} {
				if !strings.Contains(c.Brief, want) {
					t.Errorf("round 1 brief lacks %q", want)
				}
			}
			if strings.Contains(c.Brief, "concrete footing") || strings.Contains(c.Brief, "does not exist") {
				t.Errorf("a note with no quote found reached the map")
			}
			if strings.Contains(c.Brief, "notes that did not fit here") {
				t.Errorf("the map says notes were folded when all fit")
			}
			if i, j := strings.Index(c.Brief, "## Map of the conversation"), strings.Index(c.Brief, "## Catalog"); i < 0 || j < 0 || i > j {
				t.Errorf("the map should come before the catalog (map at %d, catalog at %d)", i, j)
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusNeedsRecall, RecallIDs: []string{"t000005#a2", "t000005#a3"}}
		}
		if !strings.Contains(c.Brief, atlasPara2) || !strings.Contains(c.Brief, atlasPara3) {
			t.Errorf("the paragraphs asked for are not in round 2 word for word")
		}
		if strings.Contains(c.Brief, "names paragraph 2 of 3 only") {
			t.Errorf("a pull should not get a hint line per paragraph")
		}
		return done(c, "Steel, 1,840 fitted.")
	}
	r, err := h.e.Ask(context.Background(), "What did we settle about the paddock entrance and what will it cost?")
	if err != nil {
		t.Fatal(err)
	}
	if round != 2 || r.Assistant.Body != "Steel, 1,840 fitted." {
		t.Fatalf("round %d answer %q", round, r.Assistant.Body)
	}
}

// One request is bounded by tokens: ids are served in the order asked, one
// that does not fit is passed over for a smaller one, and the reader is
// told which were left out.
func TestAtlasPullIsBoundedByTokens(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	long := strings.Repeat("The drainage survey lists every culvert along the north wall with its diameter and condition. ", 12)
	atlasConversation(t, h, 30, map[int]string{
		4: atlasPara2,
		6: strings.TrimSpace(long),
		8: atlasPara3,
	})
	if _, err := h.e.ImportNotes([]IdeaNote{{Title: "North gate becomes steel", Kind: "decision",
		Evidence: []NoteEvidence{{Turn: "t000004", Quote: "replace the north gate"}}}}, nil); err != nil {
		t.Fatal(err)
	}
	h.e.Cfg.AtlasPullTokens = 160
	focus := vault.Turn{Seq: 31}
	add, notes := h.e.recallRequest([]string{"t000004#a1", "t000006", "t000008#a1", "[t000004#a1],"}, nil, focus, nil)
	if strings.Join(add, " ") != "t000004#a1 t000008#a1" {
		t.Fatalf("put back %v, notes %v", add, notes)
	}
	joined := strings.Join(notes, "\n")
	if !strings.Contains(joined, "Not put back, because the 160 tokens one request brings back were spent") || !strings.Contains(joined, "t000006") {
		t.Errorf("the reader is not told what was left out: %v", notes)
	}
	if !strings.Contains(joined, "t000004#a1 was already put back") {
		t.Errorf("an id written with brackets should be read as the id: %v", notes)
	}
	// what a search finds is paid for from the same tokens
	add, notes = h.e.recallRequest(nil, []string{"drainage survey culvert diameter"}, focus, nil)
	if joined := strings.Join(notes, "\n"); len(add) != 0 || !strings.Contains(joined, "t000006") || !strings.Contains(joined, "not put back, because the 160 tokens") {
		t.Errorf("a search hit larger than the request's tokens: put back %v, notes %v", add, notes)
	}
	h.e.Cfg.AtlasPullTokens = 8000
	if add, notes = h.e.recallRequest(nil, []string{"drainage survey culvert diameter"}, focus, nil); len(add) == 0 || !strings.HasPrefix(add[0], "t000006") {
		t.Errorf("a search hit that fits: put back %v, notes %v", add, notes)
	}
	// without notes there is no map, and a request is bounded by count as before
	h2 := newHarness(t, 200_000, nil)
	atlasConversation(t, h2, 30, map[int]string{6: strings.TrimSpace(long)})
	h2.e.Cfg.AtlasPullTokens = 120
	if add, _ := h2.e.recallRequest([]string{"t000006"}, nil, focus, nil); len(add) != 1 {
		t.Errorf("a workspace with no notes should recall as before, got %v", add)
	}
}

// With atlas on, the indexer is asked for idea notes and shown the topics
// in use; what it writes is kept only through quotes found in the batch.
func TestIndexerWritesCheckedIdeaNotes(t *testing.T) {
	var briefs []string
	h := newHarness(t, 200_000, nil)
	h.e.Cfg.Atlas = config.AtlasOn
	h.reply = func(c call) any {
		if c.Kind != packet.KindIndex {
			return done(c, "Noted.")
		}
		briefs = append(briefs, c.Brief)
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed",
			Cards: []packet.IndexCard{{Turn: "t000001", Summary: "The north gate decision."}},
			Ideas: []packet.IdeaOp{
				{Title: "North gate becomes steel because the oak frame rotted", Kind: "decision", Topics: []string{"North gate"},
					Evidence: []packet.IdeaEvidence{{Turn: "t000001", Quote: "replace the north gate with a steel one"}}},
				{Title: "The budget was doubled", Kind: "change",
					Evidence: []packet.IdeaEvidence{{Turn: "t000001", Quote: "the budget is doubled"}}},
			}}
	}
	if _, err := h.e.Ask(context.Background(), atlasPara2); err != nil {
		t.Fatal(err)
	}
	h.e.Wait()
	if len(briefs) == 0 {
		t.Fatal("no index run")
	}
	for _, want := range []string{"- ideas: notes for the map of the conversation", "Existing note topics"} {
		if !strings.Contains(briefs[0], want) {
			t.Errorf("index brief lacks %q", want)
		}
	}
	notes := h.e.V.Notes()
	if len(notes) != 1 || notes[0].Title != "North gate becomes steel because the oak frame rotted" || notes[0].Topics[0] != "north gate" {
		t.Fatalf("notes stored: %+v", notes)
	}
	if got := notes[0].Atoms(); len(got) != 1 || got[0] != "t000001#a1" {
		t.Errorf("pinned to %v", got)
	}

	// the default asks the indexer for no notes
	briefs = nil
	h2 := newHarness(t, 200_000, nil)
	h2.reply = func(c call) any {
		if c.Kind == packet.KindIndex {
			briefs = append(briefs, c.Brief)
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		return done(c, "Noted.")
	}
	if _, err := h2.e.Ask(context.Background(), atlasPara2); err != nil {
		t.Fatal(err)
	}
	h2.e.Wait()
	if len(briefs) == 0 || strings.Contains(briefs[0], "notes for the map of the conversation") || strings.Contains(briefs[0], "Existing note topics") {
		t.Errorf("atlas auto should not ask the indexer for notes")
	}
}

// A packet never shows a note about a turn at or after the message it is
// built for, a packet that holds the whole conversation has no map, and
// atlas off shows none.
func TestAtlasMapOnlyWhereItBelongs(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	atlasConversation(t, h, 40, map[int]string{5: atlasPara2, 30: atlasPara3})
	if _, err := h.e.ImportNotes([]IdeaNote{
		{Title: "North gate becomes steel", Kind: "decision", Evidence: []NoteEvidence{{Turn: "t000005", Quote: "replace the north gate"}}},
		{Title: "Halvorsen quoted 1,840", Kind: "fact", Evidence: []NoteEvidence{{Turn: "t000030", Quote: "came to 1,840"}}},
	}, nil); err != nil {
		t.Fatal(err)
	}
	notes := h.e.V.Notes()
	items, st := packet.BuildMap(notes, nil, 30, "gate", 40000, 8000)
	if st.Shown != 1 || len(items) != 2 || strings.Contains(items[0].Text+items[1].Text, "1,840") {
		t.Fatalf("a packet for turn 30 shows a note about turn 30: %+v", st)
	}
	sized, err := h.e.DryPacket("What did the gate cost?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sized.Brief, "## Map of the conversation") || !strings.Contains(sized.Brief, "(fact) Halvorsen quoted 1,840") {
		t.Fatalf("no map in a selected packet")
	}
	h.e.Cfg.Atlas = config.AtlasOff
	if sized, _ = h.e.DryPacket("What did the gate cost?"); strings.Contains(sized.Brief, "## Map of the conversation") {
		t.Errorf("atlas off still shows the map")
	}
	h.e.Cfg.Atlas = config.AtlasAuto
	h.e.Cfg.WholeHistoryTokens = 60000
	sized, err = h.e.DryPacket("What did the gate cost?")
	if err != nil {
		t.Fatal(err)
	}
	if !sized.WholeHistory || strings.Contains(sized.Brief, "## Map of the conversation") {
		t.Errorf("a packet with every turn needs no map (whole %v)", sized.WholeHistory)
	}
}

// Past its budget the map shows the lines ranked for the message and names
// the rest by topic; the notes op lists a topic's lines.
func TestAtlasMapFoldsByTopicAndNotesOpOpensIt(t *testing.T) {
	h := newHarness(t, 400_000, nil)
	special := map[int]string{}
	var in []IdeaNote
	for i := 1; i <= 60; i++ {
		subject, topic := "irrigation valve", "irrigation"
		if i%2 == 0 {
			subject, topic = "greenhouse heater", "greenhouse"
		}
		special[i] = fmt.Sprintf("Entry %d: the %s reading was %d units after the morning check, logged by hand.", i, subject, 100+i)
		in = append(in, IdeaNote{Title: fmt.Sprintf("The %s reading on entry %d was %d units", subject, i, 100+i), Kind: "result", Topics: []string{topic},
			Evidence: []NoteEvidence{{Turn: vault.TurnID(i), Quote: fmt.Sprintf("reading was %d units", 100+i)}}})
	}
	atlasConversation(t, h, 70, special)
	if res, err := h.e.ImportNotes(in, nil); err != nil || res.New != 60 {
		t.Fatalf("import %+v %v", res, err)
	}
	h.e.Cfg.AtlasMapTokens = 900
	sized, err := h.e.DryPacket("What was the greenhouse heater reading on entry 42?")
	if err != nil {
		t.Fatal(err)
	}
	shown, folds, tokens := packet.MapStatsOf(sized.Items)
	if shown == 0 || shown >= 60 || folds == 0 || tokens > 900 {
		t.Fatalf("map: %d shown, %d fold lines, %d tokens", shown, folds, tokens)
	}
	for _, want := range []string{
		"(result) The greenhouse heater reading on entry 42 was 142 units", // the best match is shown
		`"irrigation": `, "more notes, turns t0000",
		`{"op": "notes", "target": "<the topic, or search words>"}`,
	} {
		if !strings.Contains(sized.Brief, want) {
			t.Errorf("brief lacks %q", want)
		}
	}
	// every note is either a line or counted under a fold line
	rep, err := h.e.Atlas()
	if err != nil || rep.Notes != 60 || rep.Shown+rep.Folded != 60 || rep.Folded == 0 || rep.Paragraphs != 60 {
		t.Fatalf("report %+v %v", rep, err)
	}

	focus := vault.Turn{Seq: 71}
	_, extras, notes := h.e.ExecuteRecallOps([]packet.RecallOp{{Op: "notes", Target: "Irrigation", Limit: 10}, {Op: "notes", Target: "zeppelin"}}, focus, nil)
	if len(extras) != 1 || strings.Count(extras[0].Text, "\n")+1 != 10 || !strings.Contains(extras[0].Text, "irrigation valve reading") ||
		strings.Contains(extras[0].Text, "greenhouse") {
		t.Fatalf("notes op: %d items %v", len(extras), notes)
	}
	if joined := strings.Join(notes, "\n"); !strings.Contains(joined, "More match") || !strings.Contains(joined, `notes("zeppelin") matched no topic and no note`) {
		t.Errorf("notes op notes: %v", notes)
	}
	// a notes op for a message early in the conversation shows no later note
	_, extras, _ = h.e.ExecuteRecallOps([]packet.RecallOp{{Op: "notes", Target: "irrigation", Limit: 40}}, vault.Turn{Seq: 8}, nil)
	if len(extras) != 1 || strings.Contains(extras[0].Text, "t000009") || !strings.Contains(extras[0].Text, "t000007") {
		t.Errorf("as-of notes op: %v", extras)
	}
}

// Defects an audit of the ledger query and plan code reproduced, each kept
// from coming back: a resume id is an id and never a path, a witness whose
// quote is not in its turn is not "verified", and plan checkpoints written
// from several goroutines stay readable.
func TestLedgerQueryIDsAndWitnessesAreChecked(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Use Postgres port 5432 for primary.")
	_ = h.e.V.PutDecision(vault.Decision{ID: "dec-ok", Statement: "Primary uses port 5432", Source: t1.ID, Quote: "port 5432", Status: "active"})
	_ = h.e.V.PutDecision(vault.Decision{ID: "dec-bad", Statement: "Primary uses port 9999", Source: t1.ID, Quote: "port 9999 on the replica", Status: "active"})
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	res, err := h.e.RunLedgerQuery(LedgerQuerySpec{Kind: "decisions", Op: "list", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	verified := map[string]bool{}
	for _, w := range res.Witnesses {
		verified[w.RecordID] = w.Verified
	}
	if !verified["dec-ok"] || verified["dec-bad"] {
		t.Errorf("witnesses verified: %v (a quote that is not in its turn must not count)", verified)
	}
	for _, id := range []string{"../../session", "../turns/t000001", "lq-../x", "", "LQ-ABC", "lq-" + strings.Repeat("a", 40)} {
		if _, err := h.e.ResumeLedgerQuery(id); err == nil || !strings.Contains(err.Error(), "bad id") {
			t.Errorf("resume id %q: %v", id, err)
		}
	}
	if _, err := h.e.ResumeLedgerQuery(res.ID); err != nil {
		t.Errorf("a real id should resume: %v", err)
	}
}

func TestPlanCheckpointSurvivesConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	cp := planCheckpoint{PlanID: "p1", FocusID: "t000001"}
	for i := 0; i < 40; i++ {
		cp.Units = append(cp.Units, packet.Unit{ID: fmt.Sprintf("u%d", i), Role: "responder", Title: strings.Repeat("unit title ", 20)})
	}
	for round := 0; round < 50; round++ {
		done := make(chan struct{})
		for w := 0; w < 8; w++ {
			go func(w int) {
				c := cp
				c.Completed = append([]string{}, fmt.Sprintf("u%d", w))
				c.Units = c.Units[:10+w*4]
				savePlanCheckpoint(dir, c)
				done <- struct{}{}
			}(w)
		}
		for w := 0; w < 8; w++ {
			<-done
		}
		if got, err := loadPlanCheckpoint(dir); err != nil || got.PlanID != "p1" || len(got.Completed) != 1 {
			t.Fatalf("round %d: checkpoint unreadable after concurrent writes: %v", round, err)
		}
	}
}

// A receipt's evidence is verified only when it names the body it came from
// and its quote is inside its span.
func TestReceiptEvidenceMustHoldItsQuote(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Set pool_size=20 for the initial canary in us-east-1.")
	_, _ = h.e.V.AppendTurn(vault.RoleAssistant, "test", "", "", "Confirmed.")
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	check := func(ev store.ReceiptEvidence) bool {
		t.Helper()
		r, err := h.e.CreateReceipt("What is pool_size?", t1.ID, "20", 0, []store.ReceiptEvidence{ev}, nil, "", "")
		if err != nil {
			t.Fatal(err)
		}
		d, err := h.e.DiffReceiptSince(r.ID)
		if err != nil {
			t.Fatal(err)
		}
		return d.OriginalVerified
	}
	whole := store.ReceiptEvidence{Turn: t1.ID, BodySHA: t1.BodySHA, Start: 0, End: len(t1.Body), Seg: vault.AtomSeg}
	good := whole
	good.Quote = "pool_size=20 for the initial canary"
	if !check(good) {
		t.Errorf("a quote inside its span should verify")
	}
	forged := whole
	forged.Quote = "pool_size=999 for production"
	if check(forged) {
		t.Errorf("a quote that is not in the turn verified")
	}
	nohash := good
	nohash.BodySHA = ""
	if check(nohash) {
		t.Errorf("evidence with no body hash verified")
	}
}
