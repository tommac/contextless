package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSearchAndEdges(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".cx"), 0o755); err != nil {
		t.Fatal(err)
	}
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	recs := []Rec{
		{ID: "t000001", Type: "turn", Path: "turns/t000001.md", SHA: "a", Bytes: 10, Seq: 1, Text: "We should migrate the billing service to Postgres."},
		{ID: "t000002", Type: "turn", Path: "turns/t000002.md", SHA: "b", Bytes: 10, Seq: 2, Text: "My dog Biscuit likes long walks."},
		{ID: "t000003", Type: "turn", Path: "turns/t000003.md", SHA: "c", Bytes: 10, Seq: 3, Text: "Postgres migration plan: dump, restore, cut over billing."},
	}
	for _, r := range recs {
		if err := s.Upsert(r); err != nil {
			t.Fatal(err)
		}
	}
	hits := s.Search("what was the billing postgres plan?", 5, "turn")
	if len(hits) < 2 {
		t.Fatalf("want 2 hits, got %v", hits)
	}
	for _, h := range hits {
		if h.ID == "t000002" {
			t.Fatalf("unrelated turn matched: %v", hits)
		}
	}
	if err := s.SetEdges("t000003", "threads", []string{"th-billing"}); err != nil {
		t.Fatal(err)
	}
	if got := s.In("th-billing", "threads"); len(got) != 1 || got[0] != "t000003" {
		t.Fatalf("In = %v", got)
	}
}

// A database made before cache accounting gains the columns on open.
func TestRunsTableGainsCacheColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	s, err := OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{"cached_tokens", "cache_write_tokens", "usd_reported", "has_usd"} {
		if _, err := s.DB.Exec(`ALTER TABLE runs DROP COLUMN ` + c); err != nil {
			t.Fatal(err)
		}
	}
	s.Close()
	if s, err = OpenFile(path); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.StartRun(Run{ID: "r1", Kind: "respond", Role: "responder", Profile: "p", Started: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := s.FinishRun(Run{ID: "r1", Status: "ok", InputTokens: 100, CachedTokens: 60, CacheWriteTokens: 30, OutputTokens: 5, USDReported: 0.25, HasUSD: true}); err != nil {
		t.Fatal(err)
	}
	r, ok := s.RunRow("r1")
	if !ok || r.CachedTokens != 60 || r.CacheWriteTokens != 30 || r.USDReported != 0.25 || !r.HasUSD {
		t.Fatalf("%+v, %v", r, ok)
	}
}

// Two processes that open a workspace at once can both decide, from an
// unlocked read, that the atom index must be reset. The second reset must
// see the first one's version and leave its index alone.
func TestResetAtomsToLeavesACurrentIndexAlone(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".cx"), 0o755); err != nil {
		t.Fatal(err)
	}
	a, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	if did, err := a.ResetAtomsTo("v1"); err != nil || !did {
		t.Fatalf("first reset: did=%v err=%v", did, err)
	}
	if err := a.IndexAtoms("t000001", 1, []AtomRec{{ID: "t000001#a1", Text: "Let's go with the walnut ones."}}, ""); err != nil {
		t.Fatal(err)
	}
	// b read the old version before a wrote the new one, and resets late
	if did, err := b.ResetAtomsTo("v1"); err != nil || did {
		t.Fatalf("late reset under the same version: did=%v err=%v, want it to do nothing", did, err)
	}
	if n := len(a.SearchAtoms("walnut", 10, 0)); n != 1 {
		t.Fatalf("walnut is in %d atoms after the late reset, want 1", n)
	}
	if n := len(b.AtomTurns()); n != 1 {
		t.Fatalf("%d turns marked indexed after the late reset, want 1", n)
	}
	// a real change of rules still clears the index
	if did, err := b.ResetAtomsTo("v2"); err != nil || !did {
		t.Fatalf("reset under new rules: did=%v err=%v", did, err)
	}
	if n := len(a.SearchAtoms("walnut", 10, 0)); n != 0 {
		t.Fatalf("walnut is in %d atoms after a real reset, want 0", n)
	}
	if n := len(a.AtomTurns()); n != 0 {
		t.Fatalf("%d turns marked indexed after a real reset, want 0", n)
	}
}
