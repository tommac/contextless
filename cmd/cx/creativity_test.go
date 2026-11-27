package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/creativity"
)

func TestWorkshopReplaysEditedConstraint(t *testing.T) {
	root := t.TempDir()
	rows, err := creativity.DemoRows("")
	if err != nil {
		t.Fatal(err)
	}
	if err := saveCreativeDemo(root, rows); err != nil {
		t.Fatal(err)
	}
	board := filepath.Join(root, "Passage cache.canvas")
	replay := func() string {
		t.Helper()
		var out bytes.Buffer
		if err := replayCreativeBoard(&out, board, ""); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}
	original := replay()
	if !strings.Contains(original, "value-first: 75") || strings.Contains(original, "admission: review") {
		t.Fatalf("unmodified board: %s", original)
	}
	if err := os.WriteFile(filepath.Join(root, "Return pattern.md"), []byte("Scans no longer return.\nTrace: before\n"), 0600); err != nil {
		t.Fatal(err)
	}
	revised := replay()
	if !strings.Contains(revised, "value-first: 9") || !strings.Contains(revised, "admission: review") || strings.Contains(revised, "resident-limit: review") {
		t.Fatalf("changed board: %s", revised)
	}
}
