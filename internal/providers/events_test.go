package providers

import (
	"strings"
	"testing"
)

func parseAll(t *testing.T, a Adapter, lines string) []Event {
	t.Helper()
	var out []Event
	for _, l := range strings.Split(lines, "\n") {
		if ev, ok := a.ParseEvent([]byte(l)); ok {
			out = append(out, ev)
		}
	}
	return out
}

const claudeStream = `{"type":"system","subtype":"init","session_id":"s1","model":"claude-opus-4-7"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Reading the brief."}],"usage":{"input_tokens":12,"cache_creation_input_tokens":14000,"cache_read_input_tokens":0,"output_tokens":20}}}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"scenes/sc-a.md"}}],"usage":{"input_tokens":30,"cache_creation_input_tokens":0,"cache_read_input_tokens":14000,"output_tokens":60}}}
{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"..."}]}}
not json at all
{"type":"rate_limit_event","rate_limit_info":{"status":"allowed_warning","rateLimitType":"five_hour","utilization":0.9}}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"t2","name":"WebFetch","input":{}}],"usage":{"input_tokens":9000,"cache_creation_input_tokens":0,"cache_read_input_tokens":14000,"output_tokens":100}}}
{"type":"result","subtype":"success","is_error":false,"duration_ms":4200,"num_turns":3,"result":"OK","total_cost_usd":0.42,"usage":{"input_tokens":9042,"cache_creation_input_tokens":14000,"cache_read_input_tokens":28000,"output_tokens":180}}`

func TestClaudeEvents(t *testing.T) {
	a := ClaudeAdapter{}
	events := parseAll(t, a, claudeStream)
	if len(events) != 7 {
		t.Fatalf("got %d events", len(events))
	}
	if events[0].Kind != KindOther || events[0].Message != "init" {
		t.Errorf("init: %+v", events[0])
	}
	if events[1].Kind != KindText || events[1].ContextTokens != 14032 {
		t.Errorf("text turn: %+v", events[1])
	}
	if events[2].Kind != KindTool || events[2].Message != "Read" || events[2].ContextTokens != 14090 {
		t.Errorf("tool turn: %+v", events[2])
	}
	if events[4].Kind != KindOther || !strings.HasPrefix(events[4].Message, "rate_limit:allowed_warning") {
		t.Errorf("warning must not be a rate_limit event: %+v", events[4])
	}
	if events[5].Kind != KindTool || events[5].Message != "WebFetch" {
		t.Errorf("web tool: %+v", events[5])
	}
	if events[6].Kind != KindUsage || events[6].Message != "result:success" {
		t.Errorf("result: %+v", events[6])
	}
	u, err := a.FinalUsage([]byte(claudeStream), events)
	if err != nil {
		t.Fatal(err)
	}
	if u.InputTokens != 9042+14000+28000 || u.CachedTokens != 28000 || u.OutputTokens != 180 || u.Turns != 3 ||
		u.WallMs != 4200 || !u.HasUSD || u.USDReported != 0.42 || u.PeakContext != 23100 {
		t.Errorf("usage %+v", u)
	}
	if a.ExitClass(0, nil, events) != Ok {
		t.Error("clean run")
	}

	compact := `{"type":"system","subtype":"compact_boundary","compact_metadata":{"trigger":"auto","pre_tokens":150000}}`
	ev, ok := a.ParseEvent([]byte(compact))
	if !ok || ev.Kind != KindCompaction {
		t.Errorf("compact boundary: %+v", ev)
	}
	rejected := `{"type":"rate_limit_event","rate_limit_info":{"status":"rejected","rateLimitType":"seven_day","resetsAt":9999999999}}`
	ev, _ = a.ParseEvent([]byte(rejected))
	if ev.Kind != KindRateLimit || ev.RetryAfterSeconds <= 0 {
		t.Errorf("rejected: %+v", ev)
	}
	apiErr := `{"type":"assistant","message":{"content":[{"type":"text","text":"API Error: 429 {\"type\":\"error\",\"error\":{\"type\":\"rate_limit_error\",\"message\":\"usage limit reached, resets at 3pm\"}}"}],"usage":{"input_tokens":1,"output_tokens":1}}}`
	ev, _ = a.ParseEvent([]byte(apiErr))
	if ev.Kind != KindError {
		t.Errorf("api error: %+v", ev)
	}
	if a.ExitClass(1, nil, []Event{ev}) != Quota {
		t.Error("429 text must be Quota")
	}
	capped := `{"type":"result","subtype":"error_max_turns","is_error":true,"num_turns":40,"usage":{"input_tokens":1}}`
	ev, _ = a.ParseEvent([]byte(capped))
	if ev.Kind != KindError || a.ExitClass(0, nil, []Event{ev}) != OkBlocked {
		t.Errorf("max turns: %+v class %s", ev, a.ExitClass(0, nil, []Event{ev}))
	}
	if _, err := a.FinalUsage([]byte("nothing\n"), nil); err != ErrNoUsage {
		t.Error("no usage must be ErrNoUsage")
	}
}

