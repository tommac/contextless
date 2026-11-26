package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

func TestZeroScanRecallAndDegradedLaneWarning(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Primary database pool_size=42 in eu-central-1.")
	_, _ = h.e.V.AppendTurn(vault.RoleAssistant, "alpha", "", "", "Confirmed pool_size=42 in eu-central-1.")
	_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Unrelated filler note about office snacks.")
	_, _ = h.e.V.AppendTurn(vault.RoleAssistant, "alpha", "", "", "Snacks noted.")
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}

	// Healthy zero-scan recall finds t1 cleanly.
	focus := vault.Turn{ID: "t000005", Seq: 5, Body: "What is pool_size?"}
	add, notes := h.e.recallRequest(nil, []string{"pool_size eu-central-1"}, focus, nil)
	foundT1 := false
	for _, id := range add {
		if strings.HasPrefix(id, t1.ID) {
			foundT1 = true
			break
		}
	}
	if !foundT1 {
		t.Fatalf("zero-scan recall missed %s: add=%v notes=%v", t1.ID, add, notes)
	}

	// Simulate SQLite FTS corruption by dropping atoms_fts.
	// Neither recallRequest nor Explain/buildPacket may silently swallow the error!
	if _, err := h.e.S.DB.Exec(`DROP TABLE atoms_fts`); err != nil {
		t.Fatal(err)
	}
	addDeg, notesDeg := h.e.recallRequest(nil, []string{"pool_size eu-central-1"}, focus, nil)
	if !strings.Contains(strings.Join(notesDeg, " "), "lane error") {
		t.Fatalf("expected degraded lane warning note in recallRequest, got %v", notesDeg)
	}
	if len(addDeg) == 0 {
		t.Fatalf("expected fallback scan to still recover %s when FTS fails", t1.ID)
	}
	sized, err := h.e.DryPacket("What is pool_size?")
	if err != nil {
		t.Fatal(err)
	}
	if sized.Index.Complete || len(sized.Index.LaneErrors) == 0 {
		t.Fatalf("expected Index.Complete=false with LaneErrors when FTS fails, got %+v", sized.Index)
	}
	if !strings.Contains(sized.Brief, "## Index coverage notice") {
		t.Fatalf("expected degraded index coverage banner in brief.md")
	}
}

func TestSnapshotPinnedOmittedCatalogVerbs(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	for i := 1; i <= 10; i++ {
		_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Synthetic archived turn.")
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	snapSeq, root, err := h.e.S.SnapshotRoot(10)
	if err != nil || snapSeq != 10 {
		t.Fatalf("SnapshotRoot: seq=%d err=%v", snapSeq, err)
	}
	catDir := filepath.Join(h.e.V.Dir, ".cx", "catalogs")
	handle, err := packet.BuildOmittedCatalog(catDir, root, 10, func(seq int) bool { return seq%2 == 1 })
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(catDir, handle.Index+".json"), []byte(handle.JSON()+"\n"), 0o644)

	focus := vault.Turn{ID: "t000011", Seq: 11}
	// 1. list_omitted paginates verified omitted turn IDs.
	_, extras, notes := h.e.ExecuteRecallOps([]packet.RecallOp{{
		Op: "list_omitted", Index: handle.Index, Snapshot: root, After: 0, Limit: 3,
	}}, focus, nil)
	if len(extras) != 1 || !strings.Contains(extras[0].Text, "t000001, t000003, t000005") {
		t.Fatalf("list_omitted page 1 failed: extras=%+v notes=%v", extras, notes)
	}
	// 2. check_omitted O(1) verified membership.
	_, _, chkNotes := h.e.ExecuteRecallOps([]packet.RecallOp{
		{Op: "check_omitted", Target: "t000003", Index: handle.Index, Snapshot: root},
		{Op: "check_omitted", Target: "t000004", Index: handle.Index, Snapshot: root},
	}, focus, nil)
	joined := strings.Join(chkNotes, "\n")
	if !strings.Contains(joined, "check_omitted(t000003") || !strings.Contains(joined, "omitted=true") || !strings.Contains(joined, "omitted=false") {
		t.Fatalf("check_omitted unexpected notes: %s", joined)
	}
	// 3. Snapshot mismatch must be rejected loudly!
	_, _, badNotes := h.e.ExecuteRecallOps([]packet.RecallOp{{
		Op: "list_omitted", Index: handle.Index, Snapshot: "0000000000000000000000000000000000000000000000000000000000000000",
	}}, focus, nil)
	if !strings.Contains(strings.Join(badNotes, " "), "snapshot mismatch") {
		t.Fatalf("expected snapshot mismatch rejection, got %v", badNotes)
	}
}

