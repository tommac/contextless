package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

func callIndex(h *harness, pred func(c call) bool) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, c := range h.calls {
		if pred(c) {
			return i
		}
	}
	return -1
}

func isUnit(id string) func(c call) bool {
	return func(c call) bool {
		return c.Kind == packet.KindUnit && strings.Contains(c.Brief, "You are unit "+id+" of")
	}
}

func readAssembly(t *testing.T, h *harness, plan string) assemblyRecord {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(h.e.V.Dir, "artifacts", plan, "assembly.json"))
	if err != nil {
		t.Fatal(err)
	}
	var a assemblyRecord
	if err := json.Unmarshal(data, &a); err != nil {
		t.Fatal(err)
	}
	return a
}

// A plan runs in waves: units with nothing to wait for run together, a
// unit that depends on another runs after it with its record in its
// packet, and every unit and the assembler get the shared decisions.
func TestPlanRunsInWavesWithUpstreamRecordsAndSharedDecisions(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	briefs := map[string]string{}
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan,
				SharedDecisions: []string{"Times are written on the 24-hour clock.", "The new flat is called Elm 214."},
				Units: []packet.Unit{
					{ID: "u1", Role: "responder", Title: "Times", Instructions: "Find when the movers come."},
					{ID: "u2", Role: "responder", Title: "Schedule", Instructions: "Write the day's schedule from u1.", DependsOn: []string{"u1"}},
					{ID: "u3", Role: "responder", Title: "Storage", Instructions: "Find the storage unit."},
				}}
		case packet.KindUnit:
			for _, id := range []string{"u1", "u2", "u3"} {
				if isUnit(id)(c) {
					briefs[id] = c.Brief
				}
			}
			if isUnit("u1")(c) {
				return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "The movers come at 07:00.",
					Claims: []packet.UnitClaim{{Claim: "The movers come at seven.", Evidence: []string{"t000003"}, Quote: "at seven in the morning", Support: "quoted"}}}
			}
			return done(c, "ok")
		case packet.KindAssemble:
			briefs["asm"] = c.Brief
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "07:00, Elm 214.", ClaimsUsed: []string{"u1.c1"}}
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), "Plan the move day."); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"u1", "u2", "u3", "asm"} {
		if !strings.Contains(briefs[id], "Decisions every unit of this plan follows") || !strings.Contains(briefs[id], "- The new flat is called Elm 214.") {
			t.Errorf("%s did not get the shared decisions", id)
		}
	}
	if strings.Contains(briefs["u1"], "### Unit u") || strings.Contains(briefs["u3"], "### Unit u") {
		t.Error("an independent unit got another unit's record")
	}
	if !strings.Contains(briefs["u2"], "### Unit u1 · responder · Times") || !strings.Contains(briefs["u2"], "u1.c1 · quoted, quote found in t000003#a3") ||
		!strings.Contains(briefs["u2"], "This unit depends on u1") {
		t.Error("the dependent unit did not get its upstream record")
	}
	if callIndex(h, isUnit("u2")) < callIndex(h, isUnit("u1")) {
		t.Error("u2 ran before the unit it depends on")
	}
	if !strings.Contains(briefs["asm"], "The engine's comparison of the unit records") {
		t.Error("the assembler did not get the engine's comparison")
	}
	plan := h.callsOf(packet.KindRespond)[0].RunID
	if r := readRecord(t, h, plan, "u2"); strings.Join(r.DependsOn, ",") != "u1" {
		t.Errorf("u2's record depends on %v", r.DependsOn)
	}
}

