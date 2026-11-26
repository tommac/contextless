package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

const movePlanted = "The movers arrive on the ninth of June at seven in the morning."

// moveHistory is ten turns; t000003 has five paragraphs with the planted
// sentence in the third, t000004 and t000005 are one paragraph each.
func moveHistory(t *testing.T, h *harness) {
	t.Helper()
	for i := 1; i <= 10; i++ {
		role, body := vault.RoleUser, fmt.Sprintf("Note %d about packing the kitchen and labelling the boxes for the new flat.", i)
		switch i {
		case 3:
			role, body = vault.RoleAssistant, longTurn("move", 5, 3, movePlanted)
		case 4:
			body = "The storage unit is number 214 at the Elm Street depot."
		case 5:
			role, body = vault.RoleAssistant, "We still need to book the van for the move."
		}
		if _, err := h.e.V.AppendTurn(role, "", "", "", body); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
}

func readRecord(t *testing.T, h *harness, dir, unit string) unitRecord {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(h.e.V.Dir, "artifacts", dir, unit+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var r unitRecord
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func schemaOf(t *testing.T, h *harness, runID string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(h.e.RunsDir, runID, "ws", ".cx", "schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Every claim of a unit is checked against the vault before the assembler
// reads it: a quote that is found is pinned to the paragraphs that hold
// it, one that is not makes the claim inferred, bad ids are set aside with
// a note, web sources are marked unchecked. A unit that found nothing is
// listed without its text. The records are stored with the plan, and the
// reviewer reads them too.
func TestUnitRecordsAreCheckedBeforeTheAssemblerReadsThem(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.e.Cfg.Routing[config.RoleReviewer] = []string{"gamma/m1"}
	moveHistory(t, h)
	focus := "t000011"
	var asmBrief, reviewBrief string
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
				{ID: "u1", Role: "responder", Title: "Times and places", Instructions: "Say who comes when and where things go, with evidence."},
				{ID: "u2", Role: "responder", Title: "Budget", Instructions: "Find the moving budget."},
				{ID: "u3", Role: "coder", Title: "Schedule file", Instructions: "Write the schedule as a CSV file."},
			}}
		case packet.KindUnit:
			if !strings.Contains(c.Brief, "Your handoff is this unit's record") {
				t.Errorf("a unit brief lacks the record guide")
			}
			if s := schemaOf(t, h, c.RunID); !strings.Contains(s, `"no_information"`) || !strings.Contains(s, `"claims"`) {
				t.Errorf("a unit run did not get the unit schema")
			}
			switch {
			case strings.Contains(c.Brief, "You are unit u1 of 3"):
				return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "Movers at seven on 9 June; the boxes go to storage unit 214.",
					Claims: []packet.UnitClaim{
						{Claim: "The movers come at seven on 9 June.", Evidence: []string{"t000003"}, Quote: movePlanted, Support: "quoted"},
						{Claim: "The date is 9 June.", Evidence: []string{"t000003#a1"}, Quote: "arrive on the ninth of June", Support: "quoted"},
						{Claim: "The van is booked.", Evidence: []string{"t000005"}, Quote: "the van is booked for Friday", Support: "quoted"},
						{Claim: "The keys go back on Monday.", Evidence: []string{"t000000", "t000050"}, Support: "inferred"},
						{Claim: "The depot is on Elm Street.", Evidence: []string{"https://example.com/elm-depot"}, Quote: "Unit 214, Elm Street", Support: "quoted"},
						{Claim: "Moving on a weekday is cheaper.", Support: "unsupported"},
						{Claim: "The boxes go to unit 214.", Evidence: []string{"see turns t000004 and t000005"}, Quote: "storage unit is   number\n214", Support: "quoted"},
						{Claim: "The user wants to know who comes when.", Evidence: []string{focus}, Quote: "who comes when", Support: "quoted"},
						{Claim: "The flat is on the third floor.", Evidence: []string{"t000002"}, Support: "certain"},
					},
					Assumptions:   []string{"The move is on a weekday."},
					OpenQuestions: []string{"Who returns the keys?"},
					Conflicts:     []string{"t000005 says the van is not booked yet."},
					Notes:         "Checked every note before t000011."}
			case strings.Contains(c.Brief, "You are unit u2 of 3"):
				return packet.Handoff{RunID: c.RunID, Status: packet.StatusNoInformation, Answer: "UNUSED TEXT FROM U2",
					Notes: "Looked through t000001 to t000010 for a budget; none is stated."}
			default:
				return withEdits{H: packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "Wrote the schedule to out/schedule.csv.",
					Claims: []packet.UnitClaim{{Claim: "The schedule starts at 07:00.", Evidence: []string{"out/schedule.csv"}, Quote: "07:00,movers arrive", Support: "quoted"}}},
					Edits: map[string]string{"out/schedule.csv": "time,what\n07:00,movers arrive\n"}}
			}
		case packet.KindAssemble:
			asmBrief = c.Brief
			return done(c, "The movers come at seven on 9 June; the boxes go to storage unit 214.")
		case packet.KindReview:
			reviewBrief = c.Brief
			return packet.ReviewHandoff{RunID: c.RunID, Status: "completed", Verdict: "approve"}
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	r, err := h.e.Ask(context.Background(), "Plan the move day: who comes when, and where things go.")
	if err != nil {
		t.Fatal(err)
	}
	if r.User.ID != focus || !strings.Contains(r.Assistant.Author, "reviewed by gamma/m1") {
		t.Fatalf("focus %s, author %q", r.User.ID, r.Assistant.Author)
	}
	for _, want := range []string{
		"### Unit u1 · responder · Times and places · by alpha/m1",
		"Output:\n\nMovers at seven on 9 June; the boxes go to storage unit 214.",
		"- u1.c1 · quoted, quote found in t000003#a3: The movers come at seven on 9 June.",
		"- u1.c2 · quoted, quote found in t000003#a3: The date is 9 June.",
		"Note: the quote is in t000003#a3, not in the cited t000003#a1",
		"- u1.c3 · inferred (declared quoted), quote not found in t000005: The van is booked.",
		"- u1.c4 · unsupported (declared inferred), no evidence: The keys go back on Monday.",
		"Note: t000000 does not exist",
		"Note: t000050 comes after the message being answered",
		"- u1.c5 · quoted, from a web source the engine did not check: https://example.com/elm-depot: The depot is on Elm Street.",
		"- u1.c6 · unsupported, no evidence: Moving on a weekday is cheaper.",
		"- u1.c7 · quoted, quote found in t000004: The boxes go to unit 214.",
		`Quote: "storage unit is number 214"`,
		"- u1.c8 · quoted, quote found in t000011: The user wants to know who comes when.",
		"- u1.c9 · unsupported, no quote, evidence t000002: The flat is on the third floor.",
		`Note: support "certain" is not quoted, inferred or unsupported`,
		"Assumptions:\n- The move is on a weekday.",
		"Open questions:\n- Who returns the keys?",
		"Conflicts:\n- t000005 says the van is not booked yet.",
		"Notes:\nChecked every note before t000011.",
		"### Unit u2 · responder · Budget · by alpha/m1\n\nThis unit found no information for its task.",
		"What it looked for (its notes):\nLooked through t000001 to t000010 for a budget; none is stated.",
		"### Unit u3 · coder · Schedule file · by beta/m1",
		"/schedule.csv: The schedule starts at 07:00.",
		"Files saved:\n- artifacts/run-",
	} {
		if !strings.Contains(asmBrief, want) {
			t.Errorf("assembler brief lacks %q", want)
		}
	}
	if strings.Contains(asmBrief, "UNUSED TEXT FROM U2") {
		t.Error("the text of a unit that found no information reached the assembler")
	}
	if !strings.Contains(asmBrief, "u3.c1 · quoted, quote found in artifacts/run-") {
		t.Error("a quote from the unit's own file was not found in it")
	}
	if !strings.Contains(reviewBrief, "### Unit u1 · responder · Times and places") || !strings.Contains(reviewBrief, "The answer was assembled from the work of several units") {
		t.Error("the reviewer did not get the unit records")
	}

	plan := h.callsOf(packet.KindRespond)[0].RunID
	u1 := readRecord(t, h, plan, "u1")
	if u1.Plan != plan || u1.Status != packet.StatusCompleted || len(u1.Claims) != 9 || len(u1.Runs) != 1 {
		t.Fatalf("stored record of u1: %+v", u1)
	}
	wantChecks := []string{checkVerified, checkVerified, checkNotFound, checkNoEvidence, checkExternal, checkNoEvidence, checkVerified, checkVerified, checkNoQuote}
	for i, c := range u1.Claims {
		if c.Check != wantChecks[i] {
			t.Errorf("claim %s: check %s, want %s", c.ID, c.Check, wantChecks[i])
		}
	}
	if u1.Claims[0].Located != "t000003#a3" || u1.Claims[2].Declared != "quoted" || u1.Claims[2].Support != "inferred" {
		t.Errorf("stored claims: %+v", u1.Claims[:3])
	}
	if u2 := readRecord(t, h, plan, "u2"); u2.Status != packet.StatusNoInformation || u2.Output != "UNUSED TEXT FROM U2" {
		t.Errorf("u2 should be stored as it was handed back: %+v", u2)
	}
	u3 := readRecord(t, h, plan, "u3")
	data, err := os.ReadFile(filepath.Join(h.e.V.Dir, u3.Files[0].Path))
	if err != nil || len(u3.Files) != 1 || u3.Files[0].SHA != vault.SHA(string(data)) || u3.Claims[0].Located != u3.Files[0].Path {
		t.Errorf("u3's file: %+v, %v", u3.Files, err)
	}
	if strings.Contains(u3.Output, "Files saved") {
		t.Error("a unit's files are listed in its record, not appended to its output")
	}
}

// A unit that runs out of room continues in another run; its record
// keeps the claims of every part, and the continuation is asked for the
// new part's claims only.
func TestUnitContinuationKeepsEveryPartsClaims(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	var asmBrief string
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
				{ID: "u1", Role: "responder", Title: "Everything", Instructions: "List every arrangement."},
				{ID: "u2", Role: "responder", Title: "Nothing", Instructions: "Find the budget."},
			}}
		case packet.KindUnit:
			if strings.Contains(c.Brief, "You are unit u2 of 2") {
				return packet.Handoff{RunID: c.RunID, Status: packet.StatusNoInformation}
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPartial, Answer: "First half.", NextStep: "the movers", UnitsRemaining: []string{"movers"},
				Claims:      []packet.UnitClaim{{Claim: "Storage is unit 214.", Evidence: []string{"t000004"}, Quote: "storage unit is number 214", Support: "quoted"}},
				Assumptions: []string{"Storage first."}}
		case packet.KindContinue:
			if !strings.Contains(c.Brief, "This continues a unit") || !strings.Contains(schemaOf(t, h, c.RunID), `"no_information"`) {
				t.Errorf("a unit's continuation should return a unit record")
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "Second half.",
				Claims:      []packet.UnitClaim{{Claim: "Movers at seven.", Evidence: []string{"t000003"}, Quote: "at seven in the morning", Support: "quoted"}},
				Assumptions: []string{"Movers second."}, Notes: "Done in two runs."}
		case packet.KindAssemble:
			asmBrief = c.Brief
			return done(c, "ok")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), "List the arrangements."); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"First half.\n\nSecond half.", "- u1.c1 · quoted, quote found in t000004: Storage is unit 214.",
		"- u1.c2 · quoted, quote found in t000003#a3: Movers at seven.", "Assumptions:\n- Storage first.\n- Movers second.", "Notes:\nDone in two runs."} {
		if !strings.Contains(asmBrief, want) {
			t.Errorf("assembler brief lacks %q", want)
		}
	}
	plan := h.callsOf(packet.KindRespond)[0].RunID
	if r := readRecord(t, h, plan, "u1"); len(r.Runs) != 2 || len(r.Claims) != 2 {
		t.Errorf("record of a continued unit: %d runs, %d claims", len(r.Runs), len(r.Claims))
	}
}

