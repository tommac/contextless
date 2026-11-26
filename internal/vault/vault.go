// Package vault stores one conversation as Markdown records with YAML
// frontmatter in a git repository. It is BookWriter's vault idea applied to
// chat: turns are the only verbatim bodies and are never rewritten; cards,
// threads, entities, facts and decisions are derived records that point at
// turns by id.
//
// Layers:
//
//	L0 session.md        purpose, standing instructions
//	L1 threads/          topics and workstreams
//	L2 entities/ facts/ decisions/
//	L3 cards/            one catalog card per turn, pinned to the turn's sha
//	L5 turns/            verbatim turns, append-only
//
// Edges are stored on one side only (card.threads, card.mentions,
// fact.source, decision.source); the store derives the inverses.
package vault

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Record types.
const (
	TypeSession  = "session"
	TypeTurn     = "turn"
	TypeCard     = "card"
	TypeThread   = "thread"
	TypeEntity   = "entity"
	TypeFact     = "fact"
	TypeDecision = "decision"
)

// Turn roles.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Session is L0.
type Session struct {
	Title        string    `yaml:"title"`
	Purpose      string    `yaml:"purpose,omitempty"`
	Created      time.Time `yaml:"created"`
	Instructions []string  `yaml:"instructions,omitempty"`
	Repos        []Repo    `yaml:"repos,omitempty"`
	Body         string    `yaml:"-"`
}

// Repo is a git repository attached to the workspace. Coder units change
// it in their own worktree and branch; the user's checkout is never
// touched.
type Repo struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	// Verify is a shell command run in the unit's worktree after the change
	// (for example "go test ./..."); empty skips verification.
	Verify string `yaml:"verify,omitempty"`
	// Allow lists the paths a unit may change (globs, "dir/**"); empty
	// allows any path except .git.
	Allow []string `yaml:"allow,omitempty"`
	// Base is the branch or commit units start from; empty is HEAD.
	Base string `yaml:"base,omitempty"`
}

// FindRepo returns the attached repository with that name.
func (s Session) FindRepo(name string) (Repo, bool) {
	for _, r := range s.Repos {
		if r.Name == name {
			return r, true
		}
	}
	return Repo{}, false
}

// Turn is one verbatim message. Nothing in it changes after it is written.
type Turn struct {
	ID      string    `yaml:"id"`
	Seq     int       `yaml:"seq"`
	Role    string    `yaml:"role"`
	Author  string    `yaml:"author,omitempty"`
	Run     string    `yaml:"run,omitempty"`
	ReplyTo string    `yaml:"reply_to,omitempty"`
	Created time.Time `yaml:"created"`
	BodySHA string    `yaml:"body_sha"`
	Body    string    `yaml:"-"`
}

// Card is the L3 catalog card for one turn. It is an index entry, never a
// stand-in for the turn: a selected turn is always sent verbatim. A card
// whose BodySHA differs from its turn's is stale and is not used.
type Card struct {
	Turn     string   `yaml:"turn"`
	BodySHA  string   `yaml:"body_sha"`
	Summary  string   `yaml:"summary"`
	Threads  []string `yaml:"threads,omitempty"`
	Mentions []string `yaml:"mentions,omitempty"`
	Keywords []string `yaml:"keywords,omitempty"`
	Run      string   `yaml:"run,omitempty"`
}

// Thread is an L1 topic or workstream.
type Thread struct {
	ID      string    `yaml:"id"`
	Title   string    `yaml:"title"`
	Status  string    `yaml:"status"` // open | resolved
	Goal    string    `yaml:"goal,omitempty"`
	Created time.Time `yaml:"created"`
	Body    string    `yaml:"-"`
}

// Entity is an L2 person, project, file, concept, organization or place.
type Entity struct {
	ID      string   `yaml:"id"`
	Name    string   `yaml:"name"`
	Kind    string   `yaml:"kind,omitempty"`
	Aliases []string `yaml:"aliases,omitempty"`
	Body    string   `yaml:"-"`
}

// Fact is a statement made in the conversation, anchored by a verbatim
// quote from its source turn.
type Fact struct {
	ID         string   `yaml:"id"`
	Statement  string   `yaml:"statement"`
	Source     string   `yaml:"source"`
	Quote      string   `yaml:"quote"`
	Entities   []string `yaml:"entities,omitempty"`
	Status     string   `yaml:"status"` // active | retracted | superseded
	Supersedes []string `yaml:"supersedes,omitempty"`
}

