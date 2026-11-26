package providers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mustEnv(t *testing.T, env []string, key, want string) {
	t.Helper()
	got, ok := envValue(env, key)
	if !ok || got != want {
		t.Errorf("env %s: got %q (present %v) want %q", key, got, ok, want)
	}
}

func mustNoEnv(t *testing.T, env []string, key string) {
	t.Helper()
	if v, ok := envValue(env, key); ok {
		t.Errorf("env %s must be absent, got %q", key, v)
	}
}

// TestClaudeJSONSchema: the schema claude gets drops the top-level
// "$schema" (Claude Code 2.1.308 knows no draft 2020-12 meta-schema) and
// the top-level combinators (the API refuses them in a tool input_schema),
// keeps everything else, and leaves a schema without them byte for byte.
func TestClaudeJSONSchema(t *testing.T) {
	full := `{"$id":"https://bookwriter.local/schemas/handoff.v1.json","$schema":"https://json-schema.org/draft/2020-12/schema",` +
		`"type":"object","required":["status"],"properties":{"status":{"enum":["completed","partial"]},` +
		`"units":{"allOf":[{"type":"array"}]}},"allOf":[{"if":{"properties":{"status":{"const":"partial"}}},"then":{"required":["units"]}}]}`
	var got map[string]any
	if err := json.Unmarshal([]byte(claudeJSONSchema(full)), &got); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"$schema", "allOf", "anyOf", "oneOf"} {
		if _, ok := got[k]; ok {
			t.Errorf("top-level %s kept: %v", k, got)
		}
	}
	props, _ := got["properties"].(map[string]any)
	units, _ := props["units"].(map[string]any)
	if got["$id"] == nil || got["type"] != "object" || props["status"] == nil || units["allOf"] == nil {
		t.Errorf("schema lost more than the top-level keys: %v", got)
	}
	for _, s := range []string{`{"type":"object"}`, `not json`, `[1,2]`} {
		if out := claudeJSONSchema(s); out != s {
			t.Errorf("claudeJSONSchema(%q) = %q, want it unchanged", s, out)
		}
	}
}

func TestClaudeArgv(t *testing.T) {
	inv := newInvocation(t, "claude")
	inv.Profile.Model = "claude-opus-4-7"
	inv.Profile.Window, inv.Profile.EffectiveWindow = 200_000, 200_000
	inv.MaxTurns = 40
	schema := filepath.Join(inv.Workspace, ".cx", "handoff.schema.json")
	os.WriteFile(schema, []byte(`{"type":"object"}`+"\n"), 0o644)
	inv.SchemaPath = schema
	t.Setenv("ANTHROPIC_API_KEY", "must-not-leak")

	argv, env, stdin, err := ClaudeAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"claude", "-p", "--permission-mode", "bypassPermissions", "--model", "claude-opus-4-7",
		"--effort", "high", "--output-format", "stream-json", "--verbose", "--no-session-persistence",
		"--max-budget-usd", "6", "--setting-sources", "project", "--disallowedTools", "WebSearch WebFetch",
		"--max-turns", "40", "--json-schema", `{"type":"object"}`}
	if !equalStrings(argv, want) {
		t.Errorf("argv\n got %q\nwant %q", argv, want)
	}
	if string(stdin) != "# Brief\n\nDo the thing.\n" {
		t.Errorf("stdin %q", stdin)
	}
	mustEnv(t, env, "DISABLE_COMPACT", "1")
	mustEnv(t, env, "DISABLE_AUTO_COMPACT", "1")
	mustEnv(t, env, "CLAUDE_CODE_MAX_CONTEXT_TOKENS", "180000")
	mustEnv(t, env, "CLAUDE_CODE_MAX_OUTPUT_TOKENS", "128000")
	mustEnv(t, env, "TMPDIR", filepath.Join(inv.RunDir, "tmp"))
	mustNoEnv(t, env, "ANTHROPIC_API_KEY")
	if _, ok := envValue(env, "PATH"); !ok {
		t.Error("PATH must be inherited")
	}

	// a one-shot call with an enforced schema gets no tools: the answer
	// leaves through StructuredOutput, which --tools "" keeps
	inv.OneShot = true
	argv, _, _, _ = ClaudeAdapter{}.BuildCommand(inv)
	if i := indexOf(argv, "--tools"); i < 0 || i+1 >= len(argv) || argv[i+1] != "" {
		t.Errorf("one-shot argv %q lacks --tools \"\"", argv)
	}
	// without a schema it keeps only the Write tool for the handoff
	saved := inv.SchemaPath
	inv.SchemaPath = ""
	argv, _, _, _ = ClaudeAdapter{}.BuildCommand(inv)
	if j := strings.Join(argv, " "); !strings.Contains(j, "--tools Write") {
		t.Errorf("one-shot argv without a schema %q lacks --tools Write", argv)
	}
	inv.SchemaPath, inv.OneShot = saved, false

	inv.Network = true
	inv.SchemaPath = ""
	inv.MaxTurns = 0
	argv, _, _, _ = ClaudeAdapter{Binary: "/opt/claude"}.BuildCommand(inv)
	if argv[0] != "/opt/claude" || strings.Contains(strings.Join(argv, " "), "disallowedTools") || strings.Contains(strings.Join(argv, " "), "max-turns") {
		t.Errorf("network argv %q", argv)
	}
	settings := ClaudeAdapter{}.ProjectSettings(inv)
	if !strings.Contains(string(settings[".claude/settings.json"]), `"autoCompactEnabled": false`) {
		t.Errorf("settings %s", settings[".claude/settings.json"])
	}
	if (ClaudeAdapter{}).InstructionFile() != "CLAUDE.md" {
		t.Error("instruction file")
	}
}

