package providers

import (
	"encoding/json"
	"strings"
	"testing"
)

// The lines below are trimmed from agy 1.5.6 runs of 2026-11-11.
const agyStream = `{"event":"init","conversation_id":"c1","init":{"model":"gemini-4-argon-high","cwd":"/ws","tools":["view_file","write_to_file","search_web"],"permission_mode":"always-proceed"}}
{"event":"step_update","step_update":{"conversation_id":"c1","step_index":0,"state":"DONE","step_type":"user_input"}}
{"event":"step_update","step_update":{"conversation_id":"c1","step_index":1,"state":"DONE","step_type":"agent_response","duration_seconds":7.27,"usage":{"input_tokens":14563,"output_tokens":168,"thinking_tokens":84,"cache_read_tokens":0,"total_tokens":14731}}}
{"event":"step_update","step_update":{"conversation_id":"c1","step_index":2,"state":"DONE","step_type":"tool","tool_name":"read_file","duration_seconds":0.03}}
{"event":"step_update","step_update":{"conversation_id":"c1","step_index":3,"state":"DONE","step_type":"agent_response","duration_seconds":8.7,"usage":{"input_tokens":2317,"output_tokens":159,"thinking_tokens":45,"cache_read_tokens":12509,"total_tokens":2476}}}
{"event":"step_update","step_update":{"conversation_id":"c1","step_index":5,"state":"ACTIVE","step_type":"agent_response","text_delta":"{\"run_id\":\"run-test\",\"status\":\"completed\"}"}}
{"event":"step_update","step_update":{"conversation_id":"c1","step_index":5,"state":"DONE","step_type":"agent_response","text_delta":"\n","duration_seconds":3.6,"usage":{"input_tokens":2288,"output_tokens":32,"thinking_tokens":10,"cache_read_tokens":12777,"total_tokens":2320}}}
{"event":"result","result":{"conversation_id":"c1","status":"SUCCESS","response":"{\"run_id\":\"run-test\",\"status\":\"completed\"}\n","duration_seconds":21.2,"num_turns":1,"usage":{"input_tokens":19168,"output_tokens":359,"thinking_tokens":139,"cache_read_tokens":25286,"total_tokens":19527}}}
`

func agyEvents(t *testing.T, stdout string) []Event {
	t.Helper()
	var evs []Event
	for _, line := range strings.Split(stdout, "\n") {
		if ev, ok := (AgyAdapter{}).ParseEvent([]byte(line)); ok {
			evs = append(evs, ev)
		}
	}
	return evs
}

