package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeAdapter drives claude 2.1.141 headless (06 §6.2, research §3).
//
// Deviation from the 06 template, recorded in the task notes: --verbose is
// passed because print mode refuses --output-format stream-json without it
// on every claude build the author has seen; the smoke verifies it.
// --max-turns is passed only when MaxTurns is set, since the flag is
// unverified on 2.1.141.
type ClaudeAdapter struct {
	// Binary is the executable; "claude" when empty.
	Binary string
}

// Provider implements Adapter.
func (ClaudeAdapter) Provider() string { return "claude" }

func (a ClaudeAdapter) bin() string {
	if a.Binary != "" {
		return a.Binary
	}
	return "claude"
}

// BuildCommand implements Adapter: stdin brief, stream-json, no session
// persistence, project settings only, no web tools unless Network, an
// optional inline schema, and the compaction switches in the environment.
func (a ClaudeAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	brief, err := readBrief(inv)
	if err != nil {
		return nil, nil, nil, err
	}
	p := inv.Profile
	argv := []string{a.bin(), "-p", "--permission-mode", "bypassPermissions", "--model", p.Model}
	if p.Effort != "" {
		argv = append(argv, "--effort", p.Effort)
	}
	argv = append(argv, "--output-format", "stream-json", "--verbose", "--no-session-persistence")
	if inv.MaxBudgetUSD > 0 {
		argv = append(argv, "--max-budget-usd", ftoa(inv.MaxBudgetUSD))
	}
	argv = append(argv, "--setting-sources", "project")
	if !inv.Network {
		argv = append(argv, "--disallowedTools", "WebSearch WebFetch")
	}
	switch {
	case inv.OneShot && inv.SchemaPath != "":
		// the brief is the whole packet and the enforced schema's
		// StructuredOutput tool carries the answer (Run adopts it as the
		// handoff), so the call gets no tools at all: a Write of the
		// handoff only cost a turn, since Write refuses a file the agent
		// has not read. Verified 2026-11-06: --tools "" keeps
		// StructuredOutput.
		argv = append(argv, "--tools", "")
	case inv.OneShot:
		// no schema to answer through: one Write for the handoff, no
		// reads, no shell
		argv = append(argv, "--tools", "Write")
	}
	if inv.MaxTurns > 0 {
		argv = append(argv, "--max-turns", itoa(inv.MaxTurns))
	}
	if inv.SchemaPath != "" {
		schema, err := readSchema(inv.SchemaPath)
		if err != nil {
			return nil, nil, nil, err
		}
		argv = append(argv, "--json-schema", claudeJSONSchema(schema))
	}
	// CLAUDE_CODE_MAX_OUTPUT_TOKENS lifts the per-response output cap (64k
	// on Opus, 32k on Sonnet by default) for a handoff that must be
	// written whole, such as a 25-chapter structure.v1; the CLI caps the
	// value to each model's own limit.
	env := baseEnv(inv, map[string]string{
		"DISABLE_COMPACT":                "1",
		"DISABLE_AUTO_COMPACT":           "1",
		"CLAUDE_CODE_MAX_CONTEXT_TOKENS": itoa(p.Ceiling()),
		"CLAUDE_CODE_MAX_OUTPUT_TOKENS":  "128000",
	})
	return argv, env, brief, nil
}

// claudeContext sums a message usage block into a context measurement.
func claudeContext(usage map[string]any) int {
	if usage == nil {
		return 0
	}
	return inum(usage, "input_tokens") + inum(usage, "cache_creation_input_tokens") +
		inum(usage, "cache_read_input_tokens") + inum(usage, "output_tokens")
}

