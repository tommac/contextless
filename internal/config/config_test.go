package config

import (
	"os"
	"path/filepath"
	"testing"
)

// relevant_tokens and whole_history_tokens are on by default, and a
// negative value in a config file turns each of them off.
func TestNegativeTokenBudgetsTurnTheRulesOff(t *testing.T) {
	d := Default()
	if d.RelevantTokens != 8000 || d.WholeHistoryTokens != 60000 {
		t.Fatalf("defaults: relevant_tokens %d, whole_history_tokens %d", d.RelevantTokens, d.WholeHistoryTokens)
	}
	path := filepath.Join(t.TempDir(), "cx.yaml")
	if err := os.WriteFile(path, []byte("relevant_tokens: -1\nwhole_history_tokens: -1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.RelevantTokens >= 0 || c.WholeHistoryTokens >= 0 {
		t.Fatalf("after -1: relevant_tokens %d, whole_history_tokens %d", c.RelevantTokens, c.WholeHistoryTokens)
	}
	// a file that does not name them keeps the defaults
	if err := os.WriteFile(path, []byte("recent_turns: 4\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if c, err = Load(path); err != nil || c.RelevantTokens != 8000 || c.WholeHistoryTokens != 60000 {
		t.Fatalf("unset: relevant_tokens %d, whole_history_tokens %d, err %v", c.RelevantTokens, c.WholeHistoryTokens, err)
	}
}