// Decision is something the user decided, anchored by a verbatim quote.
// Active decisions travel in every packet (tier 0).
type Decision struct {
	ID         string   `yaml:"id"`
	Statement  string   `yaml:"statement"`
	Source     string   `yaml:"source"`
	Quote      string   `yaml:"quote"`
	Threads    []string `yaml:"threads,omitempty"`
	Status     string   `yaml:"status"` // active | superseded
	Supersedes []string `yaml:"supersedes,omitempty"`
}

// Vault is one conversation directory.
type Vault struct {
	Dir string
	mu  sync.Mutex
}

// ErrNoVault is returned by Open for a directory without session.md.
var ErrNoVault = errors.New("vault: not a ContextlessAI workspace (no session.md)")

var dirs = []string{"turns", "cards", "threads", "entities", "facts", "decisions", "artifacts", ".cx"}

// Create makes a new vault, initializes git, and commits the base.
func Create(dir, title, purpose string) (*Vault, error) {
	if _, err := os.Stat(filepath.Join(dir, "session.md")); err == nil {
		return nil, fmt.Errorf("vault: %s already exists", dir)
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			return nil, err
		}
	}
	v := &Vault{Dir: dir}
	s := Session{Title: title, Purpose: purpose, Created: time.Now().UTC().Truncate(time.Second)}
	if err := v.PutSession(s); err != nil {
		return nil, err
	}
	gi := ".cx/\nartifacts/**/ws/\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gi), 0o644); err != nil {
		return nil, err
	}
	v.gitInit()
	v.Commit("Create workspace: " + title)
	return v, nil
}

// Open opens an existing vault.
func Open(dir string) (*Vault, error) {
	if _, err := os.Stat(filepath.Join(dir, "session.md")); err != nil {
		return nil, ErrNoVault
	}
	for _, d := range dirs {
		_ = os.MkdirAll(filepath.Join(dir, d), 0o755)
	}
	return &Vault{Dir: dir}, nil
}

// SHA returns the hex sha256 of a body.
func SHA(body string) string {
	h := sha256.Sum256([]byte(body))
	return hex.EncodeToString(h[:])
}

// NewID returns prefix-xxxxxxxx with 8 random base32 characters.
func NewID(prefix string) string {
	const alphabet = "abcdefghijkmnpqrstuvwxyz23456789"
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return prefix + "-" + string(b)
}

// TurnID renders the id of turn seq.
func TurnID(seq int) string { return fmt.Sprintf("t%06d", seq) }

var turnIDRe = regexp.MustCompile(`^t(\d{6,})$`)

// TurnSeq parses a turn id.
func TurnSeq(id string) (int, bool) {
	m := turnIDRe.FindStringSubmatch(id)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	return n, err == nil
}

// --- frontmatter codec ---

func encode(meta any, body string) ([]byte, error) {
	y, err := yaml.Marshal(meta)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	b.Write(y)
	b.WriteString("---\n")
	b.WriteString(body)
	if body != "" && !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.Bytes(), nil
}

// decode splits frontmatter from body. The body is returned exactly as
// stored minus the single trailing newline encode adds.
func decode(data []byte, meta any) (string, error) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return "", errors.New("vault: record has no frontmatter")
	}
	rest := s[4:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		if strings.HasSuffix(rest, "\n---") {
			end = len(rest) - 4
		} else {
			return "", errors.New("vault: unterminated frontmatter")
		}
	}
	if err := yaml.Unmarshal([]byte(rest[:end]), meta); err != nil {
		return "", err
	}
	body := ""
	if end+5 <= len(rest) {
		body = rest[end+5:]
	}
	body = strings.TrimSuffix(body, "\n")
	return body, nil
}

func (v *Vault) write(rel string, meta any, body string) error {
	data, err := encode(meta, body)
	if err != nil {
		return err
	}
	full := filepath.Join(v.Dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	tmp := full + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, full)
}

func (v *Vault) read(rel string, meta any) (string, error) {
	data, err := os.ReadFile(filepath.Join(v.Dir, rel))
	if err != nil {
		return "", err
	}
	return decode(data, meta)
}

func (v *Vault) list(sub string) []string {
	files, _ := v.recordFiles(sub)
	return files
}