// A plan of one unit, or of units that would only run one after another,
// buys nothing: the engine answers the message in one call, says why, and
// offers that call no plan.
func TestPlansThatBuyNothingAreAnsweredInOneCall(t *testing.T) {
	for name, units := range map[string][]packet.Unit{
		"one unit": {{ID: "u1", Role: "responder", Title: "All", Instructions: "Do it all."}},
		"chain": {
			{ID: "u1", Role: "responder", Title: "First", Instructions: "a"},
			{ID: "u2", Role: "responder", Title: "Second", Instructions: "b", DependsOn: []string{"u1"}},
			{ID: "u3", Role: "responder", Title: "Third", Instructions: "c", DependsOn: []string{"u2"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, 200_000, nil)
			round := 0
			h.reply = func(c call) any {
				if c.Kind != packet.KindRespond {
					if c.Kind == packet.KindIndex {
						return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
					}
					t.Errorf("unexpected %s call", c.Kind)
					return done(c, "x")
				}
				round++
				if round == 1 {
					return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: units}
				}
				if !strings.Contains(c.Brief, "Why this is one call") || !strings.Contains(c.Brief, "was proposed for this message and not run") {
					t.Error("the single call is not told why")
				}
				if strings.Contains(schemaOf(t, h, c.RunID), `"plan"`) || strings.Contains(c.Brief, "- plan:") {
					t.Error("the single call was offered a plan again")
				}
				return done(c, "Answered in one call.")
			}
			r, err := h.e.Ask(context.Background(), "Do the thing.")
			if err != nil {
				t.Fatal(err)
			}
			if r.Assistant.Body != "Answered in one call." || round != 2 {
				t.Fatalf("answer %q after %d respond calls", r.Assistant.Body, round)
			}
		})
	}
}

func TestPlanUnitsValidatesDependencies(t *testing.T) {
	units, refuse, notes := planUnits([]packet.Unit{
		{ID: "u1", Role: "responder", DependsOn: []string{"u2"}}, // later unit: dropped
		{ID: "u2", Role: "nonsense", DependsOn: []string{"u1", "u1", "zz"}},
		{ID: "u3", Role: "coder", Repo: "proj"},
		{ID: "u4", Role: "coder", Repo: "proj"},
		{ID: "u5", Role: "responder", DependsOn: []string{"u2", "u4"}},
	}, 8)
	if refuse != "" {
		t.Fatalf("refused: %s", refuse)
	}
	got := map[string]string{}
	for _, u := range units {
		got[u.ID] = strings.Join(u.DependsOn, "+")
	}
	if got["u1"] != "" || got["u2"] != "u1" || got["u3"] != "" || got["u4"] != "u3" || got["u5"] != "u2+u4" || units[1].Role != "responder" {
		t.Fatalf("dependencies %v, roles %v", got, units[1].Role)
	}
	if len(notes) != 3 {
		t.Errorf("notes: %q", notes)
	}
	var waves []string
	for _, w := range planWaves(units) {
		var ids []string
		for _, i := range w {
			ids = append(ids, units[i].ID)
		}
		waves = append(waves, strings.Join(ids, ","))
	}
	if strings.Join(waves, " | ") != "u1,u3 | u2,u4 | u5" {
		t.Fatalf("waves %v", waves)
	}
	// one coder unit changing a repository is a plan, not a single call
	if _, refuse, _ := planUnits([]packet.Unit{{ID: "u1", Repo: "proj"}}, 8); refuse != "" {
		t.Fatalf("a repository unit was refused: %s", refuse)
	}
}

