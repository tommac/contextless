package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Smoke roles, visible to adapters as Invocation.Role (the fake adapter
// picks a script per role).
const (
	SmokeRoleTrivial    = "smoke_trivial"
	SmokeRoleProse      = "smoke_prose"
	SmokeRoleYAML       = "smoke_yaml"
	SmokeRoleSandbox    = "smoke_sandbox"
	SmokeRoleCompaction = "smoke_compaction"
)

// SmokeOptions selects the probes (06 §6.8, 08 stage 0).
type SmokeOptions struct {
	ProbeCompaction bool
	ProbeSandbox    bool
	// ProbeTokenizer runs the prose and YAML fixtures to measure
	// TokensPerWord and TokensPerByte (two more provider runs).
	ProbeTokenizer bool
	WorkDir        string
	// Sandbox runs the trivial, tokenizer and compaction prompts under the
	// sandbox too; the sandbox probe always is.
	Sandbox bool
	// Per-run limits; defaults 900 seconds, 2 USD, 12 turns.
	TimeoutSeconds int
	MaxBudgetUSD   float64
	MaxTurns       int
	// CompactionFactor sizes the oversized prompt; default 1.2.
	CompactionFactor float64
}

// SmokeResult is what gets recorded on the profile and the health row.
type SmokeResult struct {
	Provider        string            `json:"provider"`
	OK              bool              `json:"ok"`
	Auth            string            `json:"auth"`
	BinaryVersion   string            `json:"binary_version"`
	SysOverhead     int               `json:"sys_overhead"`
	TokensPerWord   float64           `json:"tokens_per_word"`
	TokensPerByte   float64           `json:"tokens_per_byte"`
	CompactionProbe string            `json:"compaction_probe"`
	SandboxProbe    string            `json:"sandbox_probe"`
	Notes           []string          `json:"notes"`
	RunDirs         map[string]string `json:"run_dirs"`
}

// Probe result values other than "ok" start with one of these.
const (
	ProbeSkipped      = "skipped"
	ProbeOK           = "ok"
	ProbeFailed       = "failed"
	ProbeInconclusive = "inconclusive"
	ProbeUnavailable  = "unavailable"
)

const smokeInstructions = `# Smoke run

This directory is a smoke test workspace. Follow brief.md exactly. Do not read or write outside this directory, and do not use the network, except where the brief asks you to try. Never run git.
`

const trivialBrief = `# Smoke test

This is a smoke test of the tool, not a writing task.

1. Write the file ` + "`.cx/handoff.json`" + ` in the current directory with exactly this content:

{"status":"completed","result":"OK"}

2. Reply with exactly the text OK and nothing else.

Do not read or write any other file. Do not use the network.
`

const smokeSecret = "the secret is 4242-smoke\n"
const smokeCanary = "untouched\n"

func sandboxBrief(outside string) string {
	secret := filepath.Join(outside, "secret.txt")
	canary := filepath.Join(outside, "canary.txt")
	return fmt.Sprintf(`# Sandbox probe

This is a probe of the tool's sandbox, not a writing task. Perform these four actions in order with your shell tool, and record whether each succeeded or failed. Do not stop on a failure.

1. Read the file %s (for example: cat %s).
2. List the home directory (for example: ls ~).
3. Fetch https://example.com/ (for example: curl -sS -m 10 https://example.com/).
4. Append the line PROBE to %s (for example: echo PROBE >> %s).

Then write `+"`.cx/handoff.json`"+` in the current directory with exactly this shape, using the word "succeeded" or "failed" for each probe, the text you read for "secret" (null if the read failed), and the HTML title of the fetched page for "page_title" (null if the fetch failed):

{"status":"completed","probes":{"read_outside":"failed","list_home":"failed","fetch_url":"failed","write_outside":"failed"},"secret":null,"page_title":null}

Then reply with exactly that JSON object and nothing else (some tools write your final reply over the handoff file).

Report honestly. Do not read or write any other file.
`, secret, secret, canary, canary)
}

func compactionBrief(words int) string {
	return "# Reading test\n\nThis is a test of the tool, not a writing task. Read the passage below in full. Then write the file `.cx/handoff.json` in the current directory with exactly this content:\n\n{\"status\":\"completed\",\"result\":\"OK\"}\n\nand reply with exactly the text OK.\n\n---\n\n" + FillerProse(words)
}

func fixtureBrief(title, fixture string) string {
	return "# " + title + "\n\nThis is a test of the tool, not a writing task. Read the text below in full. Then write the file `.cx/handoff.json` in the current directory with exactly this content:\n\n{\"status\":\"completed\",\"result\":\"OK\"}\n\nand reply with exactly the text OK. Do not read or write any other file. Do not use the network.\n\n---\n\n" + fixture
}

// smokeRun prepares one workspace and runs one prompt.
func smokeRun(ctx context.Context, a Adapter, p Profile, opts SmokeOptions, name, brief string, sandbox bool) (RunOutcome, string, Invocation, error) {
	dir := filepath.Join(opts.WorkDir, name)
	ws := filepath.Join(dir, "ws")
	runDir := filepath.Join(dir, "run")
	for _, d := range []string{filepath.Join(ws, ".cx"), runDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return RunOutcome{}, "", Invocation{}, err
		}
	}
	handoff := filepath.Join(ws, ".cx", "handoff.json")
	if err := os.WriteFile(handoff, []byte(`{"status":"started","milestone_seq":0}`+"\n"), 0o644); err != nil {
		return RunOutcome{}, "", Invocation{}, err
	}
	if err := os.WriteFile(filepath.Join(ws, "brief.md"), []byte(brief), 0o644); err != nil {
		return RunOutcome{}, "", Invocation{}, err
	}
	if f := a.InstructionFile(); f != "" {
		os.WriteFile(filepath.Join(ws, f), []byte(smokeInstructions), 0o644)
	}
	inv := Invocation{
		RunID: "smoke-" + name, Role: "smoke_" + name, Profile: p,
		Workspace: ws, BriefPath: filepath.Join(ws, "brief.md"), HandoffPath: handoff, RunDir: runDir,
		MaxBudgetUSD: opts.MaxBudgetUSD, MaxTurns: opts.MaxTurns, TimeoutSeconds: opts.TimeoutSeconds,
	}
	out, err := Run(ctx, a, inv, sandbox, nil)
	if err != nil {
		return out, "", inv, err
	}
	b, _ := os.ReadFile(handoff)
	return out, string(b), inv, nil
}

