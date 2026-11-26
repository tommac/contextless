package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// parentEnvAllowlist is what the child inherits from the orchestrator.
var parentEnvAllowlist = []string{"PATH", "HOME", "LANG", "LC_ALL", "LC_CTYPE", "USER", "LOGNAME", "SHELL", "TERM"}

// baseEnv builds the allowlisted child environment: the parent allowlist,
// TMPDIR and zsh's TMPPREFIX inside RunDir (zsh puts here-document files at
// $TMPPREFIX, default /tmp/zsh, which the sandbox does not allow; stage 0
// finding), extra adapter entries, then inv.Env last. Keys are unique and
// the result is sorted for stable argv.json files.
func baseEnv(inv Invocation, extra map[string]string) []string {
	m := map[string]string{}
	for _, k := range parentEnvAllowlist {
		if v, ok := os.LookupEnv(k); ok {
			m[k] = v
		}
	}
	if inv.RunDir != "" {
		m["TMPDIR"] = filepath.Join(inv.RunDir, "tmp")
		m["TMPPREFIX"] = filepath.Join(inv.RunDir, "tmp", "zsh")
	}
	for k, v := range extra {
		m[k] = v
	}
	for k, v := range inv.Env {
		m[k] = v
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+m[k])
	}
	return out
}

// envValue reads one key from an env list.
func envValue(env []string, key string) (string, bool) {
	for _, kv := range env {
		if strings.HasPrefix(kv, key+"=") {
			return kv[len(key)+1:], true
		}
	}
	return "", false
}

// redactEnv replaces the values of secret-looking keys for env.redacted.json.
func redactEnv(env []string) map[string]string {
	out := map[string]string{}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		upper := strings.ToUpper(k)
		for _, needle := range []string{"KEY", "TOKEN", "SECRET", "PASSWORD", "CREDENTIAL", "AUTH"} {
			if strings.Contains(upper, needle) {
				v = "<redacted>"
				break
			}
		}
		out[k] = v
	}
	return out
}

// readBrief reads the brief for stdin delivery.
func readBrief(inv Invocation) ([]byte, error) {
	if inv.BriefPath == "" {
		return nil, fmt.Errorf("providers: %s: BriefPath is empty", inv.RunID)
	}
	b, err := os.ReadFile(inv.BriefPath)
	if err != nil {
		return nil, fmt.Errorf("providers: %s: brief: %w", inv.RunID, err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return nil, fmt.Errorf("providers: %s: brief %s is empty", inv.RunID, inv.BriefPath)
	}
	return b, nil
}

// readSchema reads a schema file for inline delivery (claude, grok).
func readSchema(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("providers: schema: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

// requireWorkspace checks the fields every adapter needs.
func requireWorkspace(inv Invocation) error {
	if inv.Workspace == "" {
		return fmt.Errorf("providers: %s: Workspace is empty", inv.RunID)
	}
	if !filepath.IsAbs(inv.Workspace) {
		return fmt.Errorf("providers: %s: Workspace %q must be absolute", inv.RunID, inv.Workspace)
	}
	return nil
}

// insideWorkspace reports whether path is inside the workspace.
func insideWorkspace(inv Invocation, path string) bool {
	rel, err := filepath.Rel(inv.Workspace, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, "../")
}

// itoa formats an int for argv.
func itoa(n int) string { return fmt.Sprintf("%d", n) }

// ftoa formats a float for argv without trailing zeros.
func ftoa(f float64) string { return fmt.Sprintf("%g", f) }
