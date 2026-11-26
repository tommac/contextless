package providers

import (
	"testing"
	"time"
)

func TestQuotaWindow(t *testing.T) {
	for _, c := range []struct {
		text string
		want int
	}{
		{"429 You've reached your 5-hour usage limit. Your quota will reset when the current 5-hour window ends.", 5 * 3600},
		{"You have hit the 1 hour rate limit", 3600},
		{"429 You've reached your weekly (7-day) usage limit. Your quota will reset when the current 7-day window ends.", 7 * 86400},
		{"weekly limit reached", 7 * 86400},
		{"rate limit exceeded; retry after 30 seconds", 0},
	} {
		if got := QuotaWindow(c.text); got != c.want {
			t.Errorf("QuotaWindow(%q) = %d, want %d", c.text, got, c.want)
		}
	}
}

func TestRetryAfter(t *testing.T) {
	for _, c := range []struct {
		text string
		want int
	}{
		{"Retry-After: 120", 120},
		{"rate limited, retry after 90 seconds", 90},
		{"usage limit reached, retry after 2 hours", 7200},
		{"retry after 5 minutes", 300},
		{"resets in 3 minutes", 180},
		{"resets in 1 hour", 3600},
		{"retry after 2 days", 2 * 86400},
		{"quota resets in 1 week", 7 * 86400},
		{"You've reached your 5-hour usage limit.", 0},
	} {
		if got := RetryAfter(c.text); got != c.want {
			t.Errorf("RetryAfter(%q) = %d, want %d", c.text, got, c.want)
		}
	}
}

func TestQuotaExcerpt(t *testing.T) {
	kimi := "EPERM: operation not permitted, watch '/Users/x'\nerror: failed to run prompt: provider.api_error: 429 You've reached your weekly (7-day) usage limit. Your quota will reset when the current 7-day window ends.\nSee log: x"
	if got, want := QuotaExcerpt(kimi), "429 You've reached your weekly (7-day) usage limit"; got != want {
		t.Errorf("kimi excerpt %q, want %q", got, want)
	}
	if got := QuotaExcerpt("API Error: 429 rate limit"); got != "API Error: 429 rate limit" {
		t.Errorf("claude excerpt %q", got)
	}
	if got := QuotaExcerpt("all fine"); got != "" {
		t.Errorf("no signature gave %q", got)
	}
}

func TestSignatureFSPermissionIsNotAuth(t *testing.T) {
	grok := "Error: Couldn't create session: Permission denied.: {\n  \"code\": \"FS_PERMISSION_DENIED\",\n  \"detail\": \"Operation not permitted (os error 1)\"\n}"
	if c := SignatureClass(grok); c == Auth {
		t.Errorf("a sandbox file permission error classified as %s", c)
	}
	if c := SignatureClass("PERMISSION_DENIED: the caller does not have permission"); c != Auth {
		t.Errorf("an API permission denial classified as %q, want auth", c)
	}
}

func TestResetAt(t *testing.T) {
	muse := "run ended with Failed: API error 429 [request_id=x]: Subscription quota exhausted. Your usage window resets at 2026-10-30T10:53:48Z. (rate_limit_error)"
	got, ok := ResetAt(muse)
	if !ok || !got.Equal(time.Date(2026, 10, 30, 10, 53, 48, 0, time.UTC)) {
		t.Fatalf("ResetAt(muse) = %v, %v", got, ok)
	}
	if _, ok := ResetAt("You've reached your weekly (7-day) usage limit."); ok {
		t.Error("a message without a reset time gave one")
	}
}
