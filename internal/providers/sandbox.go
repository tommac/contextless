package providers

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// SandboxExec is the seatbelt launcher.
const SandboxExec = "/usr/bin/sandbox-exec"

// SandboxSpec is the data the profile is rendered from. SandboxProfile
// builds one from an Invocation; Run adds the provider defaults.
type SandboxSpec struct {
	RunID, Provider, Role string
	Workspace, RunDir     string
	// Binary is the resolved CLI executable; its directory is readable too.
	Binary string
	// ReadDirs are read-only (config and auth directories, extra binaries).
	ReadDirs []string
	// WriteDirs are read-write besides the workspace and RunDir.
	WriteDirs []string
	// NetworkHosts: "localhost:N" and "*:N" become rules; hostnames are
	// recorded as comments and enforced by the proxy (package doc).
	NetworkHosts []string
	// AllNetwork allows every connection (researcher role).
	AllNetwork bool
	// LocalListen lets the CLI bind and accept on 127.0.0.1 at any port and
	// connect back to it: agy runs its agent as a local server on a random
	// port and its client talks to it there. Outside connections still go
	// only through the proxy.
	LocalListen bool
}

// systemReadPaths are readable in every profile: the OS, system libraries,
// Homebrew's read-only trees and /dev. The root directory itself must be a
// literal rule or the dynamic loader aborts (verified). /private/tmp is not
// readable: it holds other programs' scratch files (and, when the runs root
// is there, the sandbox probe's secret); TMPDIR points inside RunDir.
var systemReadPaths = []string{
	"/usr", "/bin", "/sbin", "/System", "/Library",
	"/private/etc", "/private/var/db", "/dev",
	"/opt/homebrew/bin", "/opt/homebrew/lib", "/opt/homebrew/Cellar", "/opt/homebrew/opt",
	"/opt/homebrew/share", "/opt/homebrew/Caskroom", "/opt/homebrew/etc",
}

// baseRules are the non-file rules every CLI needs: fork and exec of
// readable binaries, signals inside the sandbox, sysctl, mach lookups (the
// keychain and system daemons), POSIX IPC, metadata of every path so lookups
// resolve, and the terminal devices.
var baseRules = []string{
	"(allow process-fork)",
	"(allow process-exec)",
	"(allow signal (target same-sandbox))",
	"(allow sysctl-read)",
	"(allow mach-lookup)",
	"(allow ipc-posix-shm)",
	"(allow ipc-posix-sem)",
	"(allow file-read-metadata)",
	`(allow file-ioctl (literal "/dev/null") (literal "/dev/tty") (regex #"^/dev/ttys[0-9]+$"))`,
}

var deviceWrites = []string{`(literal "/dev/null")`, `(literal "/dev/tty")`, `(regex #"^/dev/ttys[0-9]+$")`}

var hostRuleRe = regexp.MustCompile(`^(\*|localhost|127\.0\.0\.1):(\d+|\*)$`)

// SandboxDefaults are a provider's config and auth directories, extra
// binaries whose directories must be readable (resolved through PATH and
// symlinks at render time), state directories that must be writable, and
// the API hosts the proxy admits. "~" expands to HOME. Stage 0 verifies and
// extends this table per CLI.
type SandboxDefaults struct {
	ReadDirs  []string
	WriteDirs []string
	Binaries  []string
	Hosts     []string
	// LocalListen: see SandboxSpec.LocalListen.
	LocalListen bool
}

