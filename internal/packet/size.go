package packet

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/tommac/contextless/internal/providers"
)

// ErrTooBig means the never-dropped records (session, decisions, focus,
// recalls, extras) do not fit any allowed profile. The engine then splits
// the work into units; it never shortens a record.
var ErrTooBig = errors.New("packet: the protected records do not fit any allowed profile")

// HandoffTokens is the allowance for the handoff wrapper around the answer.
const HandoffTokens = 1000

// Budget is the cost arithmetic of one packet (BookWriter 04 §4.5, reduced
// to what a one-shot chat call costs: system overhead, brief, output).
type Budget struct {
	Profile         string `json:"profile"`
	Window          int    `json:"window"`
	EffectiveWindow int    `json:"effective_window"`
	SysOverhead     int    `json:"sys_overhead"`
	Brief           int    `json:"brief"`
	Output          int    `json:"output"`
	Planned         int    `json:"planned"`
	Cap             int    `json:"cap"`
	KillLine        int    `json:"kill_line"`
	// Escalated is true when the routed (first) profile could not hold the
	// packet and a larger one was chosen instead of dropping records.
	Escalated bool `json:"escalated"`
}

// Fits reports whether the planned total is within the cap.
func (b Budget) Fits() bool { return b.Planned <= b.Cap }

// IndexStatus captures the snapshot boundary and search-index integrity for a
// compiled packet so degraded SQLite lanes or index lag are never silently
// mistaken for absence in the archive.
type IndexStatus struct {
	SnapshotSeq  int      `json:"snapshot_seq,omitempty"`
	SnapshotRoot string   `json:"snapshot_root,omitempty"`
	IndexedUpTo  int      `json:"indexed_up_to,omitempty"`
	Complete     bool     `json:"complete"`
	LaneErrors   []string `json:"lane_errors,omitempty"`
}

// Sized is a packet ready to write.
type Sized struct {
	Profile providers.Profile
	Items   []Item
	Dropped []Item
	Budget  Budget
	Brief   string
	// WholeHistory is true when every earlier turn is in the packet
	// verbatim, so nothing was selected or left out.
	WholeHistory bool
	// Index reports snapshot coverage and any degraded search lanes.
	Index IndexStatus
	// OmittedHandle holds the compact external .cxi catalog JSON when offloaded.
	OmittedHandle string
}

// Tokens estimates tokens for n bytes on a profile.
func Tokens(p providers.Profile, n int) int {
	tpb := p.TokensPerByte
	if tpb <= 0 {
		tpb = providers.DefaultTokensPerByte
	}
	return int(math.Ceil(float64(n) * tpb))
}

func budgetFor(p providers.Profile, brief string, output int) Budget {
	return Budget{
		Profile: p.ID, Window: p.Window, EffectiveWindow: p.EffectiveWindow, SysOverhead: p.SysOverhead,
		Brief: Tokens(p, len(brief)), Output: output + HandoffTokens,
		Planned: p.SysOverhead + Tokens(p, len(brief)) + output + HandoffTokens,
		Cap:     p.PlanCap(), KillLine: p.KillLine(),
	}
}

