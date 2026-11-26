package providers

import (
	"strings"
	"testing"
)

// A real CLI about to run with the sandbox off gets a warning; the scripted
// fake and the direct HTTP adapters start no local agent and get none.
func TestUnconfinedWarningOnlyForRealCLIsWithoutASandbox(t *testing.T) {
	if w := UnconfinedWarning(ClaudeAdapter{}, false); !strings.Contains(w, "permission prompts bypassed") || !strings.Contains(w, "SECURITY.md") {
		t.Errorf("a real CLI with the sandbox off: %q", w)
	}
	if w := UnconfinedWarning(ClaudeAdapter{}, true); w != "" {
		t.Errorf("with the sandbox on there is nothing to warn about: %q", w)
	}
	if w := UnconfinedWarning(&FakeAdapter{}, false); w != "" {
		t.Errorf("the scripted fake is not an agent: %q", w)
	}
	if w := UnconfinedWarning(DirectAPIAdapter{Kind: "anthropic-api"}, false); w != "" {
		t.Errorf("a direct HTTP adapter starts no local agent: %q", w)
	}
	if !strings.Contains(ErrNoSandbox.Error(), "sandbox: false") {
		t.Errorf("the refusal should say how to go on: %v", ErrNoSandbox)
	}
}
