package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DirectRunner is implemented by adapters that invoke a vendor's HTTP API
// directly with SSE streaming and explicit prompt-cache block boundaries,
// bypassing local CLI startup overhead and macOS-only seatbelt dependencies
// for one-shot roles (responder, indexer, reviewer, assembler).
type DirectRunner interface {
	RunDirect(ctx context.Context, inv Invocation, onEvent func(Event)) (RunOutcome, error)
}

// DirectAPIAdapter speaks directly to Anthropic Messages, OpenAI Chat
// Completions, or Google Gemini SSE APIs using pure Go stdlib net/http.
// It automatically splits brief.md into a cache-stable prefix block and a
// volatile suffix block so prompt caches hit deterministically across runs.
type DirectAPIAdapter struct {
	// Kind is "anthropic-api", "openai-api", or "gemini-api".
	Kind string
	// BaseURL overrides the default API endpoint (useful for proxies or tests).
	BaseURL string
	// APIKey overrides the default environment variable lookup.
	APIKey string
	// Client overrides http.DefaultClient (used in tests for in-memory transports).
	Client *http.Client
}

func (a DirectAPIAdapter) Provider() string { return a.Kind }

func (a DirectAPIAdapter) BuildCommand(Invocation) ([]string, []string, []byte, error) {
	return nil, nil, nil, fmt.Errorf("providers: %s runs directly over HTTP, not as a CLI command", a.Kind)
}
func (a DirectAPIAdapter) ParseEvent([]byte) (Event, bool) { return Event{}, false }
func (a DirectAPIAdapter) FinalUsage([]byte, []Event) (Usage, error) {
	return Usage{}, ErrNoUsage
}
func (a DirectAPIAdapter) ExitClass(exitCode int, stderr []byte, events []Event) ExitClass {
	return classifyExit(exitCode, stderr, events, nil)
}
func (a DirectAPIAdapter) InstructionFile() string              { return "" }
func (a DirectAPIAdapter) ProjectSettings(Invocation) map[string][]byte { return nil }

// SplitBriefForCache splits a ContextlessAI brief into its stable prefix
// (Rules, Session, Decisions, Catalog, Recent) and its volatile suffix
// (chosen records, task, run id, focus message) at the exact section boundary.
func SplitBriefForCache(brief string) (stable, volatile string) {
	for _, marker := range []string{
		"\n## Conversation records chosen for this message",
		"\n## Supplied records",
		"\n## Task\n",
	} {
		if idx := strings.Index(brief, marker); idx > 0 {
			return brief[:idx], brief[idx:]
		}
	}
	return "", brief
}

func (a DirectAPIAdapter) apiKey() (string, error) {
	if a.APIKey != "" {
		return a.APIKey, nil
	}
	var envName string
	switch a.Kind {
	case "anthropic-api":
		envName = "ANTHROPIC_API_KEY"
	case "openai-api":
		envName = "OPENAI_API_KEY"
	case "gemini-api":
		envName = "GEMINI_API_KEY"
	default:
		return "", fmt.Errorf("providers: unknown direct API kind %q", a.Kind)
	}
	k := strings.TrimSpace(os.Getenv(envName))
	if k == "" {
		return "", fmt.Errorf("providers: %s is not set for %s", envName, a.Kind)
	}
	return k, nil
}

