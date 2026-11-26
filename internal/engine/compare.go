package engine

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/vault"
)

// cx compare measures plans against single calls on the same messages,
// the measurement that decides whether the plan path earns its cost. Each
// message is answered once per arm, every trial on its own scratch copy of
// the workspace, so the arms see the same history and nothing reaches the
// real conversation. The answers are written blind (letters in an order
// set per message), with a key holding each arm's billed tokens and time,
// for a person to score before unblinding. Unmatched comparisons flatter
// multi-agent systems, so a budget-matched arm (a single call on a profile
// the user names, usually more effort or a larger window) can be added.

// TrialResult is one arm's answer to one message, with what it cost.
type TrialResult struct {
	Arm     string `json:"arm"`
	Answer  string `json:"-"`
	Author  string `json:"author,omitempty"`
	Planned bool   `json:"planned"`
	Units   int    `json:"units"`
	Runs    int    `json:"runs"`
	// Input is the whole prompt side; Uncached, CacheRead and CacheWrite
	// are its parts as billed. USD is what the CLIs reported, for USDRuns
	// of the Runs (0 on subscription CLIs that report no price).
	Input      int     `json:"input_tokens"`
	Uncached   int     `json:"uncached_tokens"`
	CacheRead  int     `json:"cache_read_tokens"`
	CacheWrite int     `json:"cache_write_tokens"`
	Output     int     `json:"output_tokens"`
	USD        float64 `json:"usd_reported"`
	USDRuns    int     `json:"usd_runs"`
	RunMs      int64   `json:"run_ms"`  // summed over the runs
	WallMs     int64   `json:"wall_ms"` // elapsed, start to answer
	Error      string  `json:"error,omitempty"`
}

// Trial appends msg as the next user message and answers it in one arm
// (ArmSingle or ArmPlan, review as configured), without storing an
// answer. Run it only on a scratch copy: the user turn, run directories
// and unit records it writes are the trial's.
func (e *Engine) Trial(ctx context.Context, msg, arm string) TrialResult {
	r := TrialResult{Arm: arm}
	user, err := e.V.AppendTurn(vault.RoleUser, "", "", "", msg)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	start := time.Now()
	o, err := e.respondAs(ctx, user, arm)
	e.indexTurn(user, nil)
	r.WallMs = time.Since(start).Milliseconds()
	r.Answer, r.Author, r.Planned, r.Runs = o.Answer, o.Author, o.Planned, len(o.Runs)
	for _, it := range o.Supplied {
		if strings.HasPrefix(it.ID, "unit-") && it.ID != "unit-comparison" {
			r.Units++
		}
	}
	for _, ri := range o.Runs {
		if row, ok := e.S.RunRow(ri.ID); ok {
			r.Input += row.InputTokens
			r.CacheRead += row.CachedTokens
			r.CacheWrite += row.CacheWriteTokens
			r.Uncached += max(row.InputTokens-row.CachedTokens-row.CacheWriteTokens, 0)
			r.Output += row.OutputTokens
			r.USD += row.USDReported
			if row.HasUSD {
				r.USDRuns++
			}
			r.RunMs += row.WallMs
		}
	}
	if err != nil {
		r.Error = err.Error()
	}
	return r
}

// CopyWorkspace copies workspace from to a new workspace to, for a trial:
// the vault, its git history and its index, without the writer's lock.
func CopyWorkspace(cfg config.Config, from, to string) error {
	src, dst := cfg.WorkspaceDir(from), cfg.WorkspaceDir(to)
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("%s exists", dst)
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		if rel == filepath.Join(".cx", "engine.lock") || !d.Type().IsRegular() {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, in); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	})
}

// ReadMessages reads the messages to compare: JSON lines with a
// "message" field, or plain text with one message per line.
func ReadMessages(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 16<<20)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "{") {
			var m struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal([]byte(line), &m); err != nil || strings.TrimSpace(m.Message) == "" {
				return nil, fmt.Errorf("%s:%d: want {\"message\": \"...\"}", path, n)
			}
			line = m.Message
		}
		out = append(out, line)
	}
	return out, sc.Err()
}

// CompareItem is one message with its arms' answers.
type CompareItem struct {
	Message string        `json:"message"`
	Trials  []TrialResult `json:"trials"`
}

