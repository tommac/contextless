package providers

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// KimiAdapter drives kimi 0.43.1 headless (06 §6.2, research §3). The
// prompt is a pointer at brief.md in the workspace; usage_update gives live
// context (used, size); compaction cannot be disabled and a
// compaction.started line taints the run. Effort is a config-level setting
// on kimi, so the profile's Effort is recorded but not passed.
type KimiAdapter struct {
	// Binary is the executable; "kimi" when empty.
	Binary string
}

// KimiPointer is the -p text; the brief must be brief.md in the workspace.
const KimiPointer = "Read brief.md in this directory and follow it exactly."

// KimiSkillsDir is the RunDir entry passed as --skills-dir. Nothing
// creates it, so the run has no skills.
const KimiSkillsDir = "kimi-no-skills"

// Provider implements Adapter.
func (KimiAdapter) Provider() string { return "kimi" }

func (a KimiAdapter) bin() string {
	if a.Binary != "" {
		return a.Binary
	}
	return "kimi"
}

// BuildCommand implements Adapter.
func (a KimiAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	if _, err := readBrief(inv); err != nil {
		return nil, nil, nil, err
	}
	if !insideWorkspace(inv, inv.BriefPath) {
		return nil, nil, nil, fmt.Errorf("providers: %s: kimi brief %s must be inside the workspace", inv.RunID, inv.BriefPath)
	}
	pointer := KimiPointer
	if base := filepath.Base(inv.BriefPath); base != "brief.md" {
		pointer = "Read " + base + " in this directory and follow it exactly."
	}
	// Prompt mode approves edits on its own; kimi refuses --auto with -p
	// ("Cannot combine --prompt with --auto", stage 0 finding).
	argv := []string{a.bin(), "-p", pointer, "-m", inv.Profile.Model, "--output-format", "stream-json"}
	if inv.RunDir != "" {
		// An explicit, empty skills directory: the run loads none of the
		// user's personal skills, and kimi does not watch the home
		// directory for them (under the sandbox that watch never becomes
		// ready and the run exits 0 with no output; stage 0 finding).
		argv = append(argv, "--skills-dir", filepath.Join(inv.RunDir, KimiSkillsDir))
	}
	return argv, baseEnv(inv, nil), nil, nil
}

var kimiCompactionRe = regexp.MustCompile(`compaction\.[a-z_]+`)

// ParseEvent implements Adapter for kimi's ACP-style stream-json. The
// exact nesting is unverified, so the parser searches the object: any
// compaction.* token is the compaction signal; an object with numeric
// "used" and "size" is a usage_update; sessionUpdate tool_call is a tool
// event; agent_message_chunk is text; an error member is an error.
func (KimiAdapter) ParseEvent(line []byte) (Event, bool) {
	m, ok := parseObject(line)
	if !ok {
		return Event{}, false
	}
	ev := Event{Kind: KindOther, Raw: string(line)}
	if tok := kimiCompactionRe.FindString(string(line)); tok != "" {
		ev.Kind = KindCompaction
		ev.Message = tok
		return ev, true
	}
	if u := findObject(m, func(o map[string]any) bool {
		_, hasUsed := num(o, "used")
		_, hasSize := num(o, "size")
		return hasUsed && hasSize
	}); u != nil {
		ev.Kind = KindUsage
		ev.ContextTokens = inum(u, "used")
		ev.Window = inum(u, "size")
		ev.Message = "usage_update"
		return ev, true
	}
	if e := findObject(m, func(o map[string]any) bool { _, ok := o["error"]; return ok }); e != nil {
		ev.Kind = KindError
		switch x := e["error"].(type) {
		case string:
			ev.Message = truncate(x, 2000)
		case map[string]any:
			ev.Message = truncate(str(x, "message", "data"), 2000)
		default:
			ev.Message = "error"
		}
		return ev, true
	}
	if t := findObject(m, func(o map[string]any) bool {
		k := str(o, "sessionUpdate", "type", "kind")
		return k == "tool_call" || k == "tool_call_update"
	}); t != nil {
		if str(t, "sessionUpdate", "type", "kind") == "tool_call" {
			ev.Kind = KindTool
			ev.Message = str(t, "title", "name", "toolName", "kind")
		} else {
			ev.Message = "tool_call_update"
		}
		return ev, true
	}
	if c := findObject(m, func(o map[string]any) bool {
		return str(o, "sessionUpdate", "type") == "agent_message_chunk"
	}); c != nil {
		ev.Kind = KindText
		ev.Message = str(sub(c, "content"), "text")
		return ev, true
	}
	ev.Message = str(m, "method", "type")
	return ev, true
}

// FinalUsage implements Adapter: peak context from usage_update; token
// counts only when the stream carries input_tokens and output_tokens
// (unverified on kimi).
func (KimiAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	u := Usage{PeakContext: PeakContext(events)}
	found := u.PeakContext > 0
	for _, line := range strings.Split(string(stdout), "\n") {
		m, ok := parseObject([]byte(line))
		if !ok {
			continue
		}
		if t := findObject(m, func(o map[string]any) bool {
			_, a := num(o, "input_tokens")
			_, b := num(o, "output_tokens")
			return a && b
		}); t != nil {
			found = true
			u.InputTokens = inum(t, "input_tokens")
			u.OutputTokens = inum(t, "output_tokens")
			if d := sub(t, "input_tokens_details"); d != nil {
				u.CachedTokens = inum(d, "cached_tokens")
			} else {
				u.CachedTokens = inum(t, "cached_tokens")
			}
		}
	}
	if !found {
		return Usage{}, ErrNoUsage
	}
	if u.InputTokens == 0 {
		u.InputTokens = u.PeakContext
	}
	return u, nil
}

// ExitClass implements Adapter. kimi copies its tools' stderr onto its own
// (stage 0 finding: the sandbox probe's curl "response 403" read as Auth),
// so only kimi's own failure is classified from stderr: the text from its
// last line starting with "error:" ("error: failed to run prompt: ...",
// then "See log: ..."), and nothing after a clean exit. The stream's
// error events always count.
func (KimiAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	return classifyExit(exitCode, kimiOwnError(exitCode, stderr), events, nil)
}

// kimiOwnError returns kimi's final error text from stderr, or nil.
func kimiOwnError(exitCode int, stderr []byte) []byte {
	if exitCode == 0 {
		return nil
	}
	text := string(stderr)
	i := strings.LastIndex("\n"+text, "\nerror:")
	if i < 0 {
		return nil
	}
	return []byte(text[i:])
}

// InstructionFile implements Adapter.
func (KimiAdapter) InstructionFile() string { return "AGENTS.md" }

// ProjectSettings implements Adapter: kimi has no verified per-project
// switch; the instruction file repeats the network prohibition.
func (KimiAdapter) ProjectSettings(inv Invocation) map[string][]byte { return nil }