// ParseEvent implements Adapter for claude stream-json lines: system
// (compact_boundary is the compaction signal), assistant (per-turn usage,
// tool_use blocks, text), user (tool results), result, rate_limit_event.
func (ClaudeAdapter) ParseEvent(line []byte) (Event, bool) {
	m, ok := parseObject(line)
	if !ok {
		return Event{}, false
	}
	ev := Event{Kind: KindOther, Raw: string(line)}
	typ := str(m, "type")
	switch typ {
	case "system":
		subtype := str(m, "subtype")
		ev.Message = subtype
		if strings.Contains(subtype, "compact") {
			ev.Kind = KindCompaction
		}
	case "assistant":
		msg := sub(m, "message")
		ev.ContextTokens = claudeContext(sub(msg, "usage"))
		ev.Kind = KindUsage
		var tools, texts []string
		for _, c := range list(msg, "content") {
			block, _ := c.(map[string]any)
			switch str(block, "type") {
			case "tool_use":
				tools = append(tools, str(block, "name"))
			case "text":
				texts = append(texts, str(block, "text"))
			}
		}
		text := strings.Join(texts, "\n")
		switch {
		case strings.HasPrefix(strings.TrimSpace(text), "API Error"):
			ev.Kind = KindError
			ev.Message = truncate(text, 2000)
		case len(tools) > 0:
			ev.Kind = KindTool
			ev.Message = strings.Join(tools, ",")
		case text != "":
			ev.Kind = KindText
			ev.Message = text
		}
	case "user":
		ev.Message = "tool_result"
	case "result":
		subtype := str(m, "subtype")
		isErr, _ := m["is_error"].(bool)
		if subtype == "success" && !isErr {
			ev.Kind = KindUsage
			ev.Message = "result:success"
		} else {
			ev.Kind = KindError
			ev.Message = truncate(subtype+": "+str(m, "result", "error"), 2000)
		}
	case "rate_limit_event":
		info := sub(m, "rate_limit_info")
		status := str(info, "status")
		ev.Message = "rate_limit:" + status + ":" + str(info, "rateLimitType", "rate_limit_type")
		if status == "rejected" || status == "blocked" {
			ev.Kind = KindRateLimit
		}
		if resets, ok := num(info, "resetsAt", "resets_at"); ok {
			if d := int(resets) - int(time.Now().Unix()); d > 0 {
				ev.RetryAfterSeconds = d
			}
		}
	default:
		if strings.Contains(typ, "compact") {
			ev.Kind = KindCompaction
			ev.Message = typ
		} else {
			ev.Message = typ
		}
	}
	return ev, true
}

// FinalUsage implements Adapter: the result line carries the totals, the
// per-turn assistant usage gives the peak context.
func (ClaudeAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	u := Usage{PeakContext: PeakContext(events)}
	res := lastObjectLine(stdout, func(m map[string]any) bool { return str(m, "type") == "result" })
	if res == nil {
		if u.PeakContext == 0 {
			return Usage{}, ErrNoUsage
		}
		u.InputTokens = u.PeakContext
		return u, nil
	}
	usage := sub(res, "usage")
	read := inum(usage, "cache_read_input_tokens")
	u.InputTokens = inum(usage, "input_tokens") + read + inum(usage, "cache_creation_input_tokens")
	u.CachedTokens = read
	u.CacheWriteTokens = inum(usage, "cache_creation_input_tokens")
	u.OutputTokens = inum(usage, "output_tokens")
	u.Turns = inum(res, "num_turns")
	u.WallMs = int64(inum(res, "duration_ms"))
	if usd, ok := num(res, "total_cost_usd"); ok {
		u.USDReported = usd
		u.HasUSD = true
	}
	return u, nil
}

// HandoffFromOutput implements handoffSource: the structured_output object
// of the last result line, which the enforced --json-schema makes the CLI
// return whole.
func (ClaudeAdapter) HandoffFromOutput(stdout []byte) ([]byte, bool) {
	res := lastObjectLine(stdout, func(m map[string]any) bool { return str(m, "type") == "result" })
	if res == nil {
		return nil, false
	}
	so, ok := res["structured_output"].(map[string]any)
	if !ok || len(so) == 0 {
		return nil, false
	}
	return jsonMarshal(so), true
}

