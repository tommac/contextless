package engine

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/tommac/contextless/internal/creativity"
	"github.com/tommac/contextless/internal/store"
)

// AnyOf lists independent sources for one choice. Separate choices are evaluated independently.
type CreativeChoice struct {
	Choice string   `json:"choice"`
	AnyOf  []string `json:"any_of"`
}
type CreativeDesign struct {
	Prototype  *creativity.Evaluation `json:"prototype,omitempty"`
	ID         string                 `json:"id"`
	Parent     int                    `json:"parent"`
	Title      string                 `json:"title"`
	Prediction string                 `json:"prediction"`
	Outcome    string                 `json:"outcome"`
	Choices    []CreativeChoice       `json:"choices"`
}
type designVersion struct {
	CreativeDesign
	Revision int                     `json:"revision"`
	Question string                  `json:"question"`
	Cards    map[string]CreativeCard `json:"sources"`
}
type CreativeImpact struct {
	ID     string         `json:"design"`
	Choice string         `json:"choice"`
	Status string         `json:"status"`
	Old    []CreativeCard `json:"old"`
	New    []CreativeCard `json:"new"`
}

var creativeID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,79}$`)

// The immutable source store and revision log live in the board vault. The compact
// reverse index references only the current revision of each design.
const creativeTables = `
CREATE TABLE IF NOT EXISTS creative_bodies (sha TEXT PRIMARY KEY, body BLOB NOT NULL);
CREATE TABLE IF NOT EXISTS creative_revisions (
    board TEXT NOT NULL, design TEXT NOT NULL, revision INTEGER NOT NULL,
    content TEXT NOT NULL, PRIMARY KEY(board,design,revision));
CREATE TABLE IF NOT EXISTS creative_refs (
    board TEXT NOT NULL, edge TEXT NOT NULL, design TEXT NOT NULL, PRIMARY KEY(board,edge,design));
