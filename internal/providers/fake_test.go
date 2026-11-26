package providers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFakeRunEndToEnd(t *testing.T) {
	inv := newInvocation(t, "fake")
	script := writeScript(t, t.TempDir(), "ok", FakeScript{
		ExitCode: 0,
		Events: []Event{
			{Kind: KindText, Message: "hello"},
			{Kind: KindUsage, ContextTokens: 1200},
			{Kind: KindTool, Message: "Read"},
			{Kind: KindUsage, ContextTokens: 2500},
		},
		Edits:   map[string]string{"scenes/sc-a.md": "---\nid: sc-a\n---\nNew body.\n"},
		Handoff: `{"status":"completed","milestone_seq":1}`,
		Usage:   &Usage{InputTokens: 4000, CachedTokens: 1000, OutputTokens: 300, Turns: 2},
		Stderr:  "fake done",
	})
	a := &FakeAdapter{Binary: fakeBin, ScriptPath: script}
	var seen []Event
	out, err := Run(context.Background(), a, inv, false, func(ev Event) { seen = append(seen, ev) })
	if err != nil {
		t.Fatal(err)
	}
	if out.ExitCode != 0 || out.ExitClass != Ok || out.Killed != "" || out.Compacted {
		t.Errorf("outcome %+v", out)
	}
	if len(seen) != 5 || len(out.Events) != 5 {
		t.Fatalf("events seen %d stored %d", len(seen), len(out.Events))
	}
	if seen[2].Kind != KindTool || seen[2].Message != "Read" || seen[3].ContextTokens != 2500 || seen[4].Message != "final" {
		t.Errorf("events %+v", seen)
	}
	if out.Usage.InputTokens != 4000 || out.Usage.CachedTokens != 1000 || out.Usage.PeakContext != 2500 || out.Usage.WallMs <= 0 {
		t.Errorf("usage %+v", out.Usage)
	}
	body, err := os.ReadFile(filepath.Join(inv.Workspace, "scenes", "sc-a.md"))
	if err != nil || !strings.Contains(string(body), "New body.") {
		t.Errorf("edit not applied: %v %q", err, body)
	}
	handoff, _ := os.ReadFile(inv.HandoffPath)
	if !strings.Contains(string(handoff), `"completed"`) {
		t.Errorf("handoff %q", handoff)
	}
	for _, name := range []string{FileStdout, FileStderr, FileEvents, FileUsage, FileArgv, FileEnv, FilePGID} {
		if _, err := os.Stat(filepath.Join(inv.RunDir, name)); err != nil {
			t.Errorf("missing %s", name)
		}
	}
	events, err := ReadEvents(out.EventsPath)
	if err != nil || len(events) != 5 {
		t.Errorf("events.jsonl: %v %d", err, len(events))
	}
	if !strings.Contains(out.StderrTail, "fake done") {
		t.Errorf("stderr tail %q", out.StderrTail)
	}
	if _, err := os.Stat(filepath.Join(inv.RunDir, "tmp")); err != nil {
		t.Error("tmp dir")
	}
	envFile, _ := os.ReadFile(filepath.Join(inv.RunDir, FileEnv))
	if !strings.Contains(string(envFile), FakeEnvScript) {
		t.Error("env.redacted.json")
	}
}

func TestFakeRunCrashAndSignatures(t *testing.T) {
	inv := newInvocation(t, "fake")
	dir := t.TempDir()
	crash := writeScript(t, dir, "crash", FakeScript{ExitCode: 3, Events: []Event{{Kind: KindUsage, ContextTokens: 10}}, Stderr: "boom"})
	out, err := Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: crash}, inv, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.ExitCode != 3 || out.ExitClass != Crash {
		t.Errorf("crash outcome %+v", out)
	}
	quota := writeScript(t, dir, "quota", FakeScript{ExitCode: 1, Events: []Event{{Kind: KindError, Message: "429 usage limit reached, resets in 3h"}}})
	out, _ = Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: quota}, inv, false, nil)
	if out.ExitClass != Quota {
		t.Errorf("quota outcome %s", out.ExitClass)
	}
	auth := writeScript(t, dir, "auth", FakeScript{ExitCode: 1, Stderr: "Error: not logged in"})
	out, _ = Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: auth}, inv, false, nil)
	if out.ExitClass != Auth {
		t.Errorf("auth outcome %s", out.ExitClass)
	}
	empty := writeScript(t, dir, "empty", FakeScript{ExitCode: 0})
	out, _ = Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: empty}, inv, false, nil)
	if out.ExitClass != Ok || !out.UsageMissing {
		t.Errorf("empty outcome %+v", out)
	}
	blocked := writeScript(t, dir, "blocked", FakeScript{ExitCode: 0, Events: []Event{{Kind: KindError, Message: "blocked: max turns"}}})
	out, _ = Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: blocked}, inv, false, nil)
	if out.ExitClass != OkBlocked {
		t.Errorf("blocked outcome %s", out.ExitClass)
	}
}