// Two coder units changing the same repository do not branch from the
// same base: the second runs after the first, on its commit.
func TestCoderUnitsOnOneRepositoryRunInTurn(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	repo := newRepo(t)
	attach(t, h, vault.Repo{Name: "proj", Path: repo, Allow: []string{"pkg/**"}})
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
				{ID: "u1", Role: "coder", Title: "Add B", Instructions: "Add pkg/b.go.", Repo: "proj"},
				{ID: "u2", Role: "coder", Title: "Add C", Instructions: "Add pkg/c.go.", Repo: "proj"},
			}}
		case packet.KindUnit:
			if isUnit("u1")(c) {
				return withEdits{H: done(c, "Added b."), Edits: map[string]string{"repo/pkg/b.go": "package pkg\n// b\n"}}
			}
			if _, err := os.Stat(filepath.Join(h.e.RunsDir, c.RunID, "ws", "repo", "pkg", "b.go")); err != nil {
				t.Error("u2 does not start from u1's commit")
			}
			return withEdits{H: done(c, "Added c."), Edits: map[string]string{"repo/pkg/c.go": "package pkg\n// c\n"}}
		case packet.KindAssemble:
			return done(c, "Both added.")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), "Add b and c to proj."); err != nil {
		t.Fatal(err)
	}
	plan := h.callsOf(packet.KindRespond)[0].RunID
	u1, u2 := readRecord(t, h, plan, "u1"), readRecord(t, h, plan, "u2")
	if len(u1.Changes) != 1 || len(u2.Changes) != 1 || u2.Changes[0].Base != u1.Changes[0].Commit || u1.Changes[0].Commit == "" {
		t.Fatalf("u1 %+v, u2 %+v", u1.Changes, u2.Changes)
	}
	if got := gitT(t, repo, "show", u2.Changes[0].Branch+":pkg/b.go"); !strings.Contains(got, "// b") {
		t.Fatal("u2's branch does not hold u1's change")
	}
}

// The reduce lists claims of different units that cite the same
// paragraphs, marks those that may state different values, pairs
// assumptions that may be the same choice, and leaves out units without a
// result.
func TestReduceComparesTheRecords(t *testing.T) {
	claim := func(id, text, at string) checkedClaim {
		return checkedClaim{ID: id, Claim: text, Check: checkVerified, Located: at, Declared: "quoted", Support: "quoted"}
	}
	recs := []unitRecord{
		{Unit: "u1", Status: packet.StatusCompleted, Claims: []checkedClaim{
			claim("u1.c1", "The movers arrive at 07:00 on 9 June.", "t000003#a3"),
			claim("u1.c2", "Storage is unit 214.", "t000004"),
		}, Assumptions: []string{"The boxes are labelled by room."}, OpenQuestions: []string{"Who returns the keys?"}},
		{Unit: "u2", Status: packet.StatusCompleted, Claims: []checkedClaim{
			claim("u2.c1", "The movers arrive at 08:00 on 9 June.", "t000003#a2-3"),
			{ID: "u2.c2", Claim: "The van is booked.", Check: checkNotFound, Declared: "quoted", Support: "inferred"},
		}, Assumptions: []string{"The boxes are labelled by floor."}, Conflicts: []string{"t000005 says the van is not booked."}},
		{Unit: "u3", Status: packet.StatusNoInformation, Claims: []checkedClaim{claim("u3.c1", "x", "t000004")}},
		{Unit: "u4", Status: statusFailed, Error: "boom"},
	}
	red := reduceRecords(recs)
	if strings.Join(red.Claims, ",") != "u1.c1,u1.c2,u2.c1,u2.c2" || strings.Join(red.Left, ",") != "u3 (no information),u4 (failed)" {
		t.Fatalf("claims %v, left %v", red.Claims, red.Left)
	}
	if len(red.Groups) != 1 || strings.Join(red.Groups[0].Claims, ",") != "u2.c1,u1.c1" || len(red.Groups[0].Differ) != 2 {
		t.Fatalf("groups %+v", red.Groups)
	}
	text := red.item().Text
	for _, want := range []string{
		"4 unit(s); 4 claim(s) to account for. No result from: u3 (no information), u4 (failed).",
		"- t000003#a2-3, t000003#a3: u2.c1, u1.c1; they may state different values (u2.c1 states 08:00, 9, june; u1.c1 states 07:00, 9, june)",
		"- u1: The boxes are labelled by room. / u2: The boxes are labelled by floor.",
		"Open questions:\n- u1: Who returns the keys?",
		"Conflicts the units reported:\n- u2: t000005 says the van is not booked.",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("comparison lacks %q:\n%s", want, text)
		}
	}
	// the same values, or one a part of the other, are not a disagreement
	if mayDiffer("The movers arrive at 07:00 on 9 June.", "The movers arrive at 07:00.") {
		t.Error("a claim that states part of another's values was marked as differing")
	}
	missing, unknown := accountFor(red.Claims, packet.Handoff{ClaimsUsed: []string{"u1.c1", " u2.c1", "u9.c9"},
		ClaimsOmitted: []packet.OmittedClaim{{ID: "u1.c2", Reason: "duplicate of u1.c1"}, {ID: "u2.c2", Reason: " "}}})
	if strings.Join(missing, ",") != "u2.c2" || strings.Join(unknown, ",") != "u9.c9" {
		t.Fatalf("missing %v, unknown %v", missing, unknown)
	}
}

