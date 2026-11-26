package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
)

// The floor is what selection achieves with no model at all: is the answer
// in the packet, for every baseline a smarter selector has to beat, on the
// same questions? Live experiments are read against it.

// Context arms: what a packet's history is made of.
const (
	ArmCurrent = "current" // the configured selection, all lanes
	ArmFull    = "full"    // every earlier turn
	ArmBM25    = "bm25"    // ranked paragraphs from full-text search alone
	ArmRecent  = "recent"  // the recent turns and the catalog, nothing ranked
	ArmTurns   = "turns"   // six ranked whole turns
)

// ContextArms lists the arms.
var ContextArms = []string{ArmCurrent, ArmFull, ArmBM25, ArmRecent, ArmTurns}

// IsContextArm reports whether arm names one.
func IsContextArm(arm string) bool {
	for _, a := range ContextArms {
		if a == arm {
			return true
		}
	}
	return false
}

// ApplyContextArm sets cfg, and the process's ranking lanes, for an arm.
// Only one arm runs at a time in a process. Every arm but current and full
// turns off the whole-history rule, so that it measures selection.
func ApplyContextArm(cfg *config.Config, arm string) error {
	packet.RankLanes = packet.Lanes{Exact: true, FullText: true, Value: true}
	cfg.FullHistory, cfg.NoRanked = false, false
	switch arm {
	case "", ArmCurrent:
	case ArmFull:
		cfg.FullHistory = true
	case ArmBM25:
		packet.RankLanes = packet.Lanes{FullText: true}
		cfg.WholeHistoryTokens = -1
		if cfg.RelevantTokens <= 0 {
			cfg.RelevantTokens = config.Default().RelevantTokens
		}
	case ArmRecent:
		cfg.NoRanked = true
		cfg.WholeHistoryTokens = -1
	case ArmTurns:
		cfg.RelevantTokens = -1
		cfg.WholeHistoryTokens = -1
	default:
		return fmt.Errorf("unknown context arm %q (%s)", arm, strings.Join(ContextArms, ", "))
	}
	return nil
}

// FloorSet is one probe file on one workspace.
type FloorSet struct {
	Workspace string
	Name      string
	Probes    []Probe
}

// FloorCell is one arm at one budget, over every set.
type FloorCell struct {
	Arm            string  `json:"arm"`
	Budget         int     `json:"budget,omitempty"`
	Probes         int     `json:"probes"`
	InPacket       int     `json:"in_packet"`
	AnyTurn        int     `json:"any_turn"`
	Search         int     `json:"search"`
	Wins           int     `json:"wins_vs_reference"`
	Losses         int     `json:"losses_vs_reference"`
	P              float64 `json:"exact_p_vs_reference"`
	MedianBrief    int     `json:"median_brief"`
	MedianRanked   int     `json:"median_ranked"`
	TokensToAnswer int     `json:"median_tokens_to_answer"` // -1: the median question is never reached
	ReplayInput    int     `json:"replay_input_per_message"`
	BuildP50       int64   `json:"build_ms_p50"`
	BuildP95       int64   `json:"build_ms_p95"`
	AlsoBad        int     `json:"also_entries_never_creditable"`
	hit            map[string]bool
}

func (c FloorCell) label() string {
	if c.Budget > 0 {
		return fmt.Sprintf("%s %dk", c.Arm, c.Budget/1000)
	}
	return c.Arm
}

