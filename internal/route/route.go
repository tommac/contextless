// Package route picks a provider profile for a role and admits runs into
// per-provider lanes. It joins BookWriter's role routing table and cooldown
// ladder with MultiAIWorkspace's machine-wide lanes: a lane is a lock file
// held with flock, so several ContextlessAI processes on one machine share
// the caps, and a crashed process frees its lane the moment it dies.
package route

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
)

// Router is the routing table plus machine-wide health and lanes.
type Router struct {
	mu       sync.Mutex
	cfg      config.Config
	profiles map[string]providers.Profile
	adapters map[string]providers.Adapter
	health   *store.Store // machine-wide provider health
	lanesDir string
	now      func() time.Time
}

// New builds a router. health is the machine-wide store (machine.db).
func New(cfg config.Config, health *store.Store) (*Router, error) {
	r := &Router{cfg: cfg, profiles: map[string]providers.Profile{}, adapters: map[string]providers.Adapter{},
		health: health, lanesDir: filepath.Join(cfg.Root, "lanes"), now: time.Now}
	for _, p := range cfg.AllProfiles() {
		r.profiles[p.ID] = p.WithDefaults()
	}
	if err := os.MkdirAll(r.lanesDir, 0o755); err != nil {
		return nil, err
	}
	return r, nil
}

// SetAdapter registers (or replaces) the adapter for a provider. Tests use
// it to put a scripted adapter under a real provider name.
func (r *Router) SetAdapter(provider string, a providers.Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[provider] = a
}

// SetProfile registers or replaces a profile.
func (r *Router) SetProfile(p providers.Profile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[p.ID] = p.WithDefaults()
}

// Adapter returns the adapter for a provider, building the stock one on
// first use with any configured binary override.
func (r *Router) Adapter(provider string) (providers.Adapter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.adapters[provider]; ok {
		return a, nil
	}
	a, err := providers.NewAdapter(provider)
	if err != nil {
		return nil, err
	}
	if bin := r.cfg.Binaries[provider]; bin != "" {
		a = withBinary(a, bin)
	}
	r.adapters[provider] = a
	return a, nil
}

func withBinary(a providers.Adapter, bin string) providers.Adapter {
	switch x := a.(type) {
	case providers.ClaudeAdapter:
		x.Binary = bin
		return x
	case providers.CodexAdapter:
		x.Binary = bin
		return x
	case providers.KimiAdapter:
		x.Binary = bin
		return x
	case providers.GeminiAdapter:
		x.Binary = bin
		return x
	case providers.AgyAdapter:
		x.Binary = bin
		return x
	case providers.GrokAdapter:
		x.Binary = bin
		return x
	case providers.MuseAdapter:
		x.Binary = bin
		return x
	}
	return a
}

// Profile returns a profile by id.
func (r *Router) Profile(id string) (providers.Profile, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.profiles[id]
	return p, ok
}

// Skipped explains why a routed profile was not offered.
type Skipped struct {
	Profile string
	Reason  string
}

// Candidates returns the role's enabled profiles in routing order, minus
// excluded providers and providers in cooldown. It also says why each
// other routed profile was skipped.
func (r *Router) Candidates(role string, excludeProviders []string) ([]providers.Profile, []Skipped) {
	ids := r.cfg.Routing[role]
	if len(ids) == 0 {
		ids = r.cfg.Routing[config.RoleResponder]
	}
	excluded := map[string]bool{}
	for _, p := range excludeProviders {
		excluded[p] = true
	}
	var out []providers.Profile
	var skipped []Skipped
	for _, id := range ids {
		r.mu.Lock()
		p, ok := r.profiles[id]
		r.mu.Unlock()
		switch {
		case !ok:
			skipped = append(skipped, Skipped{id, "unknown profile"})
		case !p.Enabled:
			skipped = append(skipped, Skipped{id, "disabled"})
		case excluded[p.Provider]:
			skipped = append(skipped, Skipped{id, "excluded for this call"})
		default:
			if until, reason := r.Cooling(p.Provider); !until.IsZero() {
				skipped = append(skipped, Skipped{id, fmt.Sprintf("cooling until %s (%s)", until.Local().Format("15:04"), reason)})
				continue
			}
			out = append(out, p)
		}
	}
	return out, skipped
}

// RankByStats reorders candidate profiles using empirical reliability and
// latency from store.Stats() when cfg.StatsRouting is enabled. Profiles with
// fewer than 5 observed runs keep a neutral prior (0.85) so new or rarely
// tried models are not starved.
func (r *Router) RankByStats(cands []providers.Profile, kind string, stats []store.Stat) []providers.Profile {
	if !r.cfg.StatsRouting || len(cands) < 2 || len(stats) == 0 {
		return cands
	}
	byProf := map[string]store.Stat{}
	maxWall := int64(1)
	for _, st := range stats {
		if kind != "" && st.Kind != kind {
			continue
		}
		byProf[st.Profile] = st
		if st.MedianWall > maxWall {
			maxWall = st.MedianWall
		}
	}
	scoreOf := func(p providers.Profile) float64 {
		st, ok := byProf[p.ID]
		if !ok || st.Runs < 5 {
			return 0.85
		}
		rate := float64(st.OK) / float64(st.Runs)
		latPen := 0.15 * float64(st.MedianWall) / float64(maxWall)
		return rate - latPen
	}
	out := append([]providers.Profile(nil), cands...)
	sort.SliceStable(out, func(i, j int) bool {
		return scoreOf(out[i]) > scoreOf(out[j])
	})
	return out
}