const codexStream = `{"type":"thread.started","thread_id":"th1"}
{"type":"turn.started"}
{"type":"item.started","item":{"id":"i1","type":"command_execution","command":"cat scenes/sc-a.md","status":"in_progress"}}
{"type":"item.completed","item":{"id":"i1","type":"command_execution","command":"cat scenes/sc-a.md","exit_code":0}}
{"type":"item.completed","item":{"id":"i2","type":"agent_message","text":"Done with unit one."}}
{"type":"item.completed","item":{"id":"i3","type":"file_change","changes":[{"path":"scenes/sc-a.md","kind":"update"}]}}
{"type":"turn.completed","usage":{"input_tokens":40000,"cached_input_tokens":30000,"output_tokens":2000,"reasoning_output_tokens":500,"total_tokens":42000}}
{"type":"turn.started"}
{"type":"item.started","item":{"id":"i4","type":"mcp_tool_call","server":"filesystem","tool":"read_file"}}
{"type":"turn.completed","usage":{"input_tokens":45000,"cached_input_tokens":40000,"output_tokens":1000,"reasoning_output_tokens":0,"total_tokens":46000}}`

func TestCodexEvents(t *testing.T) {
	a := CodexAdapter{}
	events := parseAll(t, a, codexStream)
	if len(events) != 10 {
		t.Fatalf("got %d events", len(events))
	}
	if events[2].Kind != KindTool || !strings.HasPrefix(events[2].Message, "command_execution: cat") {
		t.Errorf("command: %+v", events[2])
	}
	if events[3].Kind != KindOther {
		t.Errorf("completed command must not double count: %+v", events[3])
	}
	if events[4].Kind != KindText || events[4].Message != "Done with unit one." {
		t.Errorf("text: %+v", events[4])
	}
	if events[5].Kind != KindTool || events[5].Message != "file_change" {
		t.Errorf("file change: %+v", events[5])
	}
	// turn.completed usage is cumulative over model calls, so it carries no
	// context size (stage 0 finding)
	if events[6].Kind != KindUsage || events[6].ContextTokens != 0 {
		t.Errorf("usage: %+v", events[6])
	}
	if events[8].Kind != KindTool || events[8].Message != "mcp_tool_call: filesystem.read_file" {
		t.Errorf("mcp: %+v", events[8])
	}
	u, err := a.FinalUsage([]byte(codexStream), events)
	if err != nil {
		t.Fatal(err)
	}
	if u.InputTokens != 85000 || u.CachedTokens != 70000 || u.OutputTokens != 3000 || u.ReasoningTokens != 500 ||
		u.Turns != 4 || u.PeakContext != 43250 || u.HasUSD {
		t.Errorf("usage %+v", u)
	}
	ev, _ := a.ParseEvent([]byte(`{"type":"item.completed","item":{"id":"i9","type":"context_compaction"}}`))
	if ev.Kind != KindCompaction {
		t.Errorf("compaction item: %+v", ev)
	}
	ev, _ = a.ParseEvent([]byte(`{"type":"turn.failed","error":{"message":"You've hit your usage limit. Try again in 2 hours"}}`))
	if ev.Kind != KindError || a.ExitClass(1, nil, []Event{ev}) != Quota {
		t.Errorf("usage limit: %+v", ev)
	}
	if a.ExitClass(1, []byte("Error: context window exceeded for this model"), nil) != ContextBlocked {
		t.Error("context window on stderr")
	}
}

