package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FakeScript is what the fake CLI replays (04 §4.15): events to print as
// JSON lines, workspace edits to apply first, a handoff to write, a sleep
// before exit, and the exit code. CompactAt inserts a compaction event
// after N events (N at or past the end appends one). EventDelayMs paces
// the stream so watchdog tests can observe a running process; Usage, when
// set, is printed as a final {"fake_usage": ...} line and wins over the
// derived usage; Stderr is printed to stderr before exit.
type FakeScript struct {
	ExitCode     int               `json:"exit_code"`
	Events       []Event           `json:"events"`
	Edits        map[string]string `json:"edits,omitempty"`
	Handoff      string            `json:"handoff,omitempty"`
	SleepMs      int               `json:"sleep_ms,omitempty"`
	CompactAt    int               `json:"compact_at,omitempty"`
	EventDelayMs int               `json:"event_delay_ms,omitempty"`
	Usage        *Usage            `json:"usage,omitempty"`
	Stderr       string            `json:"stderr,omitempty"`
	// Replace applies substring replacements to existing workspace files
	// (path to an ordered list), after Edits. Each Old must occur; the first
	// occurrence is replaced.
	Replace map[string][]FakeReplace `json:"replace,omitempty"`
	// HandoffMerge overlays the Handoff JSON object onto the stub handoff the
	// packet wrote (keeping run_id and attempt), instead of replacing it, so
	// a hand-written script passes the attempt fence.
	HandoffMerge bool `json:"handoff_merge,omitempty"`
}

// FakeReplace is one substring replacement in a fake script.
type FakeReplace struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// Environment variables the fake CLI reads.
const (
	FakeEnvScript    = "BW_FAKE_SCRIPT"
	FakeEnvWorkspace = "BW_FAKE_WORKSPACE"
	FakeEnvHandoff   = "BW_FAKE_HANDOFF"
	FakeEnvRunDir    = "BW_FAKE_RUNDIR"
)

// LoadFakeScript reads a script file.
func LoadFakeScript(path string) (FakeScript, error) {
	var s FakeScript
	b, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, fmt.Errorf("providers: fake script %s: %w", path, err)
	}
	return s, nil
}

// WriteFakeScript writes a script file.
func WriteFakeScript(path string, s FakeScript) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, jsonMarshal(s), 0o644)
}

// FakeAdapter runs the fake CLI (internal/providers/fakecli). Scripts
// maps an Invocation.Role to a script path and wins over ScriptPath, so one
// adapter can answer the smoke's several roles differently.
type FakeAdapter struct {
	Binary     string
	ScriptPath string
	Scripts    map[string]string
}

// Provider implements Adapter.
func (*FakeAdapter) Provider() string { return "fake" }

func (a *FakeAdapter) script(inv Invocation) string {
	if p, ok := a.Scripts[inv.Role]; ok {
		return p
	}
	return a.ScriptPath
}

// BuildCommand implements Adapter.
func (a *FakeAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	if a.Binary == "" {
		return nil, nil, nil, errors.New("providers: fake adapter has no Binary; build it with BuildFakeCLI")
	}
	script := a.script(inv)
	if script == "" {
		return nil, nil, nil, fmt.Errorf("providers: fake adapter has no script for role %q", inv.Role)
	}
	var stdin []byte
	if inv.BriefPath != "" {
		stdin, _ = os.ReadFile(inv.BriefPath)
	}
	env := baseEnv(inv, map[string]string{
		FakeEnvScript:    script,
		FakeEnvWorkspace: inv.Workspace,
		FakeEnvHandoff:   inv.handoffPath(),
		FakeEnvRunDir:    inv.RunDir,
	})
	return []string{a.Binary}, env, stdin, nil
}

// SandboxDirs lets the sandbox read the binary and the scripts.
func (a *FakeAdapter) SandboxDirs() []string {
	var dirs []string
	if a.Binary != "" {
		dirs = append(dirs, filepath.Dir(a.Binary))
	}
	if a.ScriptPath != "" {
		dirs = append(dirs, filepath.Dir(a.ScriptPath))
	}
	for _, p := range a.Scripts {
		dirs = append(dirs, filepath.Dir(p))
	}
	return dirs
}

// ParseEvent implements Adapter: each line is an Event in JSON, or the
// final {"fake_usage": ...} line.
func (*FakeAdapter) ParseEvent(line []byte) (Event, bool) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] != '{' {
		return Event{}, false
	}
	if bytes.Contains(line, []byte(`"fake_usage"`)) {
		return Event{Kind: KindUsage, Message: "final", Raw: string(line)}, true
	}
	var ev Event
	if err := json.Unmarshal(line, &ev); err != nil || ev.Kind == "" {
		return Event{}, false
	}
	ev.Raw = string(line)
	return ev, true
}

// FinalUsage implements Adapter: the fake_usage line when present,
// otherwise peak context from the usage events.
func (*FakeAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	if m := lastObjectLine(stdout, func(m map[string]any) bool { return m["fake_usage"] != nil }); m != nil {
		b, _ := json.Marshal(m["fake_usage"])
		var u Usage
		if err := json.Unmarshal(b, &u); err == nil {
			if u.PeakContext == 0 {
				u.PeakContext = PeakContext(events)
			}
			return u, nil
		}
	}
	u := Usage{PeakContext: PeakContext(events)}
	for _, ev := range events {
		if ev.Kind == KindUsage {
			u.Turns++
		}
	}
	if u.PeakContext == 0 && u.Turns == 0 {
		return Usage{}, ErrNoUsage
	}
	u.InputTokens = u.PeakContext
	return u, nil
}

// ExitClass implements Adapter.
func (*FakeAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	return classifyExit(exitCode, stderr, events, func(ev Event) bool {
		return ev.Kind == KindError && strings.HasPrefix(ev.Message, "blocked:")
	})
}

// InstructionFile implements Adapter.
func (*FakeAdapter) InstructionFile() string { return "AGENTS.md" }

// ProjectSettings implements Adapter.
func (*FakeAdapter) ProjectSettings(inv Invocation) map[string][]byte { return nil }

// FakeCLIPackage is the import path of the fake CLI.
const FakeCLIPackage = "github.com/tommac/contextless/internal/providers/fakecli"

// BuildFakeCLI compiles the fake CLI into outDir and returns its path. It
// finds the module root from this source file, so tests in any package can
// call it.
func BuildFakeCLI(ctx context.Context, outDir string) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("providers: cannot locate the module root")
	}
	root := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", errors.New("providers: go.mod not found above " + file)
		}
		root = parent
	}
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if _, err := os.Stat(goBin); err != nil {
		goBin = "go"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	out := filepath.Join(outDir, "fakecli")
	cmd := exec.CommandContext(ctx, goBin, "build", "-o", out, FakeCLIPackage)
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("providers: build fakecli: %w\n%s", err, b)
	}
	return out, nil
}