func TestFakeRunCompactionKill(t *testing.T) {
	inv := newInvocation(t, "fake")
	script := writeScript(t, t.TempDir(), "compact", FakeScript{
		ExitCode:  0,
		Events:    []Event{{Kind: KindUsage, ContextTokens: 100}, {Kind: KindUsage, ContextTokens: 200}, {Kind: KindUsage, ContextTokens: 300}},
		CompactAt: 2,
		SleepMs:   20000,
		Handoff:   `{"status":"completed"}`,
	})
	start := time.Now()
	out, err := Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: script}, inv, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Killed != KilledCompaction || out.ExitClass != TaintedCompaction || !out.Compacted {
		t.Errorf("outcome %+v", out)
	}
	if time.Since(start) > 10*time.Second {
		t.Error("the compaction kill must not wait for the sleep")
	}
	if out.ExitCode != -1 {
		t.Errorf("a killed process has exit code -1, got %d", out.ExitCode)
	}
}

func TestFakeRunKillLine(t *testing.T) {
	inv := newInvocation(t, "fake")
	// kill line is 0.85 * 10000 = 8500
	script := writeScript(t, t.TempDir(), "killline", FakeScript{
		ExitCode: 0,
		Events:   []Event{{Kind: KindUsage, ContextTokens: 5000}, {Kind: KindUsage, ContextTokens: 8600}},
		SleepMs:  20000,
	})
	out, err := Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: script}, inv, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Killed != KilledKillLine || out.ExitClass != ContextBlocked || out.Usage.PeakContext != 8600 {
		t.Errorf("outcome %+v", out)
	}
}

func TestFakeRunTimeoutAndCancel(t *testing.T) {
	inv := newInvocation(t, "fake")
	inv.TimeoutSeconds = 1
	script := writeScript(t, t.TempDir(), "slow", FakeScript{ExitCode: 0, SleepMs: 30000})
	out, err := Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: script}, inv, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Killed != KilledTimeout || out.ExitClass != Timeout {
		t.Errorf("timeout outcome %+v", out)
	}

	inv.TimeoutSeconds = 60
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(500 * time.Millisecond); cancel() }()
	start := time.Now()
	out, err = Run(ctx, &FakeAdapter{Binary: fakeBin, ScriptPath: script}, inv, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Killed != KilledCancelled || out.ExitClass != Crash || time.Since(start) > 10*time.Second {
		t.Errorf("cancel outcome %+v after %s", out, time.Since(start))
	}
}

func TestFakeRunStall(t *testing.T) {
	inv := newInvocation(t, "fake")
	inv.StallSeconds = 1
	inv.TimeoutSeconds = 60
	script := writeScript(t, t.TempDir(), "stall", FakeScript{ExitCode: 0, Events: []Event{{Kind: KindText, Message: "hi"}}, SleepMs: 30000})
	out, err := Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: script}, inv, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Killed != KilledStalled || out.ExitClass != Stalled {
		t.Errorf("stall outcome %+v", out)
	}
}

func TestFakeRunMarker(t *testing.T) {
	inv := newInvocation(t, "fake")
	script := writeScript(t, t.TempDir(), "marker", FakeScript{ExitCode: 0, Events: []Event{{Kind: KindUsage, ContextTokens: 10}}})
	// The gemini PreCompress hook would write this; simulate it with an edit
	// that lands in RunDir through the workspace-relative path.
	if err := os.WriteFile(CompactionMarkerPath(inv), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: script}, inv, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !out.Compacted || out.ExitClass != TaintedCompaction || out.Killed != "" {
		t.Errorf("marker outcome %+v", out)
	}
	last := out.Events[len(out.Events)-1]
	if last.Kind != KindCompaction || !strings.HasPrefix(last.Message, "marker:") {
		t.Errorf("marker event %+v", last)
	}
}

func TestFakeScriptsPerRole(t *testing.T) {
	inv := newInvocation(t, "fake")
	dir := t.TempDir()
	a := &FakeAdapter{Binary: fakeBin,
		ScriptPath: writeScript(t, dir, "default", FakeScript{ExitCode: 0, Handoff: "default"}),
		Scripts:    map[string]string{"judge": writeScript(t, dir, "judge", FakeScript{ExitCode: 0, Handoff: "judge"})},
	}
	inv.Role = "judge"
	if _, err := Run(context.Background(), a, inv, false, nil); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(inv.HandoffPath); string(b) != "judge" {
		t.Errorf("role script not used: %q", b)
	}
	inv.Role = "drafter"
	Run(context.Background(), a, inv, false, nil)
	if b, _ := os.ReadFile(inv.HandoffPath); string(b) != "default" {
		t.Errorf("default script not used: %q", b)
	}
	if _, err := Run(context.Background(), &FakeAdapter{}, inv, false, nil); err == nil {
		t.Error("a fake without a binary must fail to start")
	}
}
