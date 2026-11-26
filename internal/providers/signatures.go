package providers

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Provider quota, auth and context signatures. Every adapter runs error
// events, rate-limit events and stderr through these; the raw text is kept
// in the events file so the health layer can log it (06 §6.3).
var (
	authRe = regexp.MustCompile(`(?i)(not logged in|not authenticated|unauthenticated|unauthorized|\b401\b|\b403\b|invalid api key|api key not valid|invalid_api_key|authentication[ _]?(failed|error)|auth_required|please (log ?in|sign in)|token (has )?expired|invalid_grant|permission_denied|login required|run .{0,24}\blogin\b)`)

	quotaRe = regexp.MustCompile(`(?i)(\b429\b|rate[ _-]?limit|usage[ _-]?limit|resets? (at|in)\b|will reset|resource_exhausted|\bquota\b|too many requests|weekly limit|5-hour limit|limit reached|insufficient_quota)`)

	contextRe = regexp.MustCompile(`(?i)(context[ _](window|length|limit)|prompt is too long|conversation is too long|exceeds the (maximum|context)|maximum context|input too large|blocking limit|context_length_exceeded|too many tokens|token limit)`)

	fsPermissionRe = regexp.MustCompile(`(?i)fs_permission_denied`)
	resetAtRe      = regexp.MustCompile(`(?i)resets? at (\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d+)?Z)`)

	retryAfterRe = regexp.MustCompile(`(?i)retry[- ]after[:\s]+(\d+)\s*(s|sec|secs|seconds?|m|min|mins|minutes?|h|hr|hrs|hours?|d|days?|w|weeks?)?\b`)
	resetsInRe   = regexp.MustCompile(`(?i)resets? in (\d+)\s*(s|sec|secs|seconds?|m|min|mins|minutes?|h|hr|hrs|hours?|d|days?|w|weeks?)\b`)
)

// SignatureClass maps provider text to Auth, Quota or ContextBlocked, or ""
// when no signature matches. Auth wins over quota, quota over context.
func SignatureClass(text string) ExitClass {
	// grok reports a sandbox file permission error as FS_PERMISSION_DENIED,
	// which the auth pattern's permission_denied would read as a login
	// problem (its first smoke under the sandbox was filed auth_required)
	text = fsPermissionRe.ReplaceAllString(text, "")
	switch {
	case authRe.MatchString(text):
		return Auth
	case quotaRe.MatchString(text):
		return Quota
	case contextRe.MatchString(text):
		return ContextBlocked
	}
	return ""
}

// quotaWindowRe matches a usage window a quota message names without a
// reset time ("You've reached your 5-hour usage limit", "your weekly
// (7-day) usage limit", kimi 2.3.3); weeklyRe catches a weekly limit that
// names no count.
var (
	quotaWindowRe = regexp.MustCompile(`(?i)(\d+)[- ](hour|day)s?\)?\s+(?:usage\s+|rate\s+)?(?:limit|window)`)
	weeklyRe      = regexp.MustCompile(`(?i)\bweekly\b`)
)

// QuotaWindow is the length in seconds of the usage window a quota message
// names ("5-hour usage limit", "weekly (7-day) usage limit"), 0 when it
// names none. A reset falls somewhere inside it, so it bounds how long a
// backoff should wait.
func QuotaWindow(text string) int {
	if m := quotaWindowRe.FindStringSubmatch(text); m != nil {
		n, _ := strconv.Atoi(m[1])
		if strings.EqualFold(m[2], "day") {
			return n * 86400
		}
		return n * 3600
	}
	if weeklyRe.MatchString(text) {
		return 7 * 86400
	}
	return 0
}

// QuotaExcerpt is the line of provider text that carries the quota
// signature, without the CLI's error prefix and cut at the first sentence
// and at 100 characters ("429 You've reached your weekly (7-day) usage
// limit" from kimi's refusal); empty when no line matches.
func QuotaExcerpt(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !quotaRe.MatchString(line) {
			continue
		}
		if i := strings.LastIndex(line, "_error: "); i >= 0 {
			line = line[i+len("_error: "):]
		}
		line = strings.TrimPrefix(line, "error: ")
		if i := strings.Index(line, ". "); i >= 0 {
			line = line[:i]
		}
		if r := []rune(line); len(r) > 100 {
			line = string(r[:100])
		}
		return line
	}
	return ""
}

// ResetAt extracts the reset time a quota message names ("Your usage
// window resets at 2026-10-30T10:53:48Z", muse 1.4.14), or false.
func ResetAt(text string) (time.Time, bool) {
	m := resetAtRe.FindStringSubmatch(text)
	if m == nil {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, m[1])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// RetryAfter extracts a retry delay in seconds from provider text, or 0:
// "Retry-After: 120", "retry after 2 hours", "resets in 3 minutes".
func RetryAfter(text string) int {
	for _, re := range []*regexp.Regexp{retryAfterRe, resetsInRe} {
		if m := re.FindStringSubmatch(text); m != nil {
			n, _ := strconv.Atoi(m[1])
			return unitSeconds(n, m[2])
		}
	}
	return 0
}

// unitSeconds scales n by a duration unit ("", s, m, h, d or w and their
// long forms); no unit means seconds.
func unitSeconds(n int, unit string) int {
	if unit == "" {
		return n
	}
	switch strings.ToLower(unit[:1]) {
	case "m":
		return n * 60
	case "h":
		return n * 3600
	case "d":
		return n * 86400
	case "w":
		return n * 7 * 86400
	}
	return n
}

// classifyExit is the shared exit classification (package doc). blocked
// reports whether an event means the provider stopped on its own cap with a
// clean exit (OkBlocked); nil means never.
func classifyExit(exitCode int, stderr []byte, events []Event, blocked func(Event) bool) ExitClass {
	var texts []string
	for _, ev := range events {
		switch ev.Kind {
		case KindCompaction:
			return TaintedCompaction
		case KindError, KindRateLimit:
			texts = append(texts, ev.Message)
		}
	}
	texts = append(texts, string(stderr))
	joined := strings.Join(texts, "\n")
	if c := SignatureClass(joined); c != "" {
		return c
	}
	for _, ev := range events {
		if ev.Kind == KindRateLimit {
			return Quota
		}
	}
	if exitCode == 0 {
		if blocked != nil {
			for _, ev := range events {
				if blocked(ev) {
					return OkBlocked
				}
			}
		}
		return Ok
	}
	return Crash
}

// HasCompaction reports whether any event is a compaction signal.
func HasCompaction(events []Event) bool {
	for _, ev := range events {
		if ev.Kind == KindCompaction {
			return true
		}
	}
	return false
}

// PeakContext returns the largest context measurement among the events.
func PeakContext(events []Event) int {
	peak := 0
	for _, ev := range events {
		if ev.ContextTokens > peak {
			peak = ev.ContextTokens
		}
	}
	return peak
}