func TestCodexArgv(t *testing.T) {
	inv := newInvocation(t, "codex")
	inv.Profile.Model = "gpt-6.1-sol"
	inv.Profile.Window, inv.Profile.EffectiveWindow = 500_000, 500_000
	inv.StrictSchemaPath = filepath.Join(inv.Workspace, ".cx", "handoff.strict.json")
	argv, env, stdin, err := CodexAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"codex", "exec", "-C", inv.Workspace, "--skip-git-repo-check", "--ephemeral",
		"--sandbox", "workspace-write", "-c", `approval_policy="never"`, "-m", "gpt-6.1-sol",
		"-c", `model_reasoning_effort="high"`, "-c", "model_auto_compact_token_limit=450000",
		"-c", `web_search="disabled"`, "-c", "features.memories=false"}
	want = append(want, codexProviderArgs...)
	for _, f := range CodexDisabledFeatures {
		want = append(want, "--disable", f)
	}
	want = append(want, "--output-schema", inv.StrictSchemaPath, "--output-last-message", inv.HandoffPath, "--json", "-")
	if !equalStrings(argv, want) {
		t.Errorf("argv\n got %q\nwant %q", argv, want)
	}
	mustEnv(t, env, "CODEX_HOME", filepath.Join(inv.RunDir, CodexHomeDir))
	mustEnv(t, env, "TMPPREFIX", filepath.Join(inv.RunDir, "tmp", "zsh"))
	// Under the seatbelt profile codex's own sandbox is off (seatbelt
	// cannot nest); the outer profile confines it.
	sandboxed := inv
	sandboxed.Sandboxed = true
	sargv, _, _, _ := CodexAdapter{}.BuildCommand(sandboxed)
	if !strings.Contains(strings.Join(sargv, " "), "--sandbox danger-full-access") {
		t.Errorf("sandboxed argv %q", sargv)
	}
	if len(stdin) == 0 {
		t.Error("codex takes the brief on stdin")
	}
	inv.Network = true
	inv.StrictSchemaPath = ""
	inv.SchemaPath = "/x/schema.json"
	argv, _, _, _ = CodexAdapter{}.BuildCommand(inv)
	j := strings.Join(argv, " ")
	// Without a strict form codex gets no --output-schema: the plain schema
	// carries a top-level allOf the API refuses with a 400.
	if !strings.Contains(j, `web_search="live"`) || strings.Contains(j, "--output-schema") {
		t.Errorf("network argv %q", argv)
	}
	if (CodexAdapter{}).InstructionFile() != "AGENTS.md" || (CodexAdapter{}).ProjectSettings(inv) != nil {
		t.Error("codex instruction file or settings")
	}
}

