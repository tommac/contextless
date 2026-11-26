package providers

import (
	"fmt"
	"math"
	"strings"
)

// Live usage sources.
const (
	LiveNone    = "none"
	LivePerTurn = "per_turn"
	LiveStream  = "stream"
)

// Prompt delivery mechanics (06 §6.6).
const (
	DeliverStdin        = "stdin"
	DeliverStdinPrepend = "stdin_prepend"
	DeliverPromptFile   = "prompt_file"
	DeliverArgvPointer  = "argv_pointer"
)

// Profile is one provider, model and effort row (06 §6.3, 04 §4.5).
type Profile struct {
	ID       string `json:"id" yaml:"id"`
	Provider string `json:"provider" yaml:"provider"`
	Model    string `json:"model" yaml:"model"`
	Effort   string `json:"effort,omitempty" yaml:"effort,omitempty"`

	Window          int `json:"window" yaml:"window"`
	EffectiveWindow int `json:"effective_window" yaml:"effective_window"`
	SysOverhead     int `json:"sys_overhead" yaml:"sys_overhead"`

	TokensPerWord float64 `json:"tokens_per_word" yaml:"tokens_per_word"`
	TokensPerByte float64 `json:"tokens_per_byte" yaml:"tokens_per_byte"`

	NativeSchema   bool   `json:"native_schema" yaml:"native_schema"`
	LiveUsage      string `json:"live_usage" yaml:"live_usage"`
	PromptDelivery string `json:"prompt_delivery" yaml:"prompt_delivery"`
	MaxConcurrent  int    `json:"max_concurrent" yaml:"max_concurrent"`
	Enabled        bool   `json:"enabled" yaml:"enabled"`

	PlanFraction  float64 `json:"plan_fraction" yaml:"plan_fraction"`
	KillFraction  float64 `json:"kill_fraction" yaml:"kill_fraction"`
	TaintFraction float64 `json:"taint_fraction" yaml:"taint_fraction"`

	// RateIn and RateOut are USD per million tokens, 0 for subscription
	// billing. Hand-maintained; see SeedProfiles.
	RateIn  float64 `json:"rate_in" yaml:"rate_in"`
	RateOut float64 `json:"rate_out" yaml:"rate_out"`
}

// Default fractions (07 §7.7).
const (
	DefaultPlanFraction       = 0.70
	DefaultPlanFractionNoLive = 0.60
	DefaultKillFraction       = 0.85
	DefaultTaintFraction      = 0.90

	// DefaultTokensPerWord is the plan's floor for prose.
	DefaultTokensPerWord = 1.5
	// DefaultTokensPerByte is the seed for structured text (about 3.3
	// bytes per token) until the smoke measures it.
	DefaultTokensPerByte = 0.30
	// MinTokensPerByte is the floor applied to a measured tokens-per-byte.
	MinTokensPerByte = 0.20
)

// ProfileID renders provider/model[/effort], the form the routing table uses.
func ProfileID(provider, model, effort string) string {
	if effort == "" {
		return provider + "/" + model
	}
	return provider + "/" + model + "/" + effort
}

var knownEfforts = map[string]bool{"none": true, "minimal": true, "low": true, "medium": true, "high": true, "xhigh": true, "max": true, "ultra": true}

// ParseProfileID splits provider/model[/effort]. Models may contain slashes
// (kimi-code/k3), so the last segment is the effort only when it is a known
// effort name.
func ParseProfileID(id string) (provider, model, effort string, ok bool) {
	parts := strings.Split(id, "/")
	if len(parts) < 2 || parts[0] == "" {
		return "", "", "", false
	}
	provider = parts[0]
	rest := parts[1:]
	if len(rest) >= 2 && knownEfforts[rest[len(rest)-1]] {
		effort = rest[len(rest)-1]
		rest = rest[:len(rest)-1]
	}
	model = strings.Join(rest, "/")
	if model == "" {
		return "", "", "", false
	}
	return provider, model, effort, true
}

// PlanCap is the largest planned total the profile accepts (RUN-PACKET-FITS).
func (p Profile) PlanCap() int { return int(math.Floor(p.PlanFraction * float64(p.EffectiveWindow))) }

// KillLine is the live context at which the watchdog kills the run.
func (p Profile) KillLine() int { return int(math.Floor(p.KillFraction * float64(p.EffectiveWindow))) }

