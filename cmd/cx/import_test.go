package main

import (
	"reflect"
	"strings"
	"testing"
)

// A Claude Code transcript keeps what the user typed and what the
// assistant wrote, and leaves out tools, thinking, compaction summaries,
// sub-agents and the harness's own records.
func TestParseClaudeCodeKeepsOnlyTheConversation(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(`
{"type":"custom-title","title":"x"}
{"type":"user","message":{"role":"user","content":"Which port does staging use?"}}
{"type":"assistant","message":{"role":"assistant","model":"m","content":[{"type":"thinking","thinking":"hmm"}]}}
{"type":"assistant","message":{"role":"assistant","model":"m","content":[{"type":"text","text":"Let me check."}]}}
{"type":"assistant","message":{"role":"assistant","model":"m","content":[{"type":"tool_use","id":"a","name":"Bash","input":{}}]}}
{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"a","content":"5433"}]}}
{"type":"assistant","message":{"role":"assistant","model":"m","content":[{"type":"text","text":"Staging listens on 5433."}]}}
{"type":"user","isMeta":true,"message":{"role":"user","content":"Caveat: local commands"}}
{"type":"user","message":{"role":"user","content":"<command-name>/model</command-name>"}}
{"type":"user","message":{"role":"user","content":"<local-command-stdout>Set model</local-command-stdout>"}}
{"type":"assistant","isSidechain":true,"message":{"role":"assistant","model":"m","content":[{"type":"text","text":"sub-agent text"}]}}
{"type":"assistant","isApiErrorMessage":true,"message":{"role":"assistant","model":"<synthetic>","content":[{"type":"text","text":"API Error"}]}}
{"type":"user","message":{"role":"user","content":[{"type":"text","text":"<system-reminder>harness</system-reminder>"},{"type":"text","text":"And production?"}]}}
{"type":"attachment","attachment":{"type":"queued_command","prompt":"<task-notification>done</task-notification>"}}
{"type":"assistant","message":{"role":"assistant","model":"m","content":[{"type":"text","text":"Production uses 5432."}]}}
{"type":"attachment","attachment":{"type":"queued_command","prompt":"Thanks, write that down."}}
{"type":"user","isCompactSummary":true,"message":{"role":"user","content":"This session is being continued..."}}
{"type":"assistant","message":{"role":"assistant","model":"m","content":[{"type":"text","text":"Noted."}]}}
`), "\n")
	if !isClaudeCode(lines) {
		t.Fatal("not detected as a Claude Code transcript")
	}
	if isClaudeCode([]string{`{"role":"user","content":"hi"}`}) {
		t.Fatal("a plain role/content line was taken for Claude Code")
	}
	want := []importMsg{
		{Role: "user", Content: "Which port does staging use?"},
		{Role: "assistant", Content: "Let me check.\n\nStaging listens on 5433."},
		{Role: "user", Content: "And production?"},
		{Role: "assistant", Content: "Production uses 5432."},
		{Role: "user", Content: "Thanks, write that down."},
		{Role: "assistant", Content: "Noted."},
	}
	if got := parseClaudeCode(lines); !reflect.DeepEqual(got, want) {
		t.Fatalf("got\n%+v\nwant\n%+v", got, want)
	}
}

// An imported Claude Code session keeps when each message was written.
func TestParseClaudeCodeKeepsMessageTimes(t *testing.T) {
	msgs := parseClaudeCode([]string{
		`{"type":"user","timestamp":"2026-11-01T10:00:00.500Z","message":{"role":"user","content":"First?"}}`,
		`{"type":"assistant","timestamp":"2026-11-01T10:00:09Z","message":{"role":"assistant","model":"m","content":[{"type":"text","text":"One."}]}}`,
		`{"type":"assistant","timestamp":"2026-11-01T10:00:20Z","message":{"role":"assistant","model":"m","content":[{"type":"text","text":"Two."}]}}`,
		`{"type":"user","timestamp":"2026-11-01T10:04:00Z","message":{"role":"user","content":"Second?"}}`,
	})
	if len(msgs) != 3 || msgs[0].Timestamp != "2026-11-01T10:00:00.500Z" || msgs[1].Timestamp != "2026-11-01T10:00:20Z" {
		t.Fatalf("%+v", msgs)
	}
	if ts := userTimes(msgs); len(ts) != 2 || ts[1].Sub(ts[0]).Seconds() != 239.5 {
		t.Fatalf("user times %v", ts)
	}
}