func TestAgyArgvAndStdin(t *testing.T) {
	inv := newInvocation(t, "agy")
	inv.Profile.Model, inv.Profile.Effort = "gemini-4-argon-high", ""
	inv.SchemaPath = "/x/handoff.schema.json"
	argv, env, stdin, err := AgyAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"agy", "--input-format", "stream-json", "--output-format", "stream-json",
		"--model", "gemini-4-argon-high", "--dangerously-skip-permissions", "--disable-slash-commands", "--print", ""}
	if !equalStrings(argv, want) {
		t.Errorf("argv\n got %q\nwant %q", argv, want)
	}
	var msg struct {
		Event   string `json:"event"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(stdin, &msg); err != nil {
		t.Fatalf("stdin is not one JSON message: %v (%q)", err, stdin)
	}
	if msg.Event != "user" || msg.Message.Role != "user" || !strings.Contains(msg.Message.Content, "Do the thing.") {
		t.Errorf("stdin message %+v", msg)
	}
	for _, kv := range env {
		if strings.HasPrefix(kv, "GEMINI_API_KEY=") || strings.HasPrefix(kv, "GOOGLE_API_KEY=") {
			t.Errorf("an API key reached agy's environment: per-use billing (%s)", strings.SplitN(kv, "=", 2)[0])
		}
	}
}

func TestAgyNoAPIKeyEvenWhenSet(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "should-not-pass")
	t.Setenv("GOOGLE_API_KEY", "should-not-pass")
	inv := newInvocation(t, "agy")
	_, env, _, err := AgyAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range env {
		if strings.Contains(kv, "should-not-pass") {
			t.Fatalf("API key passed to agy: %s", strings.SplitN(kv, "=", 2)[0])
		}
	}
}

func TestAgyEventsAndUsage(t *testing.T) {
	evs := agyEvents(t, agyStream)
	calls, tools := 0, 0
	for _, ev := range evs {
		if ev.Kind == KindUsage && ev.Message == "call" {
			calls++
		}
		if ev.Kind == KindTool {
			tools++
		}
		if ev.Kind == KindError || ev.Kind == KindCompaction {
			t.Errorf("unexpected %s event: %s", ev.Kind, ev.Message)
		}
	}
	if calls != 3 || tools != 1 {
		t.Errorf("calls %d tools %d, want 3 and 1", calls, tools)
	}
	u, err := AgyAdapter{}.FinalUsage([]byte(agyStream), evs)
	if err != nil {
		t.Fatal(err)
	}
	if u.InputTokens != 19168+25286 || u.CachedTokens != 25286 || u.OutputTokens != 359 || u.ReasoningTokens != 139 {
		t.Errorf("usage %+v", u)
	}
	if u.Turns != 3 {
		t.Errorf("turns %d, want 3 model calls", u.Turns)
	}
	// the largest call is the last, its cached prompt counted:
	// 2288 + 12777 + 32 + 10
	if u.PeakContext != 15107 {
		t.Errorf("peak %d, want 15107", u.PeakContext)
	}
	if u.HasUSD || u.USDReported != 0 {
		t.Error("agy runs on the plan; no dollars")
	}
	if c := (AgyAdapter{}).ExitClass(0, nil, evs); c != Ok {
		t.Errorf("exit class %s", c)
	}
}

func TestAgyHandoffFromResponse(t *testing.T) {
	data, ok := AgyAdapter{}.HandoffFromOutput([]byte(agyStream))
	if !ok {
		t.Fatal("no handoff adopted from a SUCCESS response holding one JSON object")
	}
	var h map[string]any
	if json.Unmarshal(data, &h) != nil || h["run_id"] != "run-test" || h["status"] != "completed" {
		t.Errorf("handoff %s", data)
	}
	failed := `{"event":"result","result":{"status":"ERROR","response":"{\"run_id\":\"run-test\",\"status\":\"completed\"}","error":"boom","usage":{}}}`
	if _, ok := (AgyAdapter{}).HandoffFromOutput([]byte(failed)); ok {
		t.Error("adopted a handoff from a failed run")
	}
}

func TestAgyErrorsClassify(t *testing.T) {
	cases := []struct {
		name   string
		stdout string
		stderr string
		code   int
		want   ExitClass
	}{
		{"quota", `{"event":"result","result":{"status":"ERROR","error":"RESOURCE_EXHAUSTED (code 429): quota exceeded","usage":{}}}`,
			`AGY_ERROR: {"short_error":"RESOURCE_EXHAUSTED (code 429)","status":"RESOURCE_EXHAUSTED","error_code":429}`, 1, Quota},
		{"auth", `{"event":"result","result":{"status":"ERROR","error":"UNAUTHENTICATED (code 401)","usage":{}}}`,
			`error: UNAUTHENTICATED (code 401)`, 1, Auth},
		{"bad request", `{"event":"result","result":{"status":"ERROR","error":"INVALID_ARGUMENT (code 400): enum[4]: cannot be empty","usage":{}}}`,
			`error: INVALID_ARGUMENT (code 400)`, 1, Crash},
		{"error with exit 0", `{"event":"result","result":{"status":"ERROR","error":"the model request failed with an unspecified error","usage":{}}}`,
			``, 0, Crash},
	}
	for _, c := range cases {
		evs := agyEvents(t, c.stdout)
		if got := (AgyAdapter{}).ExitClass(c.code, []byte(c.stderr), evs); got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
	}
}
