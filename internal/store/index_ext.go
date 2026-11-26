package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// EdgeSet is one record's outgoing edges for a single field.
type EdgeSet struct {
	Src   string
	Field string
	Dsts  []string
}

// TurnMeta caches immutable metadata of a turn so packet selection,
// pending checks, and history budgeting need not read every turn file
// from disk.
type TurnMeta struct {
	Seq       int
	ID        string
	Role      string
	Author    string
	ReplyTo   string
	Created   time.Time
	BodySHA   string
	Bytes     int
	Preview   string
	TopTerms  []string
	AtomCount int
	PartCount int
}

// ValuePosting maps a normalized context stem to a value occurrence within
// four tokens in the same sentence of an atom, matching packet.valueScore
// with exact parity without scanning every turn body at query time.
type ValuePosting struct {
	Term    string
	TurnSeq int
	AtomK   int
	OccID   int
	ValStem string
}

// ValuePostingHit is one atom's best value-lane match count from the index.
type ValuePostingHit struct {
	TurnSeq int
	AtomK   int
	Score   int
}

// AtomSignal records free behavioral telemetry from live runs: selection,
// page faults (needs_recall), and quote-verified citations.
type AtomSignal struct {
	ID        int64
	FocusTurn string
	AtomID    string
	Signal    string // selected | recalled | cited_verified | cited_unverified
	RankPos   int
	Lane      string
	Created   time.Time
}

