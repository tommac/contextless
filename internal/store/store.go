// Package store is the machine's state: a SQLite database at
// <vault>/.cx/state.db holding a rebuildable index of the vault (records,
// edges, full-text search), the run log, and provider health. It is never
// committed and models never see it.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps the database.
type Store struct {
	DB *sql.DB
	// ReadOnly is set when the database file cannot be written: the
	// schema could not be brought up to date and the index is used as
	// it is.
	ReadOnly bool
}

const schema = `
CREATE TABLE IF NOT EXISTS records (
  search_key INTEGER PRIMARY KEY,
  id TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL,
  path TEXT NOT NULL,
  sha TEXT NOT NULL,
  bytes INTEGER NOT NULL,
  seq INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS edges (
  src TEXT NOT NULL,
  dst TEXT NOT NULL,
  field TEXT NOT NULL,
  PRIMARY KEY (src, dst, field)
);
CREATE INDEX IF NOT EXISTS edges_dst ON edges(dst);
CREATE VIRTUAL TABLE IF NOT EXISTS fts USING fts5(id UNINDEXED, type UNINDEXED, text);
CREATE VIRTUAL TABLE IF NOT EXISTS atoms_fts USING fts5(id UNINDEXED, turn UNINDEXED, seq UNINDEXED, text, card, tokenize='porter unicode61');
CREATE TABLE IF NOT EXISTS atom_turns (
  turn TEXT PRIMARY KEY,
  n INTEGER NOT NULL,
  card TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS turn_meta (
  seq INTEGER PRIMARY KEY,
  id TEXT NOT NULL UNIQUE,
  role TEXT NOT NULL,
  author TEXT NOT NULL DEFAULT '',
  reply_to TEXT NOT NULL DEFAULT '',
  created TEXT NOT NULL DEFAULT '',
  body_sha TEXT NOT NULL DEFAULT '',
  bytes INTEGER NOT NULL DEFAULT 0,
  preview TEXT NOT NULL DEFAULT '',
  top_terms TEXT NOT NULL DEFAULT '',
  atom_count INTEGER NOT NULL DEFAULT 0,
  part_count INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS value_postings (
  term TEXT NOT NULL,
  turn_seq INTEGER NOT NULL,
  atom_k INTEGER NOT NULL,
  occ_id INTEGER NOT NULL,
  val_stem TEXT NOT NULL,
  PRIMARY KEY (term, turn_seq, atom_k, occ_id)
);
CREATE INDEX IF NOT EXISTS value_postings_turn ON value_postings(turn_seq);
CREATE TABLE IF NOT EXISTS key_history (
  key TEXT NOT NULL,
  turn_seq INTEGER NOT NULL,
  atom_k INTEGER NOT NULL,
  raw_value TEXT NOT NULL,
  valid_from_seq INTEGER NOT NULL,
  valid_until_seq INTEGER NOT NULL DEFAULT 0,
  superseded_by_seq INTEGER NOT NULL DEFAULT 0,
  quote TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (key, turn_seq, atom_k)
);
CREATE INDEX IF NOT EXISTS key_history_lookup ON key_history(key, valid_from_seq);
CREATE TABLE IF NOT EXISTS rare_keys (
  token TEXT NOT NULL,
  turn_seq INTEGER NOT NULL,
  atom_k INTEGER NOT NULL,
  PRIMARY KEY (token, turn_seq, atom_k)
);
CREATE INDEX IF NOT EXISTS rare_keys_turn ON rare_keys(turn_seq);
CREATE TABLE IF NOT EXISTS atom_signals (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  focus_turn TEXT NOT NULL,
  atom_id TEXT NOT NULL,
  signal TEXT NOT NULL,
  rank_pos INTEGER NOT NULL DEFAULT 0,
  lane TEXT NOT NULL DEFAULT '',
  created TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS atom_signals_atom ON atom_signals(atom_id);
CREATE INDEX IF NOT EXISTS atom_signals_focus ON atom_signals(focus_turn);
CREATE TABLE IF NOT EXISTS receipts (
  id TEXT PRIMARY KEY,
  question TEXT NOT NULL,
  answer_turn TEXT NOT NULL DEFAULT '',
  answer_text TEXT NOT NULL DEFAULT '',
  snapshot_seq INTEGER NOT NULL,
  snapshot_root TEXT NOT NULL,
  as_of_seq INTEGER NOT NULL DEFAULT 0,
  evidence_json TEXT NOT NULL DEFAULT '[]',
  unresolved_scope TEXT NOT NULL DEFAULT '',
  parent_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'current',
  stale_reason TEXT NOT NULL DEFAULT '',
  stale_turns_json TEXT NOT NULL DEFAULT '[]',
  created TEXT NOT NULL,
  updated TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS receipt_predicates (
  receipt_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  key TEXT NOT NULL,
  value TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (receipt_id, kind, key)
);
CREATE INDEX IF NOT EXISTS receipt_predicates_lookup ON receipt_predicates(kind, key);
CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS runs (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  role TEXT NOT NULL,
  profile TEXT NOT NULL,
  focus TEXT NOT NULL DEFAULT '',
  parent TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  exit_class TEXT NOT NULL DEFAULT '',
  handoff_status TEXT NOT NULL DEFAULT '',
  planned_tokens INTEGER NOT NULL DEFAULT 0,
  cap_tokens INTEGER NOT NULL DEFAULT 0,
  peak_context INTEGER NOT NULL DEFAULT 0,
  input_tokens INTEGER NOT NULL DEFAULT 0,
  cached_tokens INTEGER NOT NULL DEFAULT 0,
  cache_write_tokens INTEGER NOT NULL DEFAULT 0,
  output_tokens INTEGER NOT NULL DEFAULT 0,
  usd_reported REAL NOT NULL DEFAULT 0,
  has_usd INTEGER NOT NULL DEFAULT 0,
  wall_ms INTEGER NOT NULL DEFAULT 0,
  dropped INTEGER NOT NULL DEFAULT 0,
  run_dir TEXT NOT NULL DEFAULT '',
  error TEXT NOT NULL DEFAULT '',
  started TEXT NOT NULL,
  ended TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS provider_health (
  provider TEXT PRIMARY KEY,
  cooldown_until TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  strikes INTEGER NOT NULL DEFAULT 0
);
`

