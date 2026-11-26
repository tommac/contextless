package engine

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

func TestPlanTopologyGateAndReactiveBarrierFreeDAG(t *testing.T) {
	// 1. Near-sequential 4-unit chain (width=2 on wave 0 then 3 serial steps -> 4 waves? No: 4 units in 3 waves = ratio 1.33 < 1.35)
	// must be refused by the Kim et al. scaling-law gate!
	serialHeavy := []packet.Unit{
		{ID: "u1", Role: "responder", Title: "A"},
		{ID: "u2", Role: "responder", Title: "B"},
		{ID: "u3", Role: "responder", Title: "C", DependsOn: []string{"u1", "u2"}},
		{ID: "u4", Role: "responder", Title: "D", DependsOn: []string{"u3"}},
		{ID: "u5", Role: "responder", Title: "E", DependsOn: []string{"u4"}},
	}
	// 5 units across 4 waves => ratio 1.25 < 1.35 => refused!
	topo := EvaluatePlanTopology(serialHeavy)
	if topo.AllowFanOut {
		t.Fatalf("expected near-serial 5-unit/4-wave plan refused by topology gate, got %+v", topo)
	}

	// Genuinely parallel 3-unit fan-out (width 3, depth 1 => ratio 3.0) MUST be allowed!
	wide := []packet.Unit{
		{ID: "u1", Role: "responder", Title: "A"},
		{ID: "u2", Role: "responder", Title: "B"},
		{ID: "u3", Role: "responder", Title: "C"},
	}
	if topoWide := EvaluatePlanTopology(wide); !topoWide.AllowFanOut || topoWide.ParallelismRatio != 3.0 {
		t.Fatalf("expected wide fan-out allowed, got %+v", topoWide)
	}

	// 2. Barrier-free reactive DAG proof:
	// Node 0 finishes fast (5ms), Node 1 is slow (120ms), Node 2 depends ONLY on Node 0!
	// Under a Pregel wave barrier, Node 2 could not start until Node 1 finished (>=120ms).
	// Under runDAG, Node 2 MUST finish BEFORE Node 1 finishes!
	var mu sync.Mutex
	var finishOrder []int
	deps := [][]int{nil, nil, {0}}
	runDAG(context.Background(), 3, deps, 4, func(i int) {
		switch i {
		case 0:
			time.Sleep(5 * time.Millisecond)
		case 1:
			time.Sleep(120 * time.Millisecond)
		case 2:
			time.Sleep(10 * time.Millisecond)
		}
		mu.Lock()
		finishOrder = append(finishOrder, i)
		mu.Unlock()
	})
	if len(finishOrder) != 3 || finishOrder[0] != 0 || finishOrder[1] != 2 || finishOrder[2] != 1 {
		t.Fatalf("expected barrier-free completion order [0 2 1], got %v", finishOrder)
	}
}

func TestVerifiedQuoteDominancePreventsClaimFlip(t *testing.T) {
	recs := []unitRecord{
		{
			Unit: "u1", Status: "completed",
			Claims: []checkedClaim{{
				ID: "u1.c1", Claim: "Canary uses pool_size=64", Quote: "pool_size=64",
				Support: packet.SupportQuoted, Check: checkVerified,
			}},
		},
		{
			Unit: "u2", Status: "completed",
			Claims: []checkedClaim{{
				ID: "u2.c1", Claim: "Canary uses pool_size=20", Quote: "",
				Support: packet.SupportUnsupported, Check: checkNoEvidence,
			}},
		},
	}
	conflicts := AuditClaimKeyConflicts(recs)
	if len(conflicts) != 1 || conflicts[0].Key != "pool_size" || conflicts[0].DominantUnit != "u1" {
		t.Fatalf("expected Verified-Quote Dominance to crown u1 on pool_size, got %+v", conflicts)
	}
}

func TestForkWorkspaceTimeTravelReconstructsHistoricalState(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	t1, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Start with pool_size=20 and port=5432.")
	_, _ = h.e.V.AppendTurn(vault.RoleAssistant, "alpha", "", "", "Configured pool_size=20.")
	t3, _ := h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Switch to pool_size=96 for holiday traffic.")

	d1 := vault.Decision{ID: "dec-1", Statement: "Use pool_size=20", Source: t1.ID, Quote: "pool_size=20", Status: "superseded"}
	d2 := vault.Decision{ID: "dec-2", Statement: "Use pool_size=96", Source: t3.ID, Quote: "pool_size=96", Status: "active", Supersedes: []string{"dec-1"}}
	_ = h.e.V.PutDecision(d1)
	_ = h.e.V.PutDecision(d2)
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}

	// Fork workspace time-traveled back to seq 2 (BEFORE turn 3 happened!).
	fres, err := ForkWorkspace(h.e.Cfg, h.e.Name, "forked-at-2", ForkOptions{AsOfSeq: 2})
	if err != nil {
		t.Fatal(err)
	}
	if fres.TurnsCopied != 2 || fres.DecisionsKept != 1 {
		t.Fatalf("unexpected ForkResult: %+v", fres)
	}

	forked, err := OpenReadOnly(h.e.Cfg, "forked-at-2")
	if err != nil {
		t.Fatal(err)
	}
	defer forked.Close()

	// In the forked timeline at seq 2, dec-1 MUST be "active" again and materialized pool_size MUST be "20"!
	decs := forked.V.Decisions()
	if len(decs) != 1 || decs[0].ID != "dec-1" || decs[0].Status != "active" {
		t.Fatalf("expected dec-1 restored to active in time-traveled fork, got %+v", decs)
	}
	kvs, err := forked.S.ListActiveKeys("pool_size", 0, 10)
	if err != nil || len(kvs) != 1 || kvs[0].RawValue != "20" {
		t.Fatalf("expected forked materialized key pool_size=20, got %+v err=%v", kvs, err)
	}
}

func TestRecallOpsKeyHistoryBridgeAndCoverWindow(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	pad := strings.Repeat("detailed canary rollout telemetry verification notes ", 6)
	_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Ticket INC-7788 root cause: wal_sync_timeout=250ms.")
	_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "", "Checkout degradation linked to INC-7788; bumped wal_sync_timeout=900ms.")
	_, _ = h.e.V.AppendTurn(vault.RoleUser, "user", "", "",
		"Subsystem alpha uses Envoy proxy. "+pad+"\n\nMiddle buffer paragraph. "+pad+"\n\nCertificate rotation uses spire_agent. "+pad)
	if err := h.e.Reindex(); err != nil {
		t.Fatal(err)
	}

	focus := vault.Turn{ID: "t000004", Seq: 4}
	ids, extras, notes := h.e.ExecuteRecallOps([]packet.RecallOp{
		{Op: "key_history", Target: "wal_sync_timeout"},
		{Op: "bridge", Target: "t000002"},
		{Op: "cover_window", Target: "Envoy spire_agent", Radius: 3},
	}, focus, nil)

	if len(extras) != 1 || !strings.Contains(extras[0].Text, "250ms") || !strings.Contains(extras[0].Text, "900ms") {
		t.Fatalf("key_history extra unexpected: %+v", extras)
	}
	joinedIDs := strings.Join(ids, ",")
	if !strings.Contains(joinedIDs, "t000001") || !strings.Contains(joinedIDs, "t000003#a1-3") {
		t.Fatalf("expected bridge to pin t000001 and cover_window to pin t000003#a1-3, got ids=%v notes=%v", ids, notes)
	}
}