// claudeCapSubtypes are result subtypes that mean a clean stop on a cap.
var claudeCapSubtypes = []string{"error_max_turns", "error_max_budget_usd", "error_max_structured_output_retries"}

// ExitClass implements Adapter.
func (ClaudeAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	return classifyExit(exitCode, stderr, events, func(ev Event) bool {
		if ev.Kind != KindError {
			return false
		}
		for _, s := range claudeCapSubtypes {
			if strings.HasPrefix(ev.Message, s) {
				return true
			}
		}
		return false
	})
}

// InstructionFile implements Adapter.
func (ClaudeAdapter) InstructionFile() string { return "CLAUDE.md" }

// claudeJSONSchema is the schema given to --json-schema: the handoff
// schema without its top-level "$schema" and its top-level allOf, anyOf
// and oneOf. Claude Code 2.1.308 validates the schema against the declared
// meta-schema and does not know draft 2020-12 ("no schema with key or ref
// https://json-schema.org/draft/2020-12/schema"), and it delivers the
// schema as a tool input_schema, which the API refuses with a top-level
// combinator ("input_schema does not support oneOf, allOf, or anyOf at
// the top level"); both verified 2026-10-30, and 2.1.141 accepted the full
// schema. The dropped combinators carry conditional rules (a partial
// result's remainder, and the like) that collect still enforces when it
// validates the handoff against the full schema, so the CLI's structured
// output checks the shape and the worker checks the rest. A schema that
// does not parse as an object is passed unchanged.
func claudeJSONSchema(schema string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(schema), &m); err != nil || m == nil {
		return schema
	}
	changed := false
	for _, k := range []string{"$schema", "allOf", "anyOf", "oneOf"} {
		if _, ok := m[k]; ok {
			delete(m, k)
			changed = true
		}
	}
	if !changed {
		return schema
	}
	b, err := json.Marshal(m)
	if err != nil {
		return schema
	}
	return string(b)
}

// ProjectSettings implements Adapter: with --setting-sources project the
// workspace settings are the only settings, and they repeat the compaction
// switch (autoCompactEnabled is verified in the binary).
func (ClaudeAdapter) ProjectSettings(inv Invocation) map[string][]byte {
	b, _ := json.MarshalIndent(map[string]any{"autoCompactEnabled": false}, "", "  ")
	return map[string][]byte{".claude/settings.json": append(b, '\n')}
}

// SandboxWriteDirs implements sandboxWriteHinter: claude's Bash tool creates
// /tmp/claude-<uid>/<workspace path with every non-alphanumeric byte as
// "-"> before each command and fails with EPERM when it cannot (verified
// 2026-10-23). Only that one directory is allowed, never /tmp/claude-<uid>
// itself, which holds other claude sessions' scratch files.
func (ClaudeAdapter) SandboxWriteDirs(inv Invocation) []string {
	if inv.Workspace == "" {
		return nil
	}
	slug := ClaudeTempSlug(inv.Workspace)
	dir := fmt.Sprintf("claude-%d", os.Getuid())
	// The Bash tool also records the shell's working directory after each
	// command in /tmp/claude-<hex>-cwd; without it every command exits 1.
	return []string{filepath.Join("/private/tmp", dir, slug), filepath.Join("/tmp", dir, slug),
		ClaudeCwdFilePattern}
}

// ClaudeCwdFilePattern matches the Bash tool's working-directory files and
// nothing else under /tmp.
const ClaudeCwdFilePattern = `re:^/(private/)?tmp/claude-[0-9a-f]+-cwd$`

// ClaudeTempDirPattern matches the per-user directory /tmp/claude-<uid>
// itself and nothing inside it: claude opens it at startup.
const ClaudeTempDirPattern = `re:^/(private/)?tmp/claude-[0-9]+$`

// ClaudeTempSlug is the directory name claude derives from a working
// directory: every byte that is not an ASCII letter or digit becomes "-".
func ClaudeTempSlug(path string) string {
	b := []byte(path)
	for i, c := range b {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			b[i] = '-'
		}
	}
	return string(b)
}