// Cooling returns the cooldown end of a provider, zero when it is usable.
func (r *Router) Cooling(provider string) (time.Time, string) {
	if r.health == nil {
		return time.Time{}, ""
	}
	until, reason, _ := r.health.Health(provider)
	if until.After(r.now()) {
		return until, reason
	}
	return time.Time{}, ""
}

// Penalize records a provider failure and sets its cooldown, following
// BookWriter's ladder (reject.go providerFault): the reset time the
// message names, else its retry-after, else 15 minutes doubling per strike
// up to an hour; a weekly window rests at least 6 hours and a named window
// at least a quarter of it. Auth failures rest 6 hours (only a human login
// fixes them); other faults 2 minutes doubling up to 30. No cooldown
// exceeds 7 days.
func (r *Router) Penalize(provider string, class providers.ExitClass, retryAfter int, text string) time.Duration {
	if r.health == nil {
		return 0
	}
	_, _, strikes := r.health.Health(provider)
	now := r.now()
	var d time.Duration
	switch class {
	case providers.Auth:
		d = 6 * time.Hour
	case providers.Quota:
		if at, ok := providers.ResetAt(text); ok && at.After(now) {
			d = at.Sub(now) + time.Minute
			break
		}
		if retryAfter <= 0 {
			retryAfter = providers.RetryAfter(text)
		}
		if retryAfter > 0 {
			d = time.Duration(retryAfter) * time.Second
			break
		}
		d = 15 * time.Minute << min(strikes, 2)
		if w := time.Duration(providers.QuotaWindow(text)) * time.Second; w > 0 {
			d = max(d, min(w/4, 6*time.Hour))
		}
	default:
		d = 2 * time.Minute << min(strikes, 4)
	}
	d = min(d, 7*24*time.Hour)
	reason := string(class)
	if ex := providers.QuotaExcerpt(text); ex != "" && class == providers.Quota {
		reason += ": " + ex
	}
	_ = r.health.Cooldown(provider, reason, now.Add(d))
	return d
}

// Healthy clears a provider's strike count after a clean run.
func (r *Router) Healthy(provider string) {
	if r.health != nil {
		r.health.ClearStrikes(provider)
	}
}

// Acquire takes one lane of a provider, waiting until one is free or ctx
// ends. The returned function releases it.
func (r *Router) Acquire(ctx context.Context, provider string) (func(), error) {
	n := r.cfg.Lanes[provider]
	if n <= 0 {
		n = 1
	}
	for {
		for i := 0; i < n; i++ {
			path := filepath.Join(r.lanesDir, fmt.Sprintf("%s.%d.lock", provider, i))
			f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
			if err != nil {
				return nil, err
			}
			if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
				_ = f.Truncate(0)
				_, _ = fmt.Fprintf(f, "pid %d since %s\n", os.Getpid(), time.Now().Format(time.RFC3339))
				return func() {
					_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
					f.Close()
				}, nil
			}
			f.Close()
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(750 * time.Millisecond):
		}
	}
}

// Table renders the routing table with each profile's state, for `cx
// providers`.
func (r *Router) Table() string {
	var b strings.Builder
	roles := make([]string, 0, len(r.cfg.Routing))
	for role := range r.cfg.Routing {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	for _, role := range roles {
		fmt.Fprintf(&b, "%s\n", role)
		for _, id := range r.cfg.Routing[role] {
			p, ok := r.profiles[id]
			state := "ok"
			switch {
			case !ok:
				state = "unknown profile"
			case !p.Enabled:
				state = "disabled"
			default:
				if until, reason := r.Cooling(p.Provider); !until.IsZero() {
					state = fmt.Sprintf("cooling until %s (%s)", until.Local().Format("Jan 2 15:04"), reason)
				} else if _, err := providers.ResolveBinary(r.binaryName(p.Provider)); err != nil {
					state = "CLI not found"
				}
			}
			window := ""
			if ok {
				window = fmt.Sprintf("%dk window, plan cap %dk", p.EffectiveWindow/1000, p.PlanCap()/1000)
			}
			fmt.Fprintf(&b, "  %-42s %-28s %s\n", id, window, state)
		}
	}
	fmt.Fprintf(&b, "\nlanes (machine-wide):")
	provs := make([]string, 0, len(r.cfg.Lanes))
	for p := range r.cfg.Lanes {
		provs = append(provs, p)
	}
	sort.Strings(provs)
	for _, p := range provs {
		if p != "fake" {
			fmt.Fprintf(&b, " %s=%d", p, r.cfg.Lanes[p])
		}
	}
	b.WriteString("\n")
	return b.String()
}

func (r *Router) binaryName(provider string) string {
	if bin := r.cfg.Binaries[provider]; bin != "" {
		return bin
	}
	return provider
}
