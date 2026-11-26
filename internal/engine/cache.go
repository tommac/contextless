package engine

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// Providers bill the part of a prompt that repeats an earlier prompt byte
// for byte at a fraction of the normal price, so what a packet costs
// depends on how much of it the previous packet already held. CacheReplay
// measures that without a model: it rebuilds the respond packet each real
// user message would have received, in order, and compares each brief with
// the one before.
//
// The result is a ceiling, not a forecast. It assumes a provider reuses any
// shared start of the prompt. Anthropic and OpenAI's GPT-5.6 and later
// document reuse only up to block or message boundaries, and the brief is
// one block; a CLI's own prefix must match too, and it can carry the
// working directory, which differs per run. Only cache fields from real
// runs show what is actually reused.

// The floor and step assumed for a reusable prefix (OpenAI's older rule).
const (
	cacheMinTokens  = 1024
	cacheStepTokens = 128
)

// ReplayCall is one replayed message.
type ReplayCall struct {
	Turn   string `json:"turn"`
	Brief  int    `json:"brief"`  // tokens sent
	Cached int    `json:"cached"` // of those, in a cacheable prefix shared with the previous brief
	Whole  bool   `json:"whole"`
}

// CacheReplay replays the last n user messages (all, if n <= 0).
func (e *Engine) CacheReplay(n int) ([]ReplayCall, error) {
	turns, err := e.V.Turns()
	if err != nil {
		return nil, err
	}
	var users []vault.Turn
	for _, t := range turns {
		if t.Role == vault.RoleUser {
			users = append(users, t)
		}
	}
	first := 0
	if n > 0 && len(users) > n+1 {
		first = len(users) - n - 1 // one more, to compare the first with
	}
	cands, _ := e.R.Candidates(config.RoleResponder, nil)
	var out []ReplayCall
	prev := ""
	for i, u := range users[first:] {
		id := "run-replay-" + u.ID
		task := packet.Task{RunID: id, Kind: packet.KindRespond, Role: config.RoleResponder,
			Instructions: respondInstructions(id), Schema: packet.RespondSchema(true)}
		sized, err := e.buildPacket(task, u, packet.SelectOptions{}, cands, packet.OutputTokens[packet.KindRespond])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", u.ID, err)
		}
		if i > 0 || first == 0 {
			c := ReplayCall{Turn: u.ID, Brief: sized.Budget.Brief, Whole: sized.WholeHistory}
			c.Cached = cacheableTokens(packet.Tokens(sized.Profile, commonPrefix(prev, sized.Brief)))
			out = append(out, c)
		}
		prev = sized.Brief
	}
	return out, nil
}

func commonPrefix(a, b string) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func cacheableTokens(t int) int {
	if t < cacheMinTokens {
		return 0
	}
	return cacheMinTokens + (t-cacheMinTokens)/cacheStepTokens*cacheStepTokens
}

// CacheReport totals a replay. read and write are the prices of a cached
// and of an uncached input token relative to a normal one.
func CacheReport(calls []ReplayCall, read, write float64) string {
	if len(calls) == 0 {
		return "no user messages to replay\n"
	}
	var sent, cached, whole int
	var briefs, shares []int
	for _, c := range calls {
		sent += c.Brief
		cached += c.Cached
		briefs = append(briefs, c.Brief)
		shares = append(shares, 100*c.Cached/max(c.Brief, 1))
		if c.Whole {
			whole++
		}
	}
	sort.Ints(shares)
	billed := read*float64(cached) + write*float64(sent-cached)
	var b strings.Builder
	fmt.Fprintf(&b, "%d messages replayed; whole history sent for %d\n", len(calls), whole)
	b.WriteString("ceiling only: assumes the provider reuses any shared start of the prompt, which real runs must confirm\n")
	fmt.Fprintf(&b, "median brief %d tokens; median share in a cached prefix %d%%\n", median(briefs), shares[len(shares)/2])
	fmt.Fprintf(&b, "tokens sent %d, of which cacheable %d (%d%%)\n", sent, cached, 100*cached/max(sent, 1))
	n := float64(len(calls))
	miss := write * float64(sent) // nothing is cached: the conversation paused past the cache's lifetime
	fmt.Fprintf(&b, "billed-equivalent input per message (read %.2f, write %.2f): %.0f if every call hits the cache, %.0f if half do, %.0f if none do\n",
		read, write, billed/n, (billed+miss)/2/n, miss/n)
	return b.String()
}

