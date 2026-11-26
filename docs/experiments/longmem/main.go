// Command longmem produces ranked evidence addresses from the pinned public LongMemEval-S release.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

const digest = "d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442"
const candidates = 24

type message struct {
	Role    string
	Content string
}

// Gold and original session identifiers are deliberately absent from this decoding type.
type question struct {
	ID       string      `json:"question_id"`
	Query    string      `json:"question"`
	Sessions [][]message `json:"haystack_sessions"`
}

// Session and message positions are zero-based. Start and End are UTF-8 byte offsets.
type address struct {
	ID                           string
	Session, Message, Start, End int
}

type ranking struct {
	QID               string
	BM25, Contextless []address
}

func rank(q question, db *store.Store) (ranking, error) {
	result := ranking{QID: q.ID}
	if err := db.ResetAtoms(); err != nil {
		return result, err
	}
	var turns []vault.Turn
	var batch []store.AtomTurn
	addresses := map[string]address{}
	for session, messages := range q.Sessions {
		for index, msg := range messages {
			seq := len(turns) + 1
			t := vault.Turn{ID: vault.TurnID(seq), Seq: seq, Role: msg.Role, Body: msg.Content, BodySHA: vault.SHA(msg.Content)}
			turns = append(turns, t)
			item := store.AtomTurn{Turn: t.ID, Seq: t.Seq}
			for _, atom := range vault.Atoms(t.ID, t.Body) {
				item.Atoms = append(item.Atoms, store.AtomRec{ID: atom.ID(), Text: atom.Text(t.Body)})
				addresses[atom.ID()] = address{atom.ID(), session, index, atom.Start, atom.End}
			}
			batch = append(batch, item)
		}
	}
	if err := db.IndexAtomTurns(batch); err != nil {
		return result, err
	}
	for _, hit := range db.SearchAtoms(q.Query, candidates, 0) {
		result.BM25 = append(result.BM25, addresses[hit.ID])
	}
	for _, hit := range packet.RankAtoms(turns, db, q.Query, 0, candidates) {
		result.Contextless = append(result.Contextless, addresses[hit.Atom.ID()])
	}
	return result, nil
}

func run(path, only string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return err
	}
	if fmt.Sprintf("%x", h.Sum(nil)) != digest {
		return fmt.Errorf("dataset SHA-256 mismatch")
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "cx-longmem-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	db, err := store.OpenFile(filepath.Join(tmp, "index.db"))
	if err != nil {
		return err
	}
	defer db.Close()
	decoder := json.NewDecoder(f)
	start, err := decoder.Token()
	if err != nil || start != json.Delim('[') {
		return fmt.Errorf("expected an array of questions")
	}
	count := 0
	out := json.NewEncoder(os.Stdout)
	for decoder.More() {
		var q question
		if err := decoder.Decode(&q); err != nil {
			return err
		}
		if only != "" && q.ID != only {
			continue
		}
		result, err := rank(q, db)
		if err != nil {
			return fmt.Errorf("rank %s: %w", q.ID, err)
		}
		if err = out.Encode(result); err != nil {
			return err
		}
		count++
	}
	if only != "" && count != 1 || only == "" && count != 500 {
		return fmt.Errorf("unexpected question count %d", count)
	}
	return nil
}

func main() {
	path := flag.String("data", "", "pinned longmemeval_s_cleaned.json")
	only := flag.String("id", "", "rank just one public question")
	flag.Parse()
	if err := run(*path, *only); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
