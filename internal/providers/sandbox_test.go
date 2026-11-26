package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSandboxProfileText(t *testing.T) {
	inv := newInvocation(t, "claude")
	inv.RunID = "run-sb"
	profile := SandboxProfile(inv, "/Users/x/.local/share/claude/versions/2.1.141", []string{"/Users/x/.claude", "/var/folders/zz/T"}, []string{"api.anthropic.com", "localhost:4711", "*:443"})
	for _, needle := range []string{
		"(version 1)", "(deny default)", `(literal "/")`, "(allow process-exec)", "(allow file-read-metadata)",
		`(subpath "/usr")`, `(subpath "/System")`,
		`(subpath "` + inv.Workspace + `")`, `(subpath "` + inv.RunDir + `")`,
		`(literal "/Users/x/.local/share/claude/versions/2.1.141")`, `(subpath "/Users/x/.local/share/claude/versions")`,
		`(subpath "/Users/x/.claude")`, `(subpath "/var/folders/zz/T")`, `(subpath "/private/var/folders/zz/T")`,
		";; host allowed through the proxy: api.anthropic.com", `(remote tcp "localhost:4711")`, `(remote tcp "*:443")`,
		`(literal "/dev/null")`, "run run-sb provider claude role drafter",
	} {
		if !strings.Contains(profile, needle) {
			t.Errorf("profile lacks %q\n%s", needle, profile)
		}
	}
	if strings.Contains(profile, "(allow network*)") {
		t.Error("a non-research run must not allow all network")
	}
	if strings.Contains(profile, `(subpath "/private/tmp")`) || strings.Contains(profile, `(subpath "/tmp")`) {
		t.Error("/private/tmp must not be readable (other programs' scratch, the probe secret)")
	}
	// writes: workspace and run dir only, not the allow dirs
	writes := profile[strings.Index(profile, "(allow file-write*"):]
	writes = writes[:strings.Index(writes, ")\n")+2]
	if strings.Contains(writes, "/Users/x/.claude") || !strings.Contains(writes, inv.Workspace) {
		t.Errorf("writes %s", writes)
	}
	inv.Network = true
	if !strings.Contains(SandboxProfile(inv, "/bin/x", nil, nil), "(allow network*)") {
		t.Error("research run allows network")
	}
	inv.Network = false
	if !strings.Contains(SandboxProfile(inv, "/bin/x", nil, nil), ";; network: none") {
		t.Error("no hosts means no network")
	}
	argv := WrapSandbox([]string{"claude", "-p"}, "/r/sandbox.sb")
	if !equalStrings(argv, []string{SandboxExec, "-f", "/r/sandbox.sb", "claude", "-p"}) {
		t.Errorf("wrap %q", argv)
	}
	// A trailing "*" is a prefix rule: a state file and its atomic-write
	// siblings, never the whole directory.
	prefixed := RenderSandbox(SandboxSpec{Workspace: inv.Workspace, RunDir: inv.RunDir,
		WriteDirs: []string{"/Users/x/.kimi-code/workspaces.json*"}})
	if !strings.Contains(prefixed, `(prefix "/Users/x/.kimi-code/workspaces.json")`) ||
		strings.Contains(prefixed, `(subpath "/Users/x/.kimi-code")`) {
		t.Errorf("prefix rule\n%s", prefixed)
	}
	d := SandboxFor("claude")
	if len(d.ReadDirs) == 0 || strings.Contains(d.ReadDirs[0], "~") {
		t.Errorf("defaults %+v", d)
	}
}

func TestHostProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "upstream ok")
	}))
	defer upstream.Close()
	u, _ := url.Parse(upstream.URL)
	p, err := StartHostProxy([]string{"127.0.0.1", "example.org"})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if !p.Allowed("api.example.org:443") || p.Allowed("example.com") || p.Allowed("notexample.org") {
		t.Error("Allowed")
	}
	proxyURL, _ := url.Parse(p.URL())
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	resp, err := client.Get(upstream.URL + "/x")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(body) != "upstream ok" {
		t.Errorf("forward: %d %q", resp.StatusCode, body)
	}
	resp, err = client.Get("http://denied.test:" + u.Port() + "/x")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden || p.DeniedCount() != 1 {
		t.Errorf("deny: %d denied %d", resp.StatusCode, p.DeniedCount())
	}
	// CONNECT tunnel to the allowed upstream
	req, _ := http.NewRequest(http.MethodConnect, "http://"+u.Host, nil)
	req.URL = &url.URL{Host: u.Host}
	resp, err = client.Transport.RoundTrip(req)
	if err == nil {
		resp.Body.Close()
	}
}