CREATE INDEX IF NOT EXISTS creative_refs_design ON creative_refs(board,design);`

func ensureCreative(s *store.Store) error { _, err := s.DB.Exec(creativeTables); return err }

func indexCreativeCards(snap CreativeCanvas) map[string]CreativeCard {
	cards := make(map[string]CreativeCard, len(snap.Cards))
	for _, card := range snap.Cards {
		cards[card.Key] = card
	}
	return cards
}

// SaveCreativeDesign atomically binds declared choices to the exact bytes in this Canvas read.
// Parent is an optimistic revision check. Repeating a successful import returns the existing revision.
func SaveCreativeDesign(s *store.Store, board string, snap CreativeCanvas, spec CreativeDesign) (int, error) {
	if board == "" || !creativeID.MatchString(spec.ID) || spec.Parent < 0 || strings.TrimSpace(spec.Title) == "" ||
		strings.TrimSpace(spec.Prediction) == "" || len(spec.Choices) == 0 {
		return 0, errors.New("a design requires an ID, title, prediction, choices and a nonnegative parent revision")
	}
	if spec.Prototype != nil {
		measured, err := creativity.Evaluate(*spec.Prototype)
		if err != nil {
			return 0, fmt.Errorf("design prototype: %w", err)
		}
		encoded, err := json.Marshal(measured)
		if err != nil {
			return 0, err
		}
		spec.Outcome = string(encoded)
	}
	cards := indexCreativeCards(snap)
	v := designVersion{CreativeDesign: spec, Revision: spec.Parent + 1, Question: snap.Question, Cards: map[string]CreativeCard{}}
	seenChoices := map[string]bool{}
	for _, choice := range spec.Choices {
		if strings.TrimSpace(choice.Choice) == "" || seenChoices[choice.Choice] || len(choice.AnyOf) == 0 {
			return 0, errors.New("each choice needs a unique description and at least one source edge")
		}
		seenChoices[choice.Choice] = true
		seen := map[string]bool{}
		role := ""
		for _, key := range choice.AnyOf {
			card, ok := cards[key]
			if !ok || seen[key] {
				return 0, fmt.Errorf("missing or repeated Canvas edge %q", key)
			}
			if role != "" && role != card.Role {
				return 0, errors.New("independent supports for one choice must have the same role")
			}
			if card.Excerpt == "" || card.Start < 0 || card.End < card.Start || card.End > len(card.Body) || card.Body[card.Start:card.End] != card.Excerpt || fmt.Sprintf("%x", sha256.Sum256([]byte(card.Body))) != card.BodySHA {
				return 0, fmt.Errorf("invalid exact source for edge %q", key)
			}
			role = card.Role
			seen[key] = true
			v.Cards[key] = card
		}
	}
	content, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	if err = ensureCreative(s); err != nil {
		return 0, err
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var latest int
	var prior string
	err = tx.QueryRow("SELECT revision,content FROM creative_revisions WHERE board=? AND design=? ORDER BY revision DESC LIMIT 1", board, spec.ID).Scan(&latest, &prior)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if latest == v.Revision && prior == string(content) {
		return latest, nil
	}
	if latest > v.Revision {
		err := tx.QueryRow("SELECT content FROM creative_revisions WHERE board=? AND design=? AND revision=?", board, spec.ID, v.Revision).Scan(&prior)
		if err != nil {
			return 0, err
		}
		if prior == string(content) {
			return v.Revision, nil
		}
	}
	if latest != spec.Parent {
		return 0, fmt.Errorf("design %s at revision %d, input parent %d", spec.ID, latest, spec.Parent)
	}
	if _, err = tx.Exec("INSERT INTO creative_revisions VALUES(?,?,?,?)", board, spec.ID, v.Revision, string(content)); err != nil {
		return 0, err
	}
	if _, err = tx.Exec("DELETE FROM creative_refs WHERE board=? AND design=?", board, spec.ID); err != nil {
		return 0, err
	}
	for key, card := range v.Cards {
		if _, err = tx.Exec("INSERT OR IGNORE INTO creative_bodies VALUES(?,?)", card.BodySHA, []byte(card.Body)); err != nil {
			return 0, err
		}
		if _, err = tx.Exec("INSERT INTO creative_refs VALUES(?,?,?)", board, key, spec.ID); err != nil {
			return 0, err
		}
	}
	return v.Revision, tx.Commit()
}

func CreativeOriginal(s *store.Store, sha string) (string, error) {
	var body string
	err := s.DB.QueryRow("SELECT body FROM creative_bodies WHERE sha=?", sha).Scan(&body)
	if err == nil && fmt.Sprintf("%x", sha256.Sum256([]byte(body))) != sha {
		return "", errors.New("historical source hash mismatch")
	}
	return body, err
}

// ReviewCreativeDesigns compares only the cited passage, its identity and its role.
// A change elsewhere in a long note retains the earlier exact version but needs no review.
func ReviewCreativeDesigns(s *store.Store, board string, snap CreativeCanvas, changedKeys []string) ([]CreativeImpact, error) {
	if err := ensureCreative(s); err != nil {
		return nil, err
	}
	args := []any{board}
	filter := ""
	if len(changedKeys) > 0 {
		filter = " AND edge IN (" + strings.TrimRight(strings.Repeat("?,", len(changedKeys)), ",") + ")"
		for _, key := range changedKeys {
			args = append(args, key)
		}
	}
	query := `SELECT v.content FROM (SELECT DISTINCT design FROM creative_refs WHERE board=?` + filter + `) r
        JOIN creative_revisions v ON v.board=? AND v.design=r.design AND v.revision=(
        SELECT MAX(revision) FROM creative_revisions WHERE board=v.board AND design=v.design) ORDER BY r.design`
	args = append(args, board)
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	current := indexCreativeCards(snap)
	var out []CreativeImpact
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var v designVersion
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			return nil, err
		}
		for _, choice := range v.Choices {
			if len(changedKeys) > 0 && !slices.ContainsFunc(choice.AnyOf, func(key string) bool { return slices.Contains(changedKeys, key) }) {
				continue
			}
			impact := CreativeImpact{ID: v.ID, Choice: choice.Choice, Status: "review"}
			unchanged := false
			for _, key := range choice.AnyOf {
				old := v.Cards[key]
				now, ok := current[key]
				if ok && old.Excerpt == now.Excerpt && old.Role == now.Role && old.Relation == now.Relation && old.Source == now.Source && old.Heading == now.Heading && old.Kind == now.Kind {
					unchanged = true
					continue
				}
				impact.Old = append(impact.Old, old)
				if ok {
					impact.New = append(impact.New, now)
				}
			}
			if len(impact.Old) == 0 {
				continue
			}
			if unchanged {
				impact.Status = "supported"
			}
			if impact.Old[0].Role == "inspiration" {
				impact.Status = "inspiration"
				for _, card := range impact.New {
					if card.Role == "constraint" {
						impact.Status = "review"
						break
					}
				}
			}
			out = append(out, impact)
		}
	}
	return out, rows.Err()
}

// CreativeBoardPath namespaces receipt edges by the physical board relative to its vault.
func CreativeBoardPath(board, vaultRoot string) (root, key string, err error) {
	if vaultRoot == "" {
		vaultRoot = filepath.Dir(board)
	}
	real := func(p string) (string, error) {
		abs, e := filepath.Abs(p)
		if e != nil {
			return "", e
		}
		return filepath.EvalSymlinks(abs)
	}
	root, err = real(vaultRoot)
	if err != nil {
		return
	}
	path, err := real(board)
	if err != nil {
		return "", "", err
	}
	key, err = filepath.Rel(root, path)
	if err != nil || key == ".." || strings.HasPrefix(key, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("board must be inside the selected vault")
	}
	return root, filepath.ToSlash(key), nil
}

// Each version is a new Markdown file. User annotations and properties on prior versions stay editable.
func ExportCreativeDesign(s *store.Store, root, board, id string, revision int) error {
	var content string
	if err := s.DB.QueryRow("SELECT content FROM creative_revisions WHERE board=? AND design=? AND revision=?", board, id, revision).Scan(&content); err != nil {
		return err
	}
	var v designVersion
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(board)))[:16]
	dir := filepath.Join(root, "Creative designs", hash)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// JSON strings are valid YAML scalars, including titles with colons and quotes.
	quoted := func(x string) string { b, _ := json.Marshal(x); return string(b) }
	var note strings.Builder
	fmt.Fprintf(&note, "---\ntype: cx-creative-design\nboard: %s\ndesign: %s\nrevision: %d\nprediction: %s\noutcome: %s\n---\n\n# %s\n\n", quoted(board), quoted(id), revision, quoted(v.Prediction), quoted(v.Outcome), v.Title)
	if revision > 1 {
		fmt.Fprintf(&note, "Previous version: [[%s-v%d]]\n\n", id, revision-1)
	}
	for _, choice := range v.Choices {
		fmt.Fprintf(&note, "## %s\n\n", choice.Choice)
		for _, key := range choice.AnyOf {
			card := v.Cards[key]
			fmt.Fprintf(&note, "- %s · %s%s · bytes %d–%d · SHA-256 `%s`\n\n> %s\n\n", card.Role, card.Source, card.Heading, card.Start, card.End, card.BodySHA, strings.ReplaceAll(card.Excerpt, "\n", "\n> "))
		}
	}
	if err := createCreativeFile(filepath.Join(dir, fmt.Sprintf("%s-v%d.md", id, revision)), note.String()); err != nil {
		return err
	}
	return createCreativeFile(filepath.Join(root, "Creative designs", "Design comparisons.base"), "filters:\n  and:\n    - 'type == \"cx-creative-design\"'\nviews:\n  - type: table\n    name: Design versions\n    order:\n      - file.name\n      - design\n      - revision\n      - prediction\n      - outcome\n      - board\n")
}

// Link a fully written temporary file into place without overwriting user edits.
func createCreativeFile(path, content string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cx-design-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.WriteString(content); err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Link(tmp.Name(), path); errors.Is(err, os.ErrExist) {
		return nil
	}
	return err
}