// TaintLine is the reported peak above which a run is tainted.
func (p Profile) TaintLine() int {
	return int(math.Floor(p.TaintFraction * float64(p.EffectiveWindow)))
}

// Ceiling is the provider's own hard limit (layer 2 of 02 §2.6): the taint
// line, 0.90 of the effective window by default. Adapters pass it to the
// CLI's context switch.
func (p Profile) Ceiling() int { return p.TaintLine() }

// WithDefaults fills zero fractions, tokenizer constants and the id.
func (p Profile) WithDefaults() Profile {
	if p.ID == "" {
		p.ID = ProfileID(p.Provider, p.Model, p.Effort)
	}
	if p.EffectiveWindow == 0 {
		p.EffectiveWindow = p.Window
	}
	if p.LiveUsage == "" {
		p.LiveUsage = LiveNone
	}
	if p.PlanFraction == 0 {
		if p.LiveUsage == LiveNone {
			p.PlanFraction = DefaultPlanFractionNoLive
		} else {
			p.PlanFraction = DefaultPlanFraction
		}
	}
	if p.KillFraction == 0 {
		p.KillFraction = DefaultKillFraction
	}
	if p.TaintFraction == 0 {
		p.TaintFraction = DefaultTaintFraction
	}
	if p.TokensPerWord == 0 {
		p.TokensPerWord = DefaultTokensPerWord
	}
	if p.TokensPerByte == 0 {
		p.TokensPerByte = DefaultTokensPerByte
	}
	if p.MaxConcurrent == 0 {
		p.MaxConcurrent = 1
	}
	return p
}

// Validate checks the fraction ordering the config validator requires
// (07 §7.7): plan in (0.3, 0.8], plan < kill < taint <= 1.
func (p Profile) Validate() error {
	if p.Provider == "" || p.Model == "" {
		return fmt.Errorf("providers: profile %q needs provider and model", p.ID)
	}
	if p.Window <= 0 || p.EffectiveWindow <= 0 || p.EffectiveWindow > p.Window {
		return fmt.Errorf("providers: profile %q has a bad window (%d) or effective window (%d)", p.ID, p.Window, p.EffectiveWindow)
	}
	if p.PlanFraction <= 0.3 || p.PlanFraction > 0.8 {
		return fmt.Errorf("providers: profile %q plan_fraction %.2f must be in (0.3, 0.8]", p.ID, p.PlanFraction)
	}
	if !(p.PlanFraction < p.KillFraction && p.KillFraction < p.TaintFraction && p.TaintFraction <= 1) {
		return fmt.Errorf("providers: profile %q needs plan < kill < taint <= 1 (got %.2f %.2f %.2f)", p.ID, p.PlanFraction, p.KillFraction, p.TaintFraction)
	}
	return nil
}

// cachedReadDiscount is the fraction of the input rate charged for tokens
// served from a prompt cache (every provider that publishes a rate card
// charges cache reads at about a tenth of the input rate).
const cachedReadDiscount = 0.10

// EstimateUSD prices a run from the profile's rate card: uncached input at
// RateIn, cached input at a tenth of RateIn, output (reasoning included) at
// RateOut. It is 0 for subscription profiles, whose rates are 0.
func EstimateUSD(p Profile, u Usage) float64 {
	cached := u.CachedTokens
	if cached > u.InputTokens {
		cached = u.InputTokens
	}
	uncached := u.InputTokens - cached
	usd := float64(uncached)*p.RateIn/1e6 +
		float64(cached)*p.RateIn*cachedReadDiscount/1e6 +
		float64(u.OutputTokens)*p.RateOut/1e6
	return math.Round(usd*1e6) / 1e6
}

// seed builds one seeded profile.
func seed(provider, model, effort string, window, effective, sys int, native bool, live, delivery string, conc int, enabled bool) Profile {
	p := Profile{
		Provider: provider, Model: model, Effort: effort,
		Window: window, EffectiveWindow: effective, SysOverhead: sys,
		NativeSchema: native, LiveUsage: live, PromptDelivery: delivery,
		MaxConcurrent: conc, Enabled: enabled,
	}
	return p.WithDefaults()
}