func TestLedgerQueryExactHistoricalResumeAndWithheld(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Use Postgres port 5432 for primary.")
	t2, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Enable TLS 1.3 on gateway.")
	d1 := vault.Decision{ID: "dec-p1", Statement: "Primary uses port 5432", Source: t1.ID, Quote: "port 5432", Status: "active"}
	d2 := vault.Decision{ID: "dec-tls", Statement: "Gateway uses TLS 1.3", Source: t2.ID, Quote: "TLS 1.3", Status: "active"}
	_ = h.e.V.PutDecision(d1)
	_ = h.e.V.PutDecision(d2)
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}

	// 1. Live query at seq 2: 2 active decisions.
	live, err := h.e.RunLedgerQuery(LedgerQuerySpec{Kind: "decisions", Op: "list", Status: "active", Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if live.Verdict != "exact" || live.Count != 2 {
		t.Fatalf("expected exact count=2, got %+v", live)
	}
	for _, w := range live.Witnesses {
		if !w.Verified {
			t.Fatalf("witness %s not SHA-verified", w.RecordID)
		}
	}

	// 2. Append turn 3 superseding dec-p1 WITHOUT reindexing first -> Strict query MUST withhold!
	t3, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Switch primary port to 6432 via PgBouncer.")
	d1.Status = "superseded"
	_ = h.e.V.PutDecision(d1)
	d3 := vault.Decision{ID: "dec-p2", Statement: "Primary uses port 6432 via PgBouncer", Source: t3.ID, Quote: "port to 6432", Status: "active", Supersedes: []string{d1.ID}}
	_ = h.e.V.PutDecision(d3)

	withheld, err := h.e.RunLedgerQuery(LedgerQuerySpec{Kind: "decisions", Op: "count", Status: "active", Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if withheld.Verdict != "withheld" || !strings.Contains(withheld.WithheldReason, "lags") {
		t.Fatalf("expected Verdict=withheld when turn 3 is unindexed, got %+v", withheld)
	}

	// 3. Reindex, then ResumeLedgerQuery(live.ID) -> exact delta showing dec-p2 added & dec-p1 superseded!
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}
	resumed, err := h.e.ResumeLedgerQuery(live.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resumed.Verdict != "exact" || resumed.Count != 2 || resumed.Delta == nil || resumed.Delta.Unchanged {
		t.Fatalf("expected exact delta after supersession, got %+v", resumed)
	}
	if len(resumed.Delta.Added) != 1 || resumed.Delta.Added[0].RecordID != "dec-p2" || len(resumed.Delta.Removed) != 1 || resumed.Delta.Removed[0].RecordID != "dec-p1" {
		t.Fatalf("unexpected delta: %+v", resumed.Delta)
	}

	// 4. Historical point-in-time query (AsOfSeq=2) MUST reconstruct dec-p1 as STILL active!
	hist, err := h.e.RunLedgerQuery(LedgerQuerySpec{Kind: "decisions", Op: "list", Filter: "port", Status: "active", AsOfSeq: 2, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if hist.Verdict != "exact" || hist.Count != 1 || hist.Witnesses[0].RecordID != "dec-p1" {
		t.Fatalf("historical AsOfSeq=2 did not reconstruct dec-p1 as active: %+v", hist)
	}
}

func TestSessionRouteAndAutoReplyEndToEnd(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	h.e.Cfg.SessionRoute = true
	h.e.Cfg.AutoReplyWindow = true
	h.e.Cfg.AdaptivePack = true
	h.e.Cfg.RecentTurns = 1

	_, _ = h.e.V.AppendTurn(vault.RoleUser, "Alice", "", "", "Did we pick Rust or Go for the telemetry daemon?")
	_, _ = h.e.V.AppendTurn(vault.RoleAssistant, "Bob", "", "t000001", "We picked Go so the single binary has zero CGo friction.")
	for i := 0; i < 4; i++ {
		_, _ = h.e.V.AppendTurn(vault.RoleUser, "Alice", "", "", "Unrelated lunch scheduling message.")
	}
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}

	plan, err := h.e.Explain("What language was picked for the telemetry daemon?", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Index.Complete || plan.Index.SnapshotSeq != 6 {
		t.Fatalf("expected complete index at seq 6, got %+v", plan.Index)
	}
	sized, err := h.e.DryPacket("What language was picked for the telemetry daemon?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sized.Brief, "zero CGo friction") {
		t.Fatalf("expected AutoReplyWindow + SessionRoute brief to include Bob's reply verbatim:\n%s", sized.Brief)
	}
}
