package engine

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tommac/contextless/internal/vault"
)

// The repository lane is MultiAIWorkspace's worker model reduced to one
// unit: a fresh git worktree on its own branch, an explicit path
// allowlist, an engine-side commit (agents never run git), a verification
// command, and a branch left for the user to merge. The user's checkout,
// index and current branch are never touched.

// RepoDirName is where the worktree sits inside the packet directory.
const RepoDirName = "repo"

// VerifyTimeout bounds the verify command.
var VerifyTimeout = 15 * time.Minute

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	s := strings.TrimSpace(out.String())
	if err != nil {
		return s, fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, s)
	}
	return s, nil
}

// RepoRoot validates a path as a git repository and returns its top level.
func RepoRoot(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return git(abs, "rev-parse", "--show-toplevel")
}

// repoBase resolves the commit units start from.
func repoBase(r vault.Repo) (string, error) {
	ref := r.Base
	if ref == "" {
		ref = "HEAD"
	}
	return git(r.Path, "rev-parse", "--verify", ref+"^{commit}")
}

func repoInstructions(r vault.Repo, runID, base string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n\nRepository %q is checked out at `%s/` in this directory (a fresh worktree of %s at commit %s, on branch `cx/%s`). Make the change there.", r.Name, RepoDirName, r.Path, short(base), runID)
	if len(r.Allow) > 0 {
		fmt.Fprintf(&b, " You may change only these paths: %s. A change anywhere else rejects the whole unit.", strings.Join(r.Allow, ", "))
	}
	b.WriteString(" Do not run git; the engine commits your changes.")
	if r.Verify != "" {
		fmt.Fprintf(&b, " After you finish, the engine runs `%s` in the checkout; you may run it yourself first if your sandbox allows.", r.Verify)
	}
	b.WriteString(" In `answer`, say what you changed and why.")
	return b.String()
}

// repoRun is one unit's worktree.
type repoRun struct {
	repo   vault.Repo
	base   string
	branch string
	wt     string
	kept   string
}

func short(sha string) string {
	if len(sha) > 9 {
		return sha[:9]
	}
	return sha
}

// openRepoRun creates the worktree and branch.
func openRepoRun(r vault.Repo, base, runID, ws string) (*repoRun, error) {
	rr := &repoRun{repo: r, base: base, branch: "cx/" + runID, wt: filepath.Join(ws, RepoDirName)}
	if _, err := git(r.Path, "worktree", "add", "-q", "-b", rr.branch, rr.wt, base); err != nil {
		return nil, err
	}
	return rr, nil
}

// discard removes the worktree and the branch (nothing worth keeping).
func (rr *repoRun) discard() {
	_, _ = git(rr.repo.Path, "worktree", "remove", "--force", rr.wt)
	_, _ = git(rr.repo.Path, "branch", "-D", rr.branch)
	_, _ = git(rr.repo.Path, "worktree", "prune")
}

// keep removes the worktree and keeps the branch.
func (rr *repoRun) keep() {
	_, _ = git(rr.repo.Path, "worktree", "remove", "--force", rr.wt)
	_, _ = git(rr.repo.Path, "worktree", "prune")
}

// repoChain carries a unit's base across its runs (recall rounds and
// continuations), so each run builds on the previous run's commit.
type repoChain struct {
	mu   sync.Mutex
	base string
}

func (c *repoChain) get(r vault.Repo) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.base == "" {
		b, err := repoBase(r)
		if err != nil {
			return "", err
		}
		c.base = b
	}
	return c.base, nil
}

func (c *repoChain) set(sha string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.base = sha
}

// repoChange is what one run of a coder unit did to its repository, as
// the engine checked, committed and verified it. A unit's record lists
// one per run; Report is the text the record shows.
type repoChange struct {
	Repo string `json:"repo"`
	Path string `json:"path"`
	// Base is the commit the run started from: the repository's base, or
	// the previous run's commit within the same unit.
	Base   string   `json:"base"`
	Branch string   `json:"branch,omitempty"` // kept for the user to merge; "" when nothing was kept
	Commit string   `json:"commit,omitempty"`
	Files  []string `json:"files,omitempty"`
	Stat   string   `json:"stat,omitempty"`
	Verify string   `json:"verify,omitempty"`
	// Verified is the verify command's result, nil when none ran.
	Verified *bool  `json:"verified,omitempty"`
	Report   string `json:"report"`
}

// finish checks, commits and verifies the unit's change and returns what
// it did (Commit is "" when nothing was kept). It always removes the
// worktree.
func (rr *repoRun) finish(title, runID, profile string) repoChange {
	ch := repoChange{Repo: rr.repo.Name, Path: rr.repo.Path, Base: rr.base}
	ch.Report = rr.finishReport(title, runID, profile, &ch)
	if rr.kept != "" {
		ch.Branch, ch.Commit = rr.branch, rr.kept
	}
	return ch
}

