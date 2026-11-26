package packet

import (
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/vault"
)

func TestProvenanceSemiringAndSubmodularOptimalityCertificate(t *testing.T) {
	var poly ProvenancePoly
	poly.Add("exact", "#1", 0.90)
	poly.Add("fts", "#2", 0.75)
	poly.Add("value", "#1", 0.80)

	if poly.CorroborationCount() != 3 {
		t.Fatalf("CorroborationCount = %d, want 3", poly.CorroborationCount())
	}
	// 1 - (0.10 * 0.25 * 0.20) = 1 - 0.005 = 0.995
	if poly.CombinedBelief() < 0.99 {
		t.Fatalf("CombinedBelief = %.3f, want >= 0.99", poly.CombinedBelief())
	}
	if !strings.Contains(poly.Format(), "3x corroborated") {
		t.Fatalf("unexpected Format: %s", poly.Format())
	}

	// Submodular selection + Tang LP optimality certificate:
	// Verify Utility <= UpperBoundOPT, OptimalityPct in [40.5, 100], and
	// best-single-item fallback beats density-greedy trap!
	qIDF := map[string]float64{"postgres": 3.0, "failover": 4.0, "quorum": 5.0}
	items := []SubmodularItem{
		{ID: "t1#a1", Tokens: 120, Prior: 0.02, Terms: map[string]float64{"postgres": 1, "failover": 1}, Shingles: WordShingles("postgres failover quorum replica promotion")},
		{ID: "t2#a1", Tokens: 140, Prior: 0.015, Terms: map[string]float64{"quorum": 1}, Shingles: WordShingles("quorum consensus witness node election")},
		{ID: "t3#a1", Tokens: 130, Prior: 0.01, Terms: map[string]float64{"postgres": 1}, Shingles: WordShingles("postgres failover quorum replica promotion")}, // duplicate shingles!
	}
	cert := CertifySubmodularSelection(qIDF, items, 280)
	if cert.Utility > cert.UpperBoundOPT+1e-6 {
		t.Fatalf("unsound upper bound: utility %.4f > upper %.4f", cert.Utility, cert.UpperBoundOPT)
	}
	if cert.OptimalityPct < 80 || cert.OptimalityPct > 100 {
		t.Fatalf("expected tight optimality certificate (>=80%%), got %.1f%% (%+v)", cert.OptimalityPct, cert)
	}
	// t1#a1 + t2#a1 (260 tokens <= 280) covers all 3 query stems; redundant t3#a1 skipped!
	joined := strings.Join(cert.SelectedIDs, ",")
	if !strings.Contains(joined, "t1#a1") || !strings.Contains(joined, "t2#a1") {
		t.Fatalf("expected complementary spans t1#a1 and t2#a1 selected, got %v", cert.SelectedIDs)
	}
}

func TestBestCoverWindowsFindsMinimalMultiAtomSpan(t *testing.T) {
	pad := strings.Repeat("filler operational telemetry details ", 8)
	// A 4-atom turn where atom 1 mentions "aurora", atom 2 is filler, atom 3 mentions "checksum", and atom 4 is filler.
	// Minimal interval covering {"aurora", "checksum"} is strictly [1..3] (not [1..4] and not single atoms!).
	body := "Primary database cluster uses aurora storage engine. " + pad + "\n\n" +
		"Intermediate network throughput metrics look normal. " + pad + "\n\n" +
		"Page verification failed with invalid checksum error. " + pad + "\n\n" +
		"Trailing sign-off paragraph without query words. " + pad
	tr := vault.Turn{ID: "t000001", Seq: 1, Role: vault.RoleUser, Body: body}
	if len(AtomsOf(tr)) != 4 {
		t.Fatalf("expected 4 atoms, got %d", len(AtomsOf(tr)))
	}

	wins := BestCoverWindows([]vault.Turn{tr}, "aurora invalid checksum", 5, 4, 3)
	if len(wins) == 0 {
		t.Fatal("expected BestCoverWindows to find multi-atom interval")
	}
	top := wins[0]
	if top.AtomStart != 1 || top.AtomEnd != 3 || top.ID != "t000001#a1-3" {
		t.Fatalf("expected minimal interval t000001#a1-3, got %+v", top)
	}
}