// SeedProfiles returns the plan's profile table (04 §4.5, 06 §6.4). Seeded
// values are starting points: the smoke run replaces SysOverhead and the
// tokenizer constants, and stage 0 verifies windows and switches. Rates are
// 0 for subscription providers (claude, codex, kimi) and hand-maintained
// list prices per million tokens for API billing (gemini); check them
// against the current price list before trusting a cost report.
func SeedProfiles() []Profile {
	const (
		k200 = 200_000
		k272 = 272_000
		k500 = 500_000
		m1   = 1_000_000
		kimi = 1_048_576
		k128 = 128_000
		k256 = 256_000
	)
	var out []Profile
	add := func(p Profile) { out = append(out, p) }

	// claude: 200k, per-turn usage, stdin, native schema, subscription.
	add(seed("claude", "claude-opus-4-7", "high", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-opus-4-7", "medium", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-sonnet-4-6", "medium", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-sonnet-4-6", "low", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-haiku-4-5", "low", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	// The 5.5 generation (2026-11-11, Tom: latest and best models): each
	// answered a one-line call on claude CLI 2.1.308. Their API window is
	// 1M; seeded at 200k like 4.7 until 1M on the subscription is verified.
	add(seed("claude", "claude-opus-5-5", "high", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-opus-5-5", "medium", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-sonnet-5-5", "medium", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-sonnet-5-5", "low", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	add(seed("claude", "claude-haiku-5-5", "low", k200, k200, 15000, true, LivePerTurn, DeliverStdin, 2, true))
	// claude [1m]: availability on the subscription unverified (06 §6.8).
	add(seed("claude", "claude-opus-4-7[1m]", "high", m1, m1, 15000, true, LivePerTurn, DeliverStdin, 1, true))
	add(seed("claude", "claude-sonnet-4-6[1m]", "medium", m1, m1, 15000, true, LivePerTurn, DeliverStdin, 1, true))

	// codex: 500k on gpt-6.1-sol, 272k on the gpt-6 and gpt-5.6 family.
	add(seed("codex", "gpt-6.1-sol", "high", k500, k500, 10000, true, LiveNone, DeliverStdin, 2, true))
	add(seed("codex", "gpt-6.1-sol", "medium", k500, k500, 10000, true, LiveNone, DeliverStdin, 2, true))
	add(seed("codex", "gpt-6-sol", "high", k272, k272, 10000, true, LiveNone, DeliverStdin, 2, true))
	add(seed("codex", "gpt-5.6-sol", "high", k272, k272, 10000, true, LiveNone, DeliverStdin, 2, true))

	// gemini: 1M window, effective 500k until compressionThreshold 1.0 is
	// verified; no live usage, so the plan fraction is 0.60; API billing.
	gp := seed("gemini", "gemini-3-pro-preview", "", m1, k500, 8000, false, LiveNone, DeliverStdinPrepend, 3, true)
	gp.RateIn, gp.RateOut = 2.00, 12.00
	add(gp)
	gf := seed("gemini", "gemini-3-flash-preview", "", m1, k500, 8000, false, LiveNone, DeliverStdinPrepend, 3, true)
	gf.RateIn, gf.RateOut = 0.30, 2.50
	add(gf)
	// gemini-3.1-pro-preview: the strongest model the gemini CLI 0.56 offers
	// (2026-10-30). $2.00 and $12.00 per million tokens for prompts up to
	// 200k (ai.google.dev/gemini-api/docs/pricing; $4 and $18 above), so the
	// book budget counts it; a profile without rates would count as free.
	g31 := seed("gemini", "gemini-3.1-pro-preview", "", m1, k500, 8000, false, LiveNone, DeliverStdinPrepend, 3, true)
	g31.RateIn, g31.RateOut = 2.00, 12.00
	add(g31)
	// Optional Gemini 4 Hydrogen profile (enable in config when desired).
	g4h := seed("gemini", "gemini-4-hydrogen", "", m1, k500, 8000, false, LiveNone, DeliverStdinPrepend, 3, false)
	g4h.RateIn, g4h.RateOut = 2.00, 12.00
	add(g4h)

	// kimi: 1M, real-time usage_update, kill at 0.80 because it compacts
	// at 0.85 and has no switch. Effort is config-level, not a flag.
	kk := seed("kimi", "kimi-code/k3", "high", kimi, kimi, 8000, false, LiveNone, DeliverArgvPointer, 2, true)
	kk.KillFraction = 0.80
	add(kk)
	kc := seed("kimi", "kimi-code/kimi-for-coding", "high", kimi, kimi, 8000, false, LiveNone, DeliverArgvPointer, 2, true)
	kc.KillFraction = 0.80
	add(kc)
	// K3.1 Pro: kimi's strongest model (~/.kimi-code/config.toml, 1M
	// context, efforts low, high, max), same kill line as K3.
	k31 := seed("kimi", "kimi-code/k3.1-pro", "high", kimi, kimi, 8000, false, LiveNone, DeliverArgvPointer, 2, true)
	k31.KillFraction = 0.80
	add(k31)
	// K3.1 Pro highspeed (2026-11-06, Tom): the same model served faster,
	// efforts low, high, max, for the roles where kimi's 4 to 7 minute
	// calls were the critical path (annotation, protected and continuity
	// judging); seeded at high, the strongest effort 06 §6.4 allows.
	k31f := seed("kimi", "kimi-code/k3.1-pro-highspeed", "high", kimi, kimi, 8000, false, LiveNone, DeliverArgvPointer, 2, true)
	k31f.KillFraction = 0.80
	add(k31f)

	// grok and muse: windows unverified, seeded 128k, disabled until smoke.
	add(seed("grok", "grok-4.5", "high", k128, k128, 8000, true, LiveNone, DeliverPromptFile, 1, false))
	add(seed("muse", "muse-spark-1.3", "high", k128, k128, 8000, true, LivePerTurn, DeliverPromptFile, 1, false))
	// The strongest of each that 06 §6.4 allows (2026-10-30; never max or
	// ultra): grok-4.7, "SpaceXAI's latest frontier model" in grok's model
	// cache (256k context), at xhigh, and muse-spark-1.3, muse's newest
	// public model, at xhigh.
	add(seed("grok", "grok-4.7", "xhigh", k256, k256, 8000, true, LiveNone, DeliverPromptFile, 1, false))
	add(seed("muse", "muse-spark-1.3", "xhigh", k128, k128, 8000, true, LivePerTurn, DeliverPromptFile, 1, false))
	// agy (2026-11-11, Tom: use the Google plan he pays for, never per-use
	// billing): model ids carry the effort tier, so there is no effort
	// field. The plan's quota, not dollars, is the limit, so the rates stay
	// 0. Windows unverified: seeded at 1M like gemini, effective 500k, as
	// for gemini until a long run measures where agy compacts. No native
	// schema (Gemini refuses the handoff schema's empty enums); per-call
	// usage arrives live on each agent_response step.
	add(seed("agy", "gemini-4-argon-high", "", m1, k500, 15000, false, LivePerTurn, DeliverStdin, 2, true))
	add(seed("agy", "gemini-4-argon-medium", "", m1, k500, 15000, false, LivePerTurn, DeliverStdin, 2, true))
	add(seed("agy", "gemini-3.8-flash-high", "", m1, k500, 15000, false, LivePerTurn, DeliverStdin, 2, true))
	add(seed("agy", "gemini-3.8-flash-medium", "", m1, k500, 15000, false, LivePerTurn, DeliverStdin, 2, true))

	// Direct HTTP API profiles (opt-in via enable: [...], cross-platform,
	// SSE streaming, explicit cache_control prefix boundaries, no CLI boot cost).
	aa := seed("anthropic-api", "claude-sonnet-5-5", "", k200, k200, 1000, false, LiveStream, DeliverStdin, 4, false)
	aa.RateIn, aa.RateOut = 3.00, 15.00
	add(aa)
	oa := seed("openai-api", "gpt-6.1-sol", "", k500, k500, 1000, false, LiveStream, DeliverStdin, 4, false)
	oa.RateIn, oa.RateOut = 2.50, 10.00
	add(oa)
	ga := seed("gemini-api", "gemini-4-hydrogen", "", m1, m1, 1000, false, LiveStream, DeliverStdin, 4, false)
	ga.RateIn, ga.RateOut = 2.00, 12.00
	add(ga)
	return out
}

// FindProfile returns the profile with the given id.
func FindProfile(profiles []Profile, id string) (Profile, bool) {
	for _, p := range profiles {
		if p.ID == id {
			return p, true
		}
	}
	return Profile{}, false
}

// ProfilesFor returns the profiles of one provider in table order.
func ProfilesFor(profiles []Profile, provider string) []Profile {
	var out []Profile
	for _, p := range profiles {
		if p.Provider == provider {
			out = append(out, p)
		}
	}
	return out
}
