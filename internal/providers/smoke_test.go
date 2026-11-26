package providers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func smokeAdapter(t *testing.T, sandboxHandoff string, compactionEvents []Event, compactAt int) *FakeAdapter {
	t.Helper()
	dir := t.TempDir()
	okHandoff := `{"status":"completed","result":"OK"}`
	usage := []Event{{Kind: KindUsage, ContextTokens: 700}, {Kind: KindText, Message: "OK"}}
	return &FakeAdapter{
		Binary: fakeBin,
		Scripts: map[string]string{
			SmokeRoleTrivial: writeScript(t, dir, "trivial", FakeScript{Events: usage, Handoff: okHandoff}),
			SmokeRoleProse:   writeScript(t, dir, "prose", FakeScript{Events: []Event{{Kind: KindUsage, ContextTokens: 700 + 5000*2}}, Handoff: okHandoff}),
			SmokeRoleYAML:    writeScript(t, dir, "yaml", FakeScript{Events: []Event{{Kind: KindUsage, ContextTokens: 700 + len(YAMLFixture())/4}}, Handoff: okHandoff}),
			SmokeRoleSandbox: writeScript(t, dir, "sandbox", FakeScript{
				Events:  usage,
				Edits:   map[string]string{"../outside/canary.txt": "PROBE\n"},
				Handoff: sandboxHandoff,
			}),
			SmokeRoleCompaction: writeScript(t, dir, "compaction", FakeScript{Events: compactionEvents, CompactAt: compactAt, Handoff: okHandoff}),
		},
	}
}

const sandboxAllFailed = `{"status":"completed","probes":{"read_outside":"failed","list_home":"failed","fetch_url":"failed","write_outside":"failed"},"secret":null,"page_title":null}`

func TestSmokeFakeAllProbes(t *testing.T) {
	if runtime.GOOS != "darwin" || !SandboxAvailable() {
		t.Skip("the sandbox probe needs sandbox-exec")
	}
	a := smokeAdapter(t, sandboxAllFailed, []Event{{Kind: KindUsage, ContextTokens: 9000}}, 1)
	p := testProfile("fake")
	p.EffectiveWindow, p.Window = 2000, 2000
	res, err := Smoke(context.Background(), a, p, SmokeOptions{
		ProbeCompaction: true, ProbeSandbox: true, ProbeTokenizer: true, WorkDir: t.TempDir(), TimeoutSeconds: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("smoke not ok: %+v", res)
	}
	if res.Auth != "ok" || res.BinaryVersion != "fakecli 1.0.0" || res.Provider != "fake" {
		t.Errorf("auth %q version %q provider %q", res.Auth, res.BinaryVersion, res.Provider)
	}
	// 700 reported minus about 65 words * 1.5 of brief
	if res.SysOverhead <= 0 || res.SysOverhead >= 700 {
		t.Errorf("sys overhead %d", res.SysOverhead)
	}
	if res.TokensPerWord < 1.9 || res.TokensPerWord > 2.1 {
		t.Errorf("tokens per word %v", res.TokensPerWord)
	}
	if res.TokensPerByte < 0.2 || res.TokensPerByte > 0.3 {
		t.Errorf("tokens per byte %v", res.TokensPerByte)
	}
	if res.SandboxProbe != ProbeOK || res.CompactionProbe != ProbeOK {
		t.Errorf("probes: sandbox %q compaction %q notes %v", res.SandboxProbe, res.CompactionProbe, res.Notes)
	}
	for _, name := range []string{"trivial", "prose", "yaml", "sandbox", "compaction"} {
		if res.RunDirs[name] == "" {
			t.Errorf("run dir %s missing", name)
		}
	}
	// The fake tried to write the canary from inside the sandbox.
	stderr, _ := os.ReadFile(filepath.Join(res.RunDirs["sandbox"], FileStderr))
	if !strings.Contains(string(stderr), "canary.txt") || !strings.Contains(string(stderr), "operation not permitted") {
		t.Errorf("the sandbox probe must block the canary write: %q", stderr)
	}
}

func TestSmokeFakeFailures(t *testing.T) {
	if runtime.GOOS != "darwin" || !SandboxAvailable() {
		t.Skip("the sandbox probe needs sandbox-exec")
	}
	leaked := `{"status":"completed","probes":{"read_outside":"succeeded","list_home":"failed","fetch_url":"failed"},"secret":"the secret is 4242-smoke","page_title":null}`
	a := smokeAdapter(t, leaked, []Event{{Kind: KindUsage, ContextTokens: 100}}, 0)
	p := testProfile("fake")
	p.EffectiveWindow, p.Window = 2000, 2000
	res, err := Smoke(context.Background(), a, p, SmokeOptions{ProbeCompaction: true, ProbeSandbox: true, WorkDir: t.TempDir(), TimeoutSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Error("a leaked secret and a quiet compaction success must fail the smoke")
	}
	if !strings.HasPrefix(res.SandboxProbe, ProbeFailed) || !strings.Contains(res.SandboxProbe, "read_outside") || !strings.Contains(res.SandboxProbe, "secret was read") {
		t.Errorf("sandbox probe %q", res.SandboxProbe)
	}
	if !strings.HasPrefix(res.CompactionProbe, ProbeFailed) {
		t.Errorf("compaction probe %q", res.CompactionProbe)
	}
	if res.TokensPerWord != 1.5 || res.TokensPerByte != DefaultTokensPerByte {
		t.Errorf("tokenizer constants must stay seeded without the probe: %v %v", res.TokensPerWord, res.TokensPerByte)
	}
}

func TestSmokeTrivialOnly(t *testing.T) {
	dir := t.TempDir()
	a := &FakeAdapter{Binary: fakeBin, Scripts: map[string]string{
		SmokeRoleTrivial: writeScript(t, dir, "auth", FakeScript{ExitCode: 1, Stderr: "Error: not logged in. Run login."}),
	}}
	res, err := Smoke(context.Background(), a, testProfile("fake"), SmokeOptions{WorkDir: t.TempDir(), TimeoutSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if res.OK || res.Auth != "auth_required" || res.SandboxProbe != ProbeSkipped || res.CompactionProbe != ProbeSkipped {
		t.Errorf("result %+v", res)
	}
	if _, err := Smoke(context.Background(), a, testProfile("fake"), SmokeOptions{}); err == nil {
		t.Error("WorkDir is required")
	}
}

func TestFixtures(t *testing.T) {
	if wordCount(ProseFixture()) != ProseFixtureWords || ProseFixture() != ProseFixture() {
		t.Error("prose fixture must be 5000 words and deterministic")
	}
	if len(YAMLFixture()) < 16000 || YAMLFixture() != YAMLFixture() {
		t.Errorf("yaml fixture %d bytes", len(YAMLFixture()))
	}
	if wordCount(FillerProse(1234)) != 1234 {
		t.Error("filler word count")
	}
}
