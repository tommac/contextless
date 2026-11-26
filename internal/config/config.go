// Package config holds ContextlessAI's settings: where workspaces live, the
// provider profiles, the role routing table, lane sizes, and the selection
// knobs. Everything has a default; a YAML file overrides any part of it.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tommac/contextless/internal/providers"

	"gopkg.in/yaml.v3"
)

// Config is the whole configuration.
type Config struct {
	// Root holds workspaces/, runs/, lanes/ and machine.db.
	Root string `yaml:"root"`
	// Sandbox runs every CLI under the macOS seatbelt profile.
	Sandbox bool `yaml:"sandbox"`

	RecentTurns   int `yaml:"recent_turns"`
	RelevantTurns int `yaml:"relevant_turns"`
	// RecentEpoch aligns the start of the recent turns to multiples of
	// this many turns, which keeps the prompt's prefix stable for the
	// provider's cache, at the price of more turns sent. Zero or less
	// slides the window every message.
	RecentEpoch int `yaml:"recent_epoch"`
	// FullHistory and NoRanked are measurement arms (engine.ApplyContextArm),
	// never set from a file: send every earlier turn, or rank nothing.
	FullHistory bool `yaml:"-"`
	NoRanked    bool `yaml:"-"`
	// RelevantTokens is the budget for ranked atoms (paragraphs of earlier
	// turns chosen for the message), counted at the default 0.30 tokens
	// per byte whatever the profile. A negative value ranks whole turns
	// instead, RelevantTurns of them; a zero in a config file leaves the
	// default in place.
	RelevantTokens int `yaml:"relevant_tokens"`
	// WholeHistoryTokens: while the earlier turns add up to at most this
	// many tokens and fit the first profile, all of them are sent verbatim
	// and nothing is selected. A negative value turns this off; a zero in
	// a config file leaves the default in place.
	WholeHistoryTokens int `yaml:"whole_history_tokens"`
	RecallRounds       int `yaml:"recall_rounds"`
	ContinuationDepth  int `yaml:"continuation_depth"`
	MaxUnits           int `yaml:"max_units"`
	// Attempts is how many profiles a call may try before it fails.
	Attempts int `yaml:"attempts"`
	// IndexBatchMax caps how many unindexed turns one index run takes.
	IndexBatchMax int `yaml:"index_batch_max"`
	// Review is when an answer is checked by a reviewer on a provider that
	// wrote none of it: off, plans (assembled multi-unit answers), always.
	Review string `yaml:"review"`

	// AdaptivePack enables source-diverse round-robin + anchor-outward excerpt
	// packing when ranked candidates overflow relevant_tokens.
	AdaptivePack bool `yaml:"adaptive_pack"`
	// SessionRoute enables two-stage episode/thread-routed candidate discovery
	// interleaved with primary atom hits before packing. The +37 on
	// LongMemEval-S and +68 on LoCoMo-10 it is credited with were measured on
	// a Python prototype that routes by the datasets' own sessions
	// (docs/experiments/longmem); this Go port, which cuts episodes at time
	// gaps, has not been run on either benchmark.
	SessionRoute bool `yaml:"session_route"`
	// AutoReplyWindow automatically co-selects the immediate reply turn (seq+1)
	// when a top-ranked selected turn ends with an unanswered question, closing
	// speaker-attribution and split-Q&A misses.
	AutoReplyWindow bool `yaml:"auto_reply_window"`
	// OmittedCatalogThreshold switches omitted Tier-1 catalog listings from
	// inline 50-turn digest lines to a compact ~215-byte snapshot-pinned
	// external paged catalog handle once this many turns are omitted (default 5000).
	OmittedCatalogThreshold int `yaml:"omitted_catalog_threshold"`
	// StableWorkspace reuses a single per-workspace directory for one-shot runs
	// so provider CLIs do not bust prompt caches via changing run-id paths.
	StableWorkspace bool `yaml:"stable_workspace"`
	// WorkingSetWindow pins/boosts atoms recalled or verified-cited in the
	// last N user turns (0 disables working-set locality).
	WorkingSetWindow int `yaml:"working_set_window"`
	// StatsRouting reorders eligible profiles using measured success rate and
	// latency from the runs table once sufficient observations exist.
	StatsRouting bool `yaml:"stats_routing"`
	// DisagreementRerun runs one reconciliation round when parallel plan units
	// cite the same text with conflicting values before assembly.
	DisagreementRerun bool `yaml:"disagreement_rerun"`
	// LaneWeights optionally overrides exact/full_text/value/working_set weights.
	LaneWeights map[string]float64 `yaml:"lane_weights"`
	// Atlas is the map of the conversation: checked idea notes shown as one
	// line each, naming the paragraphs they point at, for the model to ask
	// for by id. "auto" (the default) shows the map when the workspace has
	// notes; "on" also has the indexer write notes for the turns it indexes
	// (more output per index run); "off" never shows it.
	Atlas string `yaml:"atlas"`
	// AtlasMapTokens is the budget for the map's lines, counted like
	// RelevantTokens. Past it, the lines ranked for the message are shown
	// and the rest are named by topic.
	AtlasMapTokens int `yaml:"atlas_map_tokens"`
	// AtlasPullTokens is how many tokens of paragraphs one recall request
	// brings back while the map is shown. A negative value sets no limit
	// beyond what fits the packet.
	AtlasPullTokens int `yaml:"atlas_pull_tokens"`

	TimeoutSeconds      int `yaml:"timeout_seconds"`
	IndexTimeoutSeconds int `yaml:"index_timeout_seconds"`
	StallSeconds        int `yaml:"stall_seconds"`

	// Lanes caps concurrent runs per provider across the whole machine.
	Lanes map[string]int `yaml:"lanes"`
	// Enable and Disable flip seeded profiles by id.
	Enable  []string `yaml:"enable"`
	Disable []string `yaml:"disable"`
	// Profiles adds profiles or replaces seeded ones with the same id.
	Profiles []providers.Profile `yaml:"profiles"`
	// Routing maps a role to profile ids in preference order.
	Routing map[string][]string `yaml:"routing"`
	// Binaries overrides CLI paths by provider.
	Binaries map[string]string `yaml:"binaries"`

	// sandboxFalse records an explicit "sandbox: false" (yaml cannot tell
	// absent from false on a plain bool).
	sandboxFalse bool
}

