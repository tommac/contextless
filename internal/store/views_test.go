package store

import (
	"path/filepath"
	"testing"
)

func TestMaterializedKeyViewsAndBitemporalTimeTravel(t *testing.T) {
	s, err := OpenFile(filepath.Join(t.TempDir(), "views.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Turn 1 sets pool_size=20 and port: 5432
	a1 := ExtractKeyAssertions(1, 1, "Initial config:\n- pool_size=20\n- port: 5432")
	if err := s.ApplyTurnViews(a1, nil); err != nil {
		t.Fatal(err)
	}
	// Turn 2 echoes pool_size=20 (assistant confirmation) and asks a hypothetical question:
	// Snodgrass interval coalescing + question filtering must preserve [1..3] cleanly!
	a2 := ExtractKeyAssertions(2, 1, "Confirmed pool_size=20.\nWhat if pool_size=999?")
	if err := s.ApplyTurnViews(a2, nil); err != nil {
		t.Fatal(err)
	}
	// Turn 4 updates pool_size=64 while leaving port untouched
	a4 := ExtractKeyAssertions(4, 1, "Bump canary pool_size = 64 after load test.")
	if err := s.ApplyTurnViews(a4, nil); err != nil {
		t.Fatal(err)
	}
	// Turn 7 updates pool_size=128
	a7 := ExtractKeyAssertions(7, 1, "Production GA --pool_size=128.")
	if err := s.ApplyTurnViews(a7, nil); err != nil {
		t.Fatal(err)
	}

	// 1. Live active view: pool_size MUST be 128 (0% stale exposure!) and port=5432.
	active, err := s.ListActiveKeys("", 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, r := range active {
		got[r.Key] = r.RawValue
	}
	if got["pool_size"] != "128" || got["port"] != "5432" {
		t.Fatalf("unexpected live active keys: %+v", got)
	}

	// 2. Time-travel query as-of turn 5: pool_size MUST be 64 (valid_until reconstructed as 0!).
	at5, err := s.ListActiveKeys("pool_size", 5, 10)
	if err != nil || len(at5) != 1 || at5[0].RawValue != "64" || at5[0].ValidUntilSeq != 0 {
		t.Fatalf("time-travel ListActiveKeys at seq 5 failed: %+v err=%v", at5, err)
	}

	// 3. Full version chain for pool_size has 3 ordered validity intervals: [1..3], [4..6], [7..0].
	chain, err := s.QueryKeyHistory("pool_size", 0)
	if err != nil || len(chain) != 3 {
		t.Fatalf("QueryKeyHistory len=%d err=%v: %+v", len(chain), err, chain)
	}
	if chain[0].ValidFromSeq != 1 || chain[0].ValidUntilSeq != 3 || chain[0].SupersededBySeq != 4 {
		t.Fatalf("bad interval 0: %+v", chain[0])
	}
	if chain[1].ValidFromSeq != 4 || chain[1].ValidUntilSeq != 6 || chain[1].SupersededBySeq != 7 {
		t.Fatalf("bad interval 1: %+v", chain[1])
	}
	if chain[2].ValidFromSeq != 7 || chain[2].ValidUntilSeq != 0 {
		t.Fatalf("bad interval 2: %+v", chain[2])
	}
}

func TestBridgeRareKeysPPRDiscovers2HopNeighbor(t *testing.T) {
	s, err := OpenFile(filepath.Join(t.TempDir(), "ppr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Turn 2 mentions ticket INC-9042 and root cause "wal_buffer_overflow" (no mention of "checkout").
	// Turn 5 mentions "checkout latency spike" and ticket INC-9042 (the bridge!).
	// Turn 8 is unrelated noise.
	if err := s.ApplyTurnViews(nil, ExtractRareTokens(2, 1, "Resolved INC-9042 by tuning wal_buffer_overflow threshold.")); err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyTurnViews(nil, ExtractRareTokens(5, 1, "Investigating checkout latency spike tracked under INC-9042.")); err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyTurnViews(nil, ExtractRareTokens(8, 1, "Unrelated ticket OPS-1111 for printer VLAN.")); err != nil {
		t.Fatal(err)
	}

	// Seeding PPR from Turn 5 ("checkout latency spike") must discover Turn 2 via rare key "inc-9042"!
	hits, err := s.BridgeRareKeysPPR([]int{5}, 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].TurnSeq != 2 {
		t.Fatalf("expected 2-hop PPR bridge to turn 2, got %+v", hits)
	}
	foundVia := false
	for _, k := range hits[0].ViaKeys {
		if k == "inc-9042" {
			foundVia = true
		}
	}
	if !foundVia {
		t.Fatalf("expected bridge via inc-9042, got %+v", hits[0].ViaKeys)
	}
}