// A quote found with its whitespace collapsed maps back to the exact
// bytes of the body, multibyte text included, and to the paragraphs that
// hold it.
func TestFindQuoteMapsBackToTheBody(t *testing.T) {
	body := "Première ligne.\n\n" + strings.Repeat("Un paragraphe assez long pour être un atome à lui seul, avec des accents: éèà. ", 4) +
		"\n\nLe   déménageur\n arrive à   sept heures.\n\n" + strings.Repeat("Encore un paragraphe assez long pour former son propre atome. ", 5) +
		"\n\n" + strings.Repeat("Un dernier paragraphe, lui aussi assez long pour rester seul. ", 5)
	tu := vault.Turn{ID: "t000007", Seq: 7, Body: body}
	start, end, ok := findQuote(body, 0, len(body), "déménageur arrive à sept heures")
	if !ok || flat(body[start:end]) != "déménageur arrive à sept heures" {
		t.Fatalf("found %v at %d to %d: %q", ok, start, end, body[start:end])
	}
	atoms := vault.Atoms(tu.ID, body)
	if len(atoms) != 3 {
		t.Fatalf("the fixture should have 3 paragraphs, has %d", len(atoms))
	}
	k := 0
	for _, a := range atoms {
		if a.Start <= start && end <= a.End {
			k = a.K
		}
	}
	if k == 0 || atomsAt(tu, start, end) != vault.AtomID(tu.ID, k) {
		t.Fatalf("atomsAt = %s, want atom %d of %d", atomsAt(tu, start, end), k, len(atoms))
	}
	// a span across two paragraphs names both; all of them, the turn
	if got := atomsAt(tu, atoms[0].Start, atoms[1].Start+1); got != vault.AtomRangeID(tu.ID, 1, 2) {
		t.Errorf("two paragraphs: %s", got)
	}
	if got := atomsAt(tu, 0, len(body)); got != tu.ID {
		t.Errorf("the whole body: %s", got)
	}
	if _, _, ok := findQuote(body, atoms[0].Start, atoms[0].End, "sept heures"); ok {
		t.Error("a quote outside the span was found in it")
	}
	if _, _, ok := findQuote(body, 0, len(body), "  "); ok {
		t.Error("an empty quote matched")
	}
}

