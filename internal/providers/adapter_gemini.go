package providers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// CompactionMarker is the file a provider hook writes into RunDir when the
// CLI is about to compact (gemini PreCompress). Run turns its presence into
// a compaction event (RUN-NO-COMPACTION: "any stream or marker").
const CompactionMarker = "COMPACTED"

// CompactionMarkerPath returns the marker path for an invocation.
func CompactionMarkerPath(inv Invocation) string {
	return filepath.Join(inv.RunDir, CompactionMarker)
}

// GeminiAdapter drives gemini 0.56.0 headless (06 §6.2, research §3). The
// brief goes on stdin and -p carries the pointer sentence; the project
// settings file disables compression and installs the PreCompress marker
// hook. Whether compressionThreshold 1.0 fully disables compression, the
// history-window keys, the stats field names and the web-tool exclusion
// key are unverified (06 §6.8); the profile's effective window stays at
// 500k until the smoke says otherwise.
type GeminiAdapter struct {
	// Binary is the executable; "gemini" when empty.
	Binary string
}

// GeminiPointer is the -p text appended to the stdin brief.
const GeminiPointer = "Follow the brief above exactly."

// GeminiHomeDir is the RunDir entry used as GEMINI_CLI_HOME.
const GeminiHomeDir = "gemini-home"

// Provider implements Adapter.
func (GeminiAdapter) Provider() string { return "gemini" }

func (a GeminiAdapter) bin() string {
	if a.Binary != "" {
		return a.Binary
	}
	return "gemini"
}

// BuildCommand implements Adapter.
func (a GeminiAdapter) BuildCommand(inv Invocation) ([]string, []string, []byte, error) {
	if err := requireWorkspace(inv); err != nil {
		return nil, nil, nil, err
	}
	brief, err := readBrief(inv)
	if err != nil {
		return nil, nil, nil, err
	}
	// --skip-trust: otherwise "Approval mode overridden to default because
	// the current folder is not trusted" (stage 0 finding).
	argv := []string{a.bin(), "-p", GeminiPointer, "--approval-mode", "yolo", "--skip-trust",
		"-m", inv.Profile.Model, "--output-format", "stream-json"}
	extra := map[string]string{}
	if key, ok := os.LookupEnv("GEMINI_API_KEY"); ok {
		extra["GEMINI_API_KEY"] = key
		// API-key auth needs nothing from ~/.gemini, so the run gets its own
		// gemini home in RunDir: no user settings, GEMINI.md memory,
		// extensions or MCP servers leak into the packet, and a sandboxed
		// run needs no access to ~/.gemini (gemini creates
		// ~/.gemini/tmp/<project> on start; stage 0 finding). Without a key
		// (OAuth login) the run uses ~/.gemini and cannot be sandboxed.
		if inv.RunDir != "" {
			extra["GEMINI_CLI_HOME"] = filepath.Join(inv.RunDir, GeminiHomeDir)
		}
	}
	return argv, baseEnv(inv, extra), brief, nil
}

// ParseEvent implements Adapter for gemini stream-json: init, message,
// tool_use, tool_result, error, result. Any type containing "compress" or a
// line mentioning chat_compression is the compaction signal.
func (GeminiAdapter) ParseEvent(line []byte) (Event, bool) {
	m, ok := parseObject(line)
	if !ok {
		return Event{}, false
	}
	ev := Event{Kind: KindOther, Raw: string(line)}
	typ := str(m, "type", "event")
	switch {
	case strings.Contains(typ, "compress") || strings.Contains(string(line), "chat_compression"):
		ev.Kind = KindCompaction
		ev.Message = typ
	case typ == "message":
		ev.Kind = KindText
		ev.Message = str(m, "content", "text")
		if role := str(m, "role"); role != "" && role != "assistant" {
			ev.Kind = KindOther
		}
	case typ == "tool_use":
		ev.Kind = KindTool
		ev.Message = str(m, "tool_name", "name", "tool")
	case typ == "error":
		ev.Kind = KindError
		msg := str(m, "message", "error")
		if e := sub(m, "error"); e != nil {
			msg = str(e, "message")
		}
		ev.Message = truncate(msg, 2000)
	case typ == "result":
		ev.Kind = KindUsage
		ev.Message = "result:" + str(m, "status")
		if e := str(m, "error"); e != "" {
			ev.Kind = KindError
			ev.Message = truncate(e, 2000)
		} else if e := sub(m, "error"); e != nil {
			ev.Kind = KindError
			ev.Message = truncate(str(e, "message"), 2000)
		}
	default:
		ev.Message = typ
	}
	return ev, true
}

