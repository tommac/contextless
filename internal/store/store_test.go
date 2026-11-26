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
