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

func TestRunDemoEndToEnd(t *testing.T) {
	res, err := RunDemo()
	if err != nil {
		t.Fatal(err)
	}
	if res.TurnsCreated != 7 {
		t.Fatalf("TurnsCreated = %d, want 7", res.TurnsCreated)
	}
	if res.LiveReceiptBefore.Status != "current" || res.LiveReceiptAfter.Status != "stale" {
		t.Fatalf("live receipt transition: before=%s after=%s", res.LiveReceiptBefore.Status, res.LiveReceiptAfter.Status)
	}
	if res.HistoricalReceipt.Status != "current" {
		t.Fatalf("historical cutoff receipt must remain current, got %s", res.HistoricalReceipt.Status)
	}
	if !res.LiveDiff.OriginalVerified || !res.LiveDiff.SnapshotMatch || len(res.LiveDiff.DeltaTurns) == 0 {
		t.Fatalf("live diff verification failed: %+v", res.LiveDiff)
	}
	if !res.HistoricalDiff.OriginalVerified || !res.HistoricalDiff.SnapshotMatch || len(res.HistoricalDiff.DeltaTurns) != 0 {
		t.Fatalf("historical diff verification failed: %+v", res.HistoricalDiff)
	}
	if len(res.OpenBlockIDs) == 0 || !strings.Contains(res.TimelineText, "superseded") {
		t.Fatalf("typed recall outputs incomplete: openBlock=%v timeline=%q", res.OpenBlockIDs, res.TimelineText)
	}
	txt := FormatDemo(res)
	if !strings.Contains(txt, "ContextlessAI demo") || !strings.Contains(txt, "pool_size=64") {
		t.Fatalf("unexpected FormatDemo output:\n%s", txt)
	}
}

func TestTypedRecallVerbs(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Let's configure the checkout service.")
	longCode := strings.Repeat("Intro paragraph explaining the parser design and invariants for checkout. ", 4) + "\n\n```go\n" +
		strings.Repeat("func step() int { return 42 }\n", 120) +
		"```\n\n" + strings.Repeat("Closing remarks after code explaining rollback and metrics. ", 4)
	t2, _ := h.e.V.AppendTurn(vault.RoleAssistant, "alpha", "", "", longCode)
	t3, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Confirmed PORT=8443 and RETENTION_DAYS=90.")
	d1 := vault.Decision{ID: vault.NewID("dec"), Statement: "Use PORT=8080 initially", Status: "superseded", Source: t1.ID, Quote: "checkout service"}
	_ = h.e.V.PutDecision(d1)
	_ = h.e.V.PutDecision(vault.Decision{ID: vault.NewID("dec"), Statement: "Switch to PORT=8443 with TLS", Status: "active", Source: t3.ID, Quote: "PORT=8443", Supersedes: []string{d1.ID}})
	_ = h.e.Reindex()

	focus := vault.Turn{Seq: 4}
	// 1. exchange
	ids, _, _ := h.e.ExecuteRecallOps([]packet.RecallOp{{Op: "exchange", Target: t2.ID, Radius: 1}}, focus, nil)
	if len(ids) != 3 {
		t.Fatalf("exchange radius=1 expected 3 turns, got %v", ids)
	}
	// 2. open_block expands a split code fence across multiple atoms
	atoms2 := packet.AtomsOf(t2)
	if len(atoms2) < 3 {
		t.Fatalf("expected long code turn to split into >=3 atoms, got %d", len(atoms2))
	}
	obIDs, _, obNotes := h.e.ExecuteRecallOps([]packet.RecallOp{{Op: "open_block", Target: vault.AtomID(t2.ID, 2)}}, focus, nil)
	if len(obIDs) != 1 || !strings.Contains(strings.Join(obNotes, " "), "code") {
		t.Fatalf("open_block failed: ids=%v notes=%v", obIDs, obNotes)
	}
	// 3. grep literal & regex
	gIDs, gExtras, _ := h.e.ExecuteRecallOps([]packet.RecallOp{
		{Op: "grep", Target: "PORT=8443"},
		{Op: "grep", Target: `/RETENTION_DAYS=\d+/`},
	}, focus, nil)
	if len(gIDs) == 0 || len(gExtras) != 2 {
		t.Fatalf("grep failed: ids=%v extras=%d", gIDs, len(gExtras))
	}
	// 4. timeline
	_, tlExtras, _ := h.e.ExecuteRecallOps([]packet.RecallOp{{Op: "timeline", Target: "PORT"}}, focus, nil)
	if len(tlExtras) != 1 || !strings.Contains(tlExtras[0].Text, "PORT=8080") || !strings.Contains(tlExtras[0].Text, "PORT=8443") {
		t.Fatalf("timeline failed: %+v", tlExtras)
	}
}

