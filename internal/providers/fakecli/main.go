// Command fakecli replays a providers.FakeScript: it applies the script's
// edits to the workspace, prints its events as JSON lines, writes the
// handoff, sleeps, and exits with the scripted code. The orchestrator's
// tests build it with providers.BuildFakeCLI and run it through the fake
// adapter, so every row of 04 §4.8 runs under go test without a provider.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/providers"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("fakecli 1.0.0")
		return
	}
	scriptPath := os.Getenv(providers.FakeEnvScript)
	if scriptPath == "" {
		fmt.Fprintln(os.Stderr, "fakecli: "+providers.FakeEnvScript+" is not set")
		os.Exit(2)
	}
	script, err := providers.LoadFakeScript(scriptPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fakecli: "+err.Error())
		os.Exit(2)
	}
	go io.Copy(io.Discard, os.Stdin)

	ws := os.Getenv(providers.FakeEnvWorkspace)
	paths := make([]string, 0, len(script.Edits))
	for p := range script.Edits {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		full := filepath.Join(ws, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "fakecli: edit %s: %v\n", p, err)
			continue
		}
		if err := os.WriteFile(full, []byte(script.Edits[p]), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "fakecli: edit %s: %v\n", p, err)
		}
	}

	rpaths := make([]string, 0, len(script.Replace))
	for p := range script.Replace {
		rpaths = append(rpaths, p)
	}
	sort.Strings(rpaths)
	for _, p := range rpaths {
		full := filepath.Join(ws, p)
		data, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fakecli: replace %s: %v\n", p, err)
			continue
		}
		text := string(data)
		for _, r := range script.Replace[p] {
			if !strings.Contains(text, r.Old) {
				fmt.Fprintf(os.Stderr, "fakecli: replace %s: %q not found\n", p, r.Old)
				continue
			}
			text = strings.Replace(text, r.Old, r.New, 1)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "fakecli: replace %s: %v\n", p, err)
		}
	}

	compaction := providers.Event{Kind: providers.KindCompaction, Message: "fake compaction"}
	emit := func(ev providers.Event) {
		b, _ := json.Marshal(ev)
		os.Stdout.Write(append(b, '\n'))
		if script.EventDelayMs > 0 {
			time.Sleep(time.Duration(script.EventDelayMs) * time.Millisecond)
		}
	}
	for i, ev := range script.Events {
		if script.CompactAt > 0 && i == script.CompactAt {
			emit(compaction)
		}
		emit(ev)
	}
	if script.CompactAt > 0 && script.CompactAt >= len(script.Events) {
		emit(compaction)
	}
	if script.Usage != nil {
		b, _ := json.Marshal(map[string]any{"fake_usage": script.Usage})
		os.Stdout.Write(append(b, '\n'))
	}

	if handoff := os.Getenv(providers.FakeEnvHandoff); handoff != "" && script.Handoff != "" {
		if err := os.MkdirAll(filepath.Dir(handoff), 0o755); err == nil {
			out := []byte(script.Handoff)
			if script.HandoffMerge {
				out = mergeHandoff(handoff, script.Handoff)
			}
			if err := os.WriteFile(handoff, out, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "fakecli: handoff: %v\n", err)
			}
		}
	}
	if script.Stderr != "" {
		fmt.Fprintln(os.Stderr, script.Stderr)
	}
	if script.SleepMs > 0 {
		time.Sleep(time.Duration(script.SleepMs) * time.Millisecond)
	}
	os.Exit(script.ExitCode)
}

// mergeHandoff overlays the scripted handoff object onto the stub at path;
// run_id and attempt always come from the stub.
func mergeHandoff(path, overlay string) []byte {
	base := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &base)
	}
	over := map[string]any{}
	if err := json.Unmarshal([]byte(overlay), &over); err != nil {
		return []byte(overlay)
	}
	runID, attempt := base["run_id"], base["attempt"]
	for k, v := range over {
		base[k] = v
	}
	if runID != nil {
		base["run_id"] = runID
	}
	if attempt != nil {
		base["attempt"] = attempt
	}
	out, _ := json.MarshalIndent(base, "", "  ")
	return out
}