// Values of Config.Atlas.
const (
	AtlasOff  = "off"
	AtlasAuto = "auto"
	AtlasOn   = "on"
)

// Roles in the routing table.
const (
	RoleResponder  = "responder"
	RoleCoder      = "coder"
	RoleIndexer    = "indexer"
	RoleLongReader = "long_reader"
	RoleResearcher = "researcher"
	RoleEditor     = "editor"
	RoleReviewer   = "reviewer"
	RoleAssembler  = "assembler"
)

// Default returns the built-in configuration. Routing follows BookWriter's
// measured provider strengths (BookWriter's notes of 2026-11-06, not in this repository):
// codex leads structured work, summaries and annotation; claude leads
// conversation and prose; kimi and Gemini (through agy, on the paid Google
// plan rather than per-use API billing) hold the 1M windows; grok is the
// networked researcher; muse edits lines.
func Default() Config {
	home, _ := os.UserHomeDir()
	const (
		opus      = "claude/claude-opus-5-5/high"
		sonnet    = "claude/claude-sonnet-5-5/medium"
		haiku     = "claude/claude-haiku-5-5/low"
		codexHigh = "codex/gpt-6.1-sol/high"
		codexMed  = "codex/gpt-6.1-sol/medium"
		kimiPro   = "kimi/kimi-code/k3.1-pro/high"
		kimiFast  = "kimi/kimi-code/k3.1-pro-highspeed/high"
		argon     = "agy/gemini-4-argon-high"
		grok      = "grok/grok-4.7/xhigh"
		muse      = "muse/muse-spark-1.3/xhigh"
	)
	return Config{
		Root:                    filepath.Join(home, ".contextless"),
		Sandbox:                 true,
		RecentTurns:             8,
		RelevantTurns:           6,
		RelevantTokens:          8000,
		WholeHistoryTokens:      60000,
		RecallRounds:            2,
		ContinuationDepth:       3,
		MaxUnits:                8,
		Attempts:                3,
		IndexBatchMax:           12,
		Review:                  "plans",
		OmittedCatalogThreshold: 5000,
		WorkingSetWindow:        4,
		Atlas:                   AtlasAuto,
		AtlasMapTokens:          40000,
		AtlasPullTokens:         8000,
		TimeoutSeconds:          1800,
		IndexTimeoutSeconds:     900,
		StallSeconds:            900,
		Lanes:                   map[string]int{"claude": 2, "codex": 2, "kimi": 2, "agy": 2, "gemini": 1, "grok": 1, "muse": 2, "fake": 4},
		Enable:                  []string{grok, muse},
		Routing: map[string][]string{
			RoleResponder:  {opus, codexHigh, kimiPro, argon},
			RoleCoder:      {codexHigh, opus, kimiPro},
			RoleIndexer:    {codexMed, haiku, kimiFast},
			RoleLongReader: {kimiPro, argon, codexHigh},
			RoleResearcher: {grok, codexHigh},
			RoleEditor:     {muse, sonnet, opus},
			RoleReviewer:   {codexHigh, opus, kimiFast, muse},
			RoleAssembler:  {opus, codexHigh, kimiPro},
		},
	}
}

