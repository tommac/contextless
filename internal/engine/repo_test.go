package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

func gitT(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func newRepo(t *testing.T) string {
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	// The engine merges into a user's repository as that user, so the test
	// repository carries its own identity: a machine with no global git
	// identity (a fresh container, a CI runner) would refuse the merge.
	gitT(t, dir, "config", "user.name", "t")
	gitT(t, dir, "config", "user.email", "t@t")
	_ = os.MkdirAll(filepath.Join(dir, "pkg"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "pkg", "a.go"), []byte("package pkg\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme\n"), 0o644)
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-q", "-m", "base")
	return dir
}

func attach(t *testing.T, h *harness, r vault.Repo) {
	s, _ := h.e.V.Session()
	s.Repos = append(s.Repos, r)
	if err := h.e.V.PutSession(s); err != nil {
		t.Fatal(err)
	}
}

func planOneCoder(c call) packet.Handoff {
	return packet.Handoff{RunID: c.RunID, Status: packet.StatusPlan, Units: []packet.Unit{
		{ID: "u1", Role: "coder", Title: "Add B", Instructions: "Add pkg/b.go that says hello.", Repo: "proj"},
	}}
}

func TestRepoUnitCommitsOnItsOwnBranchAndVerifies(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	repo := newRepo(t)
	headBefore := gitT(t, repo, "rev-parse", "HEAD")
	attach(t, h, vault.Repo{Name: "proj", Path: repo, Allow: []string{"pkg/**"}, Verify: "grep -q hello pkg/b.go"})
	var unitAnswer string
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			if !strings.Contains(c.Brief, "Attached repositories") || !strings.Contains(c.Brief, "may change only pkg/**") {
				t.Errorf("session does not list the attached repo")
			}
			return planOneCoder(c)
		case packet.KindUnit:
			if !strings.Contains(c.Brief, "is checked out at `repo/`") {
				t.Errorf("coder brief lacks the checkout instructions")
			}
			return withEdits{H: done(c, "Added pkg/b.go."), Edits: map[string]string{"repo/pkg/b.go": "package pkg\n\n// hello\n"}}
		case packet.KindAssemble:
			unitAnswer = c.Brief
			return done(c, "Done: see branch.")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), "Add a hello file to the proj repo."); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(unitAnswer, "Changes are on branch `cx/run-") || !strings.Contains(unitAnswer, "passed") {
		t.Fatalf("unit report missing branch or verify result:\n%s", unitAnswer)
	}
	branches := gitT(t, repo, "branch", "--list", "cx/*")
	if strings.Count(branches, "cx/run-") != 1 {
		t.Fatalf("branches: %q", branches)
	}
	br := strings.TrimSpace(strings.TrimPrefix(branches, "*"))
	if got := gitT(t, repo, "show", br+":pkg/b.go"); !strings.Contains(got, "hello") {
		t.Fatalf("branch content %q", got)
	}
	// the user's checkout is untouched
	if gitT(t, repo, "rev-parse", "HEAD") != headBefore || gitT(t, repo, "status", "--porcelain") != "" {
		t.Fatal("the user's checkout changed")
	}
	if wt := gitT(t, repo, "worktree", "list"); strings.Count(wt, "\n") != 0 {
		t.Fatalf("worktree left behind: %s", wt)
	}
	// the unit's record carries the change as fields, not only as text
	rec := readRecord(t, h, h.callsOf(packet.KindRespond)[0].RunID, "u1")
	if len(rec.Changes) != 1 {
		t.Fatalf("record changes: %+v", rec.Changes)
	}
	ch := rec.Changes[0]
	if ch.Repo != "proj" || ch.Branch != br || ch.Commit != gitT(t, repo, "rev-parse", br) || ch.Base != headBefore ||
		ch.Verified == nil || !*ch.Verified || strings.Join(ch.Files, ",") != "pkg/b.go" {
		t.Fatalf("recorded change: %+v", ch)
	}
	if strings.Contains(rec.Output, "Changes are on branch") || !strings.Contains(unitAnswer, "Repository proj:\nChanges are on branch") {
		t.Fatal("the change report belongs to the record, under its repository")
	}
}

func TestRepoUnitOutsideAllowlistIsRejected(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	repo := newRepo(t)
	attach(t, h, vault.Repo{Name: "proj", Path: repo, Allow: []string{"pkg/**"}})
	var unitAnswer string
	h.reply = func(c call) any {
		switch c.Kind {
		case packet.KindRespond:
			return planOneCoder(c)
		case packet.KindUnit:
			return withEdits{H: done(c, "Edited."), Edits: map[string]string{"repo/pkg/b.go": "package pkg\n", "repo/README.md": "rewritten\n"}}
		case packet.KindAssemble:
			unitAnswer = c.Brief
			return done(c, "ok")
		}
		return packet.IndexHandoff{RunID: c.RunID, Status: "completed"}
	}
	if _, err := h.e.Ask(context.Background(), "Change things."); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(unitAnswer, "REJECTED") || !strings.Contains(unitAnswer, "README.md") {
		t.Fatalf("want a rejection naming README.md:\n%s", unitAnswer)
	}
	if b := gitT(t, repo, "branch", "--list", "cx/*"); b != "" {
		t.Fatalf("a rejected unit left a branch: %q", b)
	}
}

func TestAllowGlobs(t *testing.T) {
	cases := []struct {
		path  string
		globs []string
		want  bool
	}{
		{"pkg/a/b.go", []string{"pkg/**"}, true},
		{"pkgx/b.go", []string{"pkg/**"}, false},
		{"x/y/z.go", []string{"**/*.go"}, true},
		{"z.go", []string{"**/*.go"}, true},
		{"docs/a.md", []string{"*.md"}, false},
		{"README.md", []string{"*.md"}, true},
		{".git/config", nil, false},
		{"anything", nil, true},
	}
	for _, c := range cases {
		if got := allowed(c.path, c.globs); got != c.want {
			t.Errorf("allowed(%q, %v) = %v", c.path, c.globs, got)
		}
	}
}
