package providers

import (
	"encoding/json"
	"strings"
)

// AgyAdapter drives Google's Antigravity CLI (agy 1.5.6, 2026-11-11): the
// terminal client Google's consumer plans (AI Pro, Ultra) cover since the
// open-source Gemini CLI stopped taking plan sign-ins on 2026-06-18. The
// gemini adapter bills a GEMINI_API_KEY per token; agy signs in with Tom's
// Google account (authMethod=consumer in its log) and draws on the plan's
// quota, with no per-call charge. The child environment carries no API key
// (parentEnvAllowlist), so a run cannot fall back to key billing.
//
// Measured on 2026-11-11:
//   - The prompt goes on stdin as one stream-json message,
//     {"event":"user","message":{"role":"user","content":...}}, with
//     --print "" (an empty argument, so -p does not take the next flag as
//     its prompt). A judge brief of several hundred KB would not fit argv.
//   - stream-json output: an "init" event, "step_update" events (each
//     agent_response step carries its own usage; tool steps name the tool),
//     and a closing "result" event with status SUCCESS or ERROR, the final
//     text as "response", an "error" string, and the run's totals.
//   - usage.input_tokens is uncached only; cache_read_tokens is separate;
//     thinking_tokens is the reasoning. num_turns counts user turns (always
//     1 in print mode), so Turns counts agent_response steps instead.
//   - --json-schema is not used: Gemini rejects the handoff schema
//     ("enum[4]: cannot be empty" for the enums that allow ""). As for
//     muse, the brief asks for the handoff as the final message and Run
//     adopts it (HandoffFromOutput) when the file is missing.
//   - Model ids carry the effort tier (gemini-4-argon-high), so profiles
//     have no effort and --effort is never passed.
//   - Web search runs server side through the same host as the model, so
//     the sandbox's host allowlist cannot block it, and neither
//     --settings {"disabledTools"} nor a permissions deny rule turned it off.
//     The brief's network sentence is the guard, as for kimi.
type AgyAdapter struct {
	// Binary is the executable; "agy" when empty.
	Binary string
}

// Provider implements Adapter.
func (AgyAdapter) Provider() string { return "agy" }

func (a AgyAdapter) bin() string {
	if a.Binary != "" {
		return a.Binary
	}
	return "agy"
}