// DefaultSandbox is the per-provider table.
var DefaultSandbox = map[string]SandboxDefaults{
	// claude (stage 0, verified 2026-10-23 after a fresh login): under this
	// profile claude starts, reads its keychain login, reaches
	// api.anthropic.com through the proxy, and passes the sandbox probe. The
	// writable paths are the ones it was denied: transcripts, session pid
	// files, the OAuth refresh lock, the updater's version lock, and the
	// Bash tool's per-session environment directory (without it every Bash
	// call failed with EPERM, so the probe's fetch never reached the proxy),
	// and the credential storage write lock with the config lock: denied the
	// storage lock, a sandboxed claude that refreshed its OAuth token could
	// not save the new one, the next refresh reused the spent token, and the
	// server revoked the whole login (2026-10-25, during a long trial). ~/.claude.json and ~/.claude/settings.json (MCP servers, hooks)
	// and ~/.local/share/claude (installed versions) stay read-only so a run
	// cannot plant config or code that a later unsandboxed claude loads.
	//
	// Claude Code 2.1.308 opens /tmp/claude-<uid> itself at startup and
	// crashes with "EPERM: operation not permitted, open '/tmp/claude-501'"
	// when it cannot (2026-10-30); a read of that one directory node is
	// enough (verified on the failed smoke's profile), so the rule matches
	// the node only, never the other sessions' directories inside it.
	"claude": {
		ReadDirs: []string{"~/.claude", "~/.claude.json", "~/.local/share/claude", "~/.local/bin", "~/Library/Keychains",
			ClaudeTempDirPattern},
		WriteDirs: []string{"~/.claude/projects", "~/.claude/sessions", "~/.claude/.oauth_refresh.lock*",
			"~/.local/state/claude/locks", "~/.claude/session-env",
			"~/.claude/.storage-write.lock*", "~/.claude.lock*",
			// the login lives in the login keychain (claude shells out to
			// `security add-generic-password`); with it read-only the write
			// fails with EPERM and a refreshed token is lost, so the next run
			// finds the old token revoked (twice during a long trial, eight
			// hours after each login). The prefix covers the keychain file
			// and securityd's temporary siblings (login.keychain-db.sb-*);
			// the directory and the other keychains stay read-only
			"~/Library/Keychains/login.keychain-db*"},
		Hosts: []string{"api.anthropic.com", "statsig.anthropic.com", "platform.claude.com"},
	},
	// codex (stage 0, 2026-10-22): seatbelt cannot nest (a nested
	// sandbox_apply fails with forbidden-sandbox-reinit under any profile
	// that denies anything), so a sandboxed codex runs with its own sandbox
	// off (danger-full-access) and this profile confines codex and every
	// tool it runs. CODEX_HOME is a per-run directory in RunDir; the only
	// file of ~/.codex it reads is auth.json (through a symlink), which a
	// token refresh rewrites. ab.chatgpt.com is admitted as a subdomain;
	// *.oaiusercontent.com (plugin assets) stays blocked.
	"codex": {
		ReadDirs:  []string{"~/.codex/auth.json"},
		WriteDirs: []string{"~/.codex/auth.json"},
		Hosts:     []string{"chatgpt.com", "api.openai.com", "auth.openai.com"},
	},
	// gemini (stage 0, 2026-10-22): API-key auth only; the adapter points
	// GEMINI_CLI_HOME into RunDir, so ~/.gemini (other projects' chats, the
	// antigravity token) is neither read nor written. node and the global
	// node_modules are readable. play.googleapis.com (telemetry) stays
	// blocked; the OAuth hosts (cloudcode-pa, oauth2) are not admitted
	// because an OAuth login lives in ~/.gemini.
	"gemini": {
		ReadDirs: []string{"~/.local/lib/node_modules", "~/.local/bin"},
		Binaries: []string{"node"},
		Hosts:    []string{"generativelanguage.googleapis.com"},
	},
	// kimi (stage 0, 2026-10-22): its state lives in ~/.kimi-code, and the
	// OAuth refresh (mid-run too) locks oauth/kimi-code and rewrites
	// credentials/, so those data paths are writable; config.toml,
	// mcp.json, plugins, skills, bin/ and updates/ stay read-only so a run
	// cannot plant code or config that a later unsandboxed kimi would load
	// (the update check and the query-store cache fail quietly). ~/Library/Caches/kimi-code holds the extracted
	// native modules. auth.kimi.com is the token refresh, code.kimi.com the
	// client config; cdn.kimi.com (updates) and telemetry-logs.kimi.com
	// stay blocked. The adapter passes --skills-dir so kimi does not watch
	// the home directory for user skills (that watch needs a readable home,
	// and without it the run exits 0 with no output). Kimi 2.5 (2026-11-11)
	// migrated workspaces.json into projects.json, guarded by a
	// .projects-write-lock, and locks sessions under locks/; without write
	// access every run failed at once with "storage write failed:
	// permission denied". All three are data, not code or config.
	"kimi": {
		ReadDirs: []string{"~/.kimi-code", "~/Library/Caches/kimi-code"},
		WriteDirs: []string{"~/.kimi-code/sessions", "~/.kimi-code/session_index.jsonl*", "~/.kimi-code/workspaces.json*",
			"~/.kimi-code/projects.json*", "~/.kimi-code/.projects-write-lock*", "~/.kimi-code/locks",
			"~/.kimi-code/credentials", "~/.kimi-code/oauth", "~/.kimi-code/logs", "~/.kimi-code/telemetry", "~/.kimi-code/file-history",
			"~/.kimi-code/user-history"},
		Hosts: []string{"api.kimi.com", "auth.kimi.com", "code.kimi.com", "api.moonshot.ai", "api.moonshot.cn"},
	},
	// grok 1.0.x (2026-10-30, first live smoke): it creates a session under
	// ~/.grok/sessions before the model runs ("Couldn't create session:
	// FS_PERMISSION_DENIED" with ~/.grok read-only), keeps its session list,
	// logs, model cache and config lock beside it, rewrites auth.json when a
	// token refreshes, and talks to cli-chat-proxy.grok.com, not api.x.ai.
	// config.toml and the rest of ~/.grok stay read-only, so a run cannot
	// change the settings of Tom's other grok sessions. Its OIDC token lasts
	// six hours and refreshes at auth.x.ai (discovery and token endpoint):
	// with that host refused, grok 1.0.70 ran unauthenticated once the token
	// lapsed and every call failed on "unknown model id" (2026-10-30).
	"grok": {
		ReadDirs: []string{"~/.grok", "~/.claude/settings.json"},
		WriteDirs: []string{"~/.grok/sessions", "~/.grok/logs", "~/.grok/active_sessions.json*", "~/.grok/active_sessions.lock*",
			"~/.grok/auth.json*", "~/.grok/models_cache.json*", "~/.grok/managed_config.lock*"},
		Hosts: []string{"api.x.ai", "cli-chat-proxy.grok.com", "auth.x.ai"},
	},
	// muse 1.4.14 (2026-10-30, first live smoke): the model API is
	// api.meta.ai (the only host a logged-in run contacts). The login is an
	// OAuth token in the login keychain (auth.json says storage: keychain,
	// service ai.meta.dev.credentials), so the keychain is readable and the
	// login keychain file writable for a refresh, as for claude; auth.json
	// is rewritten beside it. Session logs are off by flag (--no-session-log),
	// so the session lock directory is never needed.
	"muse": {
		ReadDirs: []string{"~/.config/muse", "~/.local/bin", "~/Library/Keychains"},
		WriteDirs: []string{"~/.config/muse/auth.json*", "~/.config/muse/.auth.json.lock*",
			"~/Library/Keychains/login.keychain-db*"},
		Hosts: []string{"api.meta.ai"},
	},
	// agy 1.5.6 (2026-11-11, first live runs): its state is
	// ~/.gemini/antigravity-cli. A print-mode run writes the conversation
	// (conversations/, brain/<id>, presence/, implicit/, the summaries
	// databases), its log (log/, the cli.log link), cache/ and crashes/,
	// and rewrites antigravity-oauth-token when the token refreshes.
	// settings.json, bin/, builtin/, plugin_data/ and updater/ stay
	// read-only, so a run can change neither Tom's agy settings nor the
	// code a later unsandboxed agy loads, and cannot update agy. The model,
	// its server-side tools and the plan quota all go through
	// daily-cloudcode-pa.googleapis.com; oauth2.googleapis.com refreshes
	// the consumer sign-in.
	"agy": {
		ReadDirs: []string{"~/.gemini", "~/.local/bin"},
		WriteDirs: []string{"~/.gemini/antigravity-cli/conversations", "~/.gemini/antigravity-cli/brain",
			"~/.gemini/antigravity-cli/presence", "~/.gemini/antigravity-cli/implicit", "~/.gemini/antigravity-cli/cache",
			"~/.gemini/antigravity-cli/crashes", "~/.gemini/antigravity-cli/log", "~/.gemini/antigravity-cli/scratch",
			"~/.gemini/antigravity-cli/knowledge", "~/.gemini/antigravity-cli/cli.log*",
			"~/.gemini/antigravity-cli/conversation_summaries.db*", "~/.gemini/antigravity-cli/jetbox_summaries_proto.pb*",
			"~/.gemini/antigravity-cli/jetski_state.pbtxt*", "~/.gemini/antigravity-cli/history.jsonl*",
			"~/.gemini/antigravity-cli/last_check.timestamp*", "~/.gemini/antigravity-cli/antigravity-oauth-token*"},
		// www.googleapis.com: the plan eligibility check (oauth2/v2/userinfo)
		// that every run makes before the first model call
		Hosts: []string{"daily-cloudcode-pa.googleapis.com", "cloudcode-pa.googleapis.com", "oauth2.googleapis.com",
			"www.googleapis.com"},
		// "Failed to start: listen tcp 127.0.0.1:0: bind: operation not
		// permitted" without it (first sandboxed smoke, 2026-11-11)
		LocalListen: true,
	},
	"fake": {},
}

