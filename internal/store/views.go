package store

import (
	"database/sql"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Incremental bitemporal key-value materialized views and rare-key co-occurrence
// hubs over the append-only conversation archive (DBSP / Gupta-Mumick style
// monotone view maintenance + EdgeMem/SPRIG rare-band bipartite PPR bridges).
//
// Zero LLM calls and zero embeddings: every row is a deterministic pure
// function of the immutable turn sequence, so Reindex and incremental append
// produce bit-for-bit identical views.

// KeyAssertion is one deterministic key=value assignment extracted from a turn atom.
type KeyAssertion struct {
	Key      string `json:"key"`
	TurnSeq  int    `json:"turn_seq"`
	AtomK    int    `json:"atom_k"`
	RawValue string `json:"raw_value"`
	Quote    string `json:"quote"`
}

// KeyStateRow is one bitemporal validity-interval row in key_history.
type KeyStateRow struct {
	Key             string `json:"key"`
	TurnSeq         int    `json:"turn_seq"`
	AtomK           int    `json:"atom_k"`
	RawValue        string `json:"raw_value"`
	ValidFromSeq    int    `json:"valid_from_seq"`
	ValidUntilSeq   int    `json:"valid_until_seq"`   // 0 = currently active
	SupersededBySeq int    `json:"superseded_by_seq"` // 0 = not superseded
	Quote           string `json:"quote,omitempty"`
}

// RareKeyRow links an identifier or rare entity token to a turn atom.
type RareKeyRow struct {
	Token   string
	TurnSeq int
	AtomK   int
}

// BridgeHit is a 2-hop unseen turn discovered via bipartite Personalized
// PageRank over rare co-occurring keys (2 <= df <= 12).
type BridgeHit struct {
	TurnSeq int      `json:"turn_seq"`
	Score   float64  `json:"score"`
	ViaKeys []string `json:"via_keys"`
}

var (
	// Matches `key = value`, `--key=value`, or config-like `key: value` tokens.
	kvEqRe    = regexp.MustCompile(`(?:^|[\s({\["'` + "`" + `,;-])([A-Za-z][A-Za-z0-9_.-]{1,42})\s*=\s*([^\s,;)\]}\"'` + "`" + `]{1,72})`)
	kvColonRe = regexp.MustCompile(`(?m)^\s*[-*]?\s*([A-Za-z][A-Za-z0-9_.-]{1,42})\s*:\s*([A-Za-z0-9_./:@+-]{1,72})\s*$`)
	identRe   = regexp.MustCompile(`[A-Za-z0-9_./:-]{3,64}`)
)

var kvKeyStop = map[string]bool{
	"http": true, "https": true, "mailto": true, "file": true, "note": true,
	"example": true, "summary": true, "question": true, "answer": true,
	"status": true, "type": true, "role": true, "the": true, "and": true,
	"for": true, "with": true, "from": true, "this": true, "that": true,
	"true": true, "false": true, "todo": true, "fixme": true,
}

func cleanKVKey(k string) (string, bool) {
	k = strings.TrimLeft(k, "-")
	k = strings.ToLower(strings.TrimSpace(k))
	if len(k) < 2 || len(k) > 42 || kvKeyStop[k] {
		return "", false
	}
	for _, r := range k {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' || r == '-') {
			return "", false
		}
	}
	return k, true
}

func cleanKVVal(v string) (string, bool) {
	v = strings.TrimSpace(strings.TrimRight(v, ".,;:!?)"))
	v = strings.Trim(v, `"'`+"`")
	if v == "" || len(v) > 72 {
		return "", false
	}
	// Reject comparison operators (`a == b` matched as `= b` or empty).
	if strings.HasPrefix(v, "=") {
		return "", false
	}
	return v, true
}