// Open opens (and creates) the database under the vault.
func Open(vaultDir string) (*Store, error) {
	return OpenFile(filepath.Join(vaultDir, ".cx", "state.db"))
}

// OpenFile opens (and creates) a database at path; the machine-wide
// provider health lives in <root>/machine.db.
func OpenFile(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	// _txlock=immediate: every transaction here writes, and a deferred
	// one that reads first fails at once when another process commits in
	// between (SQLite does not wait on a lock upgrade). Taking the write
	// lock at BEGIN makes the busy timeout apply.
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_txlock=immediate")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	st := &Store{DB: db}
	for try := 0; ; try++ {
		_, err = db.Exec(schema)
		if err == nil {
			break
		}
		msg := err.Error()
		if strings.Contains(msg, "readonly") || strings.Contains(msg, "read-only") {
			// a database that cannot be written still serves the
			// commands that only read it
			st.ReadOnly = true
			break
		}
		if try < 20 && (strings.Contains(msg, "locked") || strings.Contains(msg, "busy")) {
			time.Sleep(time.Duration(25*(try+1)) * time.Millisecond)
			continue
		}
		db.Close()
		return nil, fmt.Errorf("store: schema: %w", err)
	}
	if !st.ReadOnly {
		st.addRunColumns()
		if err := st.migrateRecordsSearchKey(); err != nil {
			db.Close()
			return nil, fmt.Errorf("store: migrate records: %w", err)
		}
	}
	return st, nil
}

