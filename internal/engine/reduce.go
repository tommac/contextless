package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// The reduce is the code half of combining unit records, run before the
// assembler writes (LCM and LOTUS keep the combine step in code; the join
// is where measured multi-agent pipelines fail most). It is deterministic
// and order-independent: records of units that found nothing or failed
// leave the reduce; claims of different units whose quotes sit in the
// same stored paragraphs are listed together, marked when they may state
// different values; assumptions of different units that may be about the
// same choice are paired; open questions and reported conflicts are
// collected; and every surviving claim id is one the assembler must
// account for. The reduce never decides who is right: it shows the
// assembler, and later the reviewer, where to look.

// reduction is the engine's comparison of a plan's unit records.
type reduction struct {
	Groups    []evidenceGroup
	Pairs     [][2]unitNote
	Open      []unitNote
	Conflicts []unitNote
	Claims    []string // every claim id the assembler accounts for
	Left      []string // units that gave no result, with why
	Units     int
}

// evidenceGroup is claims of at least two units whose quotes lie in
// overlapping paragraphs of one turn.
type evidenceGroup struct {
	Turn   string
	At     []string // where each quote is, as located
	Claims []string
	// Differ lists the claims that may state different values, with the
	// values each states.
	Differ []string
}

type unitNote struct{ Unit, Text string }

// maxAssumptionPairs bounds the pairs listed; more say little more.
const maxAssumptionPairs = 20

func reduceRecords(recs []unitRecord) reduction {
	red := reduction{Units: len(recs)}
	type pinned struct {
		unit  string
		claim checkedClaim
		a, b  int // atoms, inclusive; a whole turn is 1 to wholeTurn
	}
	const wholeTurn = 1 << 30
	byTurn := map[string][]pinned{}
	var live []unitRecord
	for _, r := range recs {
		switch {
		case r.Error != "":
			red.Left = append(red.Left, r.Unit+" (failed)")
			continue
		case r.Status == packet.StatusNoInformation:
			red.Left = append(red.Left, r.Unit+" (no information)")
			continue
		}
		live = append(live, r)
		for _, c := range r.Claims {
			red.Claims = append(red.Claims, c.ID)
			if c.Check != checkVerified {
				continue
			}
			ref, ok := vault.ParseRef(c.Located)
			if !ok || ref.Part > 0 {
				continue // a quote in a file the unit wrote is its own
			}
			p := pinned{unit: r.Unit, claim: c, a: 1, b: wholeTurn}
			if ref.Atom > 0 {
				p.a, p.b = ref.Atom, ref.AtomEnd
			}
			byTurn[ref.Turn] = append(byTurn[ref.Turn], p)
		}
		for _, q := range r.OpenQuestions {
			red.Open = append(red.Open, unitNote{r.Unit, q})
		}
		for _, q := range r.Conflicts {
			red.Conflicts = append(red.Conflicts, unitNote{r.Unit, q})
		}
	}

	turns := make([]string, 0, len(byTurn))
	for t := range byTurn {
		turns = append(turns, t)
	}
	sort.Strings(turns)
	for _, t := range turns {
		ps := byTurn[t]
		sort.SliceStable(ps, func(i, j int) bool { return ps[i].a < ps[j].a })
		flush := func(g []pinned) {
			units := map[string]bool{}
			for _, p := range g {
				units[p.unit] = true
			}
			if len(units) < 2 {
				return
			}
			eg := evidenceGroup{Turn: t}
			seenAt := map[string]bool{}
			for _, p := range g {
				eg.Claims = append(eg.Claims, p.claim.ID)
				if !seenAt[p.claim.Located] {
					seenAt[p.claim.Located] = true
					eg.At = append(eg.At, p.claim.Located)
				}
			}
			differ := map[string]bool{}
			for i := range g {
				for j := i + 1; j < len(g); j++ {
					if g[i].unit != g[j].unit && mayDiffer(g[i].claim.Claim, g[j].claim.Claim) {
						differ[g[i].claim.ID], differ[g[j].claim.ID] = true, true
					}
				}
			}
			for _, p := range g {
				if differ[p.claim.ID] {
					eg.Differ = append(eg.Differ, fmt.Sprintf("%s states %s", p.claim.ID, strings.Join(packet.Values(p.claim.Claim), ", ")))
				}
			}
			red.Groups = append(red.Groups, eg)
		}
		var g []pinned
		end := 0
		for _, p := range ps {
			if len(g) > 0 && p.a > end {
				flush(g)
				g = nil
			}
			if len(g) == 0 || p.b > end {
				end = max(end, p.b)
			}
			g = append(g, p)
		}
		flush(g)
	}

	// assumptions of different units that share the words of one choice
	for i := 0; i < len(live) && len(red.Pairs) < maxAssumptionPairs; i++ {
		for j := i + 1; j < len(live) && len(red.Pairs) < maxAssumptionPairs; j++ {
			for _, x := range live[i].Assumptions {
				for _, y := range live[j].Assumptions {
					if len(red.Pairs) < maxAssumptionPairs && flat(strings.ToLower(x)) != flat(strings.ToLower(y)) && sharedWords(x, y) >= 2 {
						red.Pairs = append(red.Pairs, [2]unitNote{{live[i].Unit, x}, {live[j].Unit, y}})
					}
				}
			}
		}
	}
	return red
}