// The assembler accounts for every claim: a reply that leaves claims out
// of both lists, or names ids that do not exist, is run once more with the
// engine's check, and the result is stored as the plan's assembly record.
func TestAssemblerMustAccountForEveryClaim(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	var asm []string
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
				{ID: "u1", Role: "responder", Title: "Times", Instructions: "a"},
				{ID: "u2", Role: "responder", Title: "Storage", Instructions: "b"},
			}}
		case packet.KindUnit:
			id := "u1"
			if isUnit("u2")(c) {
				id = "u2"
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: id + " done.", Claims: []packet.UnitClaim{
				{Claim: "Storage is unit 214.", Evidence: []string{"t000004"}, Quote: "storage unit is number 214", Support: "quoted"},
				{Claim: "Movers at seven.", Evidence: []string{"t000003"}, Quote: "at seven in the morning", Support: "quoted"},
			}}
		case packet.KindAssemble:
			asm = append(asm, c.Brief)
			if !strings.Contains(schemaOf(t, h, c.RunID), `"claims_omitted"`) {
				t.Error("the assembler did not get the assembly schema")
			}
			if len(asm) == 1 {
				return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "First reply.",
					ClaimsUsed: []string{"u1.c1", "u7.c1"}, ClaimsOmitted: []packet.OmittedClaim{{ID: "u2.c1", Reason: "duplicate of u1.c1"}}}
			}
			return packet.Handoff{RunID: c.RunID, Status: packet.StatusCompleted, Answer: "Second reply.",
				ClaimsUsed: []string{"u1.c1", "u1.c2"}, ClaimsOmitted: []packet.OmittedClaim{{ID: "u2.c1", Reason: "duplicate of u1.c1"}, {ID: "u2.c2", Reason: "duplicate of u1.c2"}}}
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	r, err := h.e.Ask(context.Background(), "Plan it.")
	if err != nil {
		t.Fatal(err)
	}
	if len(asm) != 2 || r.Assistant.Body != "Second reply." {
		t.Fatalf("%d assembler calls, answer %q", len(asm), r.Assistant.Body)
	}
	for _, want := range []string{"Your previous reply (verbatim)", "First reply.",
		"These claims are in neither claims_used nor claims_omitted with a reason: u1.c2, u2.c2.", "These ids name no claim of the records: u7.c1.",
		"- t000003#a3: u1.c2, u2.c2", "- t000004: u1.c1, u2.c1"} {
		if !strings.Contains(asm[1], want) {
			t.Errorf("second assembler brief lacks %q", want)
		}
	}
	a := readAssembly(t, h, h.callsOf(packet.KindRespond)[0].RunID)
	if !a.Retried || a.Claims != 4 || len(a.Missing) != 0 || len(a.Unknown) != 0 || len(a.Used) != 2 || len(a.Omitted) != 2 || a.SharedEvidence != 2 || len(a.Runs) != 2 {
		t.Fatalf("assembly record %+v", a)
	}
}

