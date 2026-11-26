package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Kill reasons recorded in RunOutcome.Killed.
const (
	KilledTimeout    = "timeout"
	KilledStalled    = "stalled"
	KilledCompaction = "compaction"
	KilledKillLine   = "kill_line"
	KilledCancelled  = "cancelled"
)

// StreamCap is the per-stream capture limit (04 §4.14).
const StreamCap = 16 << 20

// Run directory file names.
const (
	FileStdout  = "stdout.log"
	FileStderr  = "stderr.log"
	FileEvents  = "events.jsonl"
	FileUsage   = "usage.json"
	FileArgv    = "argv.json"
	FileEnv     = "env.redacted.json"
	FilePGID    = "pgid"
	FileSandbox = "sandbox.sb"
)

// Watchdog timings; tests shorten them.
var (
	stallPollInterval = 10 * time.Second
	killGrace         = 2 * time.Second
	pipeCloseGrace    = 5 * time.Second
	stderrTailBytes   = 64 << 10
)

// RunOutcome is everything Run learned about one process. A run that
// failed is still an outcome; Run returns an error only when it could not
// start the process.
type RunOutcome struct {
	ExitCode   int       `json:"exit_code"`
	StdoutPath string    `json:"stdout_path"`
	StderrPath string    `json:"stderr_path"`
	EventsPath string    `json:"events_path"`
	UsagePath  string    `json:"usage_path"`
	Events     []Event   `json:"-"`
	Usage      Usage     `json:"usage"`
	ExitClass  ExitClass `json:"exit_class"`
	// Killed names the watchdog that killed the run, or "".
	Killed string `json:"killed,omitempty"`
	// Compacted is true when any stream event or the marker file signaled
	// compaction.
	Compacted bool `json:"compacted"`
	// UsageMissing is true when the adapter found no usage (ErrNoUsage).
	UsageMissing bool   `json:"usage_missing"`
	WallMs       int64  `json:"wall_ms"`
	StderrTail   string `json:"stderr_tail,omitempty"`
	WaitError    string `json:"wait_error,omitempty"`
	ProxyDenied  int    `json:"proxy_denied"`
	// HandoffAdopted is true when the handoff file was written from the
	// CLI's structured output after the exit (handoffSource), because the
	// agent left no terminal handoff of its own.
	HandoffAdopted bool `json:"handoff_adopted,omitempty"`
	PGID           int  `json:"pgid"`
	// ProfilePath is the sandbox profile written for a sandboxed run.
	ProfilePath string `json:"profile_path,omitempty"`
}

// sandboxHinter is implemented by adapters whose files live outside the
// provider defaults (the fake adapter's binary and scripts).
type sandboxHinter interface{ SandboxDirs() []string }

// sandboxWriteHinter is implemented by adapters whose CLI writes to a
// per-workspace path outside RunDir that the profile must allow (claude's
// Bash tool directory under /tmp/claude-<uid>).
type sandboxWriteHinter interface {
	SandboxWriteDirs(inv Invocation) []string
}

// runHooks is implemented by adapters that stage files in RunDir before the
// spawn and reconcile them after the process exits (codex's per-run
// CODEX_HOME).
type runHooks interface {
	PrepareRun(inv Invocation) error
	FinishRun(inv Invocation)
}

// handoffSource is implemented by adapters whose CLI returns the handoff in
// its own output when a schema is enforced (claude's structured_output on
// the result line). Run adopts that JSON as the handoff file when the agent
// left none, or left one that never reached a terminal status: a one-shot
// claude judge holds only the Write tool, and Write refuses to overwrite a
// file the agent has not read, so its completed handoff reached the file
// never and the structured output always.
type handoffSource interface {
	HandoffFromOutput(stdout []byte) ([]byte, bool)
}

// terminalHandoff reports whether the handoff at path already carries a
// status past "started" (partial, completed, blocked, failed).
func terminalHandoff(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var h struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(data, &h) != nil {
		return false
	}
	return h.Status != "" && h.Status != "started"
}

