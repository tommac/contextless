package packet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Task kinds.
const (
	KindRespond  = "respond"
	KindUnit     = "unit"
	KindAssemble = "assemble"
	KindContinue = "continue"
	KindIndex    = "index"
	KindReview   = "review"
	KindRevise   = "revise"
)

// Handoff statuses.
const (
	StatusStarted     = "started"
	StatusCompleted   = "completed"
	StatusPartial     = "partial"
	StatusNeedsRecall = "needs_recall"
	StatusPlan        = "plan"
	StatusBlocked     = "blocked"
	// StatusNoInformation is a unit that found nothing for its task in
	// its packet and tools. It is a result, not a failure: the unit says
	// so instead of inventing an answer from unrelated material.
	StatusNoInformation = "no_information"
)

// Support says how a unit's claim stands to its evidence, as the unit
// declares it. The engine checks the quote and lowers a claim whose quote
// it cannot find.
const (
	SupportQuoted      = "quoted"      // the quote, copied from the evidence, states the claim
	SupportInferred    = "inferred"    // the claim follows from the evidence but is not stated there
	SupportUnsupported = "unsupported" // the unit's own judgment or general knowledge
)

// UnitClaim is one statement of fact a unit's output relies on, with
// where it is stated: turn, paragraph or part ids of the conversation,
// files the unit wrote, or the URLs of web sources.
type UnitClaim struct {
	Claim    string   `json:"claim"`
	Evidence []string `json:"evidence"`
	Quote    string   `json:"quote"`
	Support  string   `json:"support"`
}

// UnitRoles are the roles a plan may assign to a unit.
var UnitRoles = []string{"responder", "coder", "long_reader", "researcher", "editor", "reviewer"}

// OutputTokens is the planned output per kind (reasoning included).
var OutputTokens = map[string]int{
	KindRespond:  24000,
	KindUnit:     24000,
	KindContinue: 24000,
	KindAssemble: 32000,
	KindIndex:    12000,
	KindReview:   10000,
	KindRevise:   32000,
}

// Unit is one piece of a plan.
type Unit struct {
	ID           string   `json:"id"`
	Role         string   `json:"role"`
	Title        string   `json:"title"`
	Instructions string   `json:"instructions"`
	Needs        []string `json:"needs"`
	// Repo names an attached repository a coder unit changes; "" for none.
	Repo string `json:"repo"`
	// DependsOn names earlier units of the plan whose records this unit
	// needs; it runs after them, with their records in its packet.
	DependsOn []string `json:"depends_on"`
}