// A trial answers a message in one arm without storing an answer, and
// reports what it cost; cx compare writes the arms' answers blind.
func TestTrialArmsAndBlindCompare(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			if strings.Contains(c.Brief, "one arm of a measurement") {
				return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
					{ID: "u1", Role: "responder", Title: "A", Instructions: "a"}, {ID: "u2", Role: "responder", Title: "B", Instructions: "b"}}}
			}
			if strings.Contains(schemaOf(t, h, c.RunID), `"plan"`) {
				t.Error("the single arm was offered a plan")
			}
			return done(c, "Single answer.")
		case packet.KindUnit:
			return done(c, "unit output")
		case packet.KindAssemble:
			return done(c, "Planned answer.")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	single := h.e.Trial(context.Background(), "When do the movers come?", ArmSingle)
	plan := h.e.Trial(context.Background(), "When do the movers come?", ArmPlan)
	if single.Answer != "Single answer." || single.Planned || single.Runs != 1 || single.Error != "" {
		t.Fatalf("single %+v", single)
	}
	if plan.Answer != "Planned answer." || !plan.Planned || plan.Units != 2 || plan.Runs < 4 {
		t.Fatalf("plan %+v", plan)
	}
	turns, _ := h.e.V.Turns()
	for _, tu := range turns {
		if tu.Role == vault.RoleAssistant && (tu.Body == "Single answer." || tu.Body == "Planned answer.") {
			t.Fatal("a trial stored its answer")
		}
	}

	dir := t.TempDir()
	items := []CompareItem{{Message: "When do the movers come?", Trials: []TrialResult{single, plan}}}
	if err := WriteCompare(dir, items); err != nil {
		t.Fatal(err)
	}
	ans, _ := os.ReadFile(filepath.Join(dir, "answers.md"))
	if !strings.Contains(string(ans), "### Answer A") || !strings.Contains(string(ans), "### Answer B") ||
		strings.Contains(string(ans), ArmSingle) || strings.Contains(string(ans), ArmPlan) {
		t.Fatalf("answers.md is not blind:\n%s", ans)
	}
	var key []struct {
		Answers map[string]TrialResult `json:"answers"`
	}
	kb, _ := os.ReadFile(filepath.Join(dir, "key.json"))
	if err := json.Unmarshal(kb, &key); err != nil || len(key) != 1 || len(key[0].Answers) != 2 {
		t.Fatalf("key %s: %v", kb, err)
	}
	order := BlindOrder("When do the movers come?", []string{ArmSingle, ArmPlan})
	if key[0].Answers["A"].Arm != []string{ArmSingle, ArmPlan}[order[0]] {
		t.Fatal("the key does not match the blind order")
	}
	if sum, _ := os.ReadFile(filepath.Join(dir, "summary.md")); !strings.Contains(string(sum), "| plan | 1 | 1 | 2 |") {
		t.Fatalf("summary:\n%s", sum)
	}
}

// A trial's scratch copy holds the whole vault and index, but not the
// writer's lock, so it opens while the original is in use.
func TestCopyWorkspaceForATrial(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	moveHistory(t, h)
	if err := CopyWorkspace(h.e.Cfg, "test", "test-copy"); err != nil {
		t.Fatal(err)
	}
	if err := CopyWorkspace(h.e.Cfg, "test", "test-copy"); err == nil {
		t.Fatal("copying onto an existing workspace should fail")
	}
	c, err := Open(h.e.Cfg, "test-copy")
	if err != nil {
		t.Fatalf("the copy does not open while the original is held: %v", err)
	}
	defer c.Close()
	a, _ := h.e.V.Turns()
	b, _ := c.V.Turns()
	if len(a) != len(b) || a[2].Body != b[2].Body {
		t.Fatal("the copy does not hold the same turns")
	}
	msgs := filepath.Join(t.TempDir(), "m.jsonl")
	_ = os.WriteFile(msgs, []byte("{\"message\": \"First, with a\\nnewline.\"}\n\nSecond as plain text.\n"), 0o644)
	got, err := ReadMessages(msgs)
	if err != nil || len(got) != 2 || got[0] != "First, with a\nnewline." || got[1] != "Second as plain text." {
		t.Fatalf("messages %q, %v", got, err)
	}
}