// ExtractKeyAssertions deterministically extracts structured key/value
// assertions (`pool_size=20`, `--port=5432`, YAML/bullet `timeout: 45s`) from
// one atom's verbatim text without any model call.
func ExtractKeyAssertions(turnSeq, atomK int, text string) []KeyAssertion {
	var out []KeyAssertion
	seen := map[string]int{} // key -> index in out (last assertion in atom wins)
	lines := strings.Split(text, "\n")
	for _, ln := range lines {
		one := strings.TrimSpace(ln)
		// Skip empty lines and interrogative questions ("What if pool_size=999?").
		if one == "" || strings.HasSuffix(one, "?") {
			continue
		}
		if len(one) > 160 {
			one = one[:160]
		}
		// Skip equality comparisons (`==`, `!=`, `<=`, `>=`).
		sanitized := strings.ReplaceAll(ln, "==", "  ")
		sanitized = strings.ReplaceAll(sanitized, "!=", "  ")
		sanitized = strings.ReplaceAll(sanitized, "<=", "  ")
		sanitized = strings.ReplaceAll(sanitized, ">=", "  ")
		for _, m := range kvEqRe.FindAllStringSubmatch(sanitized, -1) {
			k, okK := cleanKVKey(m[1])
			v, okV := cleanKVVal(m[2])
			if !okK || !okV {
				continue
			}
			row := KeyAssertion{Key: k, TurnSeq: turnSeq, AtomK: atomK, RawValue: v, Quote: one}
			if idx, exists := seen[k]; exists {
				out[idx] = row
			} else {
				seen[k] = len(out)
				out = append(out, row)
			}
		}
	}
	for _, m := range kvColonRe.FindAllStringSubmatch(text, -1) {
		k, okK := cleanKVKey(m[1])
		v, okV := cleanKVVal(m[2])
		if !okK || !okV {
			continue
		}
		row := KeyAssertion{Key: k, TurnSeq: turnSeq, AtomK: atomK, RawValue: v, Quote: strings.TrimSpace(m[0])}
		if idx, exists := seen[k]; exists {
			out[idx] = row
		} else {
			seen[k] = len(out)
			out = append(out, row)
		}
	}
	return out
}

// ExtractRareTokens extracts technical identifiers, ticket codes, paths,
// version strings, and mid-sentence proper nouns for rare-key co-occurrence
// bridge discovery (EdgeMem / SPRIG).
func ExtractRareTokens(turnSeq, atomK int, text string) []RareKeyRow {
	seen := map[string]bool{}
	var out []RareKeyRow
	for _, raw := range identRe.FindAllString(text, -1) {
		tok := strings.Trim(raw, ".,:;-_/")
		if len(tok) < 3 || len(tok) > 48 {
			continue
		}
		hasDigit, hasUpper, hasLower, hasSep := false, false, false, false
		for i, r := range tok {
			switch {
			case unicode.IsDigit(r):
				hasDigit = true
			case unicode.IsUpper(r):
				if i > 0 {
					hasUpper = true
				}
			case unicode.IsLower(r):
				hasLower = true
			case r == '_' || r == '/' || r == '-':
				hasSep = true
			}
		}
		// Keep structured identifiers (digits, snake/path/kebab, camelCase).
		if !(hasDigit || hasSep || (hasUpper && hasLower)) {
			continue
		}
		norm := strings.ToLower(tok)
		if kvKeyStop[norm] || seen[norm] {
			continue
		}
		seen[norm] = true
		out = append(out, RareKeyRow{Token: norm, TurnSeq: turnSeq, AtomK: atomK})
	}
	return out
}