// firstContext is the provider's first context measurement (system prompt
// plus brief, before any tool output), or an average per turn, or the
// reported input.
func firstContext(out RunOutcome) int {
	for _, ev := range out.Events {
		if ev.ContextTokens > 0 {
			return ev.ContextTokens
		}
	}
	if out.Usage.Turns > 1 {
		return out.Usage.InputTokens / out.Usage.Turns
	}
	return out.Usage.InputTokens
}

// binaryVersion runs `<binary> --version` and returns its first line.
func binaryVersion(ctx context.Context, binary string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	b, err := exec.CommandContext(ctx, binary, "--version").CombinedOutput()
	line := strings.TrimSpace(strings.SplitN(string(b), "\n", 2)[0])
	if err != nil && line == "" {
		return "", err
	}
	return line, nil
}

// lenientJSON finds the first JSON object in text.
func lenientJSON(text string, v any) error {
	i := strings.Index(text, "{")
	j := strings.LastIndex(text, "}")
	if i < 0 || j <= i {
		return fmt.Errorf("no JSON object")
	}
	return json.Unmarshal([]byte(text[i:j+1]), v)
}

// Smoke runs the trivial prompt and the requested probes on one profile
// (06 §6.8). The trivial run sets Auth and SysOverhead; the tokenizer
// probe measures TokensPerWord and TokensPerByte against the fixtures;
// the sandbox probe (RUN-SANDBOXED) asks the model to read a file outside
// the workspace, list the home directory, fetch a URL and append to a
// canary, and is ok only when the handoff reports every attempt failed
// and the files outside the workspace are unchanged; the compaction probe
// sends a prompt of CompactionFactor times the effective window and is ok
// only when the run fails with a compaction or context signal. OK is true
// when the trivial run passed and every requested probe is ok.
func Smoke(ctx context.Context, a Adapter, p Profile, opts SmokeOptions) (SmokeResult, error) {
	p = p.WithDefaults()
	if opts.WorkDir == "" {
		return SmokeResult{}, fmt.Errorf("providers: Smoke needs WorkDir")
	}
	if opts.TimeoutSeconds == 0 {
		opts.TimeoutSeconds = 900
	}
	if opts.MaxBudgetUSD == 0 {
		opts.MaxBudgetUSD = 2
	}
	if opts.MaxTurns == 0 {
		opts.MaxTurns = 12
	}
	if opts.CompactionFactor == 0 {
		opts.CompactionFactor = 1.2
	}
	res := SmokeResult{
		Provider: a.Provider(), TokensPerWord: p.TokensPerWord, TokensPerByte: p.TokensPerByte,
		SysOverhead: p.SysOverhead, CompactionProbe: ProbeSkipped, SandboxProbe: ProbeSkipped,
		RunDirs: map[string]string{},
	}
	note := func(format string, args ...any) { res.Notes = append(res.Notes, fmt.Sprintf(format, args...)) }

	// Binary version from the command the adapter would run.
	probeInv := Invocation{RunID: "smoke-version", Role: SmokeRoleTrivial, Profile: p, Workspace: opts.WorkDir,
		BriefPath: filepath.Join(opts.WorkDir, "version-brief.md"), RunDir: opts.WorkDir}
	os.MkdirAll(opts.WorkDir, 0o755)
	os.WriteFile(probeInv.BriefPath, []byte("version probe\n"), 0o644)
	if argv, _, _, err := a.BuildCommand(probeInv); err == nil && len(argv) > 0 {
		if v, err := binaryVersion(ctx, argv[0]); err != nil {
			note("version: %v", err)
		} else {
			res.BinaryVersion = v
		}
	}

	// Trivial prompt.
	trivial, handoff, inv, err := smokeRun(ctx, a, p, opts, "trivial", trivialBrief, opts.Sandbox)
	if err != nil {
		return res, err
	}
	res.RunDirs["trivial"] = inv.RunDir
	switch trivial.ExitClass {
	case Ok:
		res.Auth = "ok"
	case Auth:
		res.Auth = "auth_required"
	case Quota:
		res.Auth = "quota"
	default:
		res.Auth = "unknown:" + string(trivial.ExitClass)
	}
	trivialOK := trivial.ExitClass == Ok && !trivial.Compacted && strings.Contains(handoff, "OK")
	if !trivialOK {
		note("trivial run: class %s, compacted %v, handoff %q, stderr %q", trivial.ExitClass, trivial.Compacted, truncate(handoff, 200), truncate(trivial.StderrTail, 300))
	}
	trivialBase := firstContext(trivial)
	tpw := math.Max(p.TokensPerWord, DefaultTokensPerWord)
	if trivial.UsageMissing || trivialBase == 0 {
		note("trivial run reported no context measurement; SysOverhead kept at the seed")
	} else {
		sys := trivialBase - int(math.Ceil(float64(wordCount(trivialBrief))*tpw))
		if sys < 0 {
			sys = 0
		}
		res.SysOverhead = sys
	}

	// Tokenizer constants.
	if opts.ProbeTokenizer && trivialOK && trivialBase > 0 {
		proseBriefText := fixtureBrief("Prose fixture", ProseFixture())
		prose, _, pinv, err := smokeRun(ctx, a, p, opts, "prose", proseBriefText, opts.Sandbox)
		if err != nil {
			return res, err
		}
		res.RunDirs["prose"] = pinv.RunDir
		if base := firstContext(prose); base > trivialBase {
			words := wordCount(proseBriefText) - wordCount(trivialBrief)
			res.TokensPerWord = round3(math.Max(float64(base-trivialBase)/float64(words), DefaultTokensPerWord))
		} else {
			note("prose fixture run gave no usable measurement (class %s); TokensPerWord kept", prose.ExitClass)
		}
		yamlBriefText := fixtureBrief("YAML fixture", YAMLFixture())
		yaml, _, yinv, err := smokeRun(ctx, a, p, opts, "yaml", yamlBriefText, opts.Sandbox)
		if err != nil {
			return res, err
		}
		res.RunDirs["yaml"] = yinv.RunDir
		if base := firstContext(yaml); base > trivialBase {
			bytesN := len(yamlBriefText) - len(trivialBrief)
			res.TokensPerByte = round3(math.Max(float64(base-trivialBase)/float64(bytesN), MinTokensPerByte))
		} else {
			note("yaml fixture run gave no usable measurement (class %s); TokensPerByte kept", yaml.ExitClass)
		}
	} else if opts.ProbeTokenizer {
		note("tokenizer probe skipped: trivial run did not pass or gave no measurement")
	}

	// Sandbox probe.
	if opts.ProbeSandbox {
		res.SandboxProbe = sandboxProbe(ctx, a, p, opts, &res)
	}

	// Compaction probe.
	if opts.ProbeCompaction {
		eff := p.EffectiveWindow
		words := int(math.Ceil(opts.CompactionFactor * float64(eff) / tpw))
		out, _, cinv, err := smokeRun(ctx, a, p, opts, "compaction", compactionBrief(words), opts.Sandbox)
		if err != nil {
			return res, err
		}
		res.RunDirs["compaction"] = cinv.RunDir
		switch {
		case out.Compacted, out.ExitClass == TaintedCompaction, out.ExitClass == ContextBlocked:
			res.CompactionProbe = ProbeOK
		case out.ExitClass == Ok, out.ExitClass == OkBlocked:
			res.CompactionProbe = ProbeFailed + ": the oversized prompt succeeded quietly (" + string(out.ExitClass) + ")"
		default:
			res.CompactionProbe = ProbeInconclusive + ": " + string(out.ExitClass)
			note("compaction probe stderr: %s", truncate(out.StderrTail, 300))
		}
	}

	res.OK = trivialOK &&
		(!opts.ProbeSandbox || res.SandboxProbe == ProbeOK) &&
		(!opts.ProbeCompaction || res.CompactionProbe == ProbeOK)
	return res, nil
}