// RunFloor measures every arm on every set: the workspaces are opened
// read-only, but opening may update their index, so use scratch copies.
// replay is how many past user messages to rebuild per workspace.
func RunFloor(cfg config.Config, sets []FloorSet, budgets []int, replay int, logf func(string)) ([]FloorCell, error) {
	defer func() { c := cfg; _ = ApplyContextArm(&c, ArmCurrent) }()
	type spec struct {
		arm    string
		budget int
	}
	specs := []spec{{ArmFull, 0}, {ArmRecent, 0}, {ArmTurns, 0}}
	for _, arm := range []string{ArmBM25, ArmCurrent} {
		for _, b := range budgets {
			specs = append(specs, spec{arm, b})
		}
	}
	var cells []FloorCell
	for _, sp := range specs {
		cell := FloorCell{Arm: sp.arm, Budget: sp.budget, hit: map[string]bool{}}
		var briefs, ranked, reach []int
		var build []int64
		replayed := map[string]bool{}
		sent, calls := 0, 0
		for _, set := range sets {
			c := cfg
			if sp.budget > 0 {
				c.RelevantTokens = sp.budget
			}
			if err := ApplyContextArm(&c, sp.arm); err != nil {
				return nil, err
			}
			if sp.arm == ArmCurrent {
				c.WholeHistoryTokens = -1 // the selection, at this budget
			}
			logf(fmt.Sprintf("%s: %s", cell.label(), set.Name))
			e, err := OpenReadOnly(c, set.Workspace)
			if err != nil {
				return nil, err
			}
			res, err := e.RunProbes(set.Probes)
			if err != nil {
				e.Close()
				return nil, err
			}
			if !replayed[set.Workspace] && replay > 0 {
				replayed[set.Workspace] = true
				rc, err := e.CacheReplay(replay)
				if err != nil {
					e.Close()
					return nil, err
				}
				for _, r := range rc {
					sent += r.Brief
					calls++
				}
			}
			e.Close()
			for _, r := range res {
				if r.Where == WhereInvalid {
					continue
				}
				cell.Probes++
				in := r.Where == WhereVerbatim
				cell.hit[set.Name+"|"+r.ID] = in
				if in {
					cell.InPacket++
				}
				if r.Any {
					cell.AnyTurn++
				}
				if r.Search {
					cell.Search++
				}
				cell.AlsoBad += len(r.AlsoBad)
				briefs, ranked = append(briefs, r.Brief), append(ranked, r.Ranked)
				build = append(build, r.BuildMs)
				t := r.TokensToAnswer
				if t == 0 {
					t = math.MaxInt
				}
				reach = append(reach, t)
			}
		}
		cell.MedianBrief, cell.MedianRanked = median(briefs), median(ranked)
		cell.TokensToAnswer = median(reach)
		if cell.TokensToAnswer == math.MaxInt {
			cell.TokensToAnswer = -1
		}
		if sp.arm == ArmFull || sp.arm == ArmRecent {
			cell.TokensToAnswer = 0 // nothing is ranked
		}
		if calls > 0 {
			cell.ReplayInput = sent / calls
		}
		sort.Slice(build, func(i, j int) bool { return build[i] < build[j] })
		if len(build) > 0 {
			cell.BuildP50, cell.BuildP95 = build[len(build)/2], build[len(build)*95/100]
		}
		cells = append(cells, cell)
	}
	compareToReference(cells, 8000)
	return cells, nil
}

// compareToReference pairs every cell with current selection at the given
// budget, probe by probe: wins, losses and an exact two-sided sign test on
// the probes where they differ (McNemar's exact test).
func compareToReference(cells []FloorCell, budget int) {
	var ref *FloorCell
	for i := range cells {
		if cells[i].Arm == ArmCurrent && cells[i].Budget == budget {
			ref = &cells[i]
		}
	}
	if ref == nil {
		return
	}
	for i := range cells {
		c := &cells[i]
		c.Wins, c.Losses, c.P = 0, 0, 1
		for id, in := range c.hit {
			switch r, ok := ref.hit[id]; {
			case !ok || r == in:
			case in:
				c.Wins++
			default:
				c.Losses++
			}
		}
		c.P = exactSignP(c.Wins, c.Losses)
	}
}

// exactSignP is the two-sided exact binomial test of a against b at one
// half.
func exactSignP(a, b int) float64 {
	n := a + b
	if n == 0 {
		return 1
	}
	k := min(a, b)
	logTerm := -float64(n) * math.Ln2 // C(n,0) / 2^n
	sum := 0.0
	for i := 0; i <= k; i++ {
		sum += math.Exp(logTerm)
		logTerm += math.Log(float64(n-i)) - math.Log(float64(i+1))
	}
	return math.Min(1, 2*sum)
}

// FloorReport renders the cells.
func FloorReport(cells []FloorCell) string {
	var b strings.Builder
	b.WriteString("| Arm | Answer in packet | Any turn stating it | Search | Against current 8k (won, lost, p) | Median brief | Median ranked | Median tokens to answer | Replay input per message | Build p50, p95 ms |\n|---|---|---|---|---|---|---|---|---|---|\n")
	for _, c := range cells {
		reach := fmt.Sprint(c.TokensToAnswer)
		switch c.TokensToAnswer {
		case 0:
			reach = "n/a"
		case -1:
			reach = "not reached"
		}
		fmt.Fprintf(&b, "| %s | %d of %d | %d | %d | %d, %d, %.2f | %d | %d | %s | %d | %d, %d |\n", c.label(), c.InPacket, c.Probes, c.AnyTurn, c.Search,
			c.Wins, c.Losses, c.P, c.MedianBrief, c.MedianRanked, reach, c.ReplayInput, c.BuildP50, c.BuildP95)
	}
	if len(cells) > 0 && cells[0].AlsoBad > 0 {
		fmt.Fprintf(&b, "\n%d also entries name a turn without an accepted answer and cannot count.\n", cells[0].AlsoBad)
	}
	return b.String()
}

// WriteFloor writes floor.md and floor.json into dir.
func WriteFloor(dir string, cells []FloorCell) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	js, _ := json.MarshalIndent(cells, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "floor.json"), append(js, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "floor.md"), []byte(FloorReport(cells)), 0o644)
}
