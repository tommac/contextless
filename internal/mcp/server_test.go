package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/engine"
	"github.com/tommac/contextless/internal/vault"
)

func TestMCPStdioServerFullLifecycle(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Sandbox = false
	cfg.WholeHistoryTokens = -1

	e, err := engine.Create(cfg, "mcptest", "MCP Vault", "Verify MCP tools and resources")
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	if _, err := e.V.AppendTurn(vault.RoleUser, "user", "", "", "Set primary pool_size=20 on port 5432."); err != nil {
		t.Fatal(err)
	}
	if _, err := e.V.AppendTurn(vault.RoleAssistant, "test", "", "", "Confirmed: pool_size=20 on port 5432."); err != nil {
		t.Fatal(err)
	}
	if err := e.Reindex(); err != nil {
		t.Fatal(err)
	}

	reqs := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"resources/list"}`,
		`{"jsonrpc":"2.0","id":4,"method":"resources/read","params":{"uri":"cx://session"}}`,
		`{"jsonrpc":"2.0","id":5,"method":"resources/read","params":{"uri":"cx://turn/t000001"}}`,
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"cx_compile_packet","arguments":{"query":"What is pool_size?"}}}`,
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"cx_explain","arguments":{"query":"What is pool_size?"}}}`,
		`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"cx_recall","arguments":{"op":"grep","target":"pool_size="}}}`,
		`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"cx_create_receipt","arguments":{"question":"What is pool_size?","answer":"pool_size is 20"}}}`,
		`{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"cx_append_turn","arguments":{"role":"user","body":"Update: pool_size=64 after load test."}}}`,
		`{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"cx_watch_status","arguments":{"status":"stale"}}}`,
		`{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"cx_stats","arguments":{}}}`,
	}

	var out bytes.Buffer
	srv := New(e)
	if err := srv.Serve(context.Background(), strings.NewReader(strings.Join(reqs, "\n")+"\n"), &out); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	// 12 requests with IDs (the notification has no response).
	if len(lines) != 12 {
		t.Fatalf("expected 12 responses, got %d:\n%s", len(lines), out.String())
	}
	byID := map[int]map[string]any{}
	for _, ln := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(ln), &m); err != nil {
			t.Fatalf("bad JSON response %q: %v", ln, err)
		}
		if m["error"] != nil {
			t.Fatalf("unexpected RPC error in %s", ln)
		}
		id := int(m["id"].(float64))
		byID[id] = m
	}

	// Verify initialize + tools/list + resources/read + stale trigger from cx_append_turn!
	b2, _ := json.Marshal(byID[2])
	if !strings.Contains(string(b2), "cx_compile_packet") || !strings.Contains(string(b2), "cx_diff_since") {
		t.Fatalf("tools/list missing expected tools: %s", b2)
	}
	b5, _ := json.Marshal(byID[5])
	if !strings.Contains(string(b5), "pool_size=20") {
		t.Fatalf("resources/read cx://turn/t000001 missing body: %s", b5)
	}
	b8, _ := json.Marshal(byID[8])
	if !strings.Contains(string(b8), "t000001") {
		t.Fatalf("cx_recall grep did not hit t000001: %s", b8)
	}
	b11, _ := json.Marshal(byID[11])
	if !strings.Contains(string(b11), "stale") {
		t.Fatalf("expected cx_append_turn to trigger stale receipt in cx_watch_status: %s", b11)
	}

	// Also verify cx_diff_since, cx_ledger_query, and cx_pack_status.
	rs, _ := e.S.ListReceipts("")
	if len(rs) != 1 {
		t.Fatalf("expected 1 receipt, got %d", len(rs))
	}
	extraReqs := strings.Join([]string{
		`{"jsonrpc":"2.0","id":13,"method":"tools/call","params":{"name":"cx_diff_since","arguments":{"receipt_id":"` + rs[0].ID + `"}}}`,
		`{"jsonrpc":"2.0","id":14,"method":"tools/call","params":{"name":"cx_ledger_query","arguments":{"kind":"turns","op":"count","filter":"pool_size","strict":true}}}`,
		`{"jsonrpc":"2.0","id":15,"method":"tools/call","params":{"name":"cx_pack_status","arguments":{"build":true,"verify":true}}}`,
		`{"jsonrpc":"2.0","id":16,"method":"tools/call","params":{"name":"cx_key_view","arguments":{"key":"pool_size"}}}`,
		`{"jsonrpc":"2.0","id":17,"method":"tools/call","params":{"name":"cx_fork_workspace","arguments":{"target":"mcpfork","as_of_seq":2}}}`,
	}, "\n") + "\n"
	out.Reset()
	if err := srv.Serve(context.Background(), strings.NewReader(extraReqs), &out); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "pool_size=64") || !strings.Contains(got, "exact") || !strings.Contains(got, "verified") || !strings.Contains(got, "mcpfork") {
		t.Fatalf("extra MCP tools output incomplete: %s", got)
	}
}

// Over MCP a fork target is a workspace name, never a path: a target with a
// separator is a directory to the configuration, and a tool caller could
// have copied the conversation anywhere writable.
func TestMCPForkTargetMustBeAPlainName(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Sandbox = false
	e, err := engine.Create(cfg, "forksrc", "Fork source", "")
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if _, err := e.V.AppendTurn(vault.RoleUser, "user", "", "", "The vault code is 31-7-22."); err != nil {
		t.Fatal(err)
	}
	if err := e.Reindex(); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	var reqs []string
	for i, target := range []string{outside + "/stolen", "../escape", ".hidden", "a/b", "ok..no"} {
		arg, _ := json.Marshal(map[string]any{"target": target})
		reqs = append(reqs, `{"jsonrpc":"2.0","id":`+string(rune('1'+i))+`,"method":"tools/call","params":{"name":"cx_fork_workspace","arguments":`+string(arg)+`}}`)
	}
	var out bytes.Buffer
	if err := New(e).Serve(context.Background(), strings.NewReader(strings.Join(reqs, "\n")+"\n"), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != len(reqs) {
		t.Fatalf("%d responses for %d requests", len(lines), len(reqs))
	}
	for i, line := range lines {
		if !strings.Contains(line, "plain workspace name") {
			t.Errorf("request %d was not refused: %s", i+1, line)
		}
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Errorf("a fork wrote outside the workspace root: %v", entries)
	}
}

// A resource uri names a turn by id. An id that is a path is refused
// instead of being read as a file.
func TestMCPTurnResourceIsNotAPath(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Sandbox = false
	e, err := engine.Create(cfg, "res", "Resources", "the session purpose line")
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if _, err := e.V.AppendTurn(vault.RoleUser, "user", "", "", "A turn."); err != nil {
		t.Fatal(err)
	}
	srv := New(e)
	// session.md is a Markdown file with front matter one level above turns/
	for _, uri := range []string{"cx://turn/../session", "cx://turn/../../res/session", "cx://turn/t000001/../../session", "cx://turn/"} {
		if text, _, err := srv.readResource(uri); err == nil {
			t.Errorf("%s was read (%d bytes)", uri, len(text))
		}
	}
	if text, _, err := srv.readResource("cx://turn/t000001"); err != nil || text != "A turn." {
		t.Errorf("a real turn: %q %v", text, err)
	}
}