type probeReport struct {
	Status    string            `json:"status"`
	Probes    map[string]string `json:"probes"`
	Secret    any               `json:"secret"`
	PageTitle any               `json:"page_title"`
}

// sandboxProbe runs RUN-SANDBOXED and returns the probe result string.
func sandboxProbe(ctx context.Context, a Adapter, p Profile, opts SmokeOptions, res *SmokeResult) string {
	if !SandboxAvailable() {
		return ProbeUnavailable + ": " + SandboxExec + " is missing"
	}
	outside := filepath.Join(opts.WorkDir, "sandbox", "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		return ProbeUnavailable + ": " + err.Error()
	}
	secretPath := filepath.Join(outside, "secret.txt")
	canaryPath := filepath.Join(outside, "canary.txt")
	os.WriteFile(secretPath, []byte(smokeSecret), 0o644)
	os.WriteFile(canaryPath, []byte(smokeCanary), 0o644)
	before, _ := os.ReadDir(outside)

	out, handoff, inv, err := smokeRun(ctx, a, p, opts, "sandbox", sandboxBrief(outside), true)
	if err != nil {
		return ProbeUnavailable + ": " + err.Error()
	}
	res.RunDirs["sandbox"] = inv.RunDir
	var problems []string
	if out.ExitClass == Crash && strings.Contains(out.StderrTail, "sandbox-exec") {
		return ProbeFailed + ": the CLI does not run under sandbox-exec: " + truncate(out.StderrTail, 300)
	}
	if out.ExitClass == Auth || out.ExitClass == Quota {
		return ProbeInconclusive + ": " + string(out.ExitClass)
	}
	if c, err := os.ReadFile(canaryPath); err != nil || !bytes.Equal(c, []byte(smokeCanary)) {
		problems = append(problems, "the canary outside the workspace changed")
	}
	if s, err := os.ReadFile(secretPath); err != nil || !bytes.Equal(s, []byte(smokeSecret)) {
		problems = append(problems, "the secret outside the workspace changed")
	}
	if after, _ := os.ReadDir(outside); len(after) != len(before) {
		problems = append(problems, "a new file appeared outside the workspace")
	}
	var report probeReport
	if err := lenientJSON(handoff, &report); err != nil {
		problems = append(problems, "handoff unreadable: "+err.Error())
	} else {
		for _, key := range []string{"read_outside", "list_home", "fetch_url"} {
			if v := strings.ToLower(report.Probes[key]); v != "failed" {
				problems = append(problems, key+" reported "+quoteOr(v, "nothing"))
			}
		}
		if v := strings.ToLower(report.Probes["write_outside"]); v == "succeeded" {
			problems = append(problems, "write_outside reported succeeded")
		}
		if s, ok := report.Secret.(string); ok && strings.Contains(s, "4242") {
			problems = append(problems, "the secret was read")
		}
		if t, ok := report.PageTitle.(string); ok && strings.Contains(strings.ToLower(t), "example") {
			problems = append(problems, "the URL was fetched")
		}
	}
	if out.ExitClass != Ok {
		problems = append(problems, "exit class "+string(out.ExitClass))
	}
	if why := keychainProbe(ctx, p, filepath.Join(inv.RunDir, FileSandbox)); why != "" {
		problems = append(problems, why)
	}
	if len(problems) > 0 {
		return ProbeFailed + ": " + strings.Join(problems, "; ")
	}
	return ProbeOK
}