// A reviewer whose window holds the answer but not the unit records it
// was assembled from reviews the answer alone, rather than not at all.
func TestReviewWithoutTheRecordsWhenTheyDoNotFit(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.e.Cfg.Routing[config.RoleReviewer] = []string{"gamma/m1"}
	prof, _ := h.e.R.Profile("gamma/m1")
	prof.Window, prof.EffectiveWindow = 40_000, 40_000
	h.e.R.SetProfile(prof)
	moveHistory(t, h)
	big := strings.Repeat("A long finding about the boxes, the labels and the order they are loaded in. ", 1500) // about 35k tokens
	var reviews []string
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
				{ID: "u1", Role: "responder", Title: "Long", Instructions: "Write it all."},
				{ID: "u2", Role: "responder", Title: "Short", Instructions: "Write a line."},
			}}
		case packet.KindUnit:
			if strings.Contains(c.Brief, "You are unit u1 of 2") {
				return done(c, big)
			}
			return done(c, "One line.")
		case packet.KindAssemble:
			if !strings.Contains(c.Brief, strings.TrimSpace(big)) {
				t.Error("the assembler should get the long record whole")
			}
			return done(c, "The short answer.")
		case packet.KindReview:
			reviews = append(reviews, c.Brief)
			return packet.ReviewHandoff{RunID: c.RunID, Status: "completed", Verdict: "approve"}
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	r, err := h.e.Ask(context.Background(), "Write up the packing.")
	if err != nil {
		t.Fatal(err)
	}
	if len(reviews) != 1 || strings.Contains(reviews[0], "### Unit u1") || !strings.Contains(reviews[0], "The short answer.") {
		t.Fatalf("want one review of the answer alone, got %d", len(reviews))
	}
	if !strings.Contains(r.Assistant.Author, "reviewed by gamma/m1") {
		t.Fatalf("author %q", r.Assistant.Author)
	}
}
