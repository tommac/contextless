package providers

import (
	"errors"
	"path/filepath"
)

// Event kinds.
const (
	KindUsage      = "usage"
	KindCompaction = "compaction"
	KindRateLimit  = "rate_limit"
	KindError      = "error"
	KindText       = "text"
	KindTool       = "tool"
	KindOther      = "other"
)

// Event is one normalized line of a provider's stream.
type Event struct {
	Kind              string `json:"kind"`
	ContextTokens     int    `json:"context_tokens,omitempty"`
	Window            int    `json:"window,omitempty"`
	RetryAfterSeconds int    `json:"retry_after_seconds,omitempty"`
	Message           string `json:"message,omitempty"`
	Raw               string `json:"raw,omitempty"`
}

// Usage is the normalized usage of one run (the run_usage row).
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	CachedTokens int `json:"cached_tokens"`
	// CacheWriteTokens is the part of InputTokens written to a prompt
	// cache, billed at a premium where the provider charges one; 0 where
	// the CLI does not report it.
	CacheWriteTokens int     `json:"cache_write_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	ReasoningTokens  int     `json:"reasoning_tokens"`
	PeakContext      int     `json:"peak_context"`
	Turns            int     `json:"turns"`
	WallMs           int64   `json:"wall_ms"`
	USDReported      float64 `json:"usd_reported"`
	HasUSD           bool    `json:"has_usd"`
}

// ExitClass classifies how a run ended, from data only.
type ExitClass string

// Exit classes. See the package documentation for the precedence.
const (
	Ok                ExitClass = "ok"
	OkBlocked         ExitClass = "ok_blocked"
	TaintedCompaction ExitClass = "tainted_compaction"
	InvalidHandoff    ExitClass = "invalid_handoff"
	ScopeViolation    ExitClass = "scope_violation"
	Timeout           ExitClass = "timeout"
	Stalled           ExitClass = "stalled"
	Quota             ExitClass = "quota"
	Auth              ExitClass = "auth"
	Crash             ExitClass = "crash"
	ContextBlocked    ExitClass = "context_blocked"
)

// ErrNoUsage is returned by FinalUsage when the output carries no usage.
var ErrNoUsage = errors.New("providers: no usage found in the provider output")

// Invocation is everything an adapter needs to build one command.
type Invocation struct {
	RunID   string
	Role    string
	Profile Profile

	// Workspace is the packet directory and the child's working directory.
	Workspace string
	// BriefPath, HandoffPath, SchemaPath and StrictSchemaPath are absolute.
	// HandoffPath defaults to <Workspace>/.cx/handoff.json; codex passes only
	// StrictSchemaPath as --output-schema and runs without one when it is
	// empty.
	BriefPath        string
	HandoffPath      string
	SchemaPath       string
	StrictSchemaPath string

	// RunDir receives the streams, the pgid file, the sandbox profile and the
	// compaction marker.
	RunDir string

	// Network is true only for the researcher role.
	Network bool

	// OneShot marks a call whose brief holds everything it needs (a gate
	// judge: every record inlined, nothing to read), so the adapter keeps
	// only what writing the handoff takes. A codex judge made 8 to 10
	// turns and 320K to 426K cumulative input for a 72K context, re-reading
	// files the brief already listed (2026-11-06).
	OneShot bool

	// Sandboxed is set by Run: the command runs under sandbox-exec. Seatbelt
	// cannot nest, so an adapter whose CLI applies its own seatbelt (codex)
	// turns that off and lets the outer profile confine it.
	Sandboxed bool

	MaxBudgetUSD   float64
	MaxTurns       int
	TimeoutSeconds int
	// StallSeconds kills the run when neither the streams nor the workspace
	// moved for that long; 0 disables the stall watchdog.
	StallSeconds int

	// Env is applied last to the child environment and wins.
	Env map[string]string
}

// Adapter is one provider's command builder and stream parser.
type Adapter interface {
	Provider() string
	BuildCommand(inv Invocation) (argv []string, env []string, stdin []byte, err error)
	// ParseEvent maps one stream line to one Event; ok is false for lines
	// that are not events (blank lines, non-JSON chatter).
	ParseEvent(line []byte) (Event, bool)
	FinalUsage(stdout []byte, eventsSeen []Event) (Usage, error)
	ExitClass(exitCode int, stderr []byte, events []Event) ExitClass
	// InstructionFile is the file the CLI auto-loads from the workspace.
	InstructionFile() string
	// ProjectSettings are provider-local files to drop into the workspace,
	// keyed by workspace-relative path.
	ProjectSettings(inv Invocation) map[string][]byte
}

// handoffPath returns HandoffPath or its default.
func (inv Invocation) handoffPath() string {
	if inv.HandoffPath != "" {
		return inv.HandoffPath
	}
	return filepath.Join(inv.Workspace, ".cx", "handoff.json")
}