// Load reads path over the defaults; a missing file is the defaults.
func Load(path string) (Config, error) {
	c := Default()
	if path == "" {
		return c, nil
	}
	data, err := os.ReadFile(expand(path))
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	var over Config
	if err := yaml.Unmarshal(data, &over); err != nil {
		return c, fmt.Errorf("config %s: %w", path, err)
	}
	merge(&c, over)
	return c, nil
}

func merge(c *Config, o Config) {
	if o.Root != "" {
		c.Root = expand(o.Root)
	}
	setInt := func(dst *int, v int) {
		if v > 0 {
			*dst = v
		}
	}
	setInt(&c.RecentTurns, o.RecentTurns)
	setInt(&c.RelevantTurns, o.RelevantTurns)
	if o.RecentEpoch != 0 {
		c.RecentEpoch = o.RecentEpoch
	}
	// a negative value turns these off, so any non-zero value is taken
	if o.RelevantTokens != 0 {
		c.RelevantTokens = o.RelevantTokens
	}
	if o.WholeHistoryTokens != 0 {
		c.WholeHistoryTokens = o.WholeHistoryTokens
	}
	setInt(&c.RecallRounds, o.RecallRounds)
	setInt(&c.ContinuationDepth, o.ContinuationDepth)
	setInt(&c.MaxUnits, o.MaxUnits)
	setInt(&c.Attempts, o.Attempts)
	setInt(&c.IndexBatchMax, o.IndexBatchMax)
	setInt(&c.TimeoutSeconds, o.TimeoutSeconds)
	setInt(&c.IndexTimeoutSeconds, o.IndexTimeoutSeconds)
	setInt(&c.StallSeconds, o.StallSeconds)
	if o.Review != "" {
		c.Review = o.Review
	}
	if o.AdaptivePack {
		c.AdaptivePack = true
	}
	if o.SessionRoute {
		c.SessionRoute = true
	}
	if o.AutoReplyWindow {
		c.AutoReplyWindow = true
	}
	if o.OmittedCatalogThreshold != 0 {
		c.OmittedCatalogThreshold = o.OmittedCatalogThreshold
	}
	if o.StableWorkspace {
		c.StableWorkspace = true
	}
	if o.WorkingSetWindow != 0 {
		c.WorkingSetWindow = o.WorkingSetWindow
	}
	if o.StatsRouting {
		c.StatsRouting = true
	}
	if o.DisagreementRerun {
		c.DisagreementRerun = true
	}
	switch a := strings.ToLower(strings.TrimSpace(o.Atlas)); a {
	case AtlasOff, AtlasAuto, AtlasOn:
		c.Atlas = a
	}
	setInt(&c.AtlasMapTokens, o.AtlasMapTokens)
	if o.AtlasPullTokens != 0 {
		c.AtlasPullTokens = o.AtlasPullTokens
	}
	if len(o.LaneWeights) > 0 {
		if c.LaneWeights == nil {
			c.LaneWeights = map[string]float64{}
		}
		for k, v := range o.LaneWeights {
			c.LaneWeights[k] = v
		}
	}
	for k, v := range o.Lanes {
		c.Lanes[k] = v
	}
	c.Enable = append(c.Enable, o.Enable...)
	c.Disable = append(c.Disable, o.Disable...)
	c.Profiles = append(c.Profiles, o.Profiles...)
	for k, v := range o.Routing {
		c.Routing[k] = v
	}
	if o.Binaries != nil {
		c.Binaries = o.Binaries
	}
	// sandbox is a bool: only an explicit "sandbox: false" turns it off
	if o.sandboxFalse {
		c.Sandbox = false
	}
}

// UnmarshalYAML notes an explicit "sandbox: false".
func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	type plain Config
	var p plain
	if err := n.Decode(&p); err != nil {
		return err
	}
	*c = Config(p)
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == "sandbox" && n.Content[i+1].Value == "false" {
			c.sandboxFalse = true
		}
	}
	return nil
}

// AllProfiles returns the seeded profiles with Enable, Disable and the
// overrides applied.
func (c Config) AllProfiles() []providers.Profile {
	seeded := providers.SeedProfiles()
	byID := map[string]int{}
	for i, p := range seeded {
		byID[p.ID] = i
	}
	for _, p := range c.Profiles {
		p = p.WithDefaults()
		if i, ok := byID[p.ID]; ok {
			seeded[i] = p
		} else {
			byID[p.ID] = len(seeded)
			seeded = append(seeded, p)
		}
	}
	for _, id := range c.Enable {
		if i, ok := byID[id]; ok {
			seeded[i].Enabled = true
		}
	}
	for _, id := range c.Disable {
		if i, ok := byID[id]; ok {
			seeded[i].Enabled = false
		}
	}
	return seeded
}