// agyStdinMessage is the one stream-json user message carrying the brief.
func agyStdinMessage(brief []byte) ([]byte, error) {
	msg := map[string]any{
		"event":   "user",
		"message": map[string]any{"role": "user", "content": string(brief)},
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// BuildCommand implements Adapter.
func (a AgyAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	brief, err := readBrief(inv)
	if err != nil {
		return nil, nil, nil, err
	}
	stdin, err := agyStdinMessage(brief)
	if err != nil {
		return nil, nil, nil, err
	}
	argv := []string{a.bin(),
		"--input-format", "stream-json", "--output-format", "stream-json",
		"--model", inv.Profile.Model,
		// the worker is the permission layer (sandbox, scope checks); an
		// approval prompt would hang a headless run
		"--dangerously-skip-permissions",
		// no slash command or skill expansion of the brief's text
		"--disable-slash-commands",
		"--print", ""}
	return argv, baseEnv(inv, nil), stdin, nil
}

// agyStep returns the step_update member of a line, or nil.
func agyStep(m map[string]any) map[string]any { return sub(m, "step_update") }

// agyContext is the context one model call held: its uncached and cached
// prompt plus its output and thinking.
func agyContext(usage map[string]any) int {
	if usage == nil {
		return 0
	}
	return inum(usage, "input_tokens") + inum(usage, "cache_read_tokens") +
		inum(usage, "output_tokens") + inum(usage, "thinking_tokens")
}

// ParseEvent implements Adapter for agy stream-json lines.
func (AgyAdapter) ParseEvent(line []byte) (Event, bool) {
	m, ok := parseObject(line)
	if !ok {
		return Event{}, false
	}
	ev := Event{Kind: KindOther, Raw: string(line)}
	typ := str(m, "event")
	switch typ {
	case "step_update":
		s := agyStep(m)
		stepType := str(s, "step_type")
		switch {
		case strings.Contains(stepType, "compact") || strings.Contains(stepType, "checkpoint"):
			ev.Kind = KindCompaction
			ev.Message = stepType
		case stepType == "error_message":
			ev.Kind = KindError
			ev.Message = truncate(str(s, "error", "message", "text"), 2000)
			if ev.Message == "" {
				ev.Message = "error_message step"
			}
		case stepType == "tool":
			ev.Kind = KindTool
			ev.Message = str(s, "tool_name")
		case stepType == "agent_response" && sub(s, "usage") != nil:
			ev.Kind = KindUsage
			ev.Message = "call"
			ev.ContextTokens = agyContext(sub(s, "usage"))
		case stepType == "agent_response":
			ev.Kind = KindText
			ev.Message = str(s, "text_delta")
		default:
			ev.Message = stepType
		}
	case "result":
		r := sub(m, "result")
		ev.Kind = KindUsage
		ev.Message = "result"
		if st := str(r, "status"); st != "" && st != "SUCCESS" {
			ev.Kind = KindError
			ev.Message = truncate(st+": "+str(r, "error"), 2000)
		}
	case "error":
		ev.Kind = KindError
		ev.Message = truncate(str(m, "error", "message"), 2000)
	default:
		ev.Message = typ
	}
	return ev, true
}

// agyResult returns the result member of the last "result" line.
func agyResult(stdout []byte) map[string]any {
	end := lastObjectLine(stdout, func(m map[string]any) bool { return str(m, "event") == "result" })
	if end == nil {
		return nil
	}
	return sub(end, "result")
}

// FinalUsage implements Adapter: the result's totals, with the cached
// prompt added to InputTokens and one turn per model call.
func (AgyAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	res := agyResult(stdout)
	usage := sub(res, "usage")
	if usage == nil {
		return Usage{}, ErrNoUsage
	}
	read := inum(usage, "cache_read_tokens")
	u := Usage{
		InputTokens:     inum(usage, "input_tokens") + read,
		CachedTokens:    read,
		OutputTokens:    inum(usage, "output_tokens"),
		ReasoningTokens: inum(usage, "thinking_tokens"),
		PeakContext:     PeakContext(events),
	}
	for _, ev := range events {
		if ev.Kind == KindUsage && ev.Message == "call" {
			u.Turns++
		}
	}
	if u.Turns == 0 {
		u.Turns = inum(res, "num_turns")
	}
	if u.PeakContext == 0 {
		u.PeakContext = u.InputTokens + u.OutputTokens
	}
	return u, nil
}

// HandoffFromOutput implements handoffSource: the final response of a
// successful run, when it is one JSON object, bare or fenced.
func (AgyAdapter) HandoffFromOutput(stdout []byte) ([]byte, bool) {
	res := agyResult(stdout)
	if res == nil || str(res, "status") != "SUCCESS" {
		return nil, false
	}
	return jsonObjectText(str(res, "response"))
}

// ExitClass implements Adapter. agy writes its error twice on stderr (a
// plain "error:" line and an AGY_ERROR JSON line with the status, such as
// RESOURCE_EXHAUSTED or UNAUTHENTICATED); the shared signatures read both.
func (AgyAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	c := classifyExit(exitCode, stderr, events, nil)
	if c == Ok {
		for _, ev := range events {
			if ev.Kind == KindError && strings.HasPrefix(ev.Message, "ERROR") {
				return Crash
			}
		}
	}
	return c
}

// InstructionFile implements Adapter: AGENTS.md, the cross-tool
// convention; the brief carries every rule the run needs regardless.
func (AgyAdapter) InstructionFile() string { return "AGENTS.md" }

// ProjectSettings implements Adapter.
func (AgyAdapter) ProjectSettings(inv Invocation) map[string][]byte { return nil }
