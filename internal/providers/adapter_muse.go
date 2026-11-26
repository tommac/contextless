package providers

import (
	"bytes"
	"strings"
)

// MuseAdapter drives muse 1.4.11 exec (06 §6.2, research §3): explicit
// compaction thresholds at 1.0, the provider context limit at the ceiling,
// a step cap, web tools off unless Network, the schema file, JSONL events.
// The usage event shape and the window are unverified; the seeded profile
// is disabled until `muse login` and the smoke pass.
type MuseAdapter struct {
	// Binary is the executable; "muse" when empty.
	Binary string
}

// OneShotMuseMaxSteps bounds a one-shot call (Invocation.OneShot): the
// answer, the handoff write and one retry.
const OneShotMuseMaxSteps = 3

// DefaultMuseMaxSteps is the template's --max-model-steps when MaxTurns is 0.
const DefaultMuseMaxSteps = 200

// Provider implements Adapter.
func (MuseAdapter) Provider() string { return "muse" }

func (a MuseAdapter) bin() string {
	if a.Binary != "" {
		return a.Binary
	}
	return "muse"
}

// BuildCommand implements Adapter.
func (a MuseAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	if _, err := readBrief(inv); err != nil {
		return nil, nil, nil, err
	}
	p := inv.Profile
	argv := []string{a.bin(), "exec", "--prompt-file", inv.BriefPath, "--yolo", "--workspace", inv.Workspace, "--model", p.Model}
	if p.Effort != "" {
		argv = append(argv, "--reasoning-effort", p.Effort)
	}
	// Muse Code 1.4.14 names -c keys without the "settings." prefix
	// ("-c settings.context_compaction.provider_context_limit_tokens:
	// refusing run: unknown key"; the bare key is accepted, 2026-10-30), and
	// it requires each threshold strictly between 0 and 1
	// ("--context-compaction-soft-threshold must be a finite fraction in
	// (0, 1)"; 1.0 was accepted up to 1.4.11). The limit below is the
	// ceiling, above the kill line, so 0.998 and 0.999 of it still never
	// compact before the worker kills the run.
	argv = append(argv,
		"--context-compaction-soft-threshold", "0.998",
		"--context-compaction-hard-threshold", "0.999",
		"-c", "context_compaction.provider_context_limit_tokens="+itoa(p.Ceiling()))
	steps := inv.MaxTurns
	if steps <= 0 {
		steps = DefaultMuseMaxSteps
		if inv.OneShot {
			steps = OneShotMuseMaxSteps
		}
	}
	argv = append(argv, "--max-model-steps", itoa(steps))
	if !inv.Network {
		argv = append(argv, "--disable-web-tools")
	}
	// Only the strict form, as for codex: muse's structured outputs refuse
	// the plain schema (an if keyword, optional properties); an empty
	// StrictSchemaPath means the contract has no strict form and muse runs
	// without --output-schema, writing the handoff file the brief asks for.
	if inv.StrictSchemaPath != "" {
		argv = append(argv, "--output-schema", inv.StrictSchemaPath)
	}
	// No session log (the run's events come on stdout; muse's session lock
	// directory is outside the sandbox) and no foreign personal context
	// (muse otherwise loads Tom's Codex personal rules and skills into the
	// run, verified 2026-10-30).
	argv = append(argv, "--no-session-log", "--no-foreign-personal-context", "--json")
	return argv, baseEnv(inv, nil), nil, nil
}

// museUsage finds a usage block in an event: the object itself or its
// "usage" member when it carries input_tokens or output_tokens.
func museUsage(m map[string]any) map[string]any {
	for _, c := range []map[string]any{sub(m, "usage"), m} {
		if c == nil {
			continue
		}
		_, a := num(c, "input_tokens")
		_, b := num(c, "output_tokens")
		if a || b {
			return c
		}
	}
	return nil
}