// WorkspaceDir is where a named workspace lives.
func (c Config) WorkspaceDir(name string) string {
	if strings.ContainsRune(name, os.PathSeparator) {
		return expand(name)
	}
	return filepath.Join(c.Root, "workspaces", name)
}

// RunsDir is where a workspace's run directories live (outside the vault).
func (c Config) RunsDir(name string) string {
	return filepath.Join(c.Root, "runs", filepath.Base(name))
}

func expand(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}

// Sample is a commented config file for `cx init-config`.
const Sample = `# ContextlessAI configuration. Every key is optional; defaults are shown.
# root: ~/.contextless          # workspaces/, runs/, lanes/, machine.db
# sandbox: true                 # run CLIs under the macOS seatbelt profile
# recent_turns: 8               # last N turns always sent verbatim
# recent_epoch: 0               # start the recent turns at a multiple of this (more tokens, a steadier prompt prefix); 0 slides
# relevant_turns: 6             # ranked whole turns, used only when relevant_tokens is off
# relevant_tokens: 8000         # budget for ranked atoms (paragraphs of earlier turns); -1 ranks whole turns
# whole_history_tokens: 60000   # send every earlier turn while they total at most this and fit; -1 is off
# recall_rounds: 2              # times a model may ask for omitted turns
# continuation_depth: 3         # partial answers continued at most N times
# max_units: 8                  # units a plan may fan out to
# review: plans                 # off | plans | always: a reviewer on another provider checks the answer
# adaptive_pack: false          # source-diverse round-robin + anchor-outward excerpt packing
# session_route: false          # two-stage episode/thread candidate routing interleaved with atoms
# auto_reply_window: false      # co-select immediate reply turn (seq+1) for question-ending anchors
# omitted_catalog_threshold: 5000 # offload large omitted turn listings into content-addressed .cxi pages
# stable_workspace: false       # reuse one workspace dir for one-shot runs to preserve prompt caches
# atlas: auto                   # the map of the conversation (checked idea notes, one line each): auto shows it
#                               # where notes exist, on also has the indexer write notes, off never shows it
# atlas_map_tokens: 40000       # budget for the map; past it the lines ranked for the message are shown
# atlas_pull_tokens: 8000       # tokens of paragraphs one recall request brings back while the map is shown
# working_set_window: 4         # boost atoms recalled or verified-cited in the last N user turns (0 disables)
# stats_routing: false          # reorder eligible profiles using observed success rate & latency
# disagreement_rerun: false     # run one reconciliation pass when parallel plan units cite conflicting claims
# lane_weights:                 # optional RRF lane weights (exact, full_text, value, working_set)
#   exact: 1.0
#   full_text: 1.0
#   value: 1.0
#   working_set: 0.5
#
# lanes:                        # concurrent runs per provider, machine-wide
#   claude: 2
#   codex: 2
#   kimi: 2
#   agy: 2
#   gemini: 1
#   grok: 1
#   muse: 2
#
# Optional profiles you can add to enable (direct HTTP API profiles require
# ANTHROPIC_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY in the environment):
#   - gemini/gemini-4-hydrogen
#   - anthropic-api/claude-sonnet-5-5
#   - openai-api/gpt-6.1-sol
#   - gemini-api/gemini-4-hydrogen
#
# enable: [grok/grok-4.7/xhigh, muse/muse-spark-1.3/xhigh]
# disable: []
#
# routing:                      # role: [profile ids in preference order]
#   responder:   [claude/claude-opus-5-5/high, codex/gpt-6.1-sol/high, kimi/kimi-code/k3.1-pro/high, agy/gemini-4-argon-high]
#   coder:       [codex/gpt-6.1-sol/high, claude/claude-opus-5-5/high, kimi/kimi-code/k3.1-pro/high]
#   indexer:     [codex/gpt-6.1-sol/medium, claude/claude-haiku-5-5/low, kimi/kimi-code/k3.1-pro-highspeed/high]
#   long_reader: [kimi/kimi-code/k3.1-pro/high, agy/gemini-4-argon-high, codex/gpt-6.1-sol/high]
#   researcher:  [grok/grok-4.7/xhigh, codex/gpt-6.1-sol/high]
#   editor:      [muse/muse-spark-1.3/xhigh, claude/claude-sonnet-5-5/medium, claude/claude-opus-5-5/high]
#   reviewer:    [codex/gpt-6.1-sol/high, claude/claude-opus-5-5/high, kimi/kimi-code/k3.1-pro-highspeed/high, muse/muse-spark-1.3/xhigh]
#   assembler:   [claude/claude-opus-5-5/high, codex/gpt-6.1-sol/high, kimi/kimi-code/k3.1-pro/high]
`
