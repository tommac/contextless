package store

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordsSearchKeyMigrationPreservesRunsAndSearch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	// Build a pre-migration legacy database where records had TEXT PRIMARY KEY
	// and runs already contained rows.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	legacyDDL := `
	CREATE TABLE records (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		path TEXT NOT NULL,
		sha TEXT NOT NULL,
		bytes INTEGER NOT NULL,
		seq INTEGER NOT NULL DEFAULT 0
	);
	CREATE VIRTUAL TABLE fts USING fts5(id UNINDEXED, type UNINDEXED, text);
	CREATE TABLE runs (
		id TEXT PRIMARY KEY,
		kind TEXT NOT NULL,
		role TEXT NOT NULL,
		profile TEXT NOT NULL,
		status TEXT NOT NULL,
		handoff_status TEXT NOT NULL DEFAULT '',
		exit_class TEXT NOT NULL DEFAULT '',
		planned_tokens INTEGER NOT NULL DEFAULT 0,
		cap_tokens INTEGER NOT NULL DEFAULT 0,
		input_tokens INTEGER NOT NULL DEFAULT 0,
		output_tokens INTEGER NOT NULL DEFAULT 0,
		peak_context INTEGER NOT NULL DEFAULT 0,
		wall_ms INTEGER NOT NULL DEFAULT 0,
		dropped INTEGER NOT NULL DEFAULT 0,
		started TEXT NOT NULL,
		ended TEXT NOT NULL DEFAULT '',
		error TEXT NOT NULL DEFAULT ''
	);
	INSERT INTO records(id,type,path,sha,bytes,seq)
	  VALUES('t000001','turn','turns/t000001.md','sha1',30,1);
	INSERT INTO fts(id,type,text) VALUES('t000001','turn','legacy kubernetes ingress rule');
	INSERT INTO runs(id,kind,role,profile,status,started) VALUES('run-legacy','respond','responder','claude/opus','ok','2026-11-01T00:00:00Z');
	`
	if _, err := db.Exec(legacyDDL); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	// Opening with current Store must migrate records to explicit search_key
	// while preserving runs and keeping FTS search working across updates.
	s, err := OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile legacy db: %v", err)
	}
	defer s.Close()

	runs := s.RecentRuns(10)
	if len(runs) != 1 || runs[0].ID != "run-legacy" {
		t.Fatalf("expected legacy run preserved, got %+v", runs)
	}
	hits, err := s.SearchE("kubernetes ingress", 5, "turn")
	if err != nil || len(hits) != 1 || hits[0].ID != "t000001" {
		t.Fatalf("legacy record not searchable after migration: hits=%+v err=%v", hits, err)
	}
	// Updating the same record must cleanly replace its FTS row without orphans.
	if err := s.Upsert(Rec{ID: "t000001", Type: "turn", Path: "turns/t000001.md", SHA: "sha2", Bytes: 28, Seq: 1, Text: "updated envoy gateway config"}); err != nil {
		t.Fatal(err)
	}
	if oldHits, err := s.SearchE("kubernetes", 5, "turn"); err != nil || len(oldHits) != 0 {
		t.Fatalf("stale FTS row survived update: %+v err=%v", oldHits, err)
	}
	if newHits, err := s.SearchE("envoy gateway", 5, "turn"); err != nil || len(newHits) != 1 {
		t.Fatalf("updated record not found: %+v err=%v", newHits, err)
	}
}

func TestSearchEErrorOnClosedDB(t *testing.T) {
	s, err := OpenFile(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	_ = s.Close()
	if _, err := s.SearchE("hello", 5); err == nil {
		t.Fatal("expected SearchE to return error on closed DB")
	}
	if _, err := s.SearchAtomsE("hello", 5, 10); err == nil {
		t.Fatal("expected SearchAtomsE to return error on closed DB")
	}
}

func TestSearchAtomsECardOnlyWithLimit1AndTwoCharTokens(t *testing.T) {
	s, err := OpenFile(filepath.Join(t.TempDir(), "atoms.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Turn 1 mentions "v2" and "go" in body; Turn 2 matches "migration" ONLY in card (not text).
	err = s.IndexAtomTurns([]AtomTurn{
		{Turn: "t000001", Seq: 1, Card: "go service", Atoms: []AtomRec{{ID: "t000001#a1", Text: "We upgraded go to v2 on db primary"}}},
		{Turn: "t000002", Seq: 2, Card: "aurora migration", Atoms: []AtomRec{{ID: "t000002#a1", Text: "Everything looks normal today."}}},
		{Turn: "t000003", Seq: 3, Card: "notes", Atoms: []AtomRec{{ID: "t000003#a1", Text: "aurora migration checklist step one"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Two-char technical tokens ("go", "v2", "db") must be searchable.
	hits, err := s.SearchAtomsE("go v2 db", 5, 10)
	if err != nil || len(hits) == 0 || hits[0].Turn != "t000001" {
		t.Fatalf("expected 2-char token match on t000001, got %+v err=%v", hits, err)
	}
	// Card-only detection must remain accurate even when limit=1 and another turn matches text.
	cardHits, err := s.SearchAtomsE("aurora migration", 1, 3)
	if err != nil || len(cardHits) != 1 || cardHits[0].Turn != "t000002" {
		t.Fatalf("expected t000002 before seq 3, got %+v err=%v", cardHits, err)
	}
	if !cardHits[0].CardOnly {
		t.Fatalf("expected t000002 CardOnly=true when matching card only, got %+v", cardHits[0])
	}
	// Exact phrase prefilter with underscore must match (FTS5 unicode61 splits on '_').
	_ = s.IndexAtomTurns([]AtomTurn{
		{Turn: "t000004", Seq: 4, Atoms: []AtomRec{{ID: "t000004#a1", Text: "Set pool_size=20 for canary"}}},
	})
	ph, err := s.SearchAtomsPhraseE("pool_size", 5, 10)
	if err != nil || len(ph) != 1 || ph[0].Turn != "t000004" {
		t.Fatalf("expected phrase match on pool_size, got %+v err=%v", ph, err)
	}
}

func TestBatchWriteValuePostingsSignalsAndReceipts(t *testing.T) {
	s, err := OpenFile(filepath.Join(t.TempDir(), "batch.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now().UTC().Truncate(time.Second)
	err = s.BatchWrite(
		[]Rec{{ID: "t000001", Type: "turn", Path: "turns/t000001.md", SHA: "sha-1", Bytes: 12, Seq: 1, Text: "pool_size=20"}},
		[]EdgeSet{{Src: "t000001", Field: "thread", Dsts: []string{"th1"}}},
		[]AtomTurn{{Turn: "t000001", Seq: 1, Atoms: []AtomRec{{ID: "t000001#a1", Text: "pool_size=20"}}}},
		[]TurnMeta{{Seq: 1, ID: "t000001", Role: "user", Created: now, BodySHA: "sha-1", Bytes: 12, Preview: "pool_size=20", AtomCount: 1}},
		[]ValuePosting{
			{Term: "pool", TurnSeq: 1, AtomK: 1, OccID: 0, ValStem: "20"},
			{Term: "size", TurnSeq: 1, AtomK: 1, OccID: 0, ValStem: "20"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta("value_postings_ver", "1"); err != nil {
		t.Fatal(err)
	}
	if !s.HasValuePostings() {
		t.Fatal("expected HasValuePostings=true")
	}

	metas, err := s.TurnMetaBefore(5)
	if err != nil || len(metas) != 1 || metas[0].ID != "t000001" {
		t.Fatalf("TurnMetaBefore: %+v err=%v", metas, err)
	}
	seq, root, err := s.SnapshotRoot(1)
	if err != nil || seq != 1 || len(root) != 64 {
		t.Fatalf("SnapshotRoot: seq=%d root=%q err=%v", seq, root, err)
	}
	vhits, err := s.SearchValuePostingsE(map[string]bool{"pool": true, "size": true}, 5)
	if err != nil || len(vhits) != 1 || vhits[0].Score != 2 {
		t.Fatalf("SearchValuePostingsE: %+v err=%v", vhits, err)
	}

	// Atom signals & working set
	_ = s.LogSignals([]AtomSignal{
		{FocusTurn: "t000002", AtomID: "t000001#a1", Signal: "recalled", Created: now},
		{FocusTurn: "t000003", AtomID: "t000001#a1", Signal: "cited_verified", Created: now},
	})
	ws, err := s.WorkingSetAtoms(4, 10)
	if err != nil || len(ws) != 1 || ws[0] != "t000001#a1" {
		t.Fatalf("WorkingSetAtoms: %+v err=%v", ws, err)
	}
	summary, err := s.SignalSummary()
	if err != nil || summary["recalled"] != 1 || summary["cited_verified"] != 1 {
		t.Fatalf("SignalSummary: %+v err=%v", summary, err)
	}

	// Receipts & predicate matching
	r := Receipt{
		ID: "rcpt-1", Question: "What is pool_size?", SnapshotSeq: 1, SnapshotRoot: root,
		Status: "current", Created: now, Updated: now,
		Predicates: []ReceiptPredicate{{Kind: "term", Key: "pool"}, {Kind: "decision", Key: "dec-1"}},
	}
	if err := s.PutReceipt(r); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.GetReceipt("rcpt-1")
	if err != nil || !ok || len(got.Predicates) != 2 {
		t.Fatalf("GetReceipt: %+v ok=%v err=%v", got, ok, err)
	}
	matched, err := s.MatchingReceiptIDs(2, [][2]string{{"term", "pool"}})
	if err != nil || len(matched) != 1 || matched[0] != "rcpt-1" {
		t.Fatalf("MatchingReceiptIDs: %+v err=%v", matched, err)
	}
}