// ParseEvent implements Adapter for muse --json lines: run_start,
// assistant_message, tool_call, tool_result, compaction, plus any event
// carrying usage keys (input_tokens, output_tokens, cached_tokens,
// reasoning_tokens, cost_micros).
func (MuseAdapter) ParseEvent(line []byte) (Event, bool) {
	m, ok := parseObject(line)
	if !ok {
		return Event{}, false
	}
	ev := Event{Kind: KindOther, Raw: string(line)}
	typ := str(m, "type", "event")
	switch {
	case strings.Contains(typ, "compact"):
		ev.Kind = KindCompaction
		ev.Message = typ
	case typ == "assistant_message":
		ev.Kind = KindText
		ev.Message = str(m, "text", "content")
	case typ == "tool_call":
		ev.Kind = KindTool
		ev.Message = str(m, "name", "tool", "tool_name")
	case typ == "error" || m["error"] != nil:
		ev.Kind = KindError
		msg := str(m, "message", "error")
		if e := sub(m, "error"); e != nil {
			msg = str(e, "message")
		}
		ev.Message = truncate(msg, 2000)
	default:
		ev.Message = typ
	}
	if u := museUsage(m); u != nil {
		ev.ContextTokens = inum(u, "input_tokens") + inum(u, "output_tokens")
		if ev.Kind == KindOther {
			ev.Kind = KindUsage
			ev.Message = "usage"
		}
	}
	return ev, true
}

// FinalUsage implements Adapter: usage blocks are summed as per-step
// usage (shape unverified); cost_micros becomes USD.
func (MuseAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	u := Usage{PeakContext: PeakContext(events)}
	found := false
	var micros float64
	for _, line := range strings.Split(string(stdout), "\n") {
		m, ok := parseObject([]byte(line))
		if !ok {
			continue
		}
		b := museUsage(m)
		if b == nil {
			continue
		}
		found = true
		u.InputTokens += inum(b, "input_tokens")
		u.CachedTokens += inum(b, "cached_tokens")
		u.OutputTokens += inum(b, "output_tokens")
		u.ReasoningTokens += inum(b, "reasoning_tokens")
		u.Turns++
		if c, ok := num(b, "cost_micros"); ok {
			micros += c
			u.HasUSD = true
		}
	}
	if !found {
		if eu, ok := museEnvelopeUsage(stdout); ok {
			eu.PeakContext = u.PeakContext
			return eu, nil
		}
		return Usage{}, ErrNoUsage
	}
	u.USDReported = micros / 1e6
	return u, nil
}

// musePayload returns the payload of a muse event-envelope line (the
// stream format since muse 1.3: schema_version, record_type, payload with
// its own kind), or nil for any other line.
func musePayload(line []byte) map[string]any {
	m, ok := parseObject(line)
	if !ok {
		return nil
	}
	return sub(m, "payload")
}

// museEnvelopeUsage reads the run_usage payload of the envelope format:
// one record listing tokens per model, calls as turns. The last record
// wins, since each one is cumulative for the run.
func museEnvelopeUsage(stdout []byte) (Usage, bool) {
	var u Usage
	found := false
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		p := musePayload(line)
		if p == nil || str(p, "kind") != "run_usage" {
			continue
		}
		found = true
		u = Usage{Turns: inum(p, "calls")}
		for _, mm := range list(p, "models") {
			b, _ := mm.(map[string]any)
			u.InputTokens += inum(b, "input_tokens")
			u.CachedTokens += inum(b, "cached_tokens")
			u.OutputTokens += inum(b, "output_tokens")
			u.ReasoningTokens += inum(b, "reasoning_tokens")
		}
	}
	return u, found
}

// HandoffFromOutput implements handoffSource: a one-shot muse call answers
// with the handoff as its final text (the run_terminal payload of a
// completed run) and often makes no tool call, so the file is never
// written. The text is adopted when it is one JSON object, bare or inside
// a code fence.
func (MuseAdapter) HandoffFromOutput(stdout []byte) ([]byte, bool) {
	var text string
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		p := musePayload(line)
		if p != nil && str(p, "kind") == "run_terminal" && str(p, "terminal") == "completed" {
			text = str(p, "text")
		}
	}
	return jsonObjectText(text)
}

// ExitClass implements Adapter.
func (MuseAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	return classifyExit(exitCode, stderr, events, func(ev Event) bool {
		return ev.Kind == KindError && strings.Contains(ev.Message, "max_model_steps")
	})
}

// InstructionFile implements Adapter: whether muse reads MUSE.md is a smoke
// item; AGENTS.md is the cross-tool convention until then.
func (MuseAdapter) InstructionFile() string { return "AGENTS.md" }

// ProjectSettings implements Adapter.
func (MuseAdapter) ProjectSettings(inv Invocation) map[string][]byte { return nil }