func (rr *repoRun) finishReport(title, runID, profile string, ch *repoChange) string {
	if _, err := git(rr.wt, "add", "-A"); err != nil {
		rr.discard()
		return "The change could not be staged: " + err.Error()
	}
	// diff against the base, so a change the agent committed itself is
	// still checked, then folded into one engine commit
	names, err := git(rr.wt, "diff", "--cached", "--name-only", rr.base)
	if err != nil {
		rr.discard()
		return "The change could not be read: " + err.Error()
	}
	var changed []string
	for _, n := range strings.Split(names, "\n") {
		if n = strings.TrimSpace(n); n != "" {
			changed = append(changed, n)
		}
	}
	ch.Files = changed
	if len(changed) == 0 {
		rr.discard()
		return fmt.Sprintf("No files in %s were changed, so no branch was kept.", rr.repo.Name)
	}
	var outside []string
	for _, n := range changed {
		if !allowed(n, rr.repo.Allow) {
			outside = append(outside, n)
		}
	}
	if len(outside) > 0 {
		rr.discard()
		return fmt.Sprintf("REJECTED: the unit changed files outside the allowed paths of %s (%s): %s. Nothing was kept.",
			rr.repo.Name, strings.Join(rr.repo.Allow, ", "), strings.Join(outside, ", "))
	}
	if head, _ := git(rr.wt, "rev-parse", "HEAD"); head != rr.base {
		_, _ = git(rr.wt, "reset", "-q", "--soft", rr.base)
	}
	stat, _ := git(rr.wt, "diff", "--cached", "--shortstat", rr.base)
	ch.Stat = strings.TrimSpace(stat)
	msg := fmt.Sprintf("%s\n\nContextlessAI run %s by %s", title, runID, profile)
	if _, err := git(rr.wt, "-c", "user.name=ContextlessAI", "-c", "user.email=contextless@localhost", "commit", "-q", "--no-verify", "-m", msg); err != nil {
		rr.discard()
		return "The change could not be committed: " + err.Error()
	}
	sha, _ := git(rr.wt, "rev-parse", "HEAD")
	rr.kept = sha
	var b strings.Builder
	fmt.Fprintf(&b, "Changes are on branch `%s` (commit %s) in %s, based on %s: %s.\nFiles: %s.",
		rr.branch, short(sha), rr.repo.Path, short(rr.base), strings.TrimSpace(stat), strings.Join(changed, ", "))
	if rr.repo.Verify != "" {
		ok, code, tail := runVerify(rr.wt, rr.repo.Verify)
		ch.Verify, ch.Verified = rr.repo.Verify, &ok
		verdict := "passed"
		if !ok {
			verdict = fmt.Sprintf("FAILED (exit %d)", code)
		}
		fmt.Fprintf(&b, "\nVerify `%s`: %s.", rr.repo.Verify, verdict)
		if tail != "" {
			fmt.Fprintf(&b, "\n```\n%s\n```", tail)
		}
	}
	fmt.Fprintf(&b, "\nNothing was merged. To take it: git -C %s merge %s", rr.repo.Path, rr.branch)
	rr.keep()
	return b.String()
}

// safeVerifyEnv strips credential/secret variables from the environment
// passed to verify commands so agent-edited tests or build scripts cannot
// exfiltrate API keys or tokens from the host environment.
func safeVerifyEnv() []string {
	denySubstr := []string{"KEY", "TOKEN", "SECRET", "PASS", "AUTH", "CREDENTIAL", "COOKIE"}
	var out []string
	for _, kv := range os.Environ() {
		k, _, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		upper := strings.ToUpper(k)
		blocked := false
		for _, bad := range denySubstr {
			if strings.Contains(upper, bad) {
				blocked = true
				break
			}
		}
		if !blocked {
			out = append(out, kv)
		}
	}
	return out
}

// runVerify runs the verify command in the worktree with a secret-stripped
// environment and its own process group so timeouts kill grandchild processes.
func runVerify(dir, command string) (ok bool, code int, tail string) {
	ctx, cancel := context.WithTimeout(context.Background(), VerifyTimeout)
	defer cancel()
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = safeVerifyEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	if err := cmd.Start(); err != nil {
		return false, -1, err.Error()
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var err error
	select {
	case err = <-done:
	case <-ctx.Done():
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			select {
			case err = <-done:
			case <-time.After(2 * time.Second):
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				err = <-done
			}
		}
		if err == nil {
			err = ctx.Err()
		}
	}
	t := strings.TrimSpace(buf.String())
	if len(t) > 3000 {
		t = "…" + t[len(t)-3000:]
	}
	if err == nil {
		return true, 0, t
	}
	if ee, isExit := err.(*exec.ExitError); isExit {
		return false, ee.ExitCode(), t
	}
	return false, -1, t + "\n" + err.Error()
}