// OmittedClaim is a unit claim an assembler left out of its reply, and why.
type OmittedClaim struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// RecallOp is a typed, deterministic, zero-LLM recall operation that goes
// beyond bare IDs and bag-of-words search:
//   - exchange:      predecessor + target + successor turns (speaker/reply attribution)
//   - open_block:    full enclosing Markdown list, table or fenced code block
//   - timeline:      chronological active + superseded facts/decisions for a topic
//   - grep:          exact literal or regex search over earlier turns with line offsets
//   - list_omitted:  snapshot-verified paginated enumeration of an external .cxi catalog
//   - check_omitted: snapshot-verified O(1) membership check in an external .cxi catalog
//   - notes:         the map lines of the notes under a topic, or matching search words
type RecallOp struct {
	Op       string `json:"op"`
	Target   string `json:"target"`
	Radius   int    `json:"radius,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Snapshot string `json:"snapshot,omitempty"` // optional pinned snapshot root; mismatched ops are rejected
	Index    string `json:"index,omitempty"`    // optional .cxi omitted catalog digest
	After    int    `json:"after,omitempty"`    // optional cursor position for list_omitted
}

// Handoff is what a respond, unit, assemble or continue run returns.
type Handoff struct {
	RunID     string   `json:"run_id"`
	Status    string   `json:"status"`
	Answer    string   `json:"answer"`
	RecallIDs []string `json:"recall_ids"`
	// RecallQueries are searches over every earlier turn, for when no
	// catalog line points to what the model needs (LCM's grep).
	RecallQueries []string `json:"recall_queries"`
	// RecallOps are typed recall verbs (exchange, open_block, timeline, grep).
	RecallOps      []RecallOp `json:"recall_ops,omitempty"`
	Units          []Unit     `json:"units"`
	UnitsDone      []string   `json:"units_done"`
	UnitsRemaining []string   `json:"units_remaining"`
	NextStep       string     `json:"next_step"`
	Notes          string     `json:"notes"`
	// A unit's record (UnitSchema): the claims its answer relies on, the
	// choices it made that another unit could have made differently, what
	// it could not settle, and evidence that disagrees with its answer.
	Claims        []UnitClaim `json:"claims,omitempty"`
	Assumptions   []string    `json:"assumptions,omitempty"`
	OpenQuestions []string    `json:"open_questions,omitempty"`
	Conflicts     []string    `json:"conflicts,omitempty"`
	// SharedDecisions are a plan's choices every unit must follow (names,
	// interfaces, formats, scope), copied into each unit's packet.
	SharedDecisions []string `json:"shared_decisions,omitempty"`
	// An assembler's accounting (AssembleSchema): the unit claims its
	// reply relies on, and every other claim with why it was left out.
	ClaimsUsed    []string       `json:"claims_used,omitempty"`
	ClaimsOmitted []OmittedClaim `json:"claims_omitted,omitempty"`
}

// recallOpsSchema is the schema of a handoff's recall_ops. Every property
// is required, as OpenAI strict mode demands; an op leaves the ones it does
// not use empty or zero. Without this entry a schema with
// additionalProperties false forbids the typed recall verbs the
// instructions offer, and a provider held to the schema can never use one.
func recallOpsSchema() map[string]any {
	str := map[string]any{"type": "string"}
	num := map[string]any{"type": "integer"}
	return map[string]any{"type": "array", "items": map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"op", "target", "radius", "limit", "snapshot", "index", "after"},
		"properties": map[string]any{
			"op": str, "target": str, "radius": num, "limit": num, "snapshot": str, "index": str, "after": num,
		},
	}}
}

// RespondSchema is the JSON schema of Handoff. Every property is required
// so the same schema is valid in OpenAI strict mode (codex, muse).
func RespondSchema(allowPlan bool) string {
	statuses := []string{StatusCompleted, StatusPartial, StatusNeedsRecall, StatusBlocked}
	if allowPlan {
		statuses = append(statuses, StatusPlan)
	}
	strs := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	s := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"run_id", "status", "answer", "recall_ids", "recall_queries", "recall_ops", "units", "shared_decisions", "units_done", "units_remaining", "next_step", "notes"},
		"properties": map[string]any{
			"run_id":         map[string]any{"type": "string"},
			"status":         map[string]any{"type": "string", "enum": statuses},
			"answer":         map[string]any{"type": "string"},
			"recall_ids":     strs,
			"recall_queries": strs,
			"recall_ops":     recallOpsSchema(),
			"units": map[string]any{"type": "array", "items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"id", "role", "title", "instructions", "needs", "repo", "depends_on"},
				"properties": map[string]any{
					"id":           map[string]any{"type": "string"},
					"role":         map[string]any{"type": "string", "enum": UnitRoles},
					"title":        map[string]any{"type": "string"},
					"instructions": map[string]any{"type": "string"},
					"needs":        strs,
					"repo":         map[string]any{"type": "string"},
					"depends_on":   strs,
				},
			}},
			"shared_decisions": strs,
			"units_done":       strs,
			"units_remaining":  strs,
			"next_step":        map[string]any{"type": "string"},
			"notes":            map[string]any{"type": "string"},
		},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// UnitSchema is the JSON schema of a unit's record: the handoff of a plan
// unit or of one part of an oversized message. Every property is required,
// as in RespondSchema, so it is valid in OpenAI strict mode.
func UnitSchema() string {
	strs := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	str := map[string]any{"type": "string"}
	s := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{"run_id", "status", "answer", "claims", "assumptions", "open_questions", "conflicts",
			"recall_ids", "recall_queries", "recall_ops", "units_remaining", "next_step", "notes"},
		"properties": map[string]any{
			"run_id": str,
			"status": map[string]any{"type": "string", "enum": []string{StatusCompleted, StatusNoInformation, StatusPartial, StatusNeedsRecall, StatusBlocked}},
			"answer": str,
			"claims": map[string]any{"type": "array", "items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"claim", "evidence", "quote", "support"},
				"properties": map[string]any{
					"claim":    str,
					"evidence": strs,
					"quote":    str,
					"support":  map[string]any{"type": "string", "enum": []string{SupportQuoted, SupportInferred, SupportUnsupported}},
				},
			}},
			"assumptions":     strs,
			"open_questions":  strs,
			"conflicts":       strs,
			"recall_ids":      strs,
			"recall_queries":  strs,
			"recall_ops":      recallOpsSchema(),
			"units_remaining": strs,
			"next_step":       str,
			"notes":           str,
		},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// AssembleSchema is the JSON schema of an assembler's handoff: the reply,
// and the accounting of every unit claim (used, or omitted with a reason).
func AssembleSchema() string {
	strs := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	str := map[string]any{"type": "string"}
	s := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{"run_id", "status", "answer", "claims_used", "claims_omitted",
			"recall_ids", "recall_queries", "recall_ops", "units_remaining", "next_step", "notes"},
		"properties": map[string]any{
			"run_id":      str,
			"status":      map[string]any{"type": "string", "enum": []string{StatusCompleted, StatusPartial, StatusNeedsRecall}},
			"answer":      str,
			"claims_used": strs,
			"claims_omitted": map[string]any{"type": "array", "items": map[string]any{
				"type": "object", "additionalProperties": false,
				"required":   []string{"id", "reason"},
				"properties": map[string]any{"id": str, "reason": str},
			}},
			"recall_ids":      strs,
			"recall_queries":  strs,
			"recall_ops":      recallOpsSchema(),
			"units_remaining": strs,
			"next_step":       str,
			"notes":           str,
		},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// IndexCard is one card proposed by the indexer.
type IndexCard struct {
	Turn     string   `json:"turn"`
	Summary  string   `json:"summary"`
	Threads  []string `json:"threads"`
	Mentions []string `json:"mentions"`
	Keywords []string `json:"keywords"`
}

// NewThread is a thread the indexer proposes; Key starts with "new:".
type NewThread struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Goal  string `json:"goal"`
}

// NewEntity is an entity the indexer proposes; Key starts with "new:".
type NewEntity struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Aliases     []string `json:"aliases"`
	Description string   `json:"description"`
}

// FactOp proposes a fact anchored by a verbatim quote. Supersedes names
// facts in force that it replaces (a value that changed).
type FactOp struct {
	Statement  string   `json:"statement"`
	Source     string   `json:"source"`
	Quote      string   `json:"quote"`
	Entities   []string `json:"entities"`
	Supersedes []string `json:"supersedes"`
}

// DecisionOp proposes a decision anchored by a verbatim quote.
type DecisionOp struct {
	Statement  string   `json:"statement"`
	Source     string   `json:"source"`
	Quote      string   `json:"quote"`
	Threads    []string `json:"threads"`
	Supersedes []string `json:"supersedes"`
}

// ThreadStatus changes a thread's status.
type ThreadStatus struct {
	Thread string `json:"thread"`
	Status string `json:"status"`
}

// IdeaOp proposes an idea note: one thing said in the conversation, in
// plain words, with the quotes that state it. The engine keeps a note only
// through the quotes it finds in the turns they cite.
type IdeaOp struct {
	Title    string         `json:"title"`
	Kind     string         `json:"kind"`
	Topics   []string       `json:"topics"`
	Evidence []IdeaEvidence `json:"evidence"`
}

// IdeaEvidence is where an idea is stated: a turn and an exact quote.
type IdeaEvidence struct {
	Turn  string `json:"turn"`
	Quote string `json:"quote"`
}

// IdeaKinds are the kinds of idea note the indexer is asked to tell apart.
var IdeaKinds = []string{"decision", "requirement", "fact", "result", "problem", "plan", "change", "question", "attempt"}

// IndexHandoff is what an index run returns: ledger operations the engine
// validates and applies.
type IndexHandoff struct {
	RunID        string         `json:"run_id"`
	Status       string         `json:"status"`
	Cards        []IndexCard    `json:"cards"`
	NewThreads   []NewThread    `json:"new_threads"`
	NewEntities  []NewEntity    `json:"new_entities"`
	Facts        []FactOp       `json:"facts"`
	Decisions    []DecisionOp   `json:"decisions"`
	ThreadStatus []ThreadStatus `json:"thread_status"`
	// Ideas are idea notes for the map of the conversation; empty unless
	// the indexer was asked for them (config atlas: on).
	Ideas []IdeaOp `json:"ideas"`
	Notes string   `json:"notes"`
}

// IndexSchema is the JSON schema of IndexHandoff.
func IndexSchema() string {
	strs := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	str := map[string]any{"type": "string"}
	obj := func(props map[string]any) map[string]any {
		req := make([]string, 0, len(props))
		for k := range props {
			req = append(req, k)
		}
		sortStrings(req)
		return map[string]any{"type": "object", "additionalProperties": false, "required": req, "properties": props}
	}
	arr := func(item map[string]any) map[string]any { return map[string]any{"type": "array", "items": item} }
	s := obj(map[string]any{
		"run_id": str,
		"status": map[string]any{"type": "string", "enum": []string{StatusCompleted, StatusBlocked}},
		"cards": arr(obj(map[string]any{
			"turn": str, "summary": str, "threads": strs, "mentions": strs, "keywords": strs,
		})),
		"new_threads":   arr(obj(map[string]any{"key": str, "title": str, "goal": str})),
		"new_entities":  arr(obj(map[string]any{"key": str, "name": str, "kind": str, "aliases": strs, "description": str})),
		"facts":         arr(obj(map[string]any{"statement": str, "source": str, "quote": str, "entities": strs, "supersedes": strs})),
		"decisions":     arr(obj(map[string]any{"statement": str, "source": str, "quote": str, "threads": strs, "supersedes": strs})),
		"thread_status": arr(obj(map[string]any{"thread": str, "status": map[string]any{"type": "string", "enum": []string{"open", "resolved"}}})),
		"ideas": arr(obj(map[string]any{
			"title": str, "kind": str, "topics": strs,
			"evidence": arr(obj(map[string]any{"turn": str, "quote": str})),
		})),
		"notes": str,
	})
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// ReviewProblem is one problem a reviewer found. Quote is copied from the
// answer verbatim, or empty for something missing.
type ReviewProblem struct {
	Quote   string `json:"quote"`
	Problem string `json:"problem"`
	Fix     string `json:"fix"`
}

// ReviewHandoff is what a review run returns. Supports bounded read-only recall
// (status "needs_recall") so the reviewer can inspect cited turns before verdict.
type ReviewHandoff struct {
	RunID         string          `json:"run_id"`
	Status        string          `json:"status"`
	Verdict       string          `json:"verdict"` // approve | revise | "" (on needs_recall)
	Problems      []ReviewProblem `json:"problems"`
	RecallIDs     []string        `json:"recall_ids,omitempty"`
	RecallQueries []string        `json:"recall_queries,omitempty"`
	RecallOps     []RecallOp      `json:"recall_ops,omitempty"`
	Notes         string          `json:"notes"`
}

// ReviewSchema is the JSON schema of ReviewHandoff.
func ReviewSchema() string {
	str := map[string]any{"type": "string"}
	strs := map[string]any{"type": "array", "items": str}
	s := map[string]any{
		"type": "object", "additionalProperties": false,
		"required": []string{"run_id", "status", "verdict", "problems", "recall_ids", "recall_queries", "recall_ops", "notes"},
		"properties": map[string]any{
			"run_id":  str,
			"status":  map[string]any{"type": "string", "enum": []string{StatusCompleted, StatusNeedsRecall}},
			"verdict": map[string]any{"type": "string", "enum": []string{"approve", "revise", ""}},
			"problems": map[string]any{"type": "array", "items": map[string]any{
				"type": "object", "additionalProperties": false,
				"required":   []string{"quote", "problem", "fix"},
				"properties": map[string]any{"quote": str, "problem": str, "fix": str},
			}},
			"recall_ids":     strs,
			"recall_queries": strs,
			"recall_ops":     recallOpsSchema(),
			"notes":          str,
		},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// ParseReviewHandoff parses a review handoff.
func ParseReviewHandoff(data []byte) (ReviewHandoff, error) {
	raw, err := ExtractJSON(data)
	if err != nil {
		return ReviewHandoff{}, err
	}
	var h ReviewHandoff
	if err := json.Unmarshal(raw, &h); err != nil {
		return h, fmt.Errorf("packet: review handoff: %w", err)
	}
	if h.Status == StatusNeedsRecall {
		if len(h.RecallIDs) == 0 && len(h.RecallQueries) == 0 && len(h.RecallOps) == 0 {
			return h, errors.New("packet: review needs_recall with no recall_ids, recall_queries or recall_ops")
		}
		return h, nil
	}
	if h.Verdict != "approve" && h.Verdict != "revise" {
		return h, fmt.Errorf("packet: review verdict %q", h.Verdict)
	}
	return h, nil
}

// ErrNoHandoff means no handoff JSON could be found.
var ErrNoHandoff = errors.New("packet: no handoff found")

// ExtractJSON finds the handoff in a file or a model's final text,
// leniently: code fences, a BOM, prose around the object. It returns the
// last top-level JSON object that has a "status" field other than
// "started".
func ExtractJSON(data []byte) ([]byte, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	var found []byte
	for i := 0; i < len(data); i++ {
		if data[i] != '{' {
			continue
		}
		dec := json.NewDecoder(bytes.NewReader(data[i:]))
		var m map[string]json.RawMessage
		if err := dec.Decode(&m); err != nil {
			continue
		}
		end := i + int(dec.InputOffset())
		if st, ok := m["status"]; ok {
			var s string
			if json.Unmarshal(st, &s) == nil && s != StatusStarted && s != "" {
				found = data[i:end]
			}
		}
		i = end - 1
	}
	if found == nil {
		return nil, ErrNoHandoff
	}
	return found, nil
}

// ParseHandoff parses a respond-family handoff and checks its shape.
func ParseHandoff(data []byte) (Handoff, error) {
	raw, err := ExtractJSON(data)
	if err != nil {
		return Handoff{}, err
	}
	var h Handoff
	if err := json.Unmarshal(raw, &h); err != nil {
		return h, fmt.Errorf("packet: handoff: %w", err)
	}
	switch h.Status {
	case StatusCompleted, StatusPartial, StatusBlocked:
		if strings.TrimSpace(h.Answer) == "" && h.Status != StatusPartial {
			return h, fmt.Errorf("packet: handoff status %s with an empty answer", h.Status)
		}
	case StatusNeedsRecall:
		if len(h.RecallIDs) == 0 && len(h.RecallQueries) == 0 && len(h.RecallOps) == 0 {
			return h, errors.New("packet: needs_recall with no recall_ids, recall_queries or recall_ops")
		}
	case StatusPlan:
		if len(h.Units) == 0 {
			return h, errors.New("packet: plan with no units")
		}
	default:
		return h, fmt.Errorf("packet: unknown handoff status %q", h.Status)
	}
	return h, nil
}

// ParseUnitHandoff parses a unit's record and checks its shape. It is
// lenient where a provider without a native schema may slip (a missing
// list is an empty one, a claim with no text is dropped); the engine
// checks the claims themselves against the vault.
func ParseUnitHandoff(data []byte) (Handoff, error) {
	raw, err := ExtractJSON(data)
	if err != nil {
		return Handoff{}, err
	}
	var h Handoff
	if err := json.Unmarshal(raw, &h); err != nil {
		return h, fmt.Errorf("packet: unit handoff: %w", err)
	}
	var claims []UnitClaim
	for _, c := range h.Claims {
		if strings.TrimSpace(c.Claim) != "" {
			claims = append(claims, c)
		}
	}
	h.Claims = claims
	switch h.Status {
	case StatusCompleted:
		if strings.TrimSpace(h.Answer) == "" && len(h.Claims) == 0 {
			return h, errors.New("packet: unit handoff completed with no answer and no claims")
		}
	case StatusBlocked:
		if strings.TrimSpace(h.Answer) == "" {
			return h, errors.New("packet: unit handoff blocked with an empty answer")
		}
	case StatusPartial, StatusNoInformation:
	case StatusNeedsRecall:
		if len(h.RecallIDs) == 0 && len(h.RecallQueries) == 0 && len(h.RecallOps) == 0 {
			return h, errors.New("packet: needs_recall with no recall_ids, recall_queries or recall_ops")
		}
	default:
		return h, fmt.Errorf("packet: unknown unit handoff status %q", h.Status)
	}
	return h, nil
}

// ParseIndexHandoff parses an index handoff.
func ParseIndexHandoff(data []byte) (IndexHandoff, error) {
	raw, err := ExtractJSON(data)
	if err != nil {
		return IndexHandoff{}, err
	}
	var h IndexHandoff
	if err := json.Unmarshal(raw, &h); err != nil {
		return h, fmt.Errorf("packet: index handoff: %w", err)
	}
	return h, nil
}