// gemini stats key families (names unverified, summed wherever they occur).
var (
	geminiInputKeys  = []string{"input_tokens", "inputTokens", "prompt_tokens", "promptTokens", "promptTokenCount", "input"}
	geminiOutputKeys = []string{"output_tokens", "outputTokens", "candidates_tokens", "candidatesTokenCount", "output"}
	geminiCachedKeys = []string{"cached_tokens", "cachedTokens", "cached_content_tokens", "cachedContentTokenCount", "cached"}
	geminiThinkKeys  = []string{"thoughts_tokens", "thoughtsTokenCount", "reasoning_tokens", "thoughts"}
)

// FinalUsage implements Adapter: the result line's stats, summed across
// models with every known key spelling.
func (GeminiAdapter) FinalUsage(stdout []byte, events []Event) (Usage, error) {
	res := lastObjectLine(stdout, func(m map[string]any) bool { return str(m, "type") == "result" })
	if res == nil {
		return Usage{}, ErrNoUsage
	}
	stats, ok := res["stats"]
	if !ok {
		return Usage{}, ErrNoUsage
	}
	u := Usage{PeakContext: PeakContext(events)}
	in, okIn := sumNumbers(stats, geminiInputKeys...)
	out, okOut := sumNumbers(stats, geminiOutputKeys...)
	if !okIn && !okOut {
		return Usage{}, ErrNoUsage
	}
	cached, _ := sumNumbers(stats, geminiCachedKeys...)
	think, _ := sumNumbers(stats, geminiThinkKeys...)
	u.InputTokens = int(in)
	u.OutputTokens = int(out)
	u.CachedTokens = int(cached)
	u.ReasoningTokens = int(think)
	for _, ev := range events {
		if ev.Kind == KindText {
			u.Turns++
		}
	}
	return u, nil
}

// ExitClass implements Adapter: exit 53 is the turn limit (OkBlocked).
func (GeminiAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	if exitCode == 53 && !HasCompaction(events) {
		if c := SignatureClass(string(stderr)); c != "" {
			return c
		}
		return OkBlocked
	}
	return classifyExit(exitCode, stderr, events, nil)
}

// InstructionFile implements Adapter.
func (GeminiAdapter) InstructionFile() string { return "GEMINI.md" }

// ProjectSettings implements Adapter: .gemini/settings.json with the
// compression threshold at 1.0, the history window sized to the effective
// window so it never truncates inside it, the PreCompress marker hook, and
// the web tools excluded unless Network.
func (GeminiAdapter) ProjectSettings(inv Invocation) map[string][]byte {
	eff := inv.Profile.EffectiveWindow
	if eff == 0 {
		eff = inv.Profile.Window
	}
	marker := CompactionMarkerPath(inv)
	settings := map[string]any{
		"model": map[string]any{"compressionThreshold": 1.0},
		"contextManagement": map[string]any{
			"historyWindow": map[string]any{"maxTokens": eff, "retainedTokens": eff},
		},
		"hooks": map[string]any{
			"PreCompress": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{
							"name":    "bw-compaction-marker",
							"type":    "command",
							"command": "/usr/bin/touch '" + strings.ReplaceAll(marker, "'", "'\\''") + "'",
							"timeout": 5000,
						},
					},
				},
			},
		},
	}
	if !inv.Network {
		settings["tools"] = map[string]any{"exclude": []string{"google_web_search", "web_fetch"}}
	}
	b, _ := json.MarshalIndent(settings, "", "  ")
	return map[string][]byte{".gemini/settings.json": append(b, '\n')}
}