// RunDirect streams one model call over HTTPS, records stdout/events/usage
// artifacts in RunDir identically to CLI runs, and writes the adopted handoff
// to inv.HandoffPath.
func (a DirectAPIAdapter) RunDirect(ctx context.Context, inv Invocation, onEvent func(Event)) (RunOutcome, error) {
	if inv.Workspace == "" || inv.RunDir == "" {
		return RunOutcome{}, errors.New("providers: RunDirect needs Workspace and RunDir")
	}
	if err := os.MkdirAll(inv.RunDir, 0o755); err != nil {
		return RunOutcome{}, err
	}
	briefBytes, err := readBrief(inv)
	if err != nil {
		return RunOutcome{}, err
	}
	out := RunOutcome{
		StdoutPath: filepath.Join(inv.RunDir, FileStdout),
		StderrPath: filepath.Join(inv.RunDir, FileStderr),
		EventsPath: filepath.Join(inv.RunDir, FileEvents),
		UsagePath:  filepath.Join(inv.RunDir, FileUsage),
	}
	key, err := a.apiKey()
	if err != nil {
		out.ExitCode = 1
		out.ExitClass = Auth
		out.StderrTail = err.Error()
		_ = os.WriteFile(out.StderrPath, []byte(err.Error()+"\n"), 0o644)
		return out, nil
	}
	if inv.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(inv.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	stable, volatile := SplitBriefForCache(string(briefBytes))
	req, err := a.buildHTTPRequest(ctx, inv, key, stable, volatile)
	if err != nil {
		return out, err
	}
	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		out.WallMs = time.Since(start).Milliseconds()
		out.ExitCode = 1
		out.StderrTail = err.Error()
		if errors.Is(err, context.DeadlineExceeded) {
			out.ExitClass = Timeout
			out.Killed = KilledTimeout
		} else {
			out.ExitClass = Crash
		}
		_ = os.WriteFile(out.StderrPath, []byte(err.Error()+"\n"), 0o644)
		return out, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		out.WallMs = time.Since(start).Milliseconds()
		out.ExitCode = resp.StatusCode
		out.StderrTail = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		_ = os.WriteFile(out.StderrPath, []byte(out.StderrTail+"\n"), 0o644)
		switch {
		case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
			out.ExitClass = Auth
		case resp.StatusCode == http.StatusTooManyRequests:
			out.ExitClass = Quota
		default:
			out.ExitClass = Crash
		}
		return out, nil
	}

	var (
		fullText strings.Builder
		events   []Event
		usage    Usage
		rawBuf   bytes.Buffer
	)
	emit := func(ev Event) {
		events = append(events, ev)
		if onEvent != nil {
			onEvent(ev)
		}
	}

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64<<10), 4<<20)
	for sc.Scan() {
		line := sc.Text()
		rawBuf.WriteString(line + "\n")
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		delta, udelta := a.parseSSEChunk([]byte(payload))
		if delta != "" {
			fullText.WriteString(delta)
			emit(Event{Kind: KindText, Message: delta})
		}
		if udelta != nil {
			mergeUsage(&usage, *udelta)
			if usage.PeakContext > 0 {
				emit(Event{Kind: KindUsage, ContextTokens: usage.PeakContext})
			}
		}
	}
	if err := sc.Err(); err != nil && !errors.Is(err, io.EOF) {
		out.StderrTail = err.Error()
		out.ExitClass = Crash
	}

	out.WallMs = time.Since(start).Milliseconds()
	usage.WallMs = out.WallMs
	usage.Turns = 1
	if usage.PeakContext == 0 {
		usage.PeakContext = usage.InputTokens + usage.OutputTokens
	}
	if inv.Profile.RateIn > 0 || inv.Profile.RateOut > 0 {
		usage.USDReported = EstimateUSD(inv.Profile, usage)
		usage.HasUSD = true
	}
	out.Usage = usage
	out.Events = events

	_ = os.WriteFile(out.StdoutPath, []byte(fullText.String()), 0o644)
	_ = os.WriteFile(out.UsagePath, jsonMarshal(usage), 0o644)
	var evBuf bytes.Buffer
	for _, ev := range events {
		if b, err := json.Marshal(ev); err == nil {
			evBuf.Write(b)
			evBuf.WriteByte('\n')
		}
	}
	_ = os.WriteFile(out.EventsPath, evBuf.Bytes(), 0o644)

	if data, ok := jsonObjectText(fullText.String()); ok && handoffForRun(data, inv.RunID) {
		if os.MkdirAll(filepath.Dir(inv.handoffPath()), 0o755) == nil && os.WriteFile(inv.handoffPath(), data, 0o644) == nil {
			out.HandoffAdopted = true
		}
	}
	if out.ExitClass == "" {
		out.ExitCode = 0
		out.ExitClass = Ok
	}
	return out, nil
}

func mergeUsage(dst *Usage, src Usage) {
	if src.InputTokens > 0 {
		dst.InputTokens = src.InputTokens
	}
	if src.CachedTokens > 0 {
		dst.CachedTokens = src.CachedTokens
	}
	if src.CacheWriteTokens > 0 {
		dst.CacheWriteTokens = src.CacheWriteTokens
	}
	if src.OutputTokens > 0 {
		dst.OutputTokens = src.OutputTokens
	}
	if src.ReasoningTokens > 0 {
		dst.ReasoningTokens = src.ReasoningTokens
	}
	tot := dst.InputTokens + dst.OutputTokens
	if tot > dst.PeakContext {
		dst.PeakContext = tot
	}
}