// keychainProbeService is the keychain item the probe writes and removes.
const keychainProbeService = "bookwriter-sandbox-probe"

// keychainProbe checks, for a provider that keeps its login in the macOS
// keychain (claude), that a keychain write succeeds under the run's
// seatbelt profile: a refreshed OAuth token that cannot be saved is lost,
// and the server then treats the next use of the old one as revoked. It
// adds a probe item through the same `security` tool the CLI uses, then
// removes it outside the sandbox. "" when fine or not applicable.
func keychainProbe(ctx context.Context, p Profile, profilePath string) string {
	if runtime.GOOS != "darwin" || p.Provider != "claude" {
		return ""
	}
	if _, err := os.Stat(profilePath); err != nil {
		return ""
	}
	defer exec.Command("/usr/bin/security", "delete-generic-password", "-s", keychainProbeService).Run()
	add := exec.CommandContext(ctx, SandboxExec, "-f", profilePath, "/usr/bin/security", "add-generic-password",
		"-a", "bookwriter", "-s", keychainProbeService, "-w", "probe", "-U")
	outb, err := add.CombinedOutput()
	msg := strings.TrimSpace(string(outb))
	if err != nil || msg != "" {
		why := truncate(msg, 200)
		if why == "" {
			why = err.Error()
		}
		return "the keychain is not writable under the sandbox, so a refreshed login would be lost: " + why
	}
	find := exec.CommandContext(ctx, "/usr/bin/security", "find-generic-password", "-s", keychainProbeService)
	if err := find.Run(); err != nil {
		return "the keychain probe item was not saved under the sandbox"
	}
	return ""
}

func quoteOr(s, empty string) string {
	if s == "" {
		return empty
	}
	return fmt.Sprintf("%q", s)
}

func wordCount(s string) int { return len(strings.Fields(s)) }

func round3(f float64) float64 { return math.Round(f*1000) / 1000 }
