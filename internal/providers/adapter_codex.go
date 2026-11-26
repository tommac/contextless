package providers

import (
	"os"
	"path/filepath"
	"strings"
)

// CodexAdapter drives codex 0.144.1 exec (06 §6.2, research §3). Values
// passed with -c are TOML: strings are quoted, numbers and booleans bare.
//
// Each run gets its own CODEX_HOME in RunDir (CodexHomeDir) holding only a
// symlink to the user's auth.json, so no user config.toml, global
// AGENTS.md, skills, plugins, MCP servers or notify hook reach the packet,
// codex's state databases and PATH-alias directory live in RunDir, and a
// sandboxed run needs nothing from ~/.codex but auth.json (stage 0
// findings, 2026-10-22).
type CodexAdapter struct {
	// Binary is the executable; "codex" when empty.
	Binary string
}

// CodexHomeDir is the RunDir entry used as CODEX_HOME.
const CodexHomeDir = "codex-home"

// codexProviderArgs define the model provider inline, since the per-run
// home has no config.toml. They mirror the openai_inline provider in the
// user's config: with codex's built-in provider gpt-6.1-sol is "not
// supported when using Codex with a ChatGPT account" (stage 0 finding).
var codexProviderArgs = []string{
	"-c", `model_provider="openai_inline"`,
	"-c", `model_providers.openai_inline.name="openai_inline"`,
	"-c", "model_providers.openai_inline.requires_openai_auth=true",
	"-c", "model_providers.openai_inline.supports_websockets=true",
}

// CodexDisabledFeatures are tool surfaces a packet never needs and that
// reach outside the workspace: ChatGPT connectors and plugins (Gmail,
// Stripe, GitHub, Calendar) act through chatgpt.com, which the sandbox
// proxy must admit for the model API; browser and computer use; image
// generation; plugin suggestions; chronicle. Stage 0 measured codex's
// context per model call falling from about 182,000 to about 6,700 tokens
// with these off.
var CodexDisabledFeatures = []string{"apps", "plugins", "remote_plugin", "browser_use", "browser_use_external",
	"in_app_browser", "computer_use", "image_generation", "tool_suggest", "chronicle"}

// CodexAuthFile is the user's codex login: $CODEX_HOME/auth.json, else
// ~/.codex/auth.json.
func CodexAuthFile() string {
	if h := os.Getenv("CODEX_HOME"); h != "" {
		return filepath.Join(h, "auth.json")
	}
	return ExpandHome("~/.codex/auth.json")
}

// PrepareRun creates the per-run CODEX_HOME with a symlink to the user's
// auth.json; a token refresh writes through it to the real file.
func (CodexAdapter) PrepareRun(inv Invocation) error {
	home := filepath.Join(inv.RunDir, CodexHomeDir)
	if err := os.MkdirAll(home, 0o700); err != nil {
		return err
	}
	link := filepath.Join(home, "auth.json")
	if _, err := os.Lstat(link); err == nil {
		return nil
	}
	real := CodexAuthFile()
	if _, err := os.Stat(real); err != nil {
		return nil // not logged in: the run reports auth_required
	}
	return os.Symlink(real, link)
}

// FinishRun carries a refreshed login back when codex replaced the symlink
// with a file instead of writing through it: refresh tokens rotate, so a
// refresh left only in RunDir would log the user out.
func (CodexAdapter) FinishRun(inv Invocation) {
	link := filepath.Join(inv.RunDir, CodexHomeDir, "auth.json")
	st, err := os.Lstat(link)
	if err != nil || st.Mode()&os.ModeSymlink != 0 || !st.Mode().IsRegular() {
		return
	}
	real := CodexAuthFile()
	if rst, err := os.Stat(real); err == nil && !st.ModTime().After(rst.ModTime()) {
		return
	}
	b, err := os.ReadFile(link)
	if err != nil || len(b) == 0 {
		return
	}
	tmp := real + ".bookwriter-tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return
	}
	if err := os.Rename(tmp, real); err != nil {
		os.Remove(tmp)
	}
}

// Provider implements Adapter.
func (CodexAdapter) Provider() string { return "codex" }

func (a CodexAdapter) bin() string {
	if a.Binary != "" {
		return a.Binary
	}
	return "codex"
}

// BuildCommand implements Adapter: exec with the workspace as root, the
// workspace-write sandbox (danger-full-access when Run wraps codex in the
// seatbelt profile: seatbelt cannot nest, and the outer profile then
// confines codex and every tool it runs), never-ask approvals, model and
// effort, the compaction limit at the ceiling, web search off unless
// Network, memories off, the inline provider, CodexDisabledFeatures off,
// the strict schema and the handoff as the last message, JSONL events,
// brief on stdin, CODEX_HOME in RunDir.
func (a CodexAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	brief, err := readBrief(inv)
	if err != nil {
		return nil, nil, nil, err
	}
	p := inv.Profile
	mode := "workspace-write"
	if inv.Sandboxed {
		mode = "danger-full-access"
	}
	argv := []string{a.bin(), "exec", "-C", inv.Workspace, "--skip-git-repo-check", "--ephemeral",
		"--sandbox", mode, "-c", `approval_policy="never"`,
		"-m", p.Model}
	if p.Effort != "" {
		argv = append(argv, "-c", `model_reasoning_effort="`+p.Effort+`"`)
	}
	argv = append(argv, "-c", "model_auto_compact_token_limit="+itoa(p.Ceiling()))
	if inv.Network {
		argv = append(argv, "-c", `web_search="live"`)
	} else {
		argv = append(argv, "-c", `web_search="disabled"`)
	}
	argv = append(argv, "-c", "features.memories=false")
	extra := map[string]string{}
	if inv.RunDir != "" {
		argv = append(argv, codexProviderArgs...)
		extra["CODEX_HOME"] = filepath.Join(inv.RunDir, CodexHomeDir)
	}
	for _, f := range CodexDisabledFeatures {
		argv = append(argv, "--disable", f)
	}
	// Only the strict form: the plain handoff schema carries the partial
	// rule as a top-level allOf, which the API refuses with a 400. An empty
	// StrictSchemaPath means the contract has no strict form and codex runs
	// without --output-schema.
	if inv.StrictSchemaPath != "" {
		argv = append(argv, "--output-schema", inv.StrictSchemaPath)
	}
	argv = append(argv, "--output-last-message", inv.handoffPath(), "--json", "-")
	return argv, baseEnv(inv, extra), brief, nil
}

