package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// Time the search stage on a preindexed corpus. Identical source IDs must
// come back in both the one-query and five-query request.
func BenchmarkRecallRequest(b *testing.B) {
	dir := b.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "turns"), 0700); err != nil {
		b.Fatal(err)
	}
	queries := []string{"amberthistle", "quartzriver", "indigospan", "malachiteslope", "coralharbor"}
	var want []string
	for seq := 1; seq <= 2048; seq++ {
		id := vault.TurnID(seq)
		body := fmt.Sprintf("Receipt #%d. Reconcile carton weights at the staging dock before close. The packing list and ledger both record a routine transfer.", seq)
		if seq >= 11 && seq < 11+len(queries) {
			body = "Calibration signal " + queries[seq-11] + " was 44 kelvin at station Delta."
			want = append(want, id)
		}
		text := fmt.Sprintf("---\nid: %s\nseq: %d\nrole: user\ncreated: 2026-01-01T00:00:00Z\nbody_sha: %s\n---\n%s\n", id, seq, vault.SHA(body), body)
		if err := os.WriteFile(filepath.Join(dir, vault.TurnPath(id)), []byte(text), 0600); err != nil {
			b.Fatal(err)
		}
	}
	db, err := store.Open(dir)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { db.Close() })
	e := &Engine{V: &vault.Vault{Dir: dir}, S: db}
	e.Cfg.RelevantTokens = 8000
	if err := e.Reindex(); err != nil {
		b.Fatal(err)
	}
	for _, count := range []int{1, 5} {
		b.Run(fmt.Sprintf("%d-query", count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				add, _ := e.recallRequest(nil, queries[:count], vault.Turn{Seq: 2049}, nil)
				if !slices.Equal(add, want[:count]) {
					b.Fatalf("returned %v, expected %v", add, want[:count])
				}
			}
		})
	}
}