// SandboxFor returns the provider's defaults with "~" expanded and the
// extra binaries resolved to their real directories.
func SandboxFor(provider string) SandboxDefaults {
	d := DefaultSandbox[provider]
	out := SandboxDefaults{Hosts: append([]string{}, d.Hosts...), LocalListen: d.LocalListen}
	for _, p := range d.ReadDirs {
		out.ReadDirs = append(out.ReadDirs, ExpandHome(p))
	}
	for _, p := range d.WriteDirs {
		out.WriteDirs = append(out.WriteDirs, ExpandHome(p))
	}
	for _, b := range d.Binaries {
		if real, err := ResolveBinary(b); err == nil {
			out.ReadDirs = append(out.ReadDirs, filepath.Dir(real))
		}
	}
	return out
}

// ExpandHome replaces a leading "~" with HOME.
func ExpandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

// ResolveBinary finds a command on PATH and follows symlinks, so the
// sandbox can allow the real file (claude and muse are launcher symlinks).
func ResolveBinary(name string) (string, error) {
	p, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real, nil
	}
	return p, nil
}

// pathVariants returns p plus its /private twin, since /var, /tmp and /etc
// are symlinks and seatbelt matches the resolved path.
func pathVariants(p string) []string {
	p = filepath.Clean(p)
	out := []string{p}
	switch {
	case strings.HasPrefix(p, "/private/"):
		out = append(out, strings.TrimPrefix(p, "/private"))
	case strings.HasPrefix(p, "/var/"), strings.HasPrefix(p, "/tmp/"), strings.HasPrefix(p, "/etc/"), p == "/var", p == "/tmp", p == "/etc":
		out = append(out, "/private"+p)
	}
	return out
}