func (a DirectAPIAdapter) buildHTTPRequest(ctx context.Context, inv Invocation, key, stable, volatile string) (*http.Request, error) {
	model := inv.Profile.Model
	switch a.Kind {
	case "anthropic-api":
		endpoint := a.BaseURL
		if endpoint == "" {
			endpoint = "https://api.anthropic.com/v1/messages"
		}
		var blocks []map[string]any
		if stable != "" {
			blocks = append(blocks, map[string]any{
				"type":          "text",
				"text":          stable,
				"cache_control": map[string]string{"type": "ephemeral"},
			})
		}
		blocks = append(blocks, map[string]any{
			"type": "text",
			"text": volatile,
		})
		body, _ := json.Marshal(map[string]any{
			"model":      model,
			"max_tokens": 16384,
			"stream":     true,
			"messages": []map[string]any{
				{"role": "user", "content": blocks},
			},
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", key)
		req.Header.Set("anthropic-version", "2023-06-01")
		return req, nil

	case "openai-api":
		endpoint := a.BaseURL
		if endpoint == "" {
			endpoint = "https://api.openai.com/v1/chat/completions"
		}
		var msgs []map[string]any
		if stable != "" {
			msgs = append(msgs, map[string]any{"role": "system", "content": stable})
		}
		msgs = append(msgs, map[string]any{"role": "user", "content": volatile})
		body, _ := json.Marshal(map[string]any{
			"model":          model,
			"stream":         true,
			"stream_options": map[string]any{"include_usage": true},
			"messages":       msgs,
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+key)
		return req, nil

	case "gemini-api":
		base := a.BaseURL
		if base == "" {
			base = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse", model)
		}
		var parts []map[string]any
		if stable != "" {
			parts = append(parts, map[string]any{"text": stable})
		}
		parts = append(parts, map[string]any{"text": volatile})
		body, _ := json.Marshal(map[string]any{
			"contents": []map[string]any{
				{"role": "user", "parts": parts},
			},
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-goog-api-key", key)
		return req, nil
	}
	return nil, fmt.Errorf("providers: unsupported direct API kind %q", a.Kind)
}

func (a DirectAPIAdapter) parseSSEChunk(data []byte) (string, *Usage) {
	m, ok := parseObject(data)
	if !ok {
		return "", nil
	}
	switch a.Kind {
	case "anthropic-api":
		typ := str(m, "type")
		switch typ {
		case "content_block_delta":
			d := sub(m, "delta")
			return str(d, "text"), nil
		case "message_start":
			if u := sub(sub(m, "message"), "usage"); u != nil {
				read := inum(u, "cache_read_input_tokens")
				write := inum(u, "cache_creation_input_tokens")
				return "", &Usage{
					InputTokens:      inum(u, "input_tokens") + read + write,
					CachedTokens:     read,
					CacheWriteTokens: write,
				}
			}
		case "message_delta":
			if u := sub(m, "usage"); u != nil {
				return "", &Usage{OutputTokens: inum(u, "output_tokens")}
			}
		}
	case "openai-api":
		var delta string
		if choices := list(m, "choices"); len(choices) > 0 {
			if c0, ok := choices[0].(map[string]any); ok {
				delta = str(sub(c0, "delta"), "content")
			}
		}
		if u := sub(m, "usage"); u != nil {
			cached := inum(sub(u, "prompt_tokens_details"), "cached_tokens")
			return delta, &Usage{
				InputTokens:  inum(u, "prompt_tokens"),
				CachedTokens: cached,
				OutputTokens: inum(u, "completion_tokens"),
			}
		}
		return delta, nil
	case "gemini-api":
		var delta strings.Builder
		for _, cand := range list(m, "candidates") {
			cm, _ := cand.(map[string]any)
			for _, part := range list(sub(cm, "content"), "parts") {
				pm, _ := part.(map[string]any)
				delta.WriteString(str(pm, "text"))
			}
		}
		if u := sub(m, "usageMetadata"); u != nil {
			return delta.String(), &Usage{
				InputTokens:     inum(u, "promptTokenCount"),
				CachedTokens:    inum(u, "cachedContentTokenCount"),
				OutputTokens:    inum(u, "candidatesTokenCount"),
				ReasoningTokens: inum(u, "thoughtsTokenCount"),
			}
		}
		return delta.String(), nil
	}
	return "", nil
}
