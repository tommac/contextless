package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var fakeBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "bw-fakecli-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	bin, err := BuildFakeCLI(ctx, dir)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	fakeBin = bin
	stallPollInterval = 200 * time.Millisecond
	killGrace = 300 * time.Millisecond
	pipeCloseGrace = 1 * time.Second
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// testProfile is a small profile so watchdog arithmetic stays readable.
func testProfile(provider string) Profile {
	return Profile{Provider: provider, Model: "m-test", Effort: "high", Window: 10000, EffectiveWindow: 10000,
		LiveUsage: LivePerTurn, Enabled: true}.WithDefaults()
}

// newInvocation builds a workspace with a brief and a run dir under t.TempDir.
func newInvocation(t *testing.T, provider string) Invocation {
	t.Helper()
	root := t.TempDir()
	ws := filepath.Join(root, "ws")
	run := filepath.Join(root, "run")
	for _, d := range []string{filepath.Join(ws, ".cx"), run} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	brief := filepath.Join(ws, "brief.md")
	if err := os.WriteFile(brief, []byte("# Brief\n\nDo the thing.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return Invocation{
		RunID: "run-test", Role: "drafter", Profile: testProfile(provider),
		Workspace: ws, BriefPath: brief, HandoffPath: filepath.Join(ws, ".cx", "handoff.json"),
		RunDir: run, MaxBudgetUSD: 6, TimeoutSeconds: 30,
	}
}

func writeScript(t *testing.T, dir, name string, s FakeScript) string {
	t.Helper()
	p := filepath.Join(dir, name+".json")
	if err := WriteFakeScript(p, s); err != nil {
		t.Fatal(err)
	}
	return p
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