// handoffForRun reports whether data is a handoff object with a status
// and, when both are known, the run's own id: output from another run, or
// a JSON value that is not a handoff, is never adopted.
func handoffForRun(data []byte, runID string) bool {
	var h struct {
		RunID  string `json:"run_id"`
		Status string `json:"status"`
	}
	if json.Unmarshal(data, &h) != nil || h.Status == "" {
		return false
	}
	return runID == "" || h.RunID == "" || h.RunID == runID
}

// jsonObjectText returns text as indented JSON when it holds exactly one
// JSON object, bare or inside a Markdown code fence.
func jsonObjectText(text string) ([]byte, bool) {
	t := strings.TrimSpace(text)
	if strings.HasPrefix(t, "```") {
		if i := strings.Index(t, "\n"); i >= 0 {
			t = t[i+1:]
		}
		t = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(t), "```"))
	}
	var m map[string]any
	if !strings.HasPrefix(t, "{") || json.Unmarshal([]byte(t), &m) != nil || len(m) == 0 {
		return nil, false
	}
	return jsonMarshal(m), true
}

// adoptHandoff writes the adapter's structured output as the handoff when
// the agent's own handoff is missing or still "started".
func adoptHandoff(a Adapter, inv Invocation, stdout []byte) bool {
	src, ok := a.(handoffSource)
	if !ok {
		return false
	}
	data, ok := src.HandoffFromOutput(stdout)
	if !ok || terminalHandoff(inv.handoffPath()) || !handoffForRun(data, inv.RunID) {
		return false
	}
	if err := os.MkdirAll(filepath.Dir(inv.handoffPath()), 0o755); err != nil {
		return false
	}
	return os.WriteFile(inv.handoffPath(), data, 0o644) == nil
}