func (v *Vault) recordFiles(sub string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(v.Dir, sub))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			out = append(out, filepath.Join(sub, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

// --- session ---

// Session reads session.md.
func (v *Vault) Session() (Session, error) {
	var s Session
	body, err := v.read("session.md", &s)
	s.Body = body
	return s, err
}

// PutSession writes session.md.
func (v *Vault) PutSession(s Session) error { return v.write("session.md", s, s.Body) }

// --- turns ---

// TurnPath is the vault-relative path of a turn.
func TurnPath(id string) string { return filepath.Join("turns", id+".md") }

// CardPath is the vault-relative path of a turn's card.
func CardPath(id string) string { return filepath.Join("cards", id+".md") }

// AppendTurn writes the next turn verbatim. The body is stored exactly as
// given, apart from trailing whitespace.
func (v *Vault) AppendTurn(role, author, run, replyTo, body string) (Turn, error) {
	return v.AppendTurnAt(role, author, run, replyTo, body, time.Now())
}

// AppendTurnAt is AppendTurn with the time the message was written, for a
// transcript imported from elsewhere.
func (v *Vault) AppendTurnAt(role, author, run, replyTo, body string, created time.Time) (Turn, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	last, err := v.lastSeq()
	if err != nil {
		return Turn{}, err
	}
	return v.appendTurn(last+1, role, author, run, replyTo, body, created)
}

// appendTurn writes a single immutable file while the caller holds v.mu.
func (v *Vault) appendTurn(seq int, role, author, run, replyTo, body string, created time.Time) (Turn, error) {
	if seq <= 0 {
		return Turn{}, errors.New("vault: turn sequence exhausted")
	}
	body = strings.TrimRight(body, " \t\r\n")
	t := Turn{
		ID: TurnID(seq), Seq: seq, Role: role, Author: author, Run: run, ReplyTo: replyTo,
		Created: created.UTC().Truncate(time.Second), BodySHA: SHA(body), Body: body,
	}
	full := filepath.Join(v.Dir, TurnPath(t.ID))
	if _, err := os.Lstat(full); err == nil {
		return Turn{}, fmt.Errorf("vault: turn %s already exists", t.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Turn{}, err
	}
	return t, v.write(TurnPath(t.ID), t, body)
}

// lastSeq also handles gaps and turn IDs wider than six digits. Scanning
// unsorted filenames in batches bounds memory even for a large vault.
func (v *Vault) lastSeq() (int, error) {
	dir, err := os.Open(filepath.Join(v.Dir, "turns"))
	if err != nil {
		return 0, err
	}
	defer dir.Close()
	last := 0
	for {
		entries, err := dir.ReadDir(1024)
		for _, entry := range entries {
			name, markdown := strings.CutSuffix(entry.Name(), ".md")
			if markdown && !entry.IsDir() {
				if seq, ok := TurnSeq(name); ok && seq > last {
					last = seq
				}
			}
		}
		if err == io.EOF {
			return last, nil
		}
		if err != nil {
			return 0, err
		}
	}
}

// LastSeq returns the highest turn sequence number on disk without reading
// any turn file.
func (v *Vault) LastSeq() (int, error) {
	return v.lastSeq()
}

// LastTurn reads and verifies only the highest-sequence turn on disk.
// ok is false when the vault has no turns yet.
func (v *Vault) LastTurn() (Turn, bool, error) {
	last, err := v.lastSeq()
	if err != nil {
		return Turn{}, false, err
	}
	if last == 0 {
		return Turn{}, false, nil
	}
	t, err := v.Turn(TurnID(last))
	if err != nil {
		return Turn{}, false, err
	}
	return t, true, nil
}

// TurnsByIDs reads and verifies only the requested turns.
func (v *Vault) TurnsByIDs(ids []string) (map[string]Turn, error) {
	out := make(map[string]Turn, len(ids))
	for _, id := range ids {
		if _, ok := out[id]; ok {
			continue
		}
		t, err := v.Turn(id)
		if err != nil {
			return nil, err
		}
		out[id] = t
	}
	return out, nil
}

// ValidTurnID reports whether id has the shape of a turn id. Ids reach Turn
// from models (recall), tool callers (the MCP server) and stored records,
// and are joined into a file path, so anything else is refused: "../../x"
// as an id read any Markdown file with front matter on the disk.
func ValidTurnID(id string) bool { return turnIDRe.MatchString(id) }

// Turn reads one turn and verifies its body sha. An id that is not a turn
// id is reported as a turn that does not exist.
func (v *Vault) Turn(id string) (Turn, error) {
	var t Turn
	if !ValidTurnID(id) {
		return t, fmt.Errorf("vault: %q is not a turn id: %w", id, fs.ErrNotExist)
	}
	body, err := v.read(TurnPath(id), &t)
	if err != nil {
		return t, err
	}
	t.Body = body
	if t.BodySHA != "" && SHA(body) != t.BodySHA {
		return t, fmt.Errorf("vault: turn %s body does not match its sha (edited on disk?)", id)
	}
	return t, nil
}

// Turns reads every turn in sequence order.
func (v *Vault) Turns() ([]Turn, error) {
	files, err := v.recordFiles("turns")
	if err != nil {
		return nil, err
	}
	var out []Turn
	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".md")
		t, err := v.Turn(id)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Seq < out[j].Seq })
	return out, nil
}

