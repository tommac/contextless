package providers

import (
	"bytes"
	"encoding/json"
	"strings"
)

// GrokAdapter drives grok headless (06 §6.2, research §3) with
// --output-format streaming-json: one JSON line per event, a "usage" line
// per model call, and a closing "end" line with the run's totals, turn
// count and cost. The end-only json format gives totals alone; read as the
// context peak they tainted six of seven grok 1.0.70 scorecards
// (2026-10-30: a 13-call card summed to 911,776 tokens on a 256k window).
// The seeded profile is disabled until `grok login` and the smoke pass.
type GrokAdapter struct {
	// Binary is the executable; "grok" when empty.
	Binary string
}

// DefaultGrokMaxTurns is the template's --max-turns when MaxTurns is 0.
const DefaultGrokMaxTurns = 60

// Provider implements Adapter.
func (GrokAdapter) Provider() string { return "grok" }

func (a GrokAdapter) bin() string {
	if a.Binary != "" {
		return a.Binary
	}
	return "grok"
}

// BuildCommand implements Adapter.
func (a GrokAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	if _, err := readBrief(inv); err != nil {
		return nil, nil, nil, err
	}
	p := inv.Profile
	argv := []string{a.bin(), "--prompt-file", inv.BriefPath, "--cwd", inv.Workspace, "--always-approve", "-m", p.Model}
	if p.Effort != "" {
		argv = append(argv, "--effort", p.Effort)
	}
	turns := inv.MaxTurns
	if turns <= 0 {
		turns = DefaultGrokMaxTurns
	}
	argv = append(argv, "--max-turns", itoa(turns))
	if !inv.Network {
		argv = append(argv, "--disable-web-search")
	}
	if inv.SchemaPath != "" {
		schema, err := readSchema(inv.SchemaPath)
		if err != nil {
			return nil, nil, nil, err
		}
		argv = append(argv, "--json-schema", schema)
	}
	// --no-memory: no cross-session memory, so Tom's other grok sessions do
	// not reach a run's packet.
	argv = append(argv, "--no-memory", "--output-format", "streaming-json")
	env := baseEnv(inv, map[string]string{"GROK_AUTO_COMPACT_THRESHOLD_PERCENT": "100"})
	return argv, env, nil, nil
}

// ParseEvent implements Adapter: only whole-line objects are recognized; a
// line mentioning auto_compact_started is the compaction signal. A "usage"
// line is one model call, and its prompt plus output is the context that
// call held; the "end" line (or the json format's single document) is the
// result.
func (GrokAdapter) ParseEvent(line []byte) (Event, bool) {
	if bytes.Contains(line, []byte("auto_compact_started")) || bytes.Contains(line, []byte("compaction_started")) {
		return Event{Kind: KindCompaction, Message: "auto_compact_started", Raw: string(line)}, true
	}
	m, ok := parseObject(line)
	if !ok {
		return Event{}, false
	}
	ev := Event{Kind: KindOther, Raw: string(line)}
	typ := str(m, "type")
	switch {
	case typ == "usage":
		ev.Kind = KindUsage
		ev.Message = "call"
		ev.ContextTokens = grokContext(sub(m, "usage"))
	case typ == "end" || m["usage"] != nil || m["total_cost_usd"] != nil:
		ev.Kind = KindUsage
		ev.Message = "result"
		if isErr, _ := m["is_error"].(bool); isErr {
			ev.Kind = KindError
			ev.Message = truncate(str(m, "result", "error"), 2000)
		}
		if stop := str(m, "stopReason", "stop_reason"); strings.Contains(stop, "max_turns") {
			ev.Kind = KindError
			ev.Message = "max_turns: the run stopped at its turn limit"
		}
	case typ == "error" || m["error"] != nil:
		ev.Kind = KindError
		msg := str(m, "error", "message")
		if e := sub(m, "error"); e != nil {
			msg = str(e, "message")
		}
		ev.Message = truncate(msg, 2000)
	default:
		ev.Message = typ
	}
	return ev, true
}

// grokContext is the context one model call held: its uncached and cached
// prompt plus its output.
func grokContext(usage map[string]any) int {
	if usage == nil {
		return 0
	}
	return inum(usage, "input_tokens") + inum(usage, "cache_read_input_tokens") +
		inum(usage, "cache_creation_input_tokens") + inum(usage, "output_tokens")
}

// grokResult finds the result document: the whole stdout as one JSON
// object (the json format), else the "end" line, else the last line with
// run totals that is not a single call's usage.
func grokResult(stdout []byte) map[string]any {
	var m map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout), &m); err == nil && m != nil {
		return m
	}
	if end := lastObjectLine(stdout, func(m map[string]any) bool { return str(m, "type") == "end" }); end != nil {
		return end
	}
	return lastObjectLine(stdout, func(m map[string]any) bool {
		return str(m, "type") != "usage" && (m["usage"] != nil || m["total_cost_usd"] != nil)
	})
}

// FinalUsage implements Adapter: usage.input_tokens is uncached only, so
// InputTokens adds the cache fields (research §3).
func (GrokAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	res := grokResult(stdout)
	if res == nil {
		return Usage{}, ErrNoUsage
	}
	usage := sub(res, "usage")
	if usage == nil {
		return Usage{}, ErrNoUsage
	}
	read := inum(usage, "cache_read_input_tokens")
	u := Usage{
		InputTokens:      inum(usage, "input_tokens") + read + inum(usage, "cache_creation_input_tokens"),
		CachedTokens:     read,
		CacheWriteTokens: inum(usage, "cache_creation_input_tokens"),
		OutputTokens:     inum(usage, "output_tokens"),
		ReasoningTokens:  inum(usage, "reasoning_tokens"),
		Turns:            inum(res, "num_turns"),
		PeakContext:      PeakContext(events),
	}
	if u.PeakContext == 0 {
		// No per-call usage line (the end-only json format): estimate the
		// per-call peak as codex does, twice the average call's prompt plus
		// its output, never more than the totals. The totals themselves are
		// a sum over calls, not a context size.
		calls := u.Turns
		if calls < 1 {
			calls = 1
		}
		peak := 2 * (u.InputTokens / calls)
		if calls == 1 || peak > u.InputTokens {
			peak = u.InputTokens
		}
		u.PeakContext = peak + u.OutputTokens/calls
	}
	if usd, ok := num(res, "total_cost_usd"); ok {
		u.USDReported = usd
		u.HasUSD = true
	}
	return u, nil
}

// ExitClass implements Adapter.
func (GrokAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	return classifyExit(exitCode, stderr, events, func(ev Event) bool {
		return ev.Kind == KindError && strings.Contains(ev.Message, "max_turns")
	})
}

// InstructionFile implements Adapter: unverified; AGENTS.md is the
// cross-tool convention.
func (GrokAdapter) InstructionFile() string { return "AGENTS.md" }

// ProjectSettings implements Adapter.
func (GrokAdapter) ProjectSettings(inv Invocation) map[string][]byte { return nil }
