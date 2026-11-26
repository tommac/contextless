package providers

import "testing"

func TestSignatureClass(t *testing.T) {
	cases := []struct {
		text string
		want ExitClass
	}{
		{"API Error: 429 rate_limit_error", Quota},
		{"You've hit your usage limit. Resets at 3pm", Quota},
		{"RESOURCE_EXHAUSTED: quota exceeded", Quota},
		{"Too Many Requests", Quota},
		{"Not logged in. Please run claude login", Auth},
		{"HTTP 401 Unauthorized", Auth},
		{"API key not valid", Auth},
		{"OAuth token has expired", Auth},
		{"PERMISSION_DENIED", Auth},
		{"prompt is too long: 210000 tokens > 200000 maximum", ContextBlocked},
		{"context_length_exceeded", ContextBlocked},
		{"The agent wrote three paragraphs.", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := SignatureClass(c.text); got != c.want {
			t.Errorf("%q: got %q want %q", c.text, got, c.want)
		}
	}
	if RetryAfter("Retry-After: 120") != 120 || RetryAfter("usage limit, resets in 2 hours") != 7200 ||
		RetryAfter("resets in 15m") != 900 || RetryAfter("nothing") != 0 {
		t.Error("RetryAfter")
	}
}

func TestClassifyPrecedence(t *testing.T) {
	compaction := Event{Kind: KindCompaction}
	quota := Event{Kind: KindError, Message: "429 too many requests"}
	auth := Event{Kind: KindError, Message: "unauthorized"}
	blocked := Event{Kind: KindError, Message: "blocked: max turns"}
	isBlocked := func(ev Event) bool { return ev.Kind == KindError && ev.Message == "blocked: max turns" }

	if classifyExit(0, nil, nil, nil) != Ok {
		t.Error("clean")
	}
	if classifyExit(1, nil, nil, nil) != Crash {
		t.Error("crash")
	}
	if classifyExit(0, nil, []Event{compaction, quota}, nil) != TaintedCompaction {
		t.Error("compaction first")
	}
	if classifyExit(1, nil, []Event{quota, auth}, nil) != Auth {
		t.Error("auth over quota")
	}
	if classifyExit(1, nil, []Event{quota}, nil) != Quota {
		t.Error("quota")
	}
	if classifyExit(1, nil, []Event{{Kind: KindRateLimit, Message: "slow down"}}, nil) != Quota {
		t.Error("rate_limit event without signature words is still Quota")
	}
	if classifyExit(1, []byte("prompt is too long"), nil, nil) != ContextBlocked {
		t.Error("context on stderr")
	}
	if classifyExit(0, nil, []Event{blocked}, isBlocked) != OkBlocked {
		t.Error("ok blocked")
	}
	if classifyExit(1, nil, []Event{blocked}, isBlocked) != Crash {
		t.Error("blocked with non-zero exit is a crash")
	}
	if classifyExit(0, nil, []Event{{Kind: KindText, Message: "I hit a rate limit in the story"}}, nil) != Ok {
		t.Error("model text is never a signature")
	}
}

// TestKimiExitClassIgnoresToolStderr: kimi echoes tool stderr, so a clean
// exit whose tools printed "403" is Ok, while kimi's own failure is not.
func TestKimiExitClassIgnoresToolStderr(t *testing.T) {
	probe := []byte("curl: (56) CONNECT tunnel failed, response 403\nls: /Users/x: Operation not permitted\n")
	if c := (KimiAdapter{}).ExitClass(0, probe, nil); c != Ok {
		t.Errorf("clean exit with tool stderr: %s", c)
	}
	own := []byte("error: failed to run prompt: OAuth request to https://auth.kimi.com/api/oauth/token failed: 401 Unauthorized\n")
	if c := (KimiAdapter{}).ExitClass(1, own, nil); c != Auth {
		t.Errorf("kimi's own auth failure: %s", c)
	}
	if c := (KimiAdapter{}).ExitClass(0, probe, []Event{{Kind: KindError, Message: "429 too many requests"}}); c != Quota {
		t.Errorf("an error event still counts after a clean exit: %s", c)
	}
	// A failed run whose tools printed "403" earlier is classified by kimi's
	// own final error line (here a sandbox denial, so a crash).
	lock := append(append([]byte{}, probe...), []byte("error: failed to run prompt: Unable to prepare OAuth refresh lock for \"kimi-code\": EPERM: operation not permitted, open '/Users/x/.kimi-code/oauth/kimi-code'\nSee log: /Users/x/.kimi-code/logs/kimi-code.log\n")...)
	if c := (KimiAdapter{}).ExitClass(1, lock, nil); c != Crash {
		t.Errorf("kimi's own non-auth failure after tool 403s: %s", c)
	}
	if c := (KimiAdapter{}).ExitClass(1, append(append([]byte{}, probe...), own...), nil); c != Auth {
		t.Errorf("kimi's own auth failure after tool 403s: %s", c)
	}
}
