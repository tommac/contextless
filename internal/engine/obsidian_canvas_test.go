package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/tommac/contextless/internal/vault"
)

func writeCreativeBoard(t *testing.T, root string, target canvasNode, label string) string {
	t.Helper()
	path := filepath.Join(root, "experiment.canvas")
	board := canvasBoard{
		Nodes: []canvasNode{{ID: "q", Type: "text", Text: "# Question\r\nWhich cache policy?"}, target},
		Edges: []canvasEdge{{ID: "borrow", FromNode: target.ID, ToNode: "q", Label: label}},
	}
	data, err := json.Marshal(board)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCreativeCanvasPinsOriginalHeading(t *testing.T) {
	root := t.TempDir()
	body := "---\r\n# Rule\r\n---\r\n~~~sample\r\n# Rule\r\n~~~\r\n" +
		"## Rule ##\r\nAct when the tick arrives.\r\n### Example\r\nCafé.\r\n\r\n## Appendix\r\nUnrelated."
	if err := os.WriteFile(filepath.Join(root, "policy.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	node := canvasNode{ID: "n1", Type: "file", File: "policy.md", Subpath: "#rule"}
	path := writeCreativeBoard(t, root, node, "Constraint: only on ticks")
	snap, err := ReadCreativeCanvas(path, "")
	if err != nil || len(snap.Cards) != 1 || snap.Question != "Which cache policy?" {
		t.Fatalf("snapshot: %+v %v", snap, err)
	}
	card := snap.Cards[0]
	want := "## Rule ##\r\nAct when the tick arrives.\r\n### Example\r\nCafé."
	if card.Key != "borrow" || card.Kind != "file" || card.Role != "constraint" || card.Source != "policy.md" || card.Heading != "#rule" ||
		card.Body != body || card.BodySHA != vault.SHA(body) || card.Excerpt != want || body[card.Start:card.End] != want {
		t.Fatalf("wrong source pin: %+v", card)
	}
	if !strings.Contains(snap.Brief(false), card.Relation) || strings.Contains(snap.Brief(true), "only on ticks") ||
		!strings.Contains(snap.Brief(true), "Act when the tick arrives.") {
		t.Fatalf("incorrect labeled or neutral brief:\n%s\n%s", snap.Brief(false), snap.Brief(true))
	}
}

func TestCreativeCanvasExcerpts(t *testing.T) {
	root := t.TempDir()
	body := "---\r\nid: Café\r\n...\r\n\n  Borrow this signal.\r\n\r\n## Evidence\r\nPrivate review."
	if err := os.WriteFile(filepath.Join(root, "note.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	path := writeCreativeBoard(t, root, canvasNode{ID: "n", Type: "file", File: "note.md"}, "")
	snap, err := ReadCreativeCanvas(path, "")
	if err != nil {
		t.Fatal(err)
	}
	card := snap.Cards[0]
	if card.Role != "inspiration" || card.Relation != "inspiration" || card.Excerpt != "Borrow this signal." ||
		body[card.Start:card.End] != card.Excerpt {
		t.Fatalf("wrong opening pin: %+v", card)
	}

	body = "  " + strings.Repeat("é", 324) + " éxtra" + strings.Repeat("é", 10)
	node := canvasNode{ID: "n", Type: "text", Text: body}
	path = writeCreativeBoard(t, root, node, "Borrow a rhythm")
	snap, err = ReadCreativeCanvas(path, "")
	if err != nil {
		t.Fatal(err)
	}
	card = snap.Cards[0]
	if card.Kind != "text" || card.Source != "n" || card.Start != 2 || card.BodySHA != vault.SHA(body) ||
		card.Excerpt != strings.Repeat("é", 324) || body[card.Start:card.End] != card.Excerpt || !utf8.ValidString(card.Excerpt) {
		t.Fatalf("wrong bounded text pin: %+v", card)
	}
}

func TestCreativeCanvasRejectsAmbiguousSource(t *testing.T) {
	root := t.TempDir()
	body := "## Duplicated\nfirst\n## Duplicated\nsecond"
	if err := os.WriteFile(filepath.Join(root, "note.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, heading := range []string{"#Duplicated", "#Missing", "#^block", "no hash"} {
		t.Run(heading, func(t *testing.T) {
			path := writeCreativeBoard(t, root, canvasNode{ID: "n", Type: "file", File: "note.md", Subpath: heading}, "")
			if _, err := ReadCreativeCanvas(path, ""); err == nil {
				t.Fatalf("accepted heading %v", heading)
			}
		})
	}
}

func TestCanvasVaultBoundary(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "other.md")
	if err := os.WriteFile(outside, []byte("other"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.md")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{outside, "../other.md", "link.md", "missing.md", "config.json"} {
		path := writeCreativeBoard(t, root, canvasNode{ID: "n", Type: "file", File: name}, "")
		if _, err := CompileObsidianCanvas(path, root); err == nil {
			t.Errorf("legacy brief accepted %q", name)
		}
		if _, err := ReadCreativeCanvas(path, root); err == nil {
			t.Errorf("creative brief accepted %q", name)
		}
		brief, err := CompileCreativeObsidianCanvas(path, root, true, false)
		if err != nil || !strings.HasSuffix(brief, "Question: Which cache policy?") {
			t.Errorf("question-only brief attempted to use source %q: %v", name, err)
		}
	}
}

func TestDefaultCanvasBriefRetainsTextFileSupport(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "schedule.txt"), []byte("Friday at 9"), 0600); err != nil {
		t.Fatal(err)
	}
	board := writeCreativeBoard(t, root, canvasNode{ID: "a", Type: "file", File: "schedule.txt"}, "prior schedule")
	got, err := CompileObsidianCanvas(board, root)
	if err != nil || !strings.Contains(got, "Friday at 9") {
		t.Fatalf("ordinary text-file brief: %s, %v", got, err)
	}
	if _, err = ReadCreativeCanvas(board, root); err == nil {
		t.Fatal("expected a Markdown note for creative excerpts")
	}
}