// WriteProjectSettings drops the adapter's provider-local files into the
// workspace. Run calls it; packet builders may call it earlier.
func WriteProjectSettings(a Adapter, inv Invocation) error {
	for rel, content := range a.ProjectSettings(inv) {
		full := filepath.Join(inv.Workspace, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Run spawns the adapter's command in its own process group with cwd
// Workspace, streams stdout line by line through ParseEvent and onEvent,
// writes stdout.log, stderr.log, events.jsonl, usage.json, argv.json,
// env.redacted.json and pgid into RunDir, and enforces TimeoutSeconds and
// StallSeconds. A compaction event or live context at the profile's kill
// line kills the run. There is no kill handle: cancel ctx to kill the run
// (Killed is then "cancelled"). With sandbox true the command runs under a
// seatbelt profile (sandbox.sb in RunDir) and, without Network, behind a
// host-allowlisting proxy.
// ErrNoSandbox is returned when the configuration asks for the sandbox and
// this machine has none. cx does not fall back to running unconfined: the
// user has to choose that.
var ErrNoSandbox = errors.New("providers: the sandbox (macOS sandbox-exec) is not available on this machine, so no CLI was started. " +
	"Provider CLIs run with their permission prompts bypassed; without a sandbox nothing confines them. " +
	"To run anyway, set `sandbox: false` in the configuration, preferably inside a container, a virtual machine or a throwaway account (see SECURITY.md)")

// UnconfinedWarning is the text to show once when a real provider CLI is
// about to run with the sandbox off, or "" when there is nothing to warn
// about: the sandbox is on, or the adapter starts no local agent (the
// scripted fake, the direct HTTP adapters).
func UnconfinedWarning(a Adapter, sandbox bool) string {
	if sandbox {
		return ""
	}
	if _, ok := a.(*FakeAdapter); ok {
		return ""
	}
	if _, ok := a.(DirectRunner); ok {
		return ""
	}
	return "warning: the sandbox is off. Provider CLIs run with their permission prompts bypassed and nothing confines them: " +
		"a model can read, change and run anything this user account can. Use a container, a virtual machine or a throwaway account (see SECURITY.md)."
}

func Run(ctx context.Context, a Adapter, inv Invocation, sandbox bool, onEvent func(Event)) (RunOutcome, error) {
	if dr, ok := a.(DirectRunner); ok {
		return dr.RunDirect(ctx, inv, onEvent)
	}
	if inv.Workspace == "" || inv.RunDir == "" {
		return RunOutcome{}, errors.New("providers: Run needs Workspace and RunDir")
	}
	if st, err := os.Stat(inv.Workspace); err != nil || !st.IsDir() {
		return RunOutcome{}, fmt.Errorf("providers: workspace %s is not a directory", inv.Workspace)
	}
	if err := os.MkdirAll(filepath.Join(inv.RunDir, "tmp"), 0o755); err != nil {
		return RunOutcome{}, err
	}
	if err := WriteProjectSettings(a, inv); err != nil {
		return RunOutcome{}, err
	}
	inv.Sandboxed = sandbox
	if h, ok := a.(runHooks); ok {
		if err := h.PrepareRun(inv); err != nil {
			return RunOutcome{}, err
		}
		defer h.FinishRun(inv)
	}
	argv, env, stdin, err := a.BuildCommand(inv)
	if err != nil {
		return RunOutcome{}, err
	}
	out := RunOutcome{
		StdoutPath: filepath.Join(inv.RunDir, FileStdout),
		StderrPath: filepath.Join(inv.RunDir, FileStderr),
		EventsPath: filepath.Join(inv.RunDir, FileEvents),
		UsagePath:  filepath.Join(inv.RunDir, FileUsage),
	}

	var proxy *HostProxy
	if sandbox {
		if runtime.GOOS != "darwin" || !SandboxAvailable() {
			return RunOutcome{}, ErrNoSandbox
		}
		bin, err := resolveArgv0(argv[0])
		if err != nil {
			return RunOutcome{}, err
		}
		defaults := SandboxFor(a.Provider())
		readDirs := append([]string{}, defaults.ReadDirs...)
		if h, ok := a.(sandboxHinter); ok {
			readDirs = append(readDirs, h.SandboxDirs()...)
		}
		hosts := append([]string{}, defaults.Hosts...)
		if !inv.Network && len(hosts) > 0 {
			proxy, err = StartHostProxy(hosts)
			if err != nil {
				return RunOutcome{}, err
			}
			defer proxy.Close()
			env = mergeEnv(env, proxyEnv(proxy.URL()))
			hosts = append(hosts, "localhost:"+itoa(proxy.Port()))
		}
		writeDirs := append([]string{}, defaults.WriteDirs...)
		if h, ok := a.(sandboxWriteHinter); ok {
			writeDirs = append(writeDirs, h.SandboxWriteDirs(inv)...)
		}
		profile := RenderSandbox(SandboxSpec{
			RunID: inv.RunID, Provider: a.Provider(), Role: inv.Role,
			Workspace: inv.Workspace, RunDir: inv.RunDir, Binary: bin,
			ReadDirs: readDirs, WriteDirs: writeDirs,
			NetworkHosts: hosts, AllNetwork: inv.Network, LocalListen: defaults.LocalListen,
		})
		out.ProfilePath = filepath.Join(inv.RunDir, FileSandbox)
		if err := os.WriteFile(out.ProfilePath, []byte(profile), 0o644); err != nil {
			return RunOutcome{}, err
		}
		argv = WrapSandbox(argv, out.ProfilePath)
	}
	os.WriteFile(filepath.Join(inv.RunDir, FileArgv), jsonMarshal(argv), 0o644)
	os.WriteFile(filepath.Join(inv.RunDir, FileEnv), jsonMarshal(redactEnv(env)), 0o644)

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = inv.Workspace
	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return RunOutcome{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return RunOutcome{}, err
	}
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return RunOutcome{}, fmt.Errorf("providers: start %s: %w", argv[0], err)
	}
	pgid := cmd.Process.Pid
	out.PGID = pgid
	os.WriteFile(filepath.Join(inv.RunDir, FilePGID), []byte(itoa(pgid)+"\n"), 0o644)

	var (
		mu        sync.Mutex
		events    []Event
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
		lastMove  atomic.Int64
		killReq   = make(chan string, 4)
	)
	lastMove.Store(time.Now().UnixNano())
	touch := func() { lastMove.Store(time.Now().UnixNano()) }
	requestKill := func(reason string) {
		select {
		case killReq <- reason:
		default:
		}
	}
	killLine := 0
	if inv.Profile.EffectiveWindow > 0 && inv.Profile.KillFraction > 0 {
		killLine = inv.Profile.KillLine()
	}

	eventsFile, err := os.Create(out.EventsPath)
	if err != nil {
		syscall.Kill(-pgid, syscall.SIGKILL)
		return RunOutcome{}, err
	}
	stdoutDone := make(chan struct{})
	go func() {
		defer close(stdoutDone)
		f, _ := os.Create(out.StdoutPath)
		defer f.Close()
		r := bufio.NewReaderSize(stdoutPipe, 1<<20)
		written := 0
		for {
			line, rerr := r.ReadBytes('\n')
			if len(line) > 0 {
				touch()
				if written < StreamCap {
					n := min(len(line), StreamCap-written)
					f.Write(line[:n])
					stdoutBuf.Write(line[:n])
					written += n
				}
				if ev, ok := a.ParseEvent(line); ok {
					mu.Lock()
					events = append(events, ev)
					mu.Unlock()
					if b, err := json.Marshal(ev); err == nil {
						eventsFile.Write(append(b, '\n'))
					}
					if onEvent != nil {
						onEvent(ev)
					}
					switch {
					case ev.Kind == KindCompaction:
						requestKill(KilledCompaction)
					case killLine > 0 && ev.ContextTokens >= killLine:
						requestKill(KilledKillLine)
					}
				}
			}
			if rerr != nil {
				return
			}
		}
	}()
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		f, _ := os.Create(out.StderrPath)
		defer f.Close()
		r := bufio.NewReaderSize(stderrPipe, 256<<10)
		written := 0
		for {
			line, rerr := r.ReadBytes('\n')
			if len(line) > 0 {
				touch()
				if written < StreamCap {
					n := min(len(line), StreamCap-written)
					f.Write(line[:n])
					written += n
				}
				stderrBuf.Write(line)
				if stderrBuf.Len() > 2*stderrTailBytes {
					tail := append([]byte{}, stderrBuf.Bytes()[stderrBuf.Len()-stderrTailBytes:]...)
					stderrBuf.Reset()
					stderrBuf.Write(tail)
				}
			}
			if rerr != nil {
				return
			}
		}
	}()
	waitDone := make(chan error, 1)
	go func() {
		<-stdoutDone
		<-stderrDone
		waitDone <- cmd.Wait()
	}()

	var timeoutC <-chan time.Time
	if inv.TimeoutSeconds > 0 {
		t := time.NewTimer(time.Duration(inv.TimeoutSeconds) * time.Second)
		defer t.Stop()
		timeoutC = t.C
	}
	stallTick := time.NewTicker(stallPollInterval)
	defer stallTick.Stop()
	var timers []*time.Timer
	killed := ""
	kill := func(reason string) {
		if killed != "" {
			return
		}
		killed = reason
		syscall.Kill(-pgid, syscall.SIGTERM)
		timers = append(timers,
			time.AfterFunc(killGrace, func() { syscall.Kill(-pgid, syscall.SIGKILL) }),
			time.AfterFunc(killGrace+pipeCloseGrace, func() { stdoutPipe.Close(); stderrPipe.Close() }),
		)
	}
	var waitErr error
loop:
	for {
		select {
		case waitErr = <-waitDone:
			break loop
		case <-ctx.Done():
			kill(KilledCancelled)
		case <-timeoutC:
			kill(KilledTimeout)
		case reason := <-killReq:
			kill(reason)
		case <-stallTick.C:
			if inv.StallSeconds <= 0 || killed != "" {
				continue
			}
			last := time.Unix(0, lastMove.Load())
			if m := newestMtime(inv.Workspace); m.After(last) {
				last = m
			}
			if time.Since(last) > time.Duration(inv.StallSeconds)*time.Second {
				kill(KilledStalled)
			}
		}
	}
	for _, t := range timers {
		t.Stop()
	}
	if killed != "" {
		// Stragglers in a killed group; a clean exit leaves none.
		syscall.Kill(-pgid, syscall.SIGKILL)
	}

	out.WallMs = time.Since(start).Milliseconds()
	out.ExitCode = -1
	if cmd.ProcessState != nil {
		out.ExitCode = cmd.ProcessState.ExitCode()
	}
	var exitErr *exec.ExitError
	if waitErr != nil && !errors.As(waitErr, &exitErr) {
		out.WaitError = waitErr.Error()
	}
	if _, err := os.Stat(CompactionMarkerPath(inv)); err == nil {
		ev := Event{Kind: KindCompaction, Message: "marker: " + CompactionMarker}
		events = append(events, ev)
		if b, err := json.Marshal(ev); err == nil {
			eventsFile.Write(append(b, '\n'))
		}
		if onEvent != nil {
			onEvent(ev)
		}
	}
	eventsFile.Close()
	out.Events = events
	out.Compacted = HasCompaction(events)
	out.StderrTail = stderrBuf.String()
	out.Killed = killed

	usage, uerr := a.FinalUsage(stdoutBuf.Bytes(), events)
	if uerr != nil {
		out.UsageMissing = true
		usage = Usage{}
	}
	if usage.PeakContext == 0 {
		usage.PeakContext = PeakContext(events)
	}
	if usage.WallMs == 0 {
		usage.WallMs = out.WallMs
	}
	out.Usage = usage
	os.WriteFile(out.UsagePath, jsonMarshal(usage), 0o644)
	out.HandoffAdopted = adoptHandoff(a, inv, stdoutBuf.Bytes())

	class := a.ExitClass(out.ExitCode, []byte(out.StderrTail), events)
	switch killed {
	case KilledTimeout:
		class = Timeout
	case KilledStalled:
		class = Stalled
	case KilledCompaction:
		class = TaintedCompaction
	case KilledKillLine:
		class = ContextBlocked
	case KilledCancelled:
		class = Crash
	}
	if out.Compacted {
		class = TaintedCompaction
	}
	out.ExitClass = class
	if proxy != nil {
		out.ProxyDenied = proxy.DeniedCount()
	}
	return out, nil
}

// resolveArgv0 finds the real file behind the command, for the sandbox.
func resolveArgv0(name string) (string, error) {
	if strings.Contains(name, "/") {
		abs, err := filepath.Abs(name)
		if err != nil {
			return "", err
		}
		if real, err := filepath.EvalSymlinks(abs); err == nil {
			return real, nil
		}
		return abs, nil
	}
	return ResolveBinary(name)
}

// mergeEnv overrides or appends entries.
func mergeEnv(env []string, extra map[string]string) []string {
	out := make([]string, 0, len(env)+len(extra))
	for _, kv := range env {
		k, _, _ := strings.Cut(kv, "=")
		if _, ok := extra[k]; ok {
			continue
		}
		out = append(out, kv)
	}
	for _, k := range sortedKeysString(extra) {
		out = append(out, k+"="+extra[k])
	}
	return out
}

func sortedKeysString(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// newestMtime walks a directory for its newest modification time.
func newestMtime(dir string) time.Time {
	var newest time.Time
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if info, err := d.Info(); err == nil && info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	return newest
}

// ReadEvents reads an events.jsonl file back.
func ReadEvents(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Event
	r := bufio.NewReaderSize(f, 1<<20)
	for {
		line, rerr := r.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			var ev Event
			if err := json.Unmarshal(line, &ev); err == nil {
				out = append(out, ev)
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				return out, nil
			}
			return out, rerr
		}
	}
}
