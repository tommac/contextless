package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// notesHistory is 40 turns of nine-paragraph notes, one of them holding
// the planted sentences at paragraphs 3 and 7.
func notesHistory(t *testing.T, h *harness) vault.Turn {
	t.Helper()
	for i := 1; i <= 40; i++ {
		role, reply, body := vault.RoleUser, "", fmt.Sprintf("What is next on the %s list for week %d?", []string{"garden", "garage", "attic"}[i%3], i)
		if i%2 == 0 {
			role, reply = vault.RoleAssistant, vault.TurnID(i-1)
			body = longTurn([]string{"garden", "garage", "attic"}[i%3], 9, 0, "")
		}
		if i == 12 {
			body = longTurn("house", 9, 3, "The heat pump warranty runs until March 2031.")
			body = strings.Replace(body, "Paragraph 7 of", "The heat pump installer was Calloway Mechanical. Paragraph 7 of", 1)
		}
		if _, err := h.e.V.AppendTurn(role, "", "", reply, body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	src, _ := h.e.V.Turn("t000012")
	return src
}

// Every excerpt holds exactly the paragraphs its id and header name, no
// paragraph of a turn is sent twice, and the pieces of a turn are in
// paragraph order, with and without recalled ids.
func TestExcerptsHoldWhatTheirIdNamesOnceAndInOrder(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	src := notesHistory(t, h)
	focus := h.e.nextTurn("When does the heat pump warranty run out, and who was the installer?")
	check := func(pinned []string, wantIDs ...string) {
		t.Helper()
		items, err := packet.Select(h.e.V, h.e.S, focus, packet.SelectOptions{Recent: 8, RelevantTokens: 8000, Pinned: pinned})
		if err != nil {
			t.Fatal(err)
		}
		sent := map[string]int{} // turn#k to copies
		got := map[string]bool{}
		excerpts := 0
		for _, it := range items {
			ref, ok := vault.ParseRef(it.ID)
			if !ok || it.Key == packet.KeyCatalog || it.Key == packet.KeyFocus {
				continue
			}
			got[it.ID] = true
			tu, _ := h.e.V.Turn(ref.Turn)
			a, b := ref.Atom, ref.AtomEnd
			if ref.Whole() {
				a, b = 1, len(vault.Atoms(tu.ID, tu.Body))
			} else {
				excerpts++
				want, _ := vault.AtomsText(tu.ID, tu.Body, a, b)
				if it.Text != want {
					t.Errorf("%s does not hold the paragraphs it names", it.ID)
				}
				if !strings.Contains(it.Header, "("+it.ID+")") {
					t.Errorf("%s: header %q names another id", it.ID, it.Header)
				}
			}
			for k := a; k <= b; k++ {
				sent[fmt.Sprintf("%s#%d", ref.Turn, k)]++
			}
		}
		if excerpts == 0 && !got["t000012"] {
			t.Fatal("no excerpt in the packet")
		}
		for k, n := range sent {
			if n > 1 {
				t.Errorf("paragraph %s is in %d records (pinned %v)", k, n, pinned)
			}
		}
		for _, id := range wantIDs {
			if !got[id] {
				t.Errorf("pinned %v: want record %s, have %v", pinned, id, got)
			}
		}
		brief := packet.Render(packet.Task{RunID: "r", Kind: packet.KindRespond}, items, nil)
		last := -1
		for _, line := range strings.Split(brief, "\n") {
			if !strings.HasPrefix(line, "### t000012 ") {
				continue
			}
			var k int
			if i := strings.Index(line, "(t000012#a"); i >= 0 {
				fmt.Sscanf(line[i:], "(t000012#a%d", &k)
			}
			if k <= last {
				t.Errorf("pieces of t000012 are out of paragraph order: %d after %d", k, last)
			}
			last = k
		}
	}
	check(nil)
	// ranges that overlap or touch come back as one run
	check([]string{"t000012#a2-4", "t000012#a4-5", "t000012#a6"}, "t000012#a2-6")
	// a recalled turn is sent once, whole
	check([]string{"t000012#a3", "t000012", "t000012#a7-8"}, "t000012")
	_ = src

	// asking again for what was put back, or for one paragraph, gets a note
	_, notes := h.e.recallRequest([]string{"t000012#a3"}, nil, focus, []string{"t000012#a3"})
	if len(notes) != 1 || !strings.Contains(notes[0], "already put back") {
		t.Errorf("a repeated id should get a note: %v", notes)
	}
	ids, notes := h.e.recallRequest([]string{"t000012#a3"}, nil, focus, nil)
	if len(ids) != 1 || len(notes) != 1 || !strings.Contains(notes[0], "ask for t000012#a2-4") {
		t.Errorf("a single paragraph should come with how to read around it: %v %v", ids, notes)
	}
}

// Several processes opening a workspace that has no atom index build it
// together without holes.
func TestConcurrentOpensBuildAWholeAtomIndex(t *testing.T) {
	h := oldWorkspace(t)
	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ro, err := OpenReadOnly(h.e.Cfg, "test")
			if err != nil {
				errs <- err.Error()
				return
			}
			defer ro.Close()
			if n := len(ro.S.AtomTurns()); n != 12 {
				errs <- fmt.Sprintf("%d of 12 turns indexed", n)
			}
			if n := len(ro.S.SearchAtoms("walnut", 10, 0)); n != 1 {
				errs <- fmt.Sprintf("walnut is in %d atoms, want 1", n)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

// A card written while the atom index was not updated (atoms off, or a
// failed write) reaches the turn's atoms on the next writing open.
func TestWritingOpenRepairsAtomsIndexedWithoutTheirCard(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	for i := 1; i <= 6; i++ {
		body := fmt.Sprintf("Filler turn %d about the compost and the garden beds, long enough to be a paragraph.", i)
		if i == 4 {
			body = "Let's go with the walnut ones, not the white oak."
		}
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	tu, _ := h.e.V.Turn("t000004")
	if err := h.e.V.PutCard(vault.Card{Turn: tu.ID, BodySHA: tu.BodySHA, Summary: "The user picks the wood for the kitchen cabinets.", Keywords: []string{"cabinets"}}); err != nil {
		t.Fatal(err)
	}
	cfg := h.e.Cfg
	h.e.Close()
	w, err := Open(cfg, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	hits := w.S.SearchAtoms("kitchen cabinets", 3, 0)
	if len(hits) == 0 || hits[0].Turn != "t000004" || !hits[0].CardOnly {
		t.Fatalf("the card's words should reach the turn, as a card-only match: %+v", hits)
	}
}

// A turn found only through its card goes in whole when it fits the
// budget, since no paragraph of it matches better than another, and is
// left to its catalog line when it does not.
func TestCardOnlyMatchIsTheWholeTurnOrItsCatalogLine(t *testing.T) {
	for _, budget := range []int{8000, 300} {
		h := newHarness(t, 200_000, nil)
		h.e.Cfg.RelevantTokens = budget
		for i := 1; i <= 30; i++ {
			body := fmt.Sprintf("Filler turn %d about the compost and the garden beds, long enough to be a paragraph.", i)
			if i == 5 {
				body = longTurn("ledger", 6, 2, "Go with the walnut ones.")
			}
			if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
				t.Fatal(err)
			}
		}
		tu, _ := h.e.V.Turn("t000005")
		if err := h.e.V.PutCard(vault.Card{Turn: tu.ID, BodySHA: tu.BodySHA, Summary: "The user picks the timber for the kitchen cabinets.", Keywords: []string{"cabinets", "timber"}}); err != nil {
			t.Fatal(err)
		}
		if err := h.e.Reindex(); err != nil {
			t.Fatal(err)
		}
		sized, err := h.e.DryPacket("Which timber did we pick for the kitchen cabinets?")
		if err != nil {
			t.Fatal(err)
		}
		whole, pieces := false, 0
		for _, it := range sized.Items {
			if ref, ok := vault.ParseRef(it.ID); ok && ref.Turn == "t000005" && it.Key == packet.KeyRelevant {
				if ref.Whole() && it.Text == tu.Body {
					whole = true
				} else {
					pieces++
				}
			}
		}
		if budget == 8000 && (!whole || pieces > 0) {
			t.Errorf("budget %d: want the turn whole, got whole=%v and %d excerpts", budget, whole, pieces)
		}
		if budget == 300 && (whole || pieces > 0) {
			t.Errorf("budget %d: a turn that does not fit should be left to its catalog line, got whole=%v and %d excerpts", budget, whole, pieces)
		}
	}
}

// An excerpt cut out of a long code block says so in its header: its
// fence is not closed inside the record.
func TestExcerptSaysWhenItCutsACodeBlock(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	var code strings.Builder
	code.WriteString("Here is the migration script.\n\n```sql\n")
	for i := 1; i <= 160; i++ {
		fmt.Fprintf(&code, "UPDATE ledger SET batch = %d WHERE region = 'north' AND id > %d;\n", i, i*100)
		if i == 80 {
			code.WriteString("ALTER TABLE ledger ADD COLUMN quokka_checksum TEXT;\n")
		}
	}
	code.WriteString("```\n\nRun it after the backup.\n")
	for i := 1; i <= 30; i++ {
		body := fmt.Sprintf("Filler turn %d about the compost and the garden beds, long enough to be a paragraph.", i)
		if i == 5 {
			body = code.String()
		}
		if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	sized, err := h.e.DryPacket("Which statement adds the quokka_checksum column?")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, it := range sized.Items {
		if strings.Contains(it.Text, "quokka_checksum") && it.Key == packet.KeyRelevant {
			found = true
			if !strings.Contains(it.Header, "begins and ends inside a long code block") {
				t.Errorf("header does not say the excerpt is cut from a code block: %q", it.Header)
			}
		}
	}
	if !found {
		t.Fatal("the matching piece of the code block is not in the packet")
	}
}

// The whole-history rule is for the answering call. Plan units and their
// assembler get a selected packet (a plan must not send the transcript
// once per unit), and a history over the limit is selected too.
func TestWholeHistoryIsNotSentToPlanUnitsOrOverTheLimit(t *testing.T) {
	build := func(limit int) *harness {
		h := newHarness(t, 200_000, nil)
		h.e.Cfg.WholeHistoryTokens = limit
		for i := 1; i <= 30; i++ {
			body := fmt.Sprintf("Note %d. ", i) + strings.Repeat(fmt.Sprintf("Line about item %d of the inventory. ", i), 40)
			if _, err := h.e.V.AppendTurn(vault.RoleUser, "", "", "", body); err != nil {
				t.Fatal(err)
			}
		}
		if err := h.e.Reindex(); err != nil {
			t.Fatal(err)
		}
		return h
	}
	h := build(60_000)
	h.reply = func(c call) any {
		whole := strings.Contains(c.Brief, "### Coverage of this packet") && !strings.Contains(c.Brief, "## Catalog")
		switch c.Kind {
		case packet.KindRespond:
			if !whole {
				t.Errorf("the answering call should get the short history whole, and be told so")
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
				{ID: "u1", Role: "coder", Title: "One", Instructions: "Do one."},
				{ID: "u2", Role: "researcher", Title: "Two", Instructions: "Do two."},
			}}
		case packet.KindUnit, packet.KindAssemble:
			if whole || !strings.Contains(c.Brief, "## Catalog") {
				t.Errorf("a %s call got the whole history", c.Kind)
			}
			return done(c, "ok")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), "Do both things."); err != nil {
		t.Fatal(err)
	}
	if len(h.callsOf(packet.KindUnit)) != 2 {
		t.Fatalf("%d unit calls", len(h.callsOf(packet.KindUnit)))
	}

	h = build(5_000) // the history is about 13k tokens
	sized, err := h.e.DryPacket("Which item was on line 7?")
	if err != nil {
		t.Fatal(err)
	}
	if sized.WholeHistory || !strings.Contains(sized.Brief, "## Catalog") {
		t.Errorf("a history over whole_history_tokens should be selected")
	}
}