// migrateRecordsSearchKey upgrades pre-search_key databases in place so
// records have an explicit INTEGER PRIMARY KEY (search_key) aligned with
// fts.rowid, avoiding full-table unindexed scans on DELETE FROM fts.
// Runs, provider_health, receipts and signals are untouched.
func (s *Store) migrateRecordsSearchKey() error {
	have := map[string]bool{}
	for _, c := range s.strings(`SELECT name FROM pragma_table_info('records')`) {
		have[c] = true
	}
	if have["search_key"] || !have["id"] {
		return nil
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmts := []string{
		`CREATE TABLE records_v2 (
		  search_key INTEGER PRIMARY KEY,
		  id TEXT NOT NULL UNIQUE,
		  type TEXT NOT NULL,
		  path TEXT NOT NULL,
		  sha TEXT NOT NULL,
		  bytes INTEGER NOT NULL,
		  seq INTEGER NOT NULL DEFAULT 0
		)`,
		`INSERT INTO records_v2(id,type,path,sha,bytes,seq) SELECT id,type,path,sha,bytes,seq FROM records`,
		`CREATE TEMP TABLE fts_mig AS SELECT r.search_key, f.id, f.type, f.text FROM fts f JOIN records_v2 r ON r.id = f.id`,
		`DELETE FROM fts`,
		`INSERT INTO fts(rowid,id,type,text) SELECT search_key, id, type, text FROM fts_mig`,
		`DROP TABLE fts_mig`,
		`DROP TABLE records`,
		`ALTER TABLE records_v2 RENAME TO records`,
	}
	for _, q := range stmts {
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// addRunColumns brings a runs table made before cache accounting up to
// date. Two processes may race here; the loser's "duplicate column" is
// harmless.
func (s *Store) addRunColumns() {
	have := map[string]bool{}
	for _, c := range s.strings(`SELECT name FROM pragma_table_info('runs')`) {
		have[c] = true
	}
	for _, c := range []struct{ name, def string }{
		{"cached_tokens", "INTEGER NOT NULL DEFAULT 0"},
		{"cache_write_tokens", "INTEGER NOT NULL DEFAULT 0"},
		{"usd_reported", "REAL NOT NULL DEFAULT 0"},
		{"has_usd", "INTEGER NOT NULL DEFAULT 0"},
	} {
		if !have[c.name] {
			_, _ = s.DB.Exec(`ALTER TABLE runs ADD COLUMN ` + c.name + ` ` + c.def)
		}
	}
}

// Close closes the database.
func (s *Store) Close() error { return s.DB.Close() }

// Rec is one indexed record.
type Rec struct {
	ID    string
	Type  string
	Path  string
	SHA   string
	Bytes int
	Seq   int
	Text  string // searchable text
}

// Upsert indexes one record and replaces its search text.
func (s *Store) Upsert(r Rec) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := upsertRecTx(tx, r); err != nil {
		return err
	}
	return tx.Commit()
}

func upsertRecTx(tx *sql.Tx, r Rec) error {
	var key int64
	if err := tx.QueryRow(`INSERT INTO records(id,type,path,sha,bytes,seq) VALUES(?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET type=excluded.type, path=excluded.path, sha=excluded.sha, bytes=excluded.bytes, seq=excluded.seq
		RETURNING search_key`,
		r.ID, r.Type, r.Path, r.SHA, r.Bytes, r.Seq).Scan(&key); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM fts WHERE rowid=?`, key); err != nil {
		return err
	}
	if r.Text != "" {
		if _, err := tx.Exec(`INSERT INTO fts(rowid,id,type,text) VALUES(?,?,?,?)`, key, r.ID, r.Type, r.Text); err != nil {
			return err
		}
	}
	return nil
}

// SetEdges replaces the outgoing edges of src for one field.
func (s *Store) SetEdges(src, field string, dsts []string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM edges WHERE src=? AND field=?`, src, field); err != nil {
		return err
	}
	for _, d := range dsts {
		if d == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT OR IGNORE INTO edges(src,dst,field) VALUES(?,?,?)`, src, d, field); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Out returns the targets of src's edges in field ("" for any field).
func (s *Store) Out(src, field string) []string {
	q := `SELECT dst FROM edges WHERE src=?`
	args := []any{src}
	if field != "" {
		q += ` AND field=?`
		args = append(args, field)
	}
	return s.strings(q+` ORDER BY dst`, args...)
}

// In returns the sources pointing at dst in field ("" for any field).
func (s *Store) In(dst, field string) []string {
	q := `SELECT src FROM edges WHERE dst=?`
	args := []any{dst}
	if field != "" {
		q += ` AND field=?`
		args = append(args, field)
	}
	return s.strings(q+` ORDER BY src`, args...)
}

func (s *Store) strings(q string, args ...any) []string {
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if rows.Scan(&v) == nil {
			out = append(out, v)
		}
	}
	return out
}

// Hit is one full-text match.
type Hit struct {
	ID    string
	Type  string
	Score float64
}

var wordRe = regexp.MustCompile(`[\p{L}\p{N}_]{2,}`)
var phraseTokenRe = regexp.MustCompile(`[\p{L}\p{N}]+`)

// stopwords keep common words from swamping the match (including 2-char
// English glue words so technical 2-char tokens like go, db, io, v2, id stay searchable).
var stopwords = map[string]bool{
	"am": true, "an": true, "as": true, "at": true, "be": true, "by": true, "do": true, "he": true,
	"if": true, "in": true, "is": true, "it": true, "me": true, "my": true, "no": true, "of": true,
	"oh": true, "ok": true, "on": true, "or": true, "so": true, "to": true, "up": true, "us": true, "we": true,
	"the": true, "and": true, "for": true, "that": true, "this": true, "with": true, "you": true,
	"are": true, "was": true, "but": true, "not": true, "have": true, "has": true, "can": true,
	"what": true, "how": true, "why": true, "when": true, "who": true, "which": true, "will": true,
	"would": true, "should": true, "could": true, "from": true, "into": true, "about": true,
	"there": true, "their": true, "then": true, "than": true, "them": true, "they": true,
	"your": true, "our": true, "its": true, "also": true, "just": true, "like": true, "any": true,
	"all": true, "out": true, "use": true, "using": true, "want": true, "need": true, "please": true,
}

// matchQuery builds the FTS5 OR query for the words of text: at most 48
// distinct tokens (keeping 2-char technical identifiers), stopwords left out.
func matchQuery(text string) string {
	seen := map[string]bool{}
	var terms []string
	for _, w := range wordRe.FindAllString(strings.ToLower(text), -1) {
		if stopwords[w] || seen[w] {
			continue
		}
		seen[w] = true
		terms = append(terms, `"`+w+`"`)
		if len(terms) >= 48 {
			break
		}
	}
	return strings.Join(terms, " OR ")
}

// phraseMatchQuery builds a lossless FTS5 phrase prefilter for an exact
// substring query. Any atom holding the exact phrase at word boundaries must
// contain these tokens in order, so FTS5 eliminates non-matching turns without
// scanning the corpus on disk.
func phraseMatchQuery(text string) string {
	toks := phraseTokenRe.FindAllString(strings.ToLower(text), -1)
	if len(toks) == 0 {
		return ""
	}
	if len(toks) > 32 {
		toks = toks[:32]
	}
	return `"` + strings.Join(toks, " ") + `"`
}

// Search runs a full-text OR query over the words of text, best first.
// types limits the record types ("" entries ignored). Errors return nil for
// backward compatibility; call SearchE when errors must be distinguished from
// an empty result.
func (s *Store) Search(text string, limit int, types ...string) []Hit {
	hits, _ := s.SearchE(text, limit, types...)
	return hits
}

// SearchE is Search with explicit error reporting so SQLite failures never
// masquerade as an empty match set.
func (s *Store) SearchE(text string, limit int, types ...string) ([]Hit, error) {
	match := matchQuery(text)
	if match == "" {
		return nil, nil
	}
	q := `SELECT id, type, bm25(fts) FROM fts WHERE fts MATCH ?`
	args := []any{match}
	if len(types) > 0 {
		q += ` AND type IN (` + strings.TrimSuffix(strings.Repeat("?,", len(types)), ",") + `)`
		for _, t := range types {
			args = append(args, t)
		}
	}
	q += ` ORDER BY bm25(fts) LIMIT ?`
	args = append(args, limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.ID, &h.Type, &h.Score); err != nil {
			return nil, err
		}
		h.Score = -h.Score // bm25 is lower-is-better
		out = append(out, h)
	}
	return out, rows.Err()
}

// Reset clears the rebuildable index (records, edges, fts, atoms, turn_meta,
// value_postings) before a full rebuild. Runs, provider health, signals and
// receipts are preserved.
func (s *Store) Reset() error {
	return s.execAll(
		`DELETE FROM records`,
		`DELETE FROM edges`,
		`DELETE FROM fts`,
		`DELETE FROM atoms_fts`,
		`DELETE FROM atom_turns`,
		`DELETE FROM turn_meta`,
		`DELETE FROM value_postings`,
		`DELETE FROM key_history`,
		`DELETE FROM rare_keys`,
		`DELETE FROM meta WHERE key='value_postings_ver'`,
	)
}

// execAll runs statements in one transaction.
func (s *Store) execAll(stmts ...string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range stmts {
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// --- atoms: the paragraph-sized pieces of turns, searched with stemming ---

// AtomLayout names how atom rows are stored; a change rebuilds the atom
// index. Layout 2: atom k of the turn with sequence number seq sits at
// rowid seq<<atomShift + k, so a turn's atoms are deleted and a search is
// bounded by rowid range, not by a scan of the table.
const AtomLayout = 2

const atomShift = 20

// AtomRec is one atom to index.
type AtomRec struct {
	ID   string
	Text string
}

// AtomTurn is one turn's atoms to index, with the turn's card text.
type AtomTurn struct {
	Turn  string
	Seq   int
	Atoms []AtomRec
	Card  string
}

// IndexAtoms indexes the atoms of one turn. card is the turn's card text
// (its summary and keywords, "" before the turn is indexed): it is
// searched with every atom of the turn, as it is with the turn itself, so
// a question in the card's words still reaches the turn's paragraphs. A
// turn is immutable, so its atoms are written once and again only when
// its card text changes. The marker row is written even for a turn with
// no atoms.
func (s *Store) IndexAtoms(turn string, seq int, atoms []AtomRec, card string) error {
	return s.IndexAtomTurns([]AtomTurn{{Turn: turn, Seq: seq, Atoms: atoms, Card: card}})
}

// IndexAtomTurns indexes several turns' atoms in one transaction.
func (s *Store) IndexAtomTurns(batch []AtomTurn) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := indexAtomTurnsTx(tx, batch); err != nil {
		return err
	}
	return tx.Commit()
}

func indexAtomTurnsTx(tx *sql.Tx, batch []AtomTurn) error {
	for _, t := range batch {
		if t.Seq < 1 || len(t.Atoms) >= 1<<atomShift {
			return fmt.Errorf("store: atoms of %s cannot be indexed (seq %d, %d atoms)", t.Turn, t.Seq, len(t.Atoms))
		}
		var old string
		switch err := tx.QueryRow(`SELECT card FROM atom_turns WHERE turn=?`, t.Turn).Scan(&old); {
		case err == sql.ErrNoRows:
		case err != nil:
			return err
		case old == t.Card:
			continue
		}
		// always clear the turn's rowid range: rows left by a failed or
		// partial write must not be doubled
		base := int64(t.Seq) << atomShift
		if _, err := tx.Exec(`DELETE FROM atoms_fts WHERE rowid > ? AND rowid < ?`, base, base+1<<atomShift); err != nil {
			return err
		}
		for k, a := range t.Atoms {
			if _, err := tx.Exec(`INSERT INTO atoms_fts(rowid,id,turn,seq,text,card) VALUES(?,?,?,?,?,?)`, base+int64(k)+1, a.ID, t.Turn, t.Seq, a.Text, t.Card); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`INSERT INTO atom_turns(turn,n,card) VALUES(?,?,?)
			ON CONFLICT(turn) DO UPDATE SET n=excluded.n, card=excluded.card`, t.Turn, len(t.Atoms), t.Card); err != nil {
			return err
		}
	}
	return nil
}

// AtomTurns returns the turns whose atoms are indexed, each with the card
// text its atoms were indexed with.
func (s *Store) AtomTurns() map[string]string {
	m, _ := s.AtomTurnsE()
	if m == nil {
		return map[string]string{}
	}
	return m
}

// AtomTurnsE is AtomTurns with explicit error reporting.
func (s *Store) AtomTurnsE() (map[string]string, error) {
	out := map[string]string{}
	rows, err := s.DB.Query(`SELECT turn, card FROM atom_turns`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t, c string
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		out[t] = c
	}
	return out, rows.Err()
}

// ResetAtoms clears the atom index (the splitting rules changed).
func (s *Store) ResetAtoms() error {
	return s.execAll(`DELETE FROM atoms_fts`, `DELETE FROM atom_turns`)
}

// Meta reads a small key-value setting of the index ("" when unset).
func (s *Store) Meta(key string) string {
	var v string
	_ = s.DB.QueryRow(`SELECT value FROM meta WHERE key=?`, key).Scan(&v)
	return v
}

// SetMeta writes a setting of the index.
func (s *Store) SetMeta(key, value string) error {
	_, err := s.DB.Exec(`INSERT INTO meta(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// AtomHit is one full-text match among atoms.
type AtomHit struct {
	ID    string
	Turn  string
	Score float64
	// CardOnly is set when the atom matched only through its turn's card
	// text: none of the query's words is in the atom itself.
	CardOnly bool
}

// SearchAtoms runs the same OR query as Search over the atoms of turns
// before beforeSeq (0 for all), best first. The atom index is stemmed
// (porter), so "rejected" meets "rejections". Call SearchAtomsE when DB
// errors must not look like an empty hit list.
func (s *Store) SearchAtoms(text string, limit, beforeSeq int) []AtomHit {
	hits, _ := s.SearchAtomsE(text, limit, beforeSeq)
	return hits
}

// SearchAtomsE is SearchAtoms with explicit error reporting and exact
// CardOnly classification (checking body match directly on the returned
// rowids rather than against a second truncated LIMIT).
func (s *Store) SearchAtomsE(text string, limit, beforeSeq int) ([]AtomHit, error) {
	match := matchQuery(text)
	if match == "" {
		return nil, nil
	}
	q := `SELECT rowid, id, turn, bm25(atoms_fts) FROM atoms_fts WHERE atoms_fts MATCH ?`
	args := []any{match}
	if beforeSeq > 0 {
		q += ` AND rowid < ?`
		args = append(args, int64(beforeSeq)<<atomShift)
	}
	q += ` ORDER BY bm25(atoms_fts) LIMIT ?`
	args = append(args, limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []AtomHit
	var rowids []int64
	for rows.Next() {
		var rid int64
		var h AtomHit
		if err := rows.Scan(&rid, &h.ID, &h.Turn, &h.Score); err != nil {
			return nil, err
		}
		h.Score = -h.Score
		hits = append(hits, h)
		rowids = append(rowids, rid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(hits) == 0 {
		return nil, nil
	}
	// Verify body-text match specifically on the returned rowids (chunked to
	// stay well under SQLite's bind parameter limit) so LIMIT truncation on a
	// second query can never falsely flag a body hit as CardOnly.
	own := map[int64]bool{}
	textMatch := "{text} : (" + match + ")"
	const chunk = 400
	for start := 0; start < len(rowids); start += chunk {
		end := min(start+chunk, len(rowids))
		ph := strings.TrimSuffix(strings.Repeat("?,", end-start), ",")
		cq := `SELECT rowid FROM atoms_fts WHERE atoms_fts MATCH ? AND rowid IN (` + ph + `)`
		cargs := make([]any, 0, 1+(end-start))
		cargs = append(cargs, textMatch)
		for _, rid := range rowids[start:end] {
			cargs = append(cargs, rid)
		}
		crows, err := s.DB.Query(cq, cargs...)
		if err != nil {
			return nil, err
		}
		for crows.Next() {
			var rid int64
			if err := crows.Scan(&rid); err != nil {
				crows.Close()
				return nil, err
			}
			own[rid] = true
		}
		if err := crows.Err(); err != nil {
			crows.Close()
			return nil, err
		}
		crows.Close()
	}
	for i := range hits {
		hits[i].CardOnly = !own[rowids[i]]
	}
	return hits, nil
}

// SearchAtomsPhraseE uses FTS5 phrase matching over atom bodies as a zero-scan
// prefilter for the exact-phrase lane, newest first.
func (s *Store) SearchAtomsPhraseE(phrase string, limit, beforeSeq int) ([]AtomHit, error) {
	pm := phraseMatchQuery(phrase)
	if pm == "" {
		return nil, nil
	}
	q := `SELECT id, turn FROM atoms_fts WHERE atoms_fts MATCH ?`
	args := []any{"{text} : " + pm}
	if beforeSeq > 0 {
		q += ` AND rowid < ?`
		args = append(args, int64(beforeSeq)<<atomShift)
	}
	q += ` ORDER BY rowid DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AtomHit
	for rows.Next() {
		var h AtomHit
		if err := rows.Scan(&h.ID, &h.Turn); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// --- runs ---

// Run is one row of the run log.
type Run struct {
	ID            string
	Kind          string
	Role          string
	Profile       string
	Focus         string
	Parent        string
	Status        string
	ExitClass     string
	HandoffStatus string
	PlannedTokens int
	CapTokens     int
	PeakContext   int
	InputTokens   int // the whole prompt side: uncached, cache reads and cache writes
	// CachedTokens and CacheWriteTokens are the parts of InputTokens read
	// from and written to a prompt cache.
	CachedTokens     int
	CacheWriteTokens int
	OutputTokens     int
	USDReported      float64
	HasUSD           bool
	WallMs           int64
	Dropped          int
	RunDir           string
	Error            string
	Started          time.Time
	Ended            time.Time
}

// StartRun inserts a run row.
func (s *Store) StartRun(r Run) error {
	_, err := s.DB.Exec(`INSERT INTO runs(id,kind,role,profile,focus,parent,status,planned_tokens,cap_tokens,dropped,run_dir,started)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.ID, r.Kind, r.Role, r.Profile, r.Focus, r.Parent, "running", r.PlannedTokens, r.CapTokens, r.Dropped, r.RunDir,
		r.Started.UTC().Format(time.RFC3339))
	return err
}

// FinishRun records a run's outcome.
func (s *Store) FinishRun(r Run) error {
	_, err := s.DB.Exec(`UPDATE runs SET status=?, exit_class=?, handoff_status=?, peak_context=?, input_tokens=?, cached_tokens=?, cache_write_tokens=?, output_tokens=?, usd_reported=?, has_usd=?, wall_ms=?, error=?, ended=? WHERE id=?`,
		r.Status, r.ExitClass, r.HandoffStatus, r.PeakContext, r.InputTokens, r.CachedTokens, r.CacheWriteTokens, r.OutputTokens, r.USDReported, r.HasUSD, r.WallMs, r.Error,
		time.Now().UTC().Format(time.RFC3339), r.ID)
	return err
}

// AbandonRunning marks runs left "running" by a dead process. Call it only
// while holding the workspace lock, when no run of this workspace can be
// live.
func (s *Store) AbandonRunning() int {
	res, err := s.DB.Exec(`UPDATE runs SET status='abandoned', error='the process running it exited', ended=? WHERE status='running'`,
		time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0
	}
	n, _ := res.RowsAffected()
	return int(n)
}

// HasRecord reports whether a record is in the index.
func (s *Store) HasRecord(id string) bool {
	var n int
	return s.DB.QueryRow(`SELECT 1 FROM records WHERE id=?`, id).Scan(&n) == nil
}

// RunRow reads one run.
func (s *Store) RunRow(id string) (Run, bool) {
	var r Run
	err := s.DB.QueryRow(`SELECT id,kind,role,profile,status,input_tokens,cached_tokens,cache_write_tokens,output_tokens,usd_reported,has_usd,wall_ms FROM runs WHERE id=?`, id).
		Scan(&r.ID, &r.Kind, &r.Role, &r.Profile, &r.Status, &r.InputTokens, &r.CachedTokens, &r.CacheWriteTokens, &r.OutputTokens, &r.USDReported, &r.HasUSD, &r.WallMs)
	return r, err == nil
}

// RecentRuns lists the last n runs, newest first.
func (s *Store) RecentRuns(n int) []Run {
	rows, err := s.DB.Query(`SELECT id,kind,role,profile,status,exit_class,handoff_status,planned_tokens,cap_tokens,peak_context,input_tokens,output_tokens,wall_ms,dropped,error,started
		FROM runs ORDER BY started DESC, id DESC LIMIT ?`, n)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Run
	for rows.Next() {
		var r Run
		var started string
		if rows.Scan(&r.ID, &r.Kind, &r.Role, &r.Profile, &r.Status, &r.ExitClass, &r.HandoffStatus, &r.PlannedTokens, &r.CapTokens,
			&r.PeakContext, &r.InputTokens, &r.OutputTokens, &r.WallMs, &r.Dropped, &r.Error, &started) == nil {
			r.Started, _ = time.Parse(time.RFC3339, started)
			out = append(out, r)
		}
	}
	return out
}

// Stat aggregates the runs of one profile and kind.
type Stat struct {
	Profile    string
	Kind       string
	Role       string
	Runs       int
	OK         int
	MedianWall int64
	AvgPlanned int
	MaxPeak    int
	Failures   map[string]int
}

// Stats aggregates every finished run by profile, kind and role.
func (s *Store) Stats() []Stat {
	rows, err := s.DB.Query(`SELECT profile, kind, role, status, exit_class, wall_ms, planned_tokens, peak_context FROM runs WHERE status != 'running' ORDER BY profile, kind, role`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	type key struct{ p, k, r string }
	agg := map[key]*Stat{}
	walls := map[key][]int64{}
	var order []key
	for rows.Next() {
		var p, k, r, st, ec string
		var wall int64
		var planned, peak int
		if rows.Scan(&p, &k, &r, &st, &ec, &wall, &planned, &peak) != nil {
			continue
		}
		kk := key{p, k, r}
		a := agg[kk]
		if a == nil {
			a = &Stat{Profile: p, Kind: k, Role: r, Failures: map[string]int{}}
			agg[kk] = a
			order = append(order, kk)
		}
		a.Runs++
		a.AvgPlanned += planned
		if peak > a.MaxPeak {
			a.MaxPeak = peak
		}
		if st == "ok" {
			a.OK++
			walls[kk] = append(walls[kk], wall)
		} else {
			reason := ec
			if reason == "" || reason == "ok" {
				reason = st
			}
			a.Failures[reason]++
		}
	}
	out := make([]Stat, 0, len(order))
	for _, kk := range order {
		a := agg[kk]
		a.AvgPlanned /= a.Runs
		if w := walls[kk]; len(w) > 0 {
			sortInt64(w)
			a.MedianWall = w[len(w)/2]
		}
		out = append(out, *a)
	}
	return out
}

func sortInt64(a []int64) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

// --- provider health ---

// Cooldown sets a provider's cooldown and bumps its strike count.
func (s *Store) Cooldown(provider, reason string, until time.Time) error {
	_, err := s.DB.Exec(`INSERT INTO provider_health(provider,cooldown_until,reason,strikes) VALUES(?,?,?,1)
		ON CONFLICT(provider) DO UPDATE SET cooldown_until=excluded.cooldown_until, reason=excluded.reason, strikes=provider_health.strikes+1`,
		provider, until.UTC().Format(time.RFC3339), reason)
	return err
}

// ClearStrikes resets a provider after a clean run.
func (s *Store) ClearStrikes(provider string) {
	_, _ = s.DB.Exec(`UPDATE provider_health SET strikes=0 WHERE provider=?`, provider)
}

// Health returns a provider's cooldown end, reason, and strikes.
func (s *Store) Health(provider string) (until time.Time, reason string, strikes int) {
	var u string
	if s.DB.QueryRow(`SELECT cooldown_until, reason, strikes FROM provider_health WHERE provider=?`, provider).Scan(&u, &reason, &strikes) != nil {
		return time.Time{}, "", 0
	}
	until, _ = time.Parse(time.RFC3339, u)
	return until, reason, strikes
}
