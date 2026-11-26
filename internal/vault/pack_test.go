package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBodyPackRoundTripCoalescedSpansAndTamperDetection(t *testing.T) {
	v, err := Create(filepath.Join(t.TempDir(), "vault"), "Pack Test", "Verify 4 KiB seekable body pack")
	if err != nil {
		t.Fatal(err)
	}
	var turns []Turn
	for i := 1; i <= 40; i++ {
		body := fmt.Sprintf("Turn %02d operational report: service=payments region=us-east-1 pool_size=%d. %s",
			i, 20+i, strings.Repeat("Latency p99 stayed below 18ms across all canary nodes during rollout. ", 6))
		tr, err := v.AppendTurn(RoleUser, "user", "", "", body)
		if err != nil {
			t.Fatal(err)
		}
		turns = append(turns, tr)
	}

	packPath := filepath.Join(v.Dir, ".cx", "bodies.cxp")
	st, err := BuildBodyPack(v, packPath, DefaultPackBlock)
	if err != nil {
		t.Fatal(err)
	}
	if st.Turns != 40 || st.BlockCount < 2 {
		t.Fatalf("unexpected pack stats: %+v", st)
	}
	if st.Ratio >= 0.70 {
		t.Fatalf("expected meaningful compression (<70%% of raw), got %.2f (%+v)", st.Ratio, st)
	}

	pk, err := OpenBodyPack(packPath)
	if err != nil {
		t.Fatal(err)
	}
	defer pk.Close()

	if err := pk.VerifyAll(); err != nil {
		t.Fatalf("VerifyAll failed: %v", err)
	}

	// Single-turn random access.
	body15, sha15, err := pk.ReadTurn(turns[14].ID)
	if err != nil || body15 != turns[14].Body || sha15 != turns[14].BodySHA {
		t.Fatalf("ReadTurn mismatch: err=%v sha=%q", err, sha15)
	}

	// Coalesced multi-citation span reads: 6 spans from consecutive short turns
	// sharing 1-2 blocks must decode strictly fewer blocks than spans requested!
	var anchors []EvidenceAnchor
	for i := 0; i < 6; i++ {
		tr := turns[i]
		anchors = append(anchors, EvidenceAnchor{
			Turn: tr.ID, BodySHA: tr.BodySHA, Start: 0, End: 42, Seg: AtomSeg,
		})
	}
	spans, rst, err := pk.ReadSpans(anchors)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 6 {
		t.Fatalf("got %d spans, want 6", len(spans))
	}
	for i, sp := range spans {
		want := turns[i].Body[:42]
		if sp.Text != want {
			t.Fatalf("span %d = %q, want %q", i, sp.Text, want)
		}
	}
	if rst.BlocksDecoded >= rst.SpansRequested {
		t.Fatalf("expected block coalescing (blocks_decoded < spans_requested), got %+v", rst)
	}

	// Tampering with a compressed payload byte must fail VerifyAll.
	_ = pk.Close()
	raw, _ := os.ReadFile(packPath)
	raw[packHeaderBytes+4] ^= 0xff
	tampered := filepath.Join(t.TempDir(), "tampered.cxp")
	_ = os.WriteFile(tampered, raw, 0o644)
	pk2, err := OpenBodyPack(tampered)
	if err == nil {
		err = pk2.VerifyAll()
		_ = pk2.Close()
	}
	if err == nil {
		t.Fatal("expected tamper error on corrupted payload block")
	}

	// Tampering with directory digest must fail OpenBodyPack immediately.
	raw2, _ := os.ReadFile(packPath)
	raw2[len(raw2)-2] ^= 0xff
	tamperedDir := filepath.Join(t.TempDir(), "tampered-dir.cxp")
	_ = os.WriteFile(tamperedDir, raw2, 0o644)
	if _, err := OpenBodyPack(tamperedDir); err == nil {
		t.Fatal("expected SHA-256 directory mismatch on corrupted directory")
	}
}

func BenchmarkBodyPackCoalescedSpans(b *testing.B) {
	v, err := Create(filepath.Join(b.TempDir(), "vault"), "Bench", "")
	if err != nil {
		b.Fatal(err)
	}
	var anchors []EvidenceAnchor
	for i := 1; i <= 200; i++ {
		body := fmt.Sprintf("Turn %03d deploy config region=us-west-2 replicas=%d. %s",
			i, i%8+1, strings.Repeat("Canary checks passed without error across all shards. ", 5))
		tr, err := v.AppendTurn(RoleUser, "user", "", "", body)
		if err != nil {
			b.Fatal(err)
		}
		if i%20 == 1 {
			anchors = append(anchors, EvidenceAnchor{Turn: tr.ID, BodySHA: tr.BodySHA, Start: 0, End: 48, Seg: AtomSeg})
		}
	}
	packPath := filepath.Join(v.Dir, ".cx", "bodies.cxp")
	if _, err := BuildBodyPack(v, packPath, DefaultPackBlock); err != nil {
		b.Fatal(err)
	}
	pk, err := OpenBodyPack(packPath)
	if err != nil {
		b.Fatal(err)
	}
	defer pk.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := pk.ReadSpans(anchors); err != nil {
			b.Fatal(err)
		}
	}
}