// ApplyTurnViews incrementally updates key_history and rare_keys in one transaction.
func (s *Store) ApplyTurnViews(assertions []KeyAssertion, rare []RareKeyRow) error {
	if (len(assertions) == 0 && len(rare) == 0) || s.ReadOnly {
		return nil
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := applyTurnViewsTx(tx, assertions, rare); err != nil {
		return err
	}
	return tx.Commit()
}

// applyTurnViewsTx incrementally updates key_history and rare_keys inside tx.
// Assumes turns are indexed in non-decreasing turn_seq order (true for both
// live append and Reindex).
func applyTurnViewsTx(tx *sql.Tx, assertions []KeyAssertion, rare []RareKeyRow) error {
	if len(assertions) > 0 {
		seenTurn := map[int]bool{}
		for _, a := range assertions {
			if !seenTurn[a.TurnSeq] {
				seenTurn[a.TurnSeq] = true
				if _, err := tx.Exec(`DELETE FROM key_history WHERE turn_seq=?`, a.TurnSeq); err != nil {
					return err
				}
			}
		}
		for _, a := range assertions {
			// Check currently active assertion for this key before a.TurnSeq.
			var prevSeq, prevAtom int
			var prevVal string
			err := tx.QueryRow(`SELECT turn_seq, atom_k, raw_value FROM key_history
				WHERE key=? AND valid_until_seq=0 AND turn_seq < ?
				ORDER BY turn_seq DESC, atom_k DESC LIMIT 1`, a.Key, a.TurnSeq).
				Scan(&prevSeq, &prevAtom, &prevVal)
			if err == nil {
				if strings.EqualFold(prevVal, a.RawValue) {
					// Snodgrass value-equivalent interval coalescing: re-confirming the
					// already-active value keeps its original [valid_from_seq..0] open
					// rather than fragmenting into artificial self-supersessions.
					continue
				}
				_, _ = tx.Exec(`UPDATE key_history SET valid_until_seq=?, superseded_by_seq=?
					WHERE key=? AND valid_until_seq=0 AND turn_seq < ?`,
					a.TurnSeq-1, a.TurnSeq, a.Key, a.TurnSeq)
			}
			if _, err := tx.Exec(`INSERT OR REPLACE INTO key_history
				(key, turn_seq, atom_k, raw_value, valid_from_seq, valid_until_seq, superseded_by_seq, quote)
				VALUES(?,?,?,?,?,0,0,?)`,
				a.Key, a.TurnSeq, a.AtomK, a.RawValue, a.TurnSeq, a.Quote); err != nil {
				return err
			}
		}
	}
	if len(rare) > 0 {
		seenTurn := map[int]bool{}
		for _, r := range rare {
			if !seenTurn[r.TurnSeq] {
				seenTurn[r.TurnSeq] = true
				if _, err := tx.Exec(`DELETE FROM rare_keys WHERE turn_seq=?`, r.TurnSeq); err != nil {
					return err
				}
			}
		}
		stmt, err := tx.Prepare(`INSERT OR IGNORE INTO rare_keys(token, turn_seq, atom_k) VALUES(?,?,?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, r := range rare {
			if _, err := stmt.Exec(r.Token, r.TurnSeq, r.AtomK); err != nil {
				return err
			}
		}
	}
	return nil
}

// QueryKeyHistory returns the full chronological evolution of key (or its
// point-in-time chain up to asOfSeq when asOfSeq > 0). When asOfSeq > 0, a row
// superseded AFTER asOfSeq is reconstructed with ValidUntilSeq=0 (active at asOfSeq)
// so stale/future exposure is strictly 0%.
func (s *Store) QueryKeyHistory(key string, asOfSeq int) ([]KeyStateRow, error) {
	k, ok := cleanKVKey(key)
	if !ok {
		k = strings.ToLower(strings.TrimSpace(key))
	}
	q := `SELECT key, turn_seq, atom_k, raw_value, valid_from_seq, valid_until_seq, superseded_by_seq, quote
		FROM key_history WHERE key=?`
	args := []any{k}
	if asOfSeq > 0 {
		q += ` AND turn_seq <= ?`
		args = append(args, asOfSeq)
	}
	q += ` ORDER BY turn_seq ASC, atom_k ASC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KeyStateRow
	for rows.Next() {
		var r KeyStateRow
		if err := rows.Scan(&r.Key, &r.TurnSeq, &r.AtomK, &r.RawValue, &r.ValidFromSeq, &r.ValidUntilSeq, &r.SupersededBySeq, &r.Quote); err != nil {
			return nil, err
		}
		if asOfSeq > 0 && r.SupersededBySeq > asOfSeq {
			r.ValidUntilSeq = 0
			r.SupersededBySeq = 0
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListActiveKeys returns the currently active (or active as-of asOfSeq)
// materialized key=value state across the archive, optionally filtered by
// substring. Achieves O(1) lookup per key with 0% stale value exposure.
func (s *Store) ListActiveKeys(filter string, asOfSeq, limit int) ([]KeyStateRow, error) {
	if limit <= 0 {
		limit = 100
	}
	filter = strings.ToLower(strings.TrimSpace(filter))
	var rows *sql.Rows
	var err error
	if asOfSeq > 0 {
		q := `SELECT key, turn_seq, atom_k, raw_value, valid_from_seq, valid_until_seq, superseded_by_seq, quote
			FROM key_history
			WHERE turn_seq <= ? AND (valid_until_seq = 0 OR valid_until_seq >= ?)`
		args := []any{asOfSeq, asOfSeq}
		if filter != "" {
			q += ` AND (key LIKE ? OR LOWER(raw_value) LIKE ?)`
			args = append(args, "%"+filter+"%", "%"+filter+"%")
		}
		q += ` ORDER BY key ASC, turn_seq DESC LIMIT ?`
		args = append(args, limit)
		rows, err = s.DB.Query(q, args...)
	} else {
		q := `SELECT key, turn_seq, atom_k, raw_value, valid_from_seq, valid_until_seq, superseded_by_seq, quote
			FROM key_history WHERE valid_until_seq = 0`
		var args []any
		if filter != "" {
			q += ` AND (key LIKE ? OR LOWER(raw_value) LIKE ?)`
			args = append(args, "%"+filter+"%", "%"+filter+"%")
		}
		q += ` ORDER BY key ASC, turn_seq DESC LIMIT ?`
		args = append(args, limit)
		rows, err = s.DB.Query(q, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KeyStateRow
	seenKey := map[string]bool{}
	for rows.Next() {
		var r KeyStateRow
		if err := rows.Scan(&r.Key, &r.TurnSeq, &r.AtomK, &r.RawValue, &r.ValidFromSeq, &r.ValidUntilSeq, &r.SupersededBySeq, &r.Quote); err != nil {
			return nil, err
		}
		if asOfSeq > 0 && r.SupersededBySeq > asOfSeq {
			r.ValidUntilSeq = 0
			r.SupersededBySeq = 0
		}
		if seenKey[r.Key] {
			continue
		}
		seenKey[r.Key] = true
		out = append(out, r)
	}
	return out, rows.Err()
}

// BridgeRareKeysPPR runs 5-iteration Personalized PageRank (alpha=0.15) seeded
// from seedTurns over the bipartite (turn <-> rare_token) graph restricted to
// rare document frequency 2 <= df <= 12 (EdgeMem + SPRIG GraphHybrid).
// Returns previously unseen 2-hop bridge turns ranked by stationary mass.
func (s *Store) BridgeRareKeysPPR(seedTurns []int, beforeSeq, maxBridges int) ([]BridgeHit, error) {
	if len(seedTurns) == 0 || maxBridges <= 0 {
		return nil, nil
	}
	seedSet := map[int]bool{}
	var uniqSeeds []int
	for _, seq := range seedTurns {
		if seq > 0 && (beforeSeq <= 0 || seq < beforeSeq) && !seedSet[seq] {
			seedSet[seq] = true
			uniqSeeds = append(uniqSeeds, seq)
		}
	}
	if len(uniqSeeds) == 0 {
		return nil, nil
	}

	// Step 1: fetch rare tokens appearing in the seed turns whose corpus df lies in [2, 12].
	ph := strings.TrimSuffix(strings.Repeat("?,", len(uniqSeeds)), ",")
	args := make([]any, 0, len(uniqSeeds)+2)
	for _, seq := range uniqSeeds {
		args = append(args, seq)
	}
	boundClause := ""
	if beforeSeq > 0 {
		boundClause = ` AND r2.turn_seq < ?`
		args = append(args, beforeSeq)
	}
	q := `SELECT r1.token, r2.turn_seq
		FROM (SELECT DISTINCT token FROM rare_keys WHERE turn_seq IN (` + ph + `)) r1
		JOIN rare_keys r2 ON r2.token = r1.token` + boundClause
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokToTurns := map[string]map[int]bool{}
	for rows.Next() {
		var tok string
		var seq int
		if err := rows.Scan(&tok, &seq); err != nil {
			return nil, err
		}
		if tokToTurns[tok] == nil {
			tokToTurns[tok] = map[int]bool{}
		}
		tokToTurns[tok][seq] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Prune to rare band 2 <= df <= 12 so high-frequency boilerplate never leaks noise.
	turnToToks := map[int][]string{}
	validTokTurns := map[string][]int{}
	for tok, tset := range tokToTurns {
		df := len(tset)
		if df < 2 || df > 12 {
			continue
		}
		ts := make([]int, 0, df)
		for seq := range tset {
			ts = append(ts, seq)
			turnToToks[seq] = append(turnToToks[seq], tok)
		}
		sort.Ints(ts)
		validTokTurns[tok] = ts
	}
	if len(validTokTurns) == 0 {
		return nil, nil
	}
	for seq := range turnToToks {
		sort.Strings(turnToToks[seq])
	}

	// Step 2: 5 iterations of Personalized PageRank with restart alpha = 0.15.
	const alpha = 0.15
	const iters = 5
	pTurn := map[int]float64{}
	restart := 1.0 / float64(len(uniqSeeds))
	for _, seq := range uniqSeeds {
		pTurn[seq] = restart
	}
	for step := 0; step < iters; step++ {
		pTok := map[string]float64{}
		for seq, mass := range pTurn {
			toks := turnToToks[seq]
			if len(toks) == 0 {
				continue
			}
			share := mass / float64(len(toks))
			for _, tok := range toks {
				pTok[tok] += share
			}
		}
		nextTurn := map[int]float64{}
		for _, seq := range uniqSeeds {
			nextTurn[seq] += alpha * restart
		}
		for tok, mass := range pTok {
			ts := validTokTurns[tok]
			// IDF-tilt rare tokens slightly (1 / df): an exact bipartite random walk
			share := (1.0 - alpha) * mass / float64(len(ts))
			for _, seq := range ts {
				nextTurn[seq] += share
			}
		}
		pTurn = nextTurn
	}

	var hits []BridgeHit
	for seq, sc := range pTurn {
		if seedSet[seq] || sc <= 0 {
			continue
		}
		// Collect shared rare keys linking this bridge turn back to any seed turn.
		var via []string
		for _, tok := range turnToToks[seq] {
			linkedSeed := false
			for _, sseq := range validTokTurns[tok] {
				if seedSet[sseq] {
					linkedSeed = true
					break
				}
			}
			if linkedSeed {
				via = append(via, tok)
			}
		}
		hits = append(hits, BridgeHit{
			TurnSeq: seq,
			Score:   math.Round(sc*1e6) / 1e6,
			ViaKeys: via,
		})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].TurnSeq > hits[j].TurnSeq
	})
	if len(hits) > maxBridges {
		hits = hits[:maxBridges]
	}
	return hits, nil
}