// ParseEvent implements Adapter for codex --json lines: thread.started,
// turn.started, item.started/updated/completed with item types
// (agent_message, reasoning, command_execution, file_change, mcp_tool_call,
// web_search, todo_list), turn.completed with usage, turn.failed, error.
// Any type or item type containing "compact" is the compaction signal.
func (CodexAdapter) ParseEvent(line []byte) (Event, bool) {
	m, ok := parseObject(line)
	if !ok {
		return Event{}, false
	}
	ev := Event{Kind: KindOther, Raw: string(line)}
	typ := str(m, "type")
	switch {
	case typ == "turn.completed":
		// turn.completed usage is cumulative over every model call in the
		// run (stage 0 finding), not a context size, so it carries no
		// ContextTokens; FinalUsage estimates the per-call peak.
		ev.Kind = KindUsage
		ev.Message = "turn.completed"
	case typ == "turn.failed" || typ == "error":
		ev.Kind = KindError
		msg := str(m, "message")
		if e := sub(m, "error"); e != nil {
			msg = str(e, "message")
		}
		ev.Message = truncate(typ+": "+msg, 2000)
	case strings.HasPrefix(typ, "item."):
		item := sub(m, "item")
		it := str(item, "type")
		ev.Message = it
		switch {
		case strings.Contains(it, "compact"):
			ev.Kind = KindCompaction
		case it == "command_execution" && typ == "item.started":
			ev.Kind = KindTool
			ev.Message = "command_execution: " + truncate(str(item, "command"), 300)
		case it == "mcp_tool_call" && typ == "item.started":
			ev.Kind = KindTool
			ev.Message = "mcp_tool_call: " + str(item, "server") + "." + str(item, "tool")
		case it == "web_search" && typ == "item.started":
			ev.Kind = KindTool
			ev.Message = "web_search: " + truncate(str(item, "query"), 300)
		case it == "file_change" && typ == "item.completed":
			ev.Kind = KindTool
			ev.Message = "file_change"
		case it == "agent_message" && typ == "item.completed":
			ev.Kind = KindText
			ev.Message = str(item, "text")
		case it == "error":
			ev.Kind = KindError
			ev.Message = truncate(str(item, "message"), 2000)
		}
	case strings.Contains(typ, "compact"):
		ev.Kind = KindCompaction
		ev.Message = typ
	default:
		ev.Message = typ
	}
	return ev, true
}

// FinalUsage implements Adapter: the sum of every turn.completed usage; no
// USD on the subscription.
func (CodexAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	u := Usage{PeakContext: PeakContext(events)}
	found := false
	calls := 1 // model calls: one plus one per tool round trip
	for _, line := range strings.Split(string(stdout), "\n") {
		m, ok := parseObject([]byte(line))
		if !ok {
			continue
		}
		if typ := str(m, "type"); typ == "item.started" || typ == "item.completed" {
			it := str(sub(m, "item"), "type")
			if (typ == "item.started" && (it == "command_execution" || it == "mcp_tool_call" || it == "web_search")) ||
				(typ == "item.completed" && it == "file_change") {
				calls++
			}
			continue
		}
		if str(m, "type") != "turn.completed" {
			continue
		}
		usage := sub(m, "usage")
		if usage == nil {
			continue
		}
		found = true
		u.InputTokens += inum(usage, "input_tokens")
		u.CachedTokens += inum(usage, "cached_input_tokens")
		u.CacheWriteTokens += inum(usage, "cache_write_input_tokens")
		u.OutputTokens += inum(usage, "output_tokens")
		u.ReasoningTokens += inum(usage, "reasoning_output_tokens")
	}
	if !found {
		return Usage{}, ErrNoUsage
	}
	// Per-call estimate: the context grows from the first call to the last,
	// so the last (peak) call is at most twice the average; never more than
	// the cumulative total. Turns counts model calls so smoke's overhead
	// measurement (input / turns) is a per-call figure.
	u.Turns = calls
	avg := u.InputTokens / calls
	peak := 2 * avg
	if calls == 1 || peak > u.InputTokens {
		peak = u.InputTokens
	}
	if est := peak + u.OutputTokens/calls; est > u.PeakContext {
		u.PeakContext = est
	}
	return u, nil
}

// ExitClass implements Adapter.
func (CodexAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	return classifyExit(exitCode, stderr, events, nil)
}

// InstructionFile implements Adapter.
func (CodexAdapter) InstructionFile() string { return "AGENTS.md" }

// ProjectSettings implements Adapter: codex takes everything through -c.
func (CodexAdapter) ProjectSettings(inv Invocation) map[string][]byte { return nil }