func TestGeminiArgv(t *testing.T) {
	inv := newInvocation(t, "gemini")
	inv.Profile.Model = "gemini-3-pro-preview"
	inv.Profile.Effort = ""
	inv.Profile.Window, inv.Profile.EffectiveWindow = 1_000_000, 500_000
	t.Setenv("GEMINI_API_KEY", "gk-test")
	argv, env, stdin, err := GeminiAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"gemini", "-p", "Follow the brief above exactly.", "--approval-mode", "yolo", "--skip-trust",
		"-m", "gemini-3-pro-preview", "--output-format", "stream-json"}
	if !equalStrings(argv, want) {
		t.Errorf("argv\n got %q\nwant %q", argv, want)
	}
	if len(stdin) == 0 {
		t.Error("gemini takes the brief on stdin")
	}
	mustEnv(t, env, "GEMINI_API_KEY", "gk-test")
	mustEnv(t, env, "GEMINI_CLI_HOME", filepath.Join(inv.RunDir, GeminiHomeDir))
	settings := string(GeminiAdapter{}.ProjectSettings(inv)[".gemini/settings.json"])
	for _, needle := range []string{`"compressionThreshold": 1`, `"PreCompress"`, CompactionMarkerPath(inv), `"maxTokens": 500000`, `"google_web_search"`} {
		if !strings.Contains(settings, needle) {
			t.Errorf("settings lack %q:\n%s", needle, settings)
		}
	}
	inv.Network = true
	if strings.Contains(string((GeminiAdapter{}).ProjectSettings(inv)[".gemini/settings.json"]), "exclude") {
		t.Error("research runs keep the web tools")
	}
	if (GeminiAdapter{}).InstructionFile() != "GEMINI.md" {
		t.Error("instruction file")
	}
}

func TestKimiArgv(t *testing.T) {
	inv := newInvocation(t, "kimi")
	inv.Profile.Model = "kimi-code/k3"
	argv, _, stdin, err := KimiAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"kimi", "-p", KimiPointer, "-m", "kimi-code/k3", "--output-format", "stream-json",
		"--skills-dir", filepath.Join(inv.RunDir, KimiSkillsDir)}
	if !equalStrings(argv, want) {
		t.Errorf("argv\n got %q\nwant %q", argv, want)
	}
	if stdin != nil {
		t.Error("kimi has no stdin")
	}
	inv.BriefPath = filepath.Join(t.TempDir(), "elsewhere.md")
	os.WriteFile(inv.BriefPath, []byte("x\n"), 0o644)
	if _, _, _, err := (KimiAdapter{}).BuildCommand(inv); err == nil {
		t.Error("a brief outside the workspace must be refused")
	}
}

func TestGrokArgv(t *testing.T) {
	inv := newInvocation(t, "grok")
	inv.Profile.Model = "grok-4.5"
	inv.Profile.Window, inv.Profile.EffectiveWindow = 128_000, 128_000
	schema := filepath.Join(inv.Workspace, "s.json")
	os.WriteFile(schema, []byte(`{"a":1}`), 0o644)
	inv.SchemaPath = schema
	argv, env, stdin, err := GrokAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"grok", "--prompt-file", inv.BriefPath, "--cwd", inv.Workspace, "--always-approve",
		"-m", "grok-4.5", "--effort", "high", "--max-turns", "60", "--disable-web-search",
		"--json-schema", `{"a":1}`, "--no-memory", "--output-format", "streaming-json"}
	if !equalStrings(argv, want) {
		t.Errorf("argv\n got %q\nwant %q", argv, want)
	}
	if stdin != nil {
		t.Error("grok has no stdin")
	}
	mustEnv(t, env, "GROK_AUTO_COMPACT_THRESHOLD_PERCENT", "100")
}