// ReceiptEvidence is one verified byte span backing a saved Query Receipt.
type ReceiptEvidence struct {
	Turn    string `json:"turn"`
	BodySHA string `json:"body_sha"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Seg     int    `json:"seg"`
	AtomRef string `json:"atom_ref,omitempty"`
	Quote   string `json:"quote,omitempty"`
}

// ReceiptPredicate declares a condition that invalidates a live Query Receipt
// when a new record satisfies it.
type ReceiptPredicate struct {
	Kind  string `json:"kind"`  // term | entity | fact | decision | absence | thread
	Key   string `json:"key"`   // normalized lookup key
	Value string `json:"value"` // optional expected/prior value or detail
}

// Receipt binds a question and answer to an immutable vault snapshot, its
// verified evidence spans, its watch predicates (including absence checks),
// and any explicitly unresolved scope.
type Receipt struct {
	ID              string             `json:"id"`
	Question        string             `json:"question"`
	AnswerTurn      string             `json:"answer_turn,omitempty"`
	AnswerText      string             `json:"answer_text,omitempty"`
	SnapshotSeq     int                `json:"snapshot_seq"`
	SnapshotRoot    string             `json:"snapshot_root"`
	AsOfSeq         int                `json:"as_of_seq,omitempty"` // 0 = live; >0 = historical cutoff
	Evidence        []ReceiptEvidence  `json:"evidence,omitempty"`
	Predicates      []ReceiptPredicate `json:"predicates,omitempty"`
	UnresolvedScope string             `json:"unresolved_scope,omitempty"`
	ParentID        string             `json:"parent_id,omitempty"`
	Status          string             `json:"status"` // current | stale | superseded
	StaleReason     string             `json:"stale_reason,omitempty"`
	StaleTurns      []string           `json:"stale_turns,omitempty"`
	Created         time.Time          `json:"created"`
	Updated         time.Time          `json:"updated"`
}

// BatchWrite atomically writes records, edges, atom FTS rows, turn metadata,
// and value postings in a single SQLite transaction. Used by Reindex and
// bulk imports to avoid per-turn transaction storms.
func (s *Store) BatchWrite(recs []Rec, edges []EdgeSet, atoms []AtomTurn, metas []TurnMeta, postings []ValuePosting) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, r := range recs {
		if err := upsertRecTx(tx, r); err != nil {
			return err
		}
	}
	for _, e := range edges {
		if _, err := tx.Exec(`DELETE FROM edges WHERE src=? AND field=?`, e.Src, e.Field); err != nil {
			return err
		}
		for _, d := range e.Dsts {
			if d == "" {
				continue
			}
			if _, err := tx.Exec(`INSERT OR IGNORE INTO edges(src,dst,field) VALUES(?,?,?)`, e.Src, d, e.Field); err != nil {
				return err
			}
		}
	}
	if len(atoms) > 0 {
		if err := indexAtomTurnsTx(tx, atoms); err != nil {
			return err
		}
	}
	for _, m := range metas {
		if err := putTurnMetaTx(tx, m); err != nil {
			return err
		}
	}
	if len(postings) > 0 {
		seenTurn := map[int]bool{}
		for _, p := range postings {
			if !seenTurn[p.TurnSeq] {
				seenTurn[p.TurnSeq] = true
				if _, err := tx.Exec(`DELETE FROM value_postings WHERE turn_seq=?`, p.TurnSeq); err != nil {
					return err
				}
			}
		}
		stmt, err := tx.Prepare(`INSERT OR IGNORE INTO value_postings(term,turn_seq,atom_k,occ_id,val_stem) VALUES(?,?,?,?,?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, p := range postings {
			if _, err := stmt.Exec(p.Term, p.TurnSeq, p.AtomK, p.OccID, p.ValStem); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func putTurnMetaTx(tx *sql.Tx, m TurnMeta) error {
	terms := strings.Join(m.TopTerms, "\x1f")
	created := ""
	if !m.Created.IsZero() {
		created = m.Created.UTC().Format(time.RFC3339)
	}
	_, err := tx.Exec(`INSERT INTO turn_meta(seq,id,role,author,reply_to,created,body_sha,bytes,preview,top_terms,atom_count,part_count)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(seq) DO UPDATE SET id=excluded.id, role=excluded.role, author=excluded.author,
		reply_to=excluded.reply_to, created=excluded.created, body_sha=excluded.body_sha, bytes=excluded.bytes,
		preview=excluded.preview, top_terms=excluded.top_terms, atom_count=excluded.atom_count, part_count=excluded.part_count`,
		m.Seq, m.ID, m.Role, m.Author, m.ReplyTo, created, m.BodySHA, m.Bytes, m.Preview, terms, m.AtomCount, max(m.PartCount, 1))
	return err
}

// PutTurnMeta stores one turn's cached metadata.
func (s *Store) PutTurnMeta(m TurnMeta) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := putTurnMetaTx(tx, m); err != nil {
		return err
	}
	return tx.Commit()
}

// TurnMetaBefore returns cached turn metadata with seq < beforeSeq (or all
// when beforeSeq <= 0), ordered by seq ascending.
func (s *Store) TurnMetaBefore(beforeSeq int) ([]TurnMeta, error) {
	q := `SELECT seq,id,role,author,reply_to,created,body_sha,bytes,preview,top_terms,atom_count,part_count FROM turn_meta`
	var args []any
	if beforeSeq > 0 {
		q += ` WHERE seq < ?`
		args = append(args, beforeSeq)
	}
	q += ` ORDER BY seq ASC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TurnMeta
	for rows.Next() {
		var m TurnMeta
		var created, terms string
		if err := rows.Scan(&m.Seq, &m.ID, &m.Role, &m.Author, &m.ReplyTo, &created, &m.BodySHA, &m.Bytes, &m.Preview, &terms, &m.AtomCount, &m.PartCount); err != nil {
			return nil, err
		}
		if created != "" {
			m.Created, _ = time.Parse(time.RFC3339, created)
		}
		if terms != "" {
			m.TopTerms = strings.Split(terms, "\x1f")
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// LastTurnMeta returns the highest-seq turn metadata row without reading files.
func (s *Store) LastTurnMeta() (TurnMeta, bool, error) {
	var m TurnMeta
	var created, terms string
	err := s.DB.QueryRow(`SELECT seq,id,role,author,reply_to,created,body_sha,bytes,preview,top_terms,atom_count,part_count
		FROM turn_meta ORDER BY seq DESC LIMIT 1`).
		Scan(&m.Seq, &m.ID, &m.Role, &m.Author, &m.ReplyTo, &created, &m.BodySHA, &m.Bytes, &m.Preview, &terms, &m.AtomCount, &m.PartCount)
	if err == sql.ErrNoRows {
		return TurnMeta{}, false, nil
	}
	if err != nil {
		return TurnMeta{}, false, err
	}
	if created != "" {
		m.Created, _ = time.Parse(time.RFC3339, created)
	}
	if terms != "" {
		m.TopTerms = strings.Split(terms, "\x1f")
	}
	return m, true, nil
}

// SnapshotRoot computes a deterministic cryptographic root digest covering
// all indexed turns up to maxSeq (0 for all) from their immutable (seq, sha)
// pairs. Receipts and external omission catalogs pin to this root.
func (s *Store) SnapshotRoot(maxSeq int) (lastSeq int, root string, err error) {
	q := `SELECT seq, body_sha FROM turn_meta`
	var args []any
	if maxSeq > 0 {
		q += ` WHERE seq <= ?`
		args = append(args, maxSeq)
	}
	q += ` ORDER BY seq ASC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()
	h := sha256.New()
	h.Write([]byte("cx-snapshot/v1\x00"))
	var buf [8]byte
	for rows.Next() {
		var seq int
		var sha string
		if err := rows.Scan(&seq, &sha); err != nil {
			return 0, "", err
		}
		lastSeq = seq
		binary.LittleEndian.PutUint64(buf[:], uint64(seq))
		h.Write(buf[:])
		h.Write([]byte(sha))
		h.Write([]byte{0})
	}
	if err := rows.Err(); err != nil {
		return 0, "", err
	}
	return lastSeq, hex.EncodeToString(h.Sum(nil)), nil
}

// IndexValuePostings replaces the value postings for one turn.
func (s *Store) IndexValuePostings(turnSeq int, postings []ValuePosting) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM value_postings WHERE turn_seq=?`, turnSeq); err != nil {
		return err
	}
	if len(postings) > 0 {
		stmt, err := tx.Prepare(`INSERT OR IGNORE INTO value_postings(term,turn_seq,atom_k,occ_id,val_stem) VALUES(?,?,?,?,?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, p := range postings {
			if _, err := stmt.Exec(p.Term, p.TurnSeq, p.AtomK, p.OccID, p.ValStem); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// HasValuePostings reports whether value postings have been built for this store.
func (s *Store) HasValuePostings() bool {
	return s.Meta("value_postings_ver") == "1"
}

// SearchValuePostingsE computes each atom's best value-lane match count over
// indexed postings, matching packet.valueScore exactly while skipping value
// occurrences whose own stem is already in queryStems.
func (s *Store) SearchValuePostingsE(queryStems map[string]bool, beforeSeq int) ([]ValuePostingHit, error) {
	if len(queryStems) == 0 {
		return nil, nil
	}
	terms := make([]string, 0, len(queryStems))
	for w := range queryStems {
		terms = append(terms, w)
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(terms)), ",")
	q := `SELECT turn_seq, atom_k, occ_id, val_stem, COUNT(DISTINCT term) AS sc
		FROM value_postings WHERE term IN (` + ph + `)`
	args := make([]any, 0, len(terms)+1)
	for _, t := range terms {
		args = append(args, t)
	}
	if beforeSeq > 0 {
		q += ` AND turn_seq < ?`
		args = append(args, beforeSeq)
	}
	q += ` GROUP BY turn_seq, atom_k, occ_id, val_stem`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type ak struct{ seq, k int }
	best := map[ak]int{}
	for rows.Next() {
		var seq, k, occ, sc int
		var valStem string
		if err := rows.Scan(&seq, &k, &occ, &valStem, &sc); err != nil {
			return nil, err
		}
		if queryStems[valStem] {
			continue
		}
		key := ak{seq, k}
		if sc > best[key] {
			best[key] = sc
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]ValuePostingHit, 0, len(best))
	for key, sc := range best {
		out = append(out, ValuePostingHit{TurnSeq: key.seq, AtomK: key.k, Score: sc})
	}
	return out, nil
}

// --- free-signal telemetry (selected / recalled / verified citations) ---

// LogSignals appends one or more atom signals in a single transaction.
func (s *Store) LogSignals(sigs []AtomSignal) error {
	if len(sigs) == 0 || s.ReadOnly {
		return nil
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO atom_signals(focus_turn,atom_id,signal,rank_pos,lane,created) VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	for _, g := range sigs {
		ts := now
		if !g.Created.IsZero() {
			ts = g.Created.UTC().Format(time.RFC3339)
		}
		if _, err := stmt.Exec(g.FocusTurn, g.AtomID, g.Signal, g.RankPos, g.Lane, ts); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// WorkingSetAtoms returns atom/turn ids recalled or verified-cited within
// the most recent windowTurns distinct focus turns, most recently used first.
func (s *Store) WorkingSetAtoms(windowTurns, limit int) ([]string, error) {
	if windowTurns <= 0 || limit <= 0 {
		return nil, nil
	}
	rows, err := s.DB.Query(`
		SELECT atom_id FROM atom_signals
		WHERE signal IN ('recalled', 'cited_verified')
		  AND focus_turn IN (
		    SELECT DISTINCT focus_turn FROM atom_signals ORDER BY id DESC LIMIT ?
		  )
		GROUP BY atom_id
		ORDER BY MAX(id) DESC
		LIMIT ?`, windowTurns, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// SignalSummary aggregates counts by signal kind across the workspace.
func (s *Store) SignalSummary() (map[string]int, error) {
	rows, err := s.DB.Query(`SELECT signal, COUNT(*) FROM atom_signals GROUP BY signal`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var sig string
		var n int
		if err := rows.Scan(&sig, &n); err != nil {
			return nil, err
		}
		out[sig] = n
	}
	return out, rows.Err()
}

// --- living grounded answers (Query Receipts) ---

// PutReceipt inserts or updates a Query Receipt and its watch predicates.
func (s *Store) PutReceipt(r Receipt) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	evJSON, err := json.Marshal(r.Evidence)
	if err != nil {
		return err
	}
	stJSON, err := json.Marshal(r.StaleTurns)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if r.Created.IsZero() {
		r.Created = now
	}
	if r.Updated.IsZero() {
		r.Updated = now
	}
	if r.Status == "" {
		r.Status = "current"
	}
	if _, err := tx.Exec(`INSERT INTO receipts(id,question,answer_turn,answer_text,snapshot_seq,snapshot_root,as_of_seq,evidence_json,unresolved_scope,parent_id,status,stale_reason,stale_turns_json,created,updated)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET question=excluded.question, answer_turn=excluded.answer_turn, answer_text=excluded.answer_text,
		snapshot_seq=excluded.snapshot_seq, snapshot_root=excluded.snapshot_root, as_of_seq=excluded.as_of_seq,
		evidence_json=excluded.evidence_json, unresolved_scope=excluded.unresolved_scope, parent_id=excluded.parent_id,
		status=excluded.status, stale_reason=excluded.stale_reason, stale_turns_json=excluded.stale_turns_json, updated=excluded.updated`,
		r.ID, r.Question, r.AnswerTurn, r.AnswerText, r.SnapshotSeq, r.SnapshotRoot, r.AsOfSeq,
		string(evJSON), r.UnresolvedScope, r.ParentID, r.Status, r.StaleReason, string(stJSON),
		r.Created.UTC().Format(time.RFC3339), r.Updated.UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM receipt_predicates WHERE receipt_id=?`, r.ID); err != nil {
		return err
	}
	for _, p := range r.Predicates {
		if strings.TrimSpace(p.Kind) == "" || strings.TrimSpace(p.Key) == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT OR REPLACE INTO receipt_predicates(receipt_id,kind,key,value) VALUES(?,?,?,?)`,
			r.ID, p.Kind, strings.ToLower(strings.TrimSpace(p.Key)), p.Value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetReceipt loads one receipt and its predicates.
func (s *Store) GetReceipt(id string) (Receipt, bool, error) {
	var r Receipt
	var evJSON, stJSON, created, updated string
	err := s.DB.QueryRow(`SELECT id,question,answer_turn,answer_text,snapshot_seq,snapshot_root,as_of_seq,evidence_json,unresolved_scope,parent_id,status,stale_reason,stale_turns_json,created,updated
		FROM receipts WHERE id=?`, id).
		Scan(&r.ID, &r.Question, &r.AnswerTurn, &r.AnswerText, &r.SnapshotSeq, &r.SnapshotRoot, &r.AsOfSeq,
			&evJSON, &r.UnresolvedScope, &r.ParentID, &r.Status, &r.StaleReason, &stJSON, &created, &updated)
	if err == sql.ErrNoRows {
		return Receipt{}, false, nil
	}
	if err != nil {
		return Receipt{}, false, err
	}
	_ = json.Unmarshal([]byte(evJSON), &r.Evidence)
	_ = json.Unmarshal([]byte(stJSON), &r.StaleTurns)
	r.Created, _ = time.Parse(time.RFC3339, created)
	r.Updated, _ = time.Parse(time.RFC3339, updated)
	preds, err := s.receiptPredicates(r.ID)
	if err != nil {
		return Receipt{}, false, err
	}
	r.Predicates = preds
	return r, true, nil
}

func (s *Store) receiptPredicates(id string) ([]ReceiptPredicate, error) {
	rows, err := s.DB.Query(`SELECT kind, key, value FROM receipt_predicates WHERE receipt_id=? ORDER BY kind, key`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReceiptPredicate
	for rows.Next() {
		var p ReceiptPredicate
		if err := rows.Scan(&p.Kind, &p.Key, &p.Value); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListReceipts returns receipts filtered by status ("" for all), newest first.
func (s *Store) ListReceipts(status string) ([]Receipt, error) {
	q := `SELECT id FROM receipts`
	var args []any
	if status != "" {
		q += ` WHERE status=?`
		args = append(args, status)
	}
	q += ` ORDER BY updated DESC, id DESC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]Receipt, 0, len(ids))
	for _, id := range ids {
		r, ok, err := s.GetReceipt(id)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, r)
		}
	}
	return out, nil
}

// MatchingReceiptIDs finds live receipts (as_of_seq == 0, status != 'superseded')
// whose snapshot_seq < newSeq and whose predicates match any (kind, key) pair.
func (s *Store) MatchingReceiptIDs(newSeq int, pairs [][2]string) ([]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	var clauses []string
	args := []any{newSeq}
	for _, p := range pairs {
		clauses = append(clauses, `(rp.kind=? AND rp.key=?)`)
		args = append(args, p[0], strings.ToLower(strings.TrimSpace(p[1])))
	}
	q := fmt.Sprintf(`SELECT DISTINCT r.id FROM receipts r
		JOIN receipt_predicates rp ON rp.receipt_id = r.id
		WHERE r.as_of_seq = 0 AND r.status != 'superseded' AND r.snapshot_seq < ?
		  AND (%s)
		ORDER BY r.id`, strings.Join(clauses, " OR "))
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