// Size fits a packet to the first profile that can hold it.
//
//  1. The profiles are tried in routing order with every record in; the
//     first that fits wins. Moving to a larger window is preferred over
//     leaving anything out, because nothing is ever compressed.
//  2. If none holds everything, the largest-window profile is used and
//     records are left out by tier: tier 1 (catalog cards, oldest first),
//     then tier 2 (the rest of the focus's threads, oldest first), tier 3
//     (turns ranked for the focus, least relevant first, then entities
//     and facts), tier 4 (recent turns, oldest first). Each left-out record is listed by id in the brief so
//     the model can recall it.
//  3. If the tier-0 records alone do not fit, ErrTooBig.
func Size(t Task, items []Item, profiles []providers.Profile, output int) (Sized, error) {
	var allowed []providers.Profile
	for _, p := range profiles {
		allowed = append(allowed, p.WithDefaults())
	}
	if len(allowed) == 0 {
		return Sized{}, fmt.Errorf("packet: no profile available for role %s", t.Role)
	}
	for i := range items {
		items[i].Bytes = len(items[i].Text) + len(items[i].Header) + 8
	}
	for i, p := range allowed {
		t.NativeSchema = p.NativeSchema
		brief := Render(t, items, nil)
		b := budgetFor(p, brief, output)
		if b.Fits() {
			b.Escalated = i > 0
			return Sized{Profile: p, Items: fill(p, items), Budget: b, Brief: brief}, nil
		}
	}

	largest := allowed[0]
	for _, p := range allowed[1:] {
		if p.EffectiveWindow > largest.EffectiveWindow {
			largest = p
		}
	}
	t.NativeSchema = largest.NativeSchema
	order := dropOrder(items)
	dropped := map[string]bool{}
	var kept, gone []Item
	split := func() {
		kept, gone = kept[:0], gone[:0]
		for _, it := range items {
			if dropped[it.ID] {
				gone = append(gone, it)
			} else {
				kept = append(kept, it)
			}
		}
	}
	// estimate first, then verify with a real render
	brief := Render(t, items, nil)
	planned := budgetFor(largest, brief, output).Planned
	cap := largest.PlanCap()
	next := 0
	for {
		for planned > cap && next < len(order) {
			it := order[next]
			next++
			dropped[it.ID] = true
			planned -= Tokens(largest, it.Bytes)
			planned += Tokens(largest, omittedCost(it))
		}
		split()
		brief = Render(t, kept, gone)
		b := budgetFor(largest, brief, output)
		if b.Fits() {
			b.Escalated = largest.ID != allowed[0].ID
			return Sized{Profile: largest, Items: fill(largest, kept), Dropped: fill(largest, gone), Budget: b, Brief: brief}, nil
		}
		if next >= len(order) {
			return Sized{Profile: largest, Items: fill(largest, kept), Dropped: fill(largest, gone), Budget: b, Brief: brief},
				fmt.Errorf("%w: %d planned tokens, cap %d on %s", ErrTooBig, b.Planned, b.Cap, largest.ID)
		}
		planned = b.Planned
	}
}

func fill(p providers.Profile, items []Item) []Item {
	out := make([]Item, len(items))
	for i, it := range items {
		it.Tokens = Tokens(p, it.Bytes)
		out[i] = it
	}
	return out
}

func omittedCost(it Item) int {
	if it.Tier == 1 {
		return 10 // folded into an id range
	}
	return len(it.Card) + 4
}

// dropOrder lists droppable items in the order they are left out.
func dropOrder(items []Item) []Item {
	var out []Item
	for _, it := range items {
		if it.Tier > 0 {
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Tier != b.Tier {
			return a.Tier < b.Tier
		}
		// ledger records (facts, entities: no sequence number) outlive
		// the turns of their tier; ranked turns go least relevant first
		if (a.Seq == 0) != (b.Seq == 0) {
			return a.Seq != 0
		}
		if a.Score != b.Score {
			return a.Score < b.Score
		}
		if a.Seq != b.Seq {
			return a.Seq < b.Seq
		}
		return a.ID < b.ID
	})
	return out
}

// Fit reports whether text of n bytes alone, plus the fixed parts of a
// brief, fits a profile: the engine uses it to choose a chunk size when it
// splits an oversized focus.
func ChunkBytes(p providers.Profile, output int, fixedBriefBytes int) int {
	p = p.WithDefaults()
	room := p.PlanCap() - p.SysOverhead - output - HandoffTokens - Tokens(p, fixedBriefBytes)
	if room <= 0 {
		return 0
	}
	tpb := p.TokensPerByte
	if tpb <= 0 {
		tpb = providers.DefaultTokensPerByte
	}
	// half the room: the chunk shares the packet with the session,
	// decisions and the request
	return int(float64(room) / tpb / 2)
}