// TestFakeRunSandboxed runs the fake CLI under the real seatbelt profile
// and checks that a write outside the workspace is refused while the
// workspace edit and the handoff land.
func TestFakeRunSandboxed(t *testing.T) {
	if runtime.GOOS != "darwin" || !SandboxAvailable() {
		t.Skip("sandbox-exec is macOS only")
	}
	inv := newInvocation(t, "fake")
	outside := filepath.Join(filepath.Dir(inv.Workspace), "outside")
	os.MkdirAll(outside, 0o755)
	canary := filepath.Join(outside, "canary.txt")
	os.WriteFile(canary, []byte("untouched"), 0o644)
	script := writeScript(t, t.TempDir(), "sandboxed", FakeScript{
		ExitCode: 0,
		Events:   []Event{{Kind: KindUsage, ContextTokens: 50}},
		Edits: map[string]string{
			"scenes/sc-a.md":        "inside\n",
			"../outside/canary.txt": "PROBE\n",
		},
		Handoff: `{"status":"completed"}`,
	})
	out, err := Run(context.Background(), &FakeAdapter{Binary: fakeBin, ScriptPath: script}, inv, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.ExitCode != 0 || out.ExitClass != Ok {
		t.Fatalf("outcome %+v", out)
	}
	if b, _ := os.ReadFile(canary); string(b) != "untouched" {
		t.Errorf("the sandbox let the fake write outside the workspace: %q", b)
	}
	if !strings.Contains(out.StderrTail, "operation not permitted") {
		t.Errorf("expected a permission error on stderr, got %q", out.StderrTail)
	}
	if b, _ := os.ReadFile(filepath.Join(inv.Workspace, "scenes", "sc-a.md")); string(b) != "inside\n" {
		t.Errorf("workspace edit missing: %q", b)
	}
	if b, _ := os.ReadFile(inv.HandoffPath); !strings.Contains(string(b), "completed") {
		t.Errorf("handoff missing: %q", b)
	}
	if _, err := os.Stat(out.ProfilePath); err != nil {
		t.Error("sandbox.sb must be written")
	}
	argv, _ := os.ReadFile(filepath.Join(inv.RunDir, FileArgv))
	if !strings.Contains(string(argv), SandboxExec) {
		t.Errorf("argv.json %s", argv)
	}
}

// TestClaudeTempSlug: claude's Bash tool directory name for a workspace,
// as observed in a live EPERM (spaces and slashes both become "-").
func TestClaudeTempSlug(t *testing.T) {
	ws := "/Users/x/Library/Application Support/bookwriter/example-book/smoke/claude-20261023T023453Z/sandbox/ws"
	want := "-Users-x-Library-Application-Support-bookwriter-example-book-smoke-claude-20261023T023453Z-sandbox-ws"
	if got := ClaudeTempSlug(ws); got != want {
		t.Fatalf("slug %q, want %q", got, want)
	}
	dirs := ClaudeAdapter{}.SandboxWriteDirs(Invocation{Workspace: ws})
	if len(dirs) != 3 || !strings.HasSuffix(dirs[0], "/"+want) || !strings.HasPrefix(dirs[0], "/private/tmp/claude-") || dirs[2] != ClaudeCwdFilePattern {
		t.Fatalf("write dirs %v", dirs)
	}
	profile := RenderSandbox(SandboxSpec{Workspace: ws, RunDir: "/r", WriteDirs: dirs})
	if !strings.Contains(profile, `(regex #"^/(private/)?tmp/claude-[0-9a-f]+-cwd$")`) || strings.Contains(profile, `(subpath "/private/tmp")`) ||
		strings.Contains(profile, `(prefix "/private/tmp/claude-")`) {
		t.Fatalf("profile does not confine /tmp to the claude paths:\n%s", profile)
	}
	// claude 2.1.308 opens /tmp/claude-<uid> at startup: the node is
	// readable, its other directories are not
	d := SandboxFor("claude")
	full := RenderSandbox(SandboxSpec{Workspace: ws, RunDir: "/r", ReadDirs: d.ReadDirs, WriteDirs: dirs})
	if !strings.Contains(full, `(regex #"^/(private/)?tmp/claude-[0-9]+$")`) || strings.Contains(full, `(subpath "/private/tmp/claude-`+fmt.Sprint(os.Getuid())+`")`) {
		t.Fatalf("profile does not grant the /tmp/claude-<uid> node alone:\n%s", full)
	}
}

// TestClaudeCanSaveRefreshedLogin: the credential storage lock and the
// config lock are writable for claude, so a token refresh inside the
// sandbox is saved; ~/.claude.json itself stays read-only.
func TestClaudeCanSaveRefreshedLogin(t *testing.T) {
	d := SandboxFor("claude")
	have := strings.Join(d.WriteDirs, " ")
	for _, want := range []string{ExpandHome("~/.claude/.storage-write.lock*"), ExpandHome("~/.claude.lock*"), ExpandHome("~/Library/Keychains/login.keychain-db*")} {
		if !strings.Contains(have, want) {
			t.Errorf("claude write list lacks %s: %v", want, d.WriteDirs)
		}
	}
	for _, w := range d.WriteDirs {
		if w == ExpandHome("~/.claude.json") || w == ExpandHome("~/.claude.json*") {
			t.Errorf("~/.claude.json must stay read-only, write list has %s", w)
		}
		if w == ExpandHome("~/Library/Keychains") || w == ExpandHome("~/Library/Keychains*") {
			t.Errorf("the keychain directory must stay read-only (only the login keychain is written), write list has %s", w)
		}
	}
}

// TestGrokCanRefreshItsLogin: grok's six-hour token refreshes at auth.x.ai
// and is saved to auth.json, so both stay reachable from the sandbox while
// config.toml stays read-only.
func TestGrokCanRefreshItsLogin(t *testing.T) {
	d := SandboxFor("grok")
	if !strings.Contains(" "+strings.Join(d.Hosts, " ")+" ", " auth.x.ai ") {
		t.Errorf("grok hosts lack auth.x.ai: %v", d.Hosts)
	}
	have := strings.Join(d.WriteDirs, " ")
	if !strings.Contains(have, ExpandHome("~/.grok/auth.json*")) {
		t.Errorf("grok write list lacks auth.json: %v", d.WriteDirs)
	}
	if strings.Contains(have, "config.toml") {
		t.Errorf("grok config.toml must stay read-only: %v", d.WriteDirs)
	}
}

// agy serves its agent on a random 127.0.0.1 port and connects back to it;
// only a LocalListen profile allows that, and outside hosts still go only
// through the proxy.
func TestAgyLocalListenAndHosts(t *testing.T) {
	d := SandboxFor("agy")
	if !d.LocalListen {
		t.Fatal("agy needs LocalListen")
	}
	on := RenderSandbox(SandboxSpec{Workspace: "/w", RunDir: "/r", NetworkHosts: []string{"localhost:5000"}, LocalListen: true})
	for _, want := range []string{`(allow network-bind (local ip "localhost:*"))`, `(allow network-inbound (local ip "localhost:*"))`, `(remote tcp "localhost:*")`} {
		if !strings.Contains(on, want) {
			t.Errorf("LocalListen profile lacks %s", want)
		}
	}
	if strings.Contains(on, `(remote tcp "*:`) || strings.Contains(on, "(allow network*)") {
		t.Error("LocalListen opened outside connections")
	}
	off := RenderSandbox(SandboxSpec{Workspace: "/w", RunDir: "/r", NetworkHosts: []string{"localhost:5000"}})
	if strings.Contains(off, "network-bind") || strings.Contains(off, `"localhost:*"`) {
		t.Error("a profile without LocalListen allows a local server")
	}
	for _, h := range []string{"daily-cloudcode-pa.googleapis.com", "oauth2.googleapis.com", "www.googleapis.com"} {
		found := false
		for _, x := range d.Hosts {
			found = found || x == h
		}
		if !found {
			t.Errorf("agy host %s missing", h)
		}
	}
	for _, w := range d.WriteDirs {
		for _, ro := range []string{"/settings.json", "/bin", "/builtin", "/updater", "/plugin_data"} {
			if strings.HasSuffix(w, ro) {
				t.Errorf("agy may write %s", w)
			}
		}
	}
}