func TestMuseArgv(t *testing.T) {
	inv := newInvocation(t, "muse")
	inv.Profile.Model = "muse-spark-1.3"
	inv.Profile.Window, inv.Profile.EffectiveWindow = 128_000, 128_000
	inv.SchemaPath = "/x/handoff.schema.json"
	inv.StrictSchemaPath = "/x/handoff.strict.json"
	inv.MaxTurns = 150
	argv, _, stdin, err := MuseAdapter{}.BuildCommand(inv)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"muse", "exec", "--prompt-file", inv.BriefPath, "--yolo", "--workspace", inv.Workspace,
		"--model", "muse-spark-1.3", "--reasoning-effort", "high",
		"--context-compaction-soft-threshold", "0.998", "--context-compaction-hard-threshold", "0.999",
		"-c", "context_compaction.provider_context_limit_tokens=115200",
		"--max-model-steps", "150", "--disable-web-tools", "--output-schema", "/x/handoff.strict.json",
		"--no-session-log", "--no-foreign-personal-context", "--json"}
	if !equalStrings(argv, want) {
		t.Errorf("argv\n got %q\nwant %q", argv, want)
	}
	if stdin != nil {
		t.Error("muse has no stdin")
	}
}

func TestBuildCommandErrors(t *testing.T) {
	inv := newInvocation(t, "claude")
	inv.BriefPath = ""
	for _, a := range []Adapter{ClaudeAdapter{}, CodexAdapter{}, GeminiAdapter{}, KimiAdapter{}, GrokAdapter{}, MuseAdapter{}} {
		if _, _, _, err := a.BuildCommand(inv); err == nil {
			t.Errorf("%s: empty brief must fail", a.Provider())
		}
	}
	for _, name := range Providers {
		if _, err := NewAdapter(name); err != nil {
			t.Error(err)
		}
	}
	if _, err := NewAdapter("nope"); err == nil {
		t.Error("unknown provider")
	}
}

// TestCodexRunHome: the per-run CODEX_HOME holds a symlink to the user's
// auth.json, and a refresh that replaced the symlink is carried back.
func TestCodexRunHome(t *testing.T) {
	userHome := t.TempDir()
	t.Setenv("CODEX_HOME", userHome)
	real := filepath.Join(userHome, "auth.json")
	os.WriteFile(real, []byte(`{"tokens":"old"}`), 0o600)
	old := time.Now().Add(-time.Hour)
	os.Chtimes(real, old, old)
	inv := newInvocation(t, "codex")
	a := CodexAdapter{}
	if err := a.PrepareRun(inv); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(inv.RunDir, CodexHomeDir, "auth.json")
	if target, err := os.Readlink(link); err != nil || target != real {
		t.Fatalf("link %q %v", target, err)
	}
	if err := a.PrepareRun(inv); err != nil {
		t.Fatal("PrepareRun must be idempotent:", err)
	}
	a.FinishRun(inv)
	if b, _ := os.ReadFile(real); string(b) != `{"tokens":"old"}` {
		t.Errorf("an untouched link must leave auth.json alone: %s", b)
	}
	os.Remove(link)
	os.WriteFile(link, []byte(`{"tokens":"new"}`), 0o600)
	a.FinishRun(inv)
	if b, _ := os.ReadFile(real); string(b) != `{"tokens":"new"}` {
		t.Errorf("a refreshed login must be carried back: %s", b)
	}
	// No login: no link, no error.
	t.Setenv("CODEX_HOME", t.TempDir())
	inv2 := newInvocation(t, "codex")
	if err := a.PrepareRun(inv2); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(inv2.RunDir, CodexHomeDir, "auth.json")); err == nil {
		t.Error("no auth.json means no link")
	}
}

func indexOf(xs []string, x string) int {
	for i, v := range xs {
		if v == x {
			return i
		}
	}
	return -1
}