func TestLivingAnswersRefreshDeltaResume(t *testing.T) {
	var refreshBrief string
	h := newHarness(t, 200_000, func(c call) any {
		if c.Kind == packet.KindIndex {
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
		if strings.Contains(c.Brief, "Delta resume for receipt") {
			refreshBrief = c.Brief
			return done(c, "Updated answer: canary pool_size is now 64 per t000004.")
		}
		return done(c, "Canary pool_size is 20.")
	})
	_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Set canary pool_size=20 for Postgres.")
	_ = h.e.Reindex()

	rcpt, _, err := h.e.WatchQuestion(context.Background(), "What is the canary pool_size?", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rcpt.Status != "current" || len(rcpt.Evidence) == 0 {
		t.Fatalf("bad initial receipt: %+v", rcpt)
	}

	// Send contradicting message via Ask; automatic CheckStaleReceipts must mark rcpt stale.
	if _, err := h.e.Ask(context.Background(), "We changed canary pool_size=64 after stress testing."); err != nil {
		t.Fatal(err)
	}
	after, ok, err := h.e.S.GetReceipt(rcpt.ID)
	if err != nil || !ok || after.Status != "stale" {
		t.Fatalf("expected receipt marked stale automatically, got %+v ok=%v err=%v", after, ok, err)
	}

	// RefreshReceipt must send only the prior receipt plus the new turn, and supersede the old receipt.
	updated, o, err := h.e.RefreshReceipt(context.Background(), rcpt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(refreshBrief, "pool_size=64") || !strings.Contains(refreshBrief, "Prior verified answer") {
		t.Fatalf("refresh brief missing expected delta resume content:\n%s", refreshBrief)
	}
	if updated.ParentID != rcpt.ID || !strings.Contains(o.Answer, "64") {
		t.Fatalf("unexpected updated receipt: %+v answer=%q", updated, o.Answer)
	}
	oldNow, _, _ := h.e.S.GetReceipt(rcpt.ID)
	if oldNow.Status != "superseded" {
		t.Fatalf("old receipt should be superseded, got %s", oldNow.Status)
	}
}

func TestTuneAndClassifyAnswerFailure(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Primary cluster endpoint is db-prod-01.internal on port 5432.")
	t2, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Staging Redis cache TTL is 300s.")
	t3, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Billing export bucket is s3://acme-billing-2026.")
	t4, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Feature flag checkout_v3 rolls out on 2026-12-01.")
	t5, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "PagerDuty escalation timeout is 15 minutes.")
	_ = h.e.Reindex()

	// Test 4-way UTILMEM failure attribution taxonomy.
	p := Probe{ID: "p1", Source: t1.ID, Question: "What is the primary endpoint?", Answer: "db-prod-01.internal"}
	sizedVerbatim := packet.Sized{Items: []packet.Item{{Key: packet.KeyRelevant, ID: t1.ID, Text: t1.Body}}}
	sizedCardOnly := packet.Sized{Items: []packet.Item{{Key: packet.KeyCatalog, ID: t1.ID, Text: t1.ID + " summary"}}}
	sizedEmpty := packet.Sized{}

	if got := ClassifyAnswerFailure(p, sizedVerbatim, "Endpoint is db-prod-01.internal"); got != FailureCorrect {
		t.Fatalf("expected FailureCorrect, got %s", got)
	}
	if got := ClassifyAnswerFailure(p, sizedVerbatim, "Endpoint is db-old.internal", "db-old.internal"); got != FailureStalenessOrAttrib {
		t.Fatalf("expected FailureStalenessOrAttrib, got %s", got)
	}
	if got := ClassifyAnswerFailure(p, sizedVerbatim, "I do not know"); got != FailureUtilizationFailure {
		t.Fatalf("expected FailureUtilizationFailure, got %s", got)
	}
	if got := ClassifyAnswerFailure(p, sizedCardOnly, "I do not know"); got != FailureNavigationMiss {
		t.Fatalf("expected FailureNavigationMiss, got %s", got)
	}
	if got := ClassifyAnswerFailure(p, sizedEmpty, "I do not know"); got != FailureRetrievalMiss {
		t.Fatalf("expected FailureRetrievalMiss, got %s", got)
	}

	// Test offline Tune with 5 probes (60/40 train/held-out split).
	probes := []Probe{
		{ID: "p1", Source: t1.ID, Question: "What is the primary cluster endpoint?", Answer: "db-prod-01.internal"},
		{ID: "p2", Source: t2.ID, Question: "What is the staging Redis TTL?", Answer: "300s"},
		{ID: "p3", Source: t3.ID, Question: "What is the billing export bucket?", Answer: "acme-billing-2026"},
		{ID: "p4", Source: t4.ID, Question: "When does checkout_v3 roll out?", Answer: "2026-12-01"},
		{ID: "p5", Source: t5.ID, Question: "What is the PagerDuty escalation timeout?", Answer: "15 minutes"},
	}
	rep, err := h.e.Tune(probes)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Evaluated < 2 || rep.Baseline.TrainTotal+rep.Baseline.HeldOutTotal != 5 {
		t.Fatalf("unexpected TuneReport: %+v", rep)
	}
	if !strings.Contains(FormatTuneReport(rep), "Offline compiler tuning") {
		t.Fatalf("bad FormatTuneReport")
	}
}

func TestResumePlanReusesCompletedUnits(t *testing.T) {
	ranUnits := map[string]int{}
	h := newHarness(t, 200_000, func(c call) any {
		if c.Kind == packet.KindUnit {
			if strings.Contains(c.Brief, "Unit u2") || strings.Contains(c.Brief, "Second part") {
				ranUnits["u2"]++
				return done(c, "Output from resumed u2")
			}
			ranUnits["u1"]++
			return done(c, "Output from u1")
		}
		if c.Kind == packet.KindAssemble {
			if !strings.Contains(c.Brief, "Precomputed output from u1") || !strings.Contains(c.Brief, "Output from resumed u2") {
				t.Errorf("assembler brief missing precomputed u1 or resumed u2:\n%s", c.Brief)
			}
			return done(c, "Assembled both u1 and u2 cleanly.")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	})
	h.e.Cfg.Review = "off"
	focus, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Do both parts.")

	planID := "run-resumable"
	dir := h.e.V.ArtifactDir(planID)
	cp := planCheckpoint{
		PlanID: planID, FocusID: focus.ID, PlanAuthor: "alpha/m1",
		Units: []packet.Unit{
			{ID: "u1", Role: "responder", Title: "First part", Instructions: "Compute part 1"},
			{ID: "u2", Role: "responder", Title: "Second part", Instructions: "Compute part 2"},
		},
	}
	savePlanCheckpoint(dir, cp)
	// Write already-completed u1 record; leave u2 missing so ResumePlan runs ONLY u2!
	u1Rec := unitRecord{Plan: planID, Unit: "u1", Role: "responder", Title: "First part", By: "alpha/m1", Status: "completed", Output: "Precomputed output from u1"}
	b, _ := json.MarshalIndent(u1Rec, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "u1.json"), b, 0o644)

	o, err := h.e.ResumePlan(context.Background(), planID)
	if err != nil {
		t.Fatal(err)
	}
	if ranUnits["u1"] != 0 || ranUnits["u2"] != 1 {
		t.Fatalf("expected u1 skipped (0) and u2 executed once (1), got %+v", ranUnits)
	}
	if !strings.Contains(o.Answer, "Assembled both") {
		t.Fatalf("unexpected assembled answer: %q", o.Answer)
	}
}

func TestReviewerBoundedRecallOneRound(t *testing.T) {
	reviewCalls := 0
	var secondReviewBrief string
	h := newHarness(t, 200_000, func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return done(c, "The port is 5432.")
		case packet.KindReview:
			reviewCalls++
			if reviewCalls == 1 {
				return packet.ReviewHandoff{
					RunID:     c.RunID,
					Status:    packet.StatusNeedsRecall,
					RecallIDs: []string{"t000001"},
					Notes:     "Need exact turn t000001 to verify port number.",
				}
			}
			secondReviewBrief = c.Brief
			return packet.ReviewHandoff{RunID: c.RunID, Status: packet.StatusCompleted, Verdict: "approve"}
		default:
			return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
		}
	})
	h.e.Cfg.Review = "always"
	h.e.Cfg.RecentTurns = 1 // ensure t000001 is omitted unless recalled
	_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Secret verification phrase: PORT_CHECK_5432.")
	_, _ = h.e.V.AppendTurn(vault.RoleAssistant, "alpha", "", "", "Noted.")
	_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Unrelated filler turn.")
	_, _ = h.e.V.AppendTurn(vault.RoleAssistant, "alpha", "", "", "Filler ack.")
	_ = h.e.Reindex()

	if _, err := h.e.Ask(context.Background(), "Confirm the port."); err != nil {
		t.Fatal(err)
	}
	if reviewCalls != 2 {
		t.Fatalf("expected exactly 2 review calls (1 recall + 1 approval), got %d", reviewCalls)
	}
	if !strings.Contains(secondReviewBrief, "PORT_CHECK_5432") {
		t.Fatalf("second reviewer brief did not receive recalled turn t000001:\n%s", secondReviewBrief)
	}
}