// TurnIDs lists the ids of the stored turns in order, from the file
// names alone: no turn file is read.
func (v *Vault) TurnIDs() []string {
	var out []string
	for _, f := range v.list("turns") {
		out = append(out, strings.TrimSuffix(filepath.Base(f), ".md"))
	}
	return out
}

// --- cards ---

// Card reads a turn's card; ok is false when there is none.
func (v *Vault) Card(turnID string) (Card, bool) {
	var c Card
	if _, err := v.read(CardPath(turnID), &c); err != nil {
		return Card{}, false
	}
	return c, true
}

// PutCard writes a card.
func (v *Vault) PutCard(c Card) error { return v.write(CardPath(c.Turn), c, "") }

// Cards reads every card keyed by turn id.
func (v *Vault) Cards() map[string]Card {
	out := map[string]Card{}
	for _, f := range v.list("cards") {
		var c Card
		if _, err := v.read(f, &c); err == nil {
			out[c.Turn] = c
		}
	}
	return out
}

// --- threads, entities, facts, decisions ---

// Threads reads every thread.
func (v *Vault) Threads() []Thread {
	var out []Thread
	for _, f := range v.list("threads") {
		var t Thread
		if body, err := v.read(f, &t); err == nil {
			t.Body = body
			out = append(out, t)
		}
	}
	return out
}

// PutThread writes a thread.
func (v *Vault) PutThread(t Thread) error {
	return v.write(filepath.Join("threads", t.ID+".md"), t, t.Body)
}

// Entities reads every entity.
func (v *Vault) Entities() []Entity {
	var out []Entity
	for _, f := range v.list("entities") {
		var e Entity
		if body, err := v.read(f, &e); err == nil {
			e.Body = body
			out = append(out, e)
		}
	}
	return out
}

// PutEntity writes an entity.
func (v *Vault) PutEntity(e Entity) error {
	return v.write(filepath.Join("entities", e.ID+".md"), e, e.Body)
}

// Facts reads every fact.
func (v *Vault) Facts() []Fact {
	var out []Fact
	for _, f := range v.list("facts") {
		var x Fact
		if _, err := v.read(f, &x); err == nil {
			out = append(out, x)
		}
	}
	return out
}

// PutFact writes a fact.
func (v *Vault) PutFact(f Fact) error { return v.write(filepath.Join("facts", f.ID+".md"), f, "") }

// Decisions reads every decision.
func (v *Vault) Decisions() []Decision {
	var out []Decision
	for _, f := range v.list("decisions") {
		var d Decision
		if _, err := v.read(f, &d); err == nil {
			out = append(out, d)
		}
	}
	return out
}

// PutDecision writes a decision.
func (v *Vault) PutDecision(d Decision) error {
	return v.write(filepath.Join("decisions", d.ID+".md"), d, "")
}

// ArtifactDir returns (and creates) the artifact directory of a run.
func (v *Vault) ArtifactDir(runID string) string {
	d := filepath.Join(v.Dir, "artifacts", runID)
	_ = os.MkdirAll(d, 0o755)
	return d
}

// --- git ---

func (v *Vault) git(args ...string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return err
	}
	cmd := exec.Command("git", append([]string{"-C", v.Dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return nil
}

func (v *Vault) gitInit() {
	if _, err := os.Stat(filepath.Join(v.Dir, ".git")); err == nil {
		return
	}
	_ = v.git("init", "-q")
}

// Commit stages everything and commits; a clean tree or a missing git is
// not an error. The vault's history is git's history.
func (v *Vault) Commit(msg string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.git("add", "-A") != nil {
		return
	}
	// gc.autoDetach=false: an automatic gc runs before Commit returns
	// instead of in a detached process that outlives it
	_ = v.git("-c", "user.name=ContextlessAI", "-c", "user.email=contextless@localhost", "-c", "gc.autoDetach=false",
		"commit", "-q", "--no-verify", "-m", msg)
}

// Lock and Unlock serialize multi-file updates (the indexer's apply).
func (v *Vault) Lock()   { v.mu.Lock() }
func (v *Vault) Unlock() { v.mu.Unlock() }
