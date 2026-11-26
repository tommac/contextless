package vault

import (
	"strings"
	"time"
)

// TranscriptMessage is one message decoded by a transcript importer.
type TranscriptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// Timestamp is when the message was written (RFC 3339), if known.
	Timestamp string `json:"timestamp,omitempty"`
}

// ImportTranscript appends messages in order under the same writer lock,
// allocating IDs from one scan of the existing filenames. On a write error
// the count is the number of complete turn files already appended.
func (v *Vault) ImportTranscript(msgs []TranscriptMessage) (int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	seq, err := v.lastSeq()
	if err != nil {
		return 0, err
	}
	n := 0
	var last string
	for _, m := range msgs {
		role := RoleUser
		switch strings.ToLower(m.Role) {
		case "assistant", "ai", "model", "bot":
			role = RoleAssistant
		case "system":
			continue
		}
		body := strings.TrimSpace(m.Content)
		if body == "" {
			continue
		}
		reply := ""
		if role == RoleAssistant {
			reply = last
		}
		created := time.Now()
		if at, err := time.Parse(time.RFC3339Nano, m.Timestamp); err == nil {
			created = at
		}
		seq++
		t, err := v.appendTurn(seq, role, "imported", "", reply, body, created)
		if err != nil {
			return n, err
		}
		last = t.ID
		n++
	}
	return n, nil
}