func TestScratchMergeVerifyMergeBranchAndSafeVerifyEnv(t *testing.T) {
	// 1. Verify safeVerifyEnv strips secret-bearing environment variables.
	t.Setenv("OPENAI_API_KEY", "sk-secret-value")
	t.Setenv("CUSTOM_AUTH_TOKEN", "tok-secret")
	for _, kv := range safeVerifyEnv() {
		if strings.HasPrefix(kv, "OPENAI_API_KEY=") || strings.HasPrefix(kv, "CUSTOM_AUTH_TOKEN=") {
			t.Fatalf("safeVerifyEnv leaked secret env var: %s", kv)
		}
	}

	// 2. Real temp git repo with two non-conflicting branches + verify command.
	repoDir := newRepo(t)
	base := gitT(t, repoDir, "rev-parse", "HEAD")

	// Create branch cx/u1 modifying pkg/b.go
	gitT(t, repoDir, "checkout", "-q", "-b", "cx/u1")
	_ = os.WriteFile(filepath.Join(repoDir, "pkg", "b.go"), []byte("package pkg\n// u1\n"), 0o644)
	gitT(t, repoDir, "add", "-A")
	gitT(t, repoDir, "commit", "-q", "-m", "add b")

	// Create branch cx/u2 from base modifying pkg/c.go
	gitT(t, repoDir, "checkout", "-q", "-b", "cx/u2", base)
	_ = os.WriteFile(filepath.Join(repoDir, "pkg", "c.go"), []byte("package pkg\n// u2\n"), 0o644)
	gitT(t, repoDir, "add", "-A")
	gitT(t, repoDir, "commit", "-q", "-m", "add c")

	// Return to main and advance main with an unrelated commit so MergeBranch tests moved HEAD!
	gitT(t, repoDir, "checkout", "-q", "main")
	_ = os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("readme v2\n"), 0o644)
	gitT(t, repoDir, "add", "-A")
	gitT(t, repoDir, "commit", "-q", "-m", "advance main")

	r := vault.Repo{Name: "proj", Path: repoDir, Verify: "test -f pkg/b.go && test -f pkg/c.go"}
	ic := ScratchMergeAndVerify(r, base, []string{"cx/u1", "cx/u2"})
	if !ic.Merged || ic.Verified == nil || !*ic.Verified || ic.Error != "" {
		t.Fatalf("ScratchMergeAndVerify failed: %+v", ic)
	}

	// MergeBranch re-verifies cx/u1 against moved main HEAD before merging into main.
	rSingle := vault.Repo{Name: "proj", Path: repoDir, Verify: "test -f pkg/b.go && grep -q v2 README.md"}
	if _, err := MergeBranch(rSingle, "cx/u1"); err != nil {
		t.Fatalf("MergeBranch failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "pkg", "b.go")); err != nil {
		t.Fatalf("expected pkg/b.go merged onto main: %v", err)
	}
}