const geminiStream = `{"type":"init","session_id":"g1","model":"gemini-3-pro-preview"}
{"type":"message","role":"assistant","content":"Reading."}
{"type":"tool_use","tool_name":"read_file","parameters":{"path":"scenes/sc-a.md"}}
{"type":"tool_result","tool_name":"read_file","status":"success"}
{"type":"message","role":"assistant","content":"OK"}
{"type":"result","status":"success","stats":{"models":{"gemini-3-pro-preview":{"tokens":{"input_tokens":90000,"output_tokens":1200,"cached_tokens":50000,"thoughts_tokens":300}}},"latency_ms":5000}}`

func TestGeminiEvents(t *testing.T) {
	a := GeminiAdapter{}
	events := parseAll(t, a, geminiStream)
	if len(events) != 6 {
		t.Fatalf("got %d events", len(events))
	}
	if events[1].Kind != KindText || events[2].Kind != KindTool || events[2].Message != "read_file" || events[5].Kind != KindUsage {
		t.Errorf("events %+v", events)
	}
	u, err := a.FinalUsage([]byte(geminiStream), events)
	if err != nil {
		t.Fatal(err)
	}
	if u.InputTokens != 90000 || u.OutputTokens != 1200 || u.CachedTokens != 50000 || u.ReasoningTokens != 300 || u.Turns != 2 {
		t.Errorf("usage %+v", u)
	}
	ev, _ := a.ParseEvent([]byte(`{"type":"error","message":"RESOURCE_EXHAUSTED: Quota exceeded for quota metric"}`))
	if ev.Kind != KindError || a.ExitClass(1, nil, []Event{ev}) != Quota {
		t.Errorf("resource exhausted: %+v", ev)
	}
	ev, _ = a.ParseEvent([]byte(`{"type":"telemetry","event":"chat_compression","tokens_before":600000,"tokens_after":200000}`))
	if ev.Kind != KindCompaction {
		t.Errorf("chat_compression: %+v", ev)
	}
	if a.ExitClass(53, nil, nil) != OkBlocked {
		t.Error("exit 53 is the turn limit")
	}
	if a.ExitClass(1, []byte("API key not valid. Please pass a valid API key."), nil) != Auth {
		t.Error("bad api key is Auth")
	}
	if _, err := a.FinalUsage([]byte(`{"type":"result","status":"success"}`), nil); err != ErrNoUsage {
		t.Error("result without stats is ErrNoUsage")
	}
}

const kimiStream = `{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":"k1","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Reading brief.md"}}}}
{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":"k1","update":{"sessionUpdate":"tool_call","toolCallId":"c1","title":"ReadFile","kind":"read","status":"pending"}}}
{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":"k1","update":{"sessionUpdate":"tool_call_update","toolCallId":"c1","status":"completed"}}}
{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":"k1","update":{"sessionUpdate":"usage_update","used":120000,"size":1048576}}}
{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":"k1","update":{"sessionUpdate":"usage_update","used":860000,"size":1048576}}}`

