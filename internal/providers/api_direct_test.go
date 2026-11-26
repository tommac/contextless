package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSplitBriefForCache(t *testing.T) {
	brief := "# Rules\nAlways verbatim.\n\n## Session\nTitle\n\n## Conversation records chosen for this message\n### t000001\nfoo\n\n## Task\nAnswer"
	stable, volatile := SplitBriefForCache(brief)
	if !strings.Contains(stable, "## Session") || strings.Contains(stable, "## Task") {
		t.Fatalf("bad stable prefix: %q", stable)
	}
	if !strings.HasPrefix(volatile, "\n## Conversation records chosen") {
		t.Fatalf("bad volatile suffix: %q", volatile)
	}
}

func TestDirectAPIAnthropicSSEAndPromptCache(t *testing.T) {
	runDir := t.TempDir()
	briefPath := filepath.Join(runDir, "brief.md")
	brief := "# Rules\nBe exact.\n\n## Recent turns\nhello\n\n## Conversation records chosen for this message\nfoo\n\n## Task\nRespond JSON."
	_ = os.WriteFile(briefPath, []byte(brief), 0o644)

	var gotAuth, gotVer string
	var gotBody map[string]any
	sse := strings.Join([]string{
		`data: {"type":"message_start","message":{"usage":{"input_tokens":120,"cache_read_input_tokens":80,"cache_creation_input_tokens":20}}}`,
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"{\"run_id\":\"run-1\",\"status\":\"completed\","}}`,
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"\"answer\":\"pool_size is 20\"}"}}`,
		`data: {"type":"message_delta","usage":{"output_tokens":18}}`,
		`data: [DONE]`,
	}, "\n") + "\n"

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotAuth = req.Header.Get("x-api-key")
		gotVer = req.Header.Get("anthropic-version")
		b, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(b, &gotBody)
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(bytes.NewBufferString(sse)),
		}, nil
	})}

	adapter := DirectAPIAdapter{Kind: "anthropic-api", APIKey: "test-key-not-real", Client: client}
	inv := Invocation{
		RunID:       "run-1",
		Workspace:   runDir,
		RunDir:      runDir,
		BriefPath:   briefPath,
		HandoffPath: filepath.Join(runDir, "handoff.json"),
		Profile:     Profile{ID: "anthropic-api/claude-sonnet-5-5", Provider: "anthropic-api", Model: "claude-sonnet-5-5", RateIn: 3, RateOut: 15}.WithDefaults(),
	}
	var streamed strings.Builder
	out, err := adapter.RunDirect(context.Background(), inv, func(ev Event) {
		if ev.Kind == KindText {
			streamed.WriteString(ev.Message)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "test-key-not-real" || gotVer == "" {
		t.Fatalf("missing anthropic headers: auth=%q ver=%q", gotAuth, gotVer)
	}
	// Verify explicit prompt-cache split in request body (`cache_control: ephemeral` on stable prefix).
	rawReq, _ := json.Marshal(gotBody)
	if !strings.Contains(string(rawReq), `"cache_control":{"type":"ephemeral"}`) {
		t.Fatalf("expected cache_control ephemeral on stable prefix, got %s", rawReq)
	}
	if out.ExitClass != Ok || !out.HandoffAdopted {
		t.Fatalf("expected Ok + HandoffAdopted, got %+v", out)
	}
	if out.Usage.CachedTokens != 80 || out.Usage.CacheWriteTokens != 20 || out.Usage.OutputTokens != 18 {
		t.Fatalf("unexpected usage accounting: %+v", out.Usage)
	}
	if !strings.Contains(streamed.String(), "pool_size is 20") {
		t.Fatalf("streamed callback missed deltas: %q", streamed.String())
	}
	if b, err := os.ReadFile(inv.HandoffPath); err != nil || !strings.Contains(string(b), "pool_size is 20") {
		t.Fatalf("handoff not written: %v %s", err, b)
	}
}

func TestDirectAPIOpenAIAndGeminiHydrogenSSE(t *testing.T) {
	cases := []struct {
		name  string
		kind  string
		model string
		sse   string
		check func(t *testing.T, req *http.Request)
	}{
		{
			name:  "openai",
			kind:  "openai-api",
			model: "gpt-6.1-sol",
			sse: strings.Join([]string{
				`data: {"choices":[{"delta":{"content":"{\"run_id\":\"run-2\",\"status\":\"completed\",\"answer\":\"ok-openai\"}"}}]}`,
				`data: {"choices":[],"usage":{"prompt_tokens":100,"completion_tokens":12,"prompt_tokens_details":{"cached_tokens":60}}}`,
				`data: [DONE]`,
			}, "\n") + "\n",
			check: func(t *testing.T, req *http.Request) {
				if req.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("bad auth header: %q", req.Header.Get("Authorization"))
				}
			},
		},
		{
			name:  "gemini-4-hydrogen",
			kind:  "gemini-api",
			model: "gemini-4-hydrogen",
			sse: strings.Join([]string{
				`data: {"candidates":[{"content":{"parts":[{"text":"{\"run_id\":\"run-2\",\"status\":\"completed\",\"answer\":\"ok-hydrogen\"}"}]}}],"usageMetadata":{"promptTokenCount":90,"candidatesTokenCount":14,"cachedContentTokenCount":50}}`,
			}, "\n") + "\n",
			check: func(t *testing.T, req *http.Request) {
				if !strings.Contains(req.URL.Path, "gemini-4-hydrogen:streamGenerateContent") {
					t.Errorf("expected gemini-4-hydrogen SSE endpoint, got %s", req.URL.String())
				}
				if req.Header.Get("x-goog-api-key") != "test-token" {
					t.Errorf("bad gemini key header: %q", req.Header.Get("x-goog-api-key"))
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			briefPath := filepath.Join(runDir, "brief.md")
			_ = os.WriteFile(briefPath, []byte("# Rules\nVerbatim.\n\n## Task\nAnswer."), 0o644)
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				tc.check(t, req)
				return &http.Response{
					StatusCode: 200,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body:       io.NopCloser(bytes.NewBufferString(tc.sse)),
				}, nil
			})}
			ad := DirectAPIAdapter{Kind: tc.kind, APIKey: "test-token", Client: client}
			inv := Invocation{
				RunID:       "run-2",
				Workspace:   runDir,
				RunDir:      runDir,
				BriefPath:   briefPath,
				HandoffPath: filepath.Join(runDir, "handoff.json"),
				Profile:     Profile{ID: tc.kind + "/" + tc.model, Provider: tc.kind, Model: tc.model}.WithDefaults(),
			}
			out, err := ad.RunDirect(context.Background(), inv, nil)
			if err != nil {
				t.Fatal(err)
			}
			if out.ExitClass != Ok || !out.HandoffAdopted || out.Usage.CachedTokens == 0 {
				t.Fatalf("unexpected outcome: %+v", out)
			}
		})
	}
}
