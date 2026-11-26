package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestImportTranscriptContinuesFromNumericTail(t *testing.T) {
	v := &Vault{Dir: t.TempDir()}
	// Two earlier canonical turns with a gap across the six-digit boundary.
	for _, seq := range []int{999998, 1000000} {
		if _, err := v.appendTurn(seq, RoleUser, "", "", "", "earlier", time.Unix(0, 0)); err != nil {
			t.Fatal(err)
		}
	}
	collision := filepath.Join(v.Dir, TurnPath(TurnID(1000002)))
	if err := os.Mkdir(collision, 0700); err != nil {
		t.Fatal(err)
	}
	const timestamp = "2026-11-20T09:30:07.123-05:00"
	messages := []TranscriptMessage{
		{Role: "system", Content: "ignore"},
		{Role: "ai", Content: "   "},
		{Role: "model", Content: "  café  ", Timestamp: timestamp},
		{Role: "assistant", Content: "second", Timestamp: timestamp},
	}
	if n, err := v.ImportTranscript(messages); n != 1 || err == nil {
		t.Fatalf("collision: wrote %d, error %v", n, err)
	}
	first, err := v.Turn("t1000001")
	if err != nil {
		t.Fatal(err)
	}
	if first.Role != RoleAssistant || first.Author != "imported" || first.Body != "café" ||
		first.BodySHA != SHA("café") || first.ReplyTo != "" || first.Created.Format(time.RFC3339) != "2026-11-20T14:30:07Z" {
		t.Fatalf("unexpected imported turn: %+v", first)
	}
	if err := os.Rename(collision, collision+".old"); err != nil {
		t.Fatal(err)
	}
	if n, err := v.ImportTranscript([]TranscriptMessage{{Role: "human", Content: "Question"}, {Role: "bot", Content: "Answer"}}); n != 2 || err != nil {
		t.Fatalf("second batch: wrote %d, error %v", n, err)
	}
	answer, err := v.Turn("t1000003")
	if err != nil || answer.ReplyTo != "t1000002" || answer.Role != RoleAssistant {
		t.Fatalf("import reply: %+v, %v", answer, err)
	}
	next, err := v.AppendTurn(RoleUser, "", "", "", "Next")
	if err != nil || next.ID != "t1000004" {
		t.Fatalf("ordinary append after import: %+v, %v", next, err)
	}
}
