package providers

import "fmt"

// Providers lists the provider names in routing order plus direct HTTP APIs and fake.
var Providers = []string{"claude", "codex", "gemini", "kimi", "grok", "muse", "agy", "anthropic-api", "openai-api", "gemini-api", "fake"}

// NewAdapter returns the default adapter for a provider name. The fake
// adapter comes back with no binary or scripts; set them before use.
func NewAdapter(provider string) (Adapter, error) {
	switch provider {
	case "claude":
		return ClaudeAdapter{}, nil
	case "codex":
		return CodexAdapter{}, nil
	case "gemini":
		return GeminiAdapter{}, nil
	case "kimi":
		return KimiAdapter{}, nil
	case "grok":
		return GrokAdapter{}, nil
	case "muse":
		return MuseAdapter{}, nil
	case "agy":
		return AgyAdapter{}, nil
	case "anthropic-api", "openai-api", "gemini-api":
		return DirectAPIAdapter{Kind: provider}, nil
	case "fake":
		return &FakeAdapter{}, nil
	}
	return nil, fmt.Errorf("providers: unknown provider %q", provider)
}

// IsLive reports whether an adapter starts a real provider CLI or direct HTTP API.
func IsLive(a Adapter) bool {
	switch a.(type) {
	case ClaudeAdapter, *ClaudeAdapter, CodexAdapter, *CodexAdapter, GeminiAdapter, *GeminiAdapter,
		KimiAdapter, *KimiAdapter, GrokAdapter, *GrokAdapter, MuseAdapter, *MuseAdapter, AgyAdapter, *AgyAdapter,
		DirectAPIAdapter, *DirectAPIAdapter:
		return true
	}
	return false
}