func quoteSB(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// ruleSet renders path rules without duplicates: subpath for directories,
// literal for files (by stat, or forced with literal), prefix for a path
// ending in "*" (a file and its atomic-write siblings, such as
// workspaces.json and workspaces.json.tmp.123), and regex for a path given
// as "re:<pattern>".
type ruleSet struct {
	seen  map[string]bool
	rules []string
}

func (r *ruleSet) add(paths []string, forceLiteral bool) {
	if r.seen == nil {
		r.seen = map[string]bool{}
	}
	for _, p := range paths {
		if p == "" || p == "*" {
			continue
		}
		if pat, ok := strings.CutPrefix(p, "re:"); ok {
			// an exact pattern, for a file whose name carries a random
			// part next to directories the profile must not open
			rule := "(regex #" + quoteSB(pat) + ")"
			if !r.seen[rule] {
				r.seen[rule] = true
				r.rules = append(r.rules, rule)
			}
			continue
		}
		kind := "subpath"
		if strings.HasSuffix(p, "*") {
			kind = "prefix"
			p = strings.TrimSuffix(p, "*")
		} else if forceLiteral {
			kind = "literal"
		} else if st, err := os.Stat(p); err == nil && !st.IsDir() {
			kind = "literal"
		}
		for _, v := range pathVariants(p) {
			rule := "(" + kind + " " + quoteSB(v) + ")"
			if !r.seen[rule] {
				r.seen[rule] = true
				r.rules = append(r.rules, rule)
			}
		}
	}
}

// pathRules renders subpath (or literal for existing files) rules.
func pathRules(paths []string) []string {
	var r ruleSet
	r.add(paths, false)
	return r.rules
}

// RenderSandbox renders a seatbelt profile from a spec.
func RenderSandbox(spec SandboxSpec) string {
	var b strings.Builder
	b.WriteString(";; contextless sandbox profile, generated by internal/providers\n")
	b.WriteString(";; run " + spec.RunID + " provider " + spec.Provider + " role " + spec.Role + "\n")
	b.WriteString("(version 1)\n(deny default)\n")
	for _, r := range baseRules {
		b.WriteString(r + "\n")
	}

	var reads ruleSet
	reads.add(systemReadPaths, false)
	if spec.Binary != "" {
		reads.add([]string{spec.Binary}, true)
		reads.add([]string{filepath.Dir(spec.Binary)}, false)
	}
	reads.add([]string{spec.Workspace, spec.RunDir}, false)
	reads.add(spec.ReadDirs, false)
	reads.add(spec.WriteDirs, false)
	b.WriteString("(allow file-read*\n  (literal \"/\")\n")
	for _, r := range reads.rules {
		b.WriteString("  " + r + "\n")
	}
	b.WriteString(")\n")

	writes := []string{spec.Workspace, spec.RunDir}
	writes = append(writes, spec.WriteDirs...)
	b.WriteString("(allow file-write*\n")
	for _, r := range pathRules(writes) {
		b.WriteString("  " + r + "\n")
	}
	for _, r := range deviceWrites {
		b.WriteString("  " + r + "\n")
	}
	b.WriteString(")\n")

	switch {
	case spec.AllNetwork:
		b.WriteString(";; network: all (research role)\n(allow network*)\n")
	default:
		var rules []string
		hosts := append([]string{}, spec.NetworkHosts...)
		sort.Strings(hosts)
		for _, h := range hosts {
			if hostRuleRe.MatchString(h) {
				h = strings.Replace(h, "127.0.0.1:", "localhost:", 1)
				rules = append(rules, "  (remote tcp "+quoteSB(h)+")")
			} else {
				b.WriteString(";; host allowed through the proxy: " + h + "\n")
			}
		}
		if spec.LocalListen {
			b.WriteString(";; local server on 127.0.0.1 (any port)\n" +
				"(allow network-bind (local ip \"localhost:*\"))\n" +
				"(allow network-inbound (local ip \"localhost:*\"))\n")
			rules = append(rules, "  (remote tcp \"localhost:*\")")
		}
		if len(rules) == 0 {
			b.WriteString(";; network: none\n")
		} else {
			b.WriteString("(allow network-outbound\n" + strings.Join(rules, "\n") + "\n)\n")
		}
	}
	return b.String()
}

// SandboxProfile renders the profile for an invocation: reads of the
// workspace, the CLI binary and its directory, allowDirs and the system
// paths; writes only to the workspace and RunDir; network to the listed
// hosts (see SandboxSpec.NetworkHosts), or everything when inv.Network.
func SandboxProfile(inv Invocation, cliBinary string, allowDirs []string, networkHosts []string) string {
	return RenderSandbox(SandboxSpec{
		RunID: inv.RunID, Provider: inv.Profile.Provider, Role: inv.Role,
		Workspace: inv.Workspace, RunDir: inv.RunDir, Binary: cliBinary,
		ReadDirs: allowDirs, NetworkHosts: networkHosts, AllNetwork: inv.Network,
	})
}

// WrapSandbox prefixes argv with the seatbelt launcher.
func WrapSandbox(argv []string, profilePath string) []string {
	out := []string{SandboxExec, "-f", profilePath}
	return append(out, argv...)
}

// SandboxAvailable reports whether sandbox-exec exists on this machine.
func SandboxAvailable() bool {
	st, err := os.Stat(SandboxExec)
	return err == nil && !st.IsDir()
}