func TestKimiEvents(t *testing.T) {
	a := KimiAdapter{}
	events := parseAll(t, a, kimiStream)
	if len(events) != 5 {
		t.Fatalf("got %d events", len(events))
	}
	if events[0].Kind != KindText || events[1].Kind != KindTool || events[1].Message != "ReadFile" || events[2].Kind != KindOther {
		t.Errorf("events %+v", events[:3])
	}
	if events[3].Kind != KindUsage || events[3].ContextTokens != 120000 || events[3].Window != 1048576 {
		t.Errorf("usage_update %+v", events[3])
	}
	u, err := a.FinalUsage([]byte(kimiStream), events)
	if err != nil {
		t.Fatal(err)
	}
	if u.PeakContext != 860000 || u.InputTokens != 860000 {
		t.Errorf("usage %+v", u)
	}
	ev, _ := a.ParseEvent([]byte(`{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"compaction.started","trigger":"auto_compact"}}}`))
	if ev.Kind != KindCompaction || ev.Message != "compaction.started" {
		t.Errorf("compaction %+v", ev)
	}
	ev, _ = a.ParseEvent([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"Not logged in. Run kimi login first."}}`))
	if ev.Kind != KindError || a.ExitClass(1, nil, []Event{ev}) != Auth {
		t.Errorf("auth %+v", ev)
	}
	rich := `{"type":"usage","input_tokens":5000,"output_tokens":700,"input_tokens_details":{"cached_tokens":4000}}`
	u, _ = a.FinalUsage([]byte(rich), nil)
	if u.InputTokens != 5000 || u.OutputTokens != 700 || u.CachedTokens != 4000 {
		t.Errorf("rich usage %+v", u)
	}
}

const grokResultJSON = `{
  "result": "OK",
  "num_turns": 4,
  "total_cost_usd": 0.0731,
  "usage": {"input_tokens": 1000, "cache_read_input_tokens": 20000, "cache_creation_input_tokens": 3000, "output_tokens": 800, "reasoning_tokens": 300, "total_tokens": 24800},
  "modelUsage": {"grok-4.5": {"costUSD": 0.0731}}
}`

// grokStream is grok 1.0.70's streaming-json for a two-call run (recorded
// 2026-10-30, signatures dropped).
const grokStream = `{"type":"available_commands","commands":[]}
{"type":"usage","usage":{"input_tokens":7612,"output_tokens":98,"cache_read_input_tokens":10112,"cache_creation_input_tokens":0,"reasoning_tokens":29}}
{"type":"tool_call","name":"read_file"}
{"type":"usage","usage":{"input_tokens":171,"output_tokens":24,"cache_read_input_tokens":17664,"cache_creation_input_tokens":0,"reasoning_tokens":22}}
{"type":"text","data":"lantern"}
{"type":"end","stopReason":"end_turn","usage":{"input_tokens":7783,"cache_read_input_tokens":27776,"cache_creation_input_tokens":0,"output_tokens":122,"reasoning_tokens":51,"total_tokens":35681},"num_turns":2,"total_cost_usd":0.01026324}
`

// TestGrokStream: each usage line is one call's context, so the peak is the
// largest call, not the run's totals; the end line gives totals and cost;
// an end at the turn limit is the cap.
func TestGrokStream(t *testing.T) {
	a := GrokAdapter{}
	var events []Event
	for _, line := range strings.Split(grokStream, "\n") {
		if ev, ok := a.ParseEvent([]byte(line)); ok {
			events = append(events, ev)
		}
	}
	u, err := a.FinalUsage([]byte(grokStream), events)
	if err != nil {
		t.Fatal(err)
	}
	if u.PeakContext != 171+17664+24 || u.InputTokens != 7783+27776 || u.CachedTokens != 27776 || u.OutputTokens != 122 ||
		u.Turns != 2 || !u.HasUSD || u.USDReported != 0.01026324 {
		t.Errorf("usage %+v", u)
	}
	if a.ExitClass(0, nil, events) != Ok {
		t.Error("a clean end is ok")
	}
	ev, _ := a.ParseEvent([]byte(`{"type":"end","stopReason":"max_turns","usage":{"input_tokens":1},"num_turns":60}`))
	if a.ExitClass(0, nil, []Event{ev}) != OkBlocked {
		t.Errorf("turn limit %+v", ev)
	}
	ev, _ = a.ParseEvent([]byte(`{"type":"error","error":{"message":"Couldn't set model"}}`))
	if ev.Kind != KindError || ev.Message != "Couldn't set model" {
		t.Errorf("error %+v", ev)
	}
}

func TestGrokEvents(t *testing.T) {
	a := GrokAdapter{}
	u, err := a.FinalUsage([]byte(grokResultJSON), nil)
	if err != nil {
		t.Fatal(err)
	}
	if u.InputTokens != 24000 || u.CachedTokens != 20000 || u.OutputTokens != 800 || u.ReasoningTokens != 300 ||
		u.Turns != 4 || !u.HasUSD || u.USDReported != 0.0731 || u.PeakContext != 12200 {
		t.Errorf("usage %+v", u)
	}
	ev, ok := a.ParseEvent([]byte(`{"event":"auto_compact_started","threshold":0.85}`))
	if !ok || ev.Kind != KindCompaction {
		t.Errorf("compaction %+v", ev)
	}
	if _, ok := a.ParseEvent([]byte(`  "result": "OK",`)); ok {
		t.Error("pretty-printed fragments are not events")
	}
	if a.ExitClass(1, []byte("Error: not authenticated. Run `grok login`."), nil) != Auth {
		t.Error("grok auth")
	}
}

const museStream = `{"type":"run_start","run_id":"m1","model":"muse-spark-1.3"}
{"type":"assistant_message","text":"Starting."}
{"type":"tool_call","name":"read_file","arguments":{"path":"scenes/sc-a.md"}}
{"type":"tool_result","name":"read_file","ok":true}
{"type":"step_end","usage":{"input_tokens":30000,"output_tokens":500,"cached_tokens":10000,"reasoning_tokens":200,"cost_micros":12000}}
{"type":"step_end","usage":{"input_tokens":32000,"output_tokens":400,"cached_tokens":30000,"reasoning_tokens":0,"cost_micros":8000}}`

func TestMuseEvents(t *testing.T) {
	a := MuseAdapter{}
	events := parseAll(t, a, museStream)
	if len(events) != 6 {
		t.Fatalf("got %d events", len(events))
	}
	if events[1].Kind != KindText || events[2].Kind != KindTool || events[2].Message != "read_file" {
		t.Errorf("events %+v", events[:3])
	}
	if events[4].Kind != KindUsage || events[4].ContextTokens != 30500 {
		t.Errorf("usage event %+v", events[4])
	}
	u, err := a.FinalUsage([]byte(museStream), events)
	if err != nil {
		t.Fatal(err)
	}
	if u.InputTokens != 62000 || u.OutputTokens != 900 || u.CachedTokens != 40000 || u.ReasoningTokens != 200 ||
		u.Turns != 2 || !u.HasUSD || u.USDReported != 0.02 || u.PeakContext != 32400 {
		t.Errorf("usage %+v", u)
	}
	ev, _ := a.ParseEvent([]byte(`{"type":"compaction","strategy":"summary-preserved-suffix/v1"}`))
	if ev.Kind != KindCompaction {
		t.Errorf("compaction %+v", ev)
	}
}

// Cache writes are billed apart from cache reads and plain input, so each
// adapter that sees them keeps them apart.
func TestUsageKeepsCacheWrites(t *testing.T) {
	claude := `{"type":"result","subtype":"success","num_turns":1,"total_cost_usd":0.5,"usage":{"input_tokens":10,"cache_creation_input_tokens":300,"cache_read_input_tokens":700,"output_tokens":5}}`
	u, err := ClaudeAdapter{}.FinalUsage([]byte(claude), nil)
	if err != nil || u.InputTokens != 1010 || u.CachedTokens != 700 || u.CacheWriteTokens != 300 || !u.HasUSD {
		t.Fatalf("claude %+v, %v", u, err)
	}
	codex := `{"type":"turn.completed","usage":{"input_tokens":1000,"cached_input_tokens":600,"cache_write_input_tokens":200,"output_tokens":5}}`
	u, err = CodexAdapter{}.FinalUsage([]byte(codex), []Event{{Kind: KindUsage}})
	if err != nil || u.InputTokens != 1000 || u.CachedTokens != 600 || u.CacheWriteTokens != 200 {
		t.Fatalf("codex %+v, %v", u, err)
	}
}
