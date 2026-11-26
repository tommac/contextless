package providers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestClaudeHandoffFromStructuredOutput: a one-shot claude judge, holding
// only the Write tool, cannot overwrite the "started" handoff it never
// read; the completed handoff arrives as structured_output on the result
// line and Run adopts it. A handoff already past "started" is kept.
func TestClaudeHandoffFromStructuredOutput(t *testing.T) {
	stdout := []byte(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"StructuredOutput","input":{}}]}}
{"type":"result","subtype":"success","num_turns":4,"usage":{"input_tokens":10,"output_tokens":5},"structured_output":{"run_id":"run-x","attempt":1,"status":"completed","milestone_seq":1,"units_done":["sc-a"],"result":{"goal_verdict":null}}}
`)
	data, ok := (ClaudeAdapter{}).HandoffFromOutput(stdout)
	if !ok {
		t.Fatal("no structured output found")
	}
	var h map[string]any
	if err := json.Unmarshal(data, &h); err != nil || h["status"] != "completed" {
		t.Fatalf("adopted handoff = %s (%v)", data, err)
	}
	if _, ok := (ClaudeAdapter{}).HandoffFromOutput([]byte(`{"type":"result","subtype":"success"}` + "\n")); ok {
		t.Fatal("a result line without structured_output must not yield a handoff")
	}

	ws := t.TempDir()
	path := filepath.Join(ws, ".cx", "handoff.json")
	inv := Invocation{Workspace: ws, HandoffPath: path}
	if !adoptHandoff(ClaudeAdapter{}, inv, stdout) {
		t.Fatal("missing handoff was not adopted")
	}
	os.WriteFile(path, []byte(`{"status":"started","milestone_seq":0}`), 0o644)
	if !adoptHandoff(ClaudeAdapter{}, inv, stdout) {
		t.Fatal("a started handoff was not replaced")
	}
	got, _ := os.ReadFile(path)
	json.Unmarshal(got, &h)
	if h["status"] != "completed" {
		t.Fatalf("handoff after adoption = %s", got)
	}
	os.WriteFile(path, []byte(`{"status":"partial","units_done":["sc-a"]}`), 0o644)
	if adoptHandoff(ClaudeAdapter{}, inv, stdout) {
		t.Fatal("a partial handoff written by the agent must be kept")
	}
	if adoptHandoff(CodexAdapter{}, inv, stdout) {
		t.Fatal("codex has no structured output to adopt")
	}
}

// TestMuseHandoffAndUsageFromEnvelope: muse 1.3 streams event envelopes; a
// one-shot judge answers with the handoff as the completed run's final
// text (in the live run-yklzcpe5 it made no tool call), and usage comes as
// one run_usage record listing tokens per model.
func TestMuseHandoffAndUsageFromEnvelope(t *testing.T) {
	env := func(payload string) string {
		return `{"schema_version":1,"record_type":"event","payload_type":"x","payload":` + payload + "}\n"
	}
	final := "```json\n{\"run_id\":\"run-m\",\"status\":\"completed\",\"milestone_seq\":1,\"units_done\":[\"sc-a\"]}\n```"
	text, _ := json.Marshal(final)
	stdout := []byte(env(`{"kind":"run_output_delta","text":"{"}`) +
		env(`{"kind":"run_usage","calls":1,"models":[{"model":"muse-spark-1.3","input_tokens":71878,"cached_tokens":22769,"output_tokens":18105,"reasoning_tokens":16606,"calls":1}]}`) +
		env(`{"kind":"run_terminal","terminal":"completed","text":`+string(text)+`,"reason":null}`))

	u, err := (MuseAdapter{}).FinalUsage(stdout, nil)
	if err != nil || u.InputTokens != 71878 || u.CachedTokens != 22769 || u.OutputTokens != 18105 || u.ReasoningTokens != 16606 || u.Turns != 1 {
		t.Fatalf("usage = %+v, %v", u, err)
	}
	data, ok := (MuseAdapter{}).HandoffFromOutput(stdout)
	if !ok {
		t.Fatal("fenced final text was not read as a handoff")
	}
	ws := t.TempDir()
	path := filepath.Join(ws, ".cx", "handoff.json")
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(`{"run_id":"run-m","status":"started"}`), 0o644)
	if adoptHandoff(MuseAdapter{}, Invocation{Workspace: ws, HandoffPath: path, RunID: "run-other"}, stdout) {
		t.Fatal("a handoff naming another run must not be adopted")
	}
	if !adoptHandoff(MuseAdapter{}, Invocation{Workspace: ws, HandoffPath: path, RunID: "run-m"}, stdout) {
		t.Fatal("the run's own completed text was not adopted")
	}
	var h map[string]any
	got, _ := os.ReadFile(path)
	if json.Unmarshal(got, &h); h["status"] != "completed" {
		t.Fatalf("handoff = %s (from %s)", got, data)
	}
	failed := []byte(env(`{"kind":"run_terminal","terminal":"failed","text":"{\"status\":\"completed\"}"}`))
	if _, ok := (MuseAdapter{}).HandoffFromOutput(failed); ok {
		t.Fatal("text of a failed run must not be adopted")
	}
	if _, ok := (MuseAdapter{}).HandoffFromOutput([]byte(env(`{"kind":"run_terminal","terminal":"completed","text":"I could not decide."}`))); ok {
		t.Fatal("prose must not be adopted")
	}
}