// BlindOrder is the order a message's answers are shown in: set per
// message by a hash, so it is stable across runs but follows no arm.
func BlindOrder(message string, arms []string) []int {
	idx := make([]int, len(arms))
	keys := make([]string, len(arms))
	for i, a := range arms {
		idx[i] = i
		sum := sha256.Sum256([]byte(message + "\x00" + a))
		keys[i] = fmt.Sprintf("%x", sum)
	}
	sort.Slice(idx, func(i, j int) bool { return keys[idx[i]] < keys[idx[j]] })
	return idx
}

// WriteCompare writes the blind answers (answers.md, with a line to score
// each), the key (key.json: which letter is which arm, with tokens, time,
// runs and units) and a cost summary by arm (summary.md, which names no
// letter).
func WriteCompare(dir string, items []CompareItem) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	type keyed struct {
		Message string                 `json:"message"`
		Answers map[string]TrialResult `json:"answers"`
	}
	var ans strings.Builder
	ans.WriteString("# Answers to score\n\nEach message was answered once per arm, on a copy of the conversation. Score each answer before opening key.json.\n")
	var key []keyed
	for i, it := range items {
		var arms []string
		for _, t := range it.Trials {
			arms = append(arms, t.Arm)
		}
		fmt.Fprintf(&ans, "\n## Message %d\n\n%s\n", i+1, it.Message)
		k := keyed{Message: it.Message, Answers: map[string]TrialResult{}}
		for n, j := range BlindOrder(it.Message, arms) {
			letter := string(rune('A' + n))
			t := it.Trials[j]
			body := t.Answer
			if t.Error != "" {
				body = "(no answer: " + t.Error + ")"
			}
			fmt.Fprintf(&ans, "\n### Answer %s\n\n%s\n\nScore (1 to 5):\n", letter, strings.TrimSpace(body))
			k.Answers[letter] = t
		}
		key = append(key, k)
	}
	if err := os.WriteFile(filepath.Join(dir, "answers.md"), []byte(ans.String()), 0o644); err != nil {
		return err
	}
	kb, _ := json.MarshalIndent(key, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "key.json"), append(kb, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(CompareSummary(items)), 0o644)
}

// CompareSummary totals each arm's cost over the messages.
func CompareSummary(items []CompareItem) string {
	type tot struct {
		n, planned, units, runs, failed int
		in, uncached, read, write, out  int
		usd                             float64
		usdRuns                         int
		runMs, wallMs                   int64
	}
	by := map[string]*tot{}
	var order []string
	for _, it := range items {
		for _, t := range it.Trials {
			x := by[t.Arm]
			if x == nil {
				x = &tot{}
				by[t.Arm] = x
				order = append(order, t.Arm)
			}
			x.n++
			x.units += t.Units
			x.runs += t.Runs
			x.in += t.Input
			x.uncached += t.Uncached
			x.read += t.CacheRead
			x.write += t.CacheWrite
			x.usd += t.USD
			x.usdRuns += t.USDRuns
			x.out += t.Output
			x.runMs += t.RunMs
			x.wallMs += t.WallMs
			if t.Planned {
				x.planned++
			}
			if t.Error != "" {
				x.failed++
			}
		}
	}
	var b strings.Builder
	b.WriteString("| Arm | Messages | Planned | Units | Runs | Input tokens | Uncached | Cache read | Cache write | Output tokens | Reported USD (runs) | Run time (s) | Wall time (s) | Failed |\n|---|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, a := range order {
		x := by[a]
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d | %d | %d | %d | %d | %.4f (%d) | %.0f | %.0f | %d |\n", a, x.n, x.planned, x.units, x.runs, x.in,
			x.uncached, x.read, x.write, x.out, x.usd, x.usdRuns,
			float64(x.runMs)/1000, float64(x.wallMs)/1000, x.failed)
	}
	b.WriteString("\nTokens are what the providers reported (0 where a CLI reports none). Planned counts trials the plan path actually fanned out.\n")
	return b.String()
}

// ProfileExists reports whether id names a profile of the configuration.
func ProfileExists(cfg config.Config, id string) bool {
	for _, p := range cfg.AllProfiles() {
		if p.ID == id {
			return true
		}
	}
	return false
}