// IntegrationCheck is the result of merging several coder branches onto a
// throwaway scratch worktree and running verify once to detect write skew.
type IntegrationCheck struct {
	Repo     string   `json:"repo"`
	Base     string   `json:"base"`
	Branches []string `json:"branches"`
	Merged   bool     `json:"merged"`
	Verified *bool    `json:"verified,omitempty"`
	Output   string   `json:"output,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// ScratchMergeAndVerify merges multiple coder branches in order onto a
// throwaway worktree rooted at base and runs the repo's verify command once,
// catching semantic write skew without touching the user's checkout or branches.
func ScratchMergeAndVerify(r vault.Repo, base string, branches []string) IntegrationCheck {
	ic := IntegrationCheck{Repo: r.Name, Base: base, Branches: branches}
	if len(branches) < 2 {
		ic.Merged = true
		return ic
	}
	if base == "" {
		var err error
		base, err = repoBase(r)
		if err != nil {
			ic.Error = err.Error()
			return ic
		}
		ic.Base = base
	}
	tmpDir, err := os.MkdirTemp("", "cx-scratch-merge-*")
	if err != nil {
		ic.Error = err.Error()
		return ic
	}
	wt := filepath.Join(tmpDir, "repo")
	scratchBranch := "cx/scratch-" + vault.NewID("m")
	defer func() {
		_, _ = git(r.Path, "worktree", "remove", "--force", wt)
		_, _ = git(r.Path, "branch", "-D", scratchBranch)
		_, _ = git(r.Path, "worktree", "prune")
		_ = os.RemoveAll(tmpDir)
	}()
	if _, err := git(r.Path, "worktree", "add", "-q", "-b", scratchBranch, wt, base); err != nil {
		ic.Error = err.Error()
		return ic
	}
	for _, br := range branches {
		if out, err := git(wt, "-c", "user.name=ContextlessAI", "-c", "user.email=contextless@localhost", "merge", "--no-ff", "-m", "Integrate "+br, br); err != nil {
			ic.Error = fmt.Sprintf("merge conflict integrating %s: %v", br, err)
			ic.Output = out
			return ic
		}
	}
	ic.Merged = true
	if r.Verify != "" {
		ok, _, tail := runVerify(wt, r.Verify)
		ic.Verified = &ok
		ic.Output = tail
		if !ok {
			ic.Error = "integrated branches failed verify: " + r.Verify
		}
	}
	return ic
}

// MergeBranch re-verifies a cx/* branch merged against the repository's
// current HEAD in a throwaway worktree before fast-forwarding/merging into
// the repository's current checkout (MAW integration step).
func MergeBranch(r vault.Repo, branch string) (string, error) {
	head, err := git(r.Path, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return "", err
	}
	if status, err := git(r.Path, "status", "--porcelain"); err != nil {
		return "", err
	} else if strings.TrimSpace(status) != "" {
		return "", fmt.Errorf("repository %s has uncommitted changes; commit or stash before merging %s", r.Name, branch)
	}
	tmpDir, err := os.MkdirTemp("", "cx-merge-verify-*")
	if err != nil {
		return "", err
	}
	wt := filepath.Join(tmpDir, "repo")
	tmpBranch := "cx/verify-" + vault.NewID("v")
	defer func() {
		_, _ = git(r.Path, "worktree", "remove", "--force", wt)
		_, _ = git(r.Path, "branch", "-D", tmpBranch)
		_, _ = git(r.Path, "worktree", "prune")
		_ = os.RemoveAll(tmpDir)
	}()
	if _, err := git(r.Path, "worktree", "add", "-q", "-b", tmpBranch, wt, head); err != nil {
		return "", err
	}
	if out, err := git(wt, "-c", "user.name=ContextlessAI", "-c", "user.email=contextless@localhost", "merge", "--no-ff", "-m", "Verify merge "+branch, branch); err != nil {
		return out, fmt.Errorf("merge %s onto moved HEAD %s failed: %w", branch, short(head), err)
	}
	if r.Verify != "" {
		ok, code, tail := runVerify(wt, r.Verify)
		if !ok {
			return tail, fmt.Errorf("re-verify `%s` on moved base %s FAILED (exit %d)", r.Verify, short(head), code)
		}
	}
	out, err := git(r.Path, "merge", "--no-edit", branch)
	if err != nil {
		return out, err
	}
	return out, nil
}

// allowed reports whether a repository path matches the allowlist. An
// empty list allows everything but .git.
func allowed(path string, patterns []string) bool {
	if path == ".git" || strings.HasPrefix(path, ".git/") {
		return false
	}
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if globRe(p).MatchString(path) {
			return true
		}
	}
	return false
}

// globRe compiles a path glob: ** crosses directories, * and ? do not, and
// "dir/**" matches everything under dir.
func globRe(p string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(p); i++ {
		switch c := p[i]; {
		case c == '*' && i+1 < len(p) && p[i+1] == '*':
			b.WriteString(".*")
			i++
			if i+1 < len(p) && p[i+1] == '/' {
				i++
				b.WriteString("/?")
			}
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}