// CachePlan is a way a provider keeps a prompt prefix: how long, and
// what writing it costs against a normal input token. A read costs
// cacheRead. These are the plans the providers document (November 2026);
// they say nothing of whether a given CLI reuses a prefix across runs.
type CachePlan struct {
	Name     string
	Lifetime time.Duration
	Write    float64
}

const cacheRead = 0.1

// CachePlans are the documented plans GapReport prices.
var CachePlans = []CachePlan{
	{"5 minutes, write 1.25x", 5 * time.Minute, 1.25},
	{"30 minutes, write 1.25x", 30 * time.Minute, 1.25},
	{"1 hour, write 2x", time.Hour, 2},
	{"12 hours, write 2.5x", 12 * time.Hour, 2.5},
}

// expectedPrefixPrice is the mean price of a stable prefix per message,
// as a multiple of the normal input price: a message that arrives within
// the lifetime reads the prefix, any other writes it again.
func expectedPrefixPrice(gaps []time.Duration, p CachePlan) float64 {
	sum := 0.0
	for _, g := range gaps {
		if g <= p.Lifetime {
			sum += cacheRead
		} else {
			sum += p.Write
		}
	}
	return sum / float64(len(gaps))
}

// keepAlivePrice is the same for a cache kept warm by a ping (a read) just
// before each lifetime ends, for up to horizon of silence.
func keepAlivePrice(gaps []time.Duration, p CachePlan, horizon time.Duration) float64 {
	every := p.Lifetime - p.Lifetime/10
	sum := 0.0
	for _, g := range gaps {
		pings := int(min(g, horizon) / every)
		sum += float64(pings) * cacheRead
		if g <= max(horizon, p.Lifetime) {
			sum += cacheRead
		} else {
			sum += p.Write
		}
	}
	return sum / float64(len(gaps))
}

// GapReport says how long a user waits between messages, per the times of
// each session's user messages, and what that pace makes a cached prefix
// cost under each documented cache lifetime: a prompt cache helps only
// when the next call comes within its lifetime. Gaps between sessions are
// not counted.
func GapReport(sessions [][]time.Time) string {
	var gaps []time.Duration
	for _, ts := range sessions {
		for i := 1; i < len(ts); i++ {
			if d := ts[i].Sub(ts[i-1]); d >= 0 {
				gaps = append(gaps, d)
			}
		}
	}
	if len(gaps) == 0 {
		return "no message times to compare (a workspace imported before times were kept has only the import time)\n"
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i] < gaps[j] })
	var b strings.Builder
	fmt.Fprintf(&b, "%d gaps between user messages in %d session(s); median %s\n", len(gaps), len(sessions), gaps[len(gaps)/2].Round(time.Second))
	for _, limit := range []time.Duration{5 * time.Minute, 30 * time.Minute, time.Hour, 12 * time.Hour, 24 * time.Hour} {
		n := sort.Search(len(gaps), func(i int) bool { return gaps[i] > limit })
		fmt.Fprintf(&b, "within %s: %d (%d%%)\n", limit, n, 100*n/len(gaps))
	}
	b.WriteString("\nprice of a stable prompt prefix per message, as a multiple of the normal input price (1.00 with no cache; a cache read costs 0.10):\n")
	for _, p := range CachePlans {
		fmt.Fprintf(&b, "  %.2f  cache kept %s\n", expectedPrefixPrice(gaps, p), p.Name)
	}
	short := CachePlans[0]
	best, bestAt := expectedPrefixPrice(gaps, short), time.Duration(0)
	for _, h := range []time.Duration{10 * time.Minute, 20 * time.Minute, 30 * time.Minute, time.Hour, 2 * time.Hour} {
		if v := keepAlivePrice(gaps, short, h); v < best {
			best, bestAt = v, h
		}
	}
	if bestAt == 0 {
		fmt.Fprintf(&b, "keep-alive pings on the %s cache do not pay at this pace: the best horizon is none\n", short.Lifetime)
	} else {
		fmt.Fprintf(&b, "keep-alive pings on the %s cache for up to %s of silence bring it to %.2f\n", short.Lifetime, bestAt, best)
	}
	return b.String()
}
