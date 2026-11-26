package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/vault"
)

type importMsg = vault.TranscriptMessage

// ccLine is the part of a Claude Code transcript line (one JSON object per
// line under ~/.claude/projects/<project>/<session>.jsonl) the importer reads.
type ccLine struct {
	Type             string `json:"type"`
	Timestamp        string `json:"timestamp"`
	IsMeta           bool   `json:"isMeta"`
	IsCompactSummary bool   `json:"isCompactSummary"`
	IsSidechain      bool   `json:"isSidechain"`
	IsAPIError       bool   `json:"isApiErrorMessage"`
	Message          *struct {
		Role    string          `json:"role"`
		Model   string          `json:"model"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	Attachment *struct {
		Type   string          `json:"type"`
		Prompt json.RawMessage `json:"prompt"`
	} `json:"attachment"`
}

type ccBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// isClaudeCode reports whether JSONL lines look like a Claude Code
// transcript: some line has a type and a message with a role.
func isClaudeCode(lines []string) bool {
	for _, line := range lines {
		var l ccLine
		if json.Unmarshal([]byte(line), &l) == nil && l.Type != "" && l.Message != nil && l.Message.Role != "" {
			return true
		}
	}
	return false
}

// harnessWrapped are blocks the Claude Code harness adds inside a user
// message; they are not the user's words and are left out.
var harnessWrapped = regexp.MustCompile(`(?s)<(system-reminder|ide_opened_file|ide_selection)>.*?</(system-reminder|ide_opened_file|ide_selection)>`)

// harnessOnly are prefixes of user records the harness writes on its own:
// slash commands and their output, shell escapes, background task notices.
var harnessOnly = []string{"<command-name>", "<command-message>", "<command-args>", "<local-command-",
	"<bash-input>", "<bash-stdout>", "<bash-stderr>", "<task-notification>", "[Request interrupted"}

// userText returns what the user typed in a user record's content, or ""
// for tool results and harness records.
func userText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) != nil {
		var blocks []ccBlock
		if json.Unmarshal(raw, &blocks) != nil {
			return ""
		}
		var parts []string
		for _, b := range blocks {
			if b.Type == "text" {
				parts = append(parts, b.Text)
			}
		}
		s = strings.Join(parts, "\n\n")
	}
	s = strings.TrimSpace(harnessWrapped.ReplaceAllString(s, ""))
	for _, p := range harnessOnly {
		if strings.HasPrefix(s, p) {
			return ""
		}
	}
	return s
}

// parseClaudeCode turns a Claude Code transcript into user and assistant
// messages. Each prompt the user typed is a user message (including
// prompts queued while the assistant was working); all assistant text
// between two prompts becomes one assistant message. Tool calls, tool
// results, thinking, compaction summaries, sub-agent (sidechain) records
// and the harness's own records are left out.
func parseClaudeCode(lines []string) []importMsg {
	var out []importMsg
	var reply []string
	replyAt := ""
	flush := func() {
		if len(reply) > 0 {
			out = append(out, importMsg{Role: "assistant", Content: strings.Join(reply, "\n\n"), Timestamp: replyAt})
			reply = nil
		}
	}
	user := func(text, at string) {
		if text == "" {
			return
		}
		flush()
		out = append(out, importMsg{Role: "user", Content: text, Timestamp: at})
	}
	for _, line := range lines {
		var l ccLine
		if json.Unmarshal([]byte(line), &l) != nil || l.IsSidechain {
			continue
		}
		switch {
		case l.Type == "user" && l.Message != nil:
			if !l.IsMeta && !l.IsCompactSummary {
				user(userText(l.Message.Content), l.Timestamp)
			}
		case l.Type == "attachment" && l.Attachment != nil && l.Attachment.Type == "queued_command":
			user(userText(l.Attachment.Prompt), l.Timestamp)
		case l.Type == "assistant" && l.Message != nil:
			if l.IsAPIError || l.Message.Model == "<synthetic>" {
				continue
			}
			var blocks []ccBlock
			if json.Unmarshal(l.Message.Content, &blocks) != nil {
				continue
			}
			for _, b := range blocks {
				if b.Type == "text" && strings.TrimSpace(b.Text) != "" {
					reply = append(reply, strings.TrimSpace(b.Text))
					replyAt = l.Timestamp
				}
			}
		}
	}
	flush()
	return out
}

// userTimes are the times of a transcript's user messages, where known.
func userTimes(msgs []importMsg) []time.Time {
	var out []time.Time
	for _, m := range msgs {
		if m.Role != "user" {
			continue
		}
		if t, err := time.Parse(time.RFC3339Nano, m.Timestamp); err == nil {
			out = append(out, t)
		}
	}
	return out
}