func TestObsidianExportPreservesHumanNotesAndCanvasBrief(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Use Postgres 17 with pgvector.")
	_ = h.e.V.PutDecision(vault.Decision{ID: vault.NewID("dec"), Statement: "Postgres 17 selected", Status: "active", Source: t1.ID, Quote: "Use Postgres 17"})
	_ = h.e.Reindex()

	outDir := filepath.Join(t.TempDir(), "obsidian")
	st1, err := h.e.ExportObsidian(outDir)
	if err != nil || st1.Turns != 1 || st1.Decisions != 1 {
		t.Fatalf("first ExportObsidian: %+v err=%v", st1, err)
	}
	// Append human note outside the machine-owned section.
	turnNote := filepath.Join(outDir, "turns", t1.ID+".md")
	orig, _ := os.ReadFile(turnNote)
	humanLine := "Human insight: double-check vacuum tuning before launch."
	_ = os.WriteFile(turnNote, append(orig, []byte(humanLine+"\n")...), 0o644)

	// Re-export must preserve the human annotation intact.
	st2, err := h.e.ExportObsidian(outDir)
	if err != nil || st2.Preserved == 0 {
		t.Fatalf("second ExportObsidian did not preserve human note: %+v err=%v", st2, err)
	}
	updated, _ := os.ReadFile(turnNote)
	if !strings.Contains(string(updated), humanLine) {
		t.Fatalf("human note was clobbered on re-export:\n%s", updated)
	}

	// Test native Go Canvas brief compiler with #Heading subpath + edge label.
	canvasDir := t.TempDir()
	noteMD := "# Overview\nGeneral intro.\n\n## Constraints\nLatency budget is 40ms p99.\n\n## Other\nIgnore this section.\n"
	_ = os.WriteFile(filepath.Join(canvasDir, "spec.md"), []byte(noteMD), 0o644)
	canvasJSON := `{
		"nodes": [
			{"id": "q1", "type": "text", "text": "# Question\nWhat is our latency budget?"},
			{"id": "f1", "type": "file", "file": "spec.md", "subpath": "#Constraints"}
		],
		"edges": [
			{"fromNode": "f1", "toNode": "q1", "label": "Hard SLO constraint"}
		]
	}`
	canvasPath := filepath.Join(canvasDir, "board.canvas")
	_ = os.WriteFile(canvasPath, []byte(canvasJSON), 0o644)

	brief, err := CompileObsidianCanvas(canvasPath, canvasDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(brief, "Hard SLO constraint") || !strings.Contains(brief, "40ms p99") {
		t.Fatalf("compiled canvas brief missing edge label or heading content:\n%s", brief)
	}
	if strings.Contains(brief, "Ignore this section") {
		t.Fatalf("canvas #Heading narrowing leaked outside section:\n%s", brief)
	}
}