// mayDiffer reports two claims that are about the same thing (two content
// words in common) and state values neither of which contains the other.
func mayDiffer(x, y string) bool {
	vx, vy := packet.Values(x), packet.Values(y)
	if len(vx) == 0 || len(vy) == 0 || subset(vx, vy) || subset(vy, vx) {
		return false
	}
	return sharedWords(x, y) >= 2
}

func subset(a, b []string) bool {
	in := map[string]bool{}
	for _, v := range b {
		in[v] = true
	}
	for _, v := range a {
		if !in[v] {
			return false
		}
	}
	return true
}

// sharedWords counts the content words two texts share, values aside.
func sharedWords(x, y string) int {
	vals := map[string]bool{}
	for _, v := range append(packet.Values(x), packet.Values(y)...) {
		vals[v] = true
	}
	wy := packet.ContentWords(y)
	n := 0
	for w := range packet.ContentWords(x) {
		if wy[w] && !vals[w] {
			n++
		}
	}
	return n
}

// item is the reduction as the supplied record the assembler and the
// reviewer read, before the unit records.
func (red reduction) item() packet.Item {
	var b strings.Builder
	fmt.Fprintf(&b, "%d unit(s); %d claim(s) to account for.", red.Units, len(red.Claims))
	if len(red.Left) > 0 {
		fmt.Fprintf(&b, " No result from: %s.", strings.Join(red.Left, ", "))
	}
	b.WriteString("\n\nClaims of different units that cite the same stored text:\n")
	if len(red.Groups) == 0 {
		b.WriteString("- none\n")
	}
	for _, g := range red.Groups {
		fmt.Fprintf(&b, "- %s: %s", strings.Join(g.At, ", "), strings.Join(g.Claims, ", "))
		if len(g.Differ) > 0 {
			fmt.Fprintf(&b, "; they may state different values (%s): settle it from the cited text", strings.Join(g.Differ, "; "))
		}
		b.WriteString("\n")
	}
	b.WriteString("\nAssumptions of different units that may be about the same choice:\n")
	if len(red.Pairs) == 0 {
		b.WriteString("- none\n")
	}
	for _, p := range red.Pairs {
		fmt.Fprintf(&b, "- %s: %s / %s: %s\n", p[0].Unit, flat(p[0].Text), p[1].Unit, flat(p[1].Text))
	}
	notes := func(title string, ns []unitNote) {
		if len(ns) == 0 {
			return
		}
		fmt.Fprintf(&b, "\n%s:\n", title)
		for _, n := range ns {
			fmt.Fprintf(&b, "- %s: %s\n", n.Unit, flat(n.Text))
		}
	}
	notes("Open questions", red.Open)
	notes("Conflicts the units reported", red.Conflicts)
	return packet.Item{Key: packet.KeyExtra, ID: "unit-comparison", Seq: 0,
		Header: "The engine's comparison of the unit records (computed by code, not by a model)", Text: strings.TrimSpace(b.String())}
}

// accountFor checks an assembler's accounting against the claims it must
// account for: the claims in neither list (an omission needs a reason),
// and the ids that name no claim.
func accountFor(claims []string, h packet.Handoff) (missing, unknown []string) {
	want := map[string]bool{}
	for _, c := range claims {
		want[c] = true
	}
	done := map[string]bool{}
	seen := map[string]bool{}
	note := func(id string) bool {
		id = strings.TrimSpace(id)
		if id == "" {
			return false
		}
		if !want[id] && !seen[id] {
			seen[id] = true
			unknown = append(unknown, id)
		}
		return want[id]
	}
	for _, id := range h.ClaimsUsed {
		if note(id) {
			done[strings.TrimSpace(id)] = true
		}
	}
	for _, o := range h.ClaimsOmitted {
		if note(o.ID) && strings.TrimSpace(o.Reason) != "" {
			done[strings.TrimSpace(o.ID)] = true
		}
	}
	for _, c := range claims {
		if !done[c] {
			missing = append(missing, c)
		}
	}
	return missing, unknown
}
