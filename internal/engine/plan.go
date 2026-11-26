package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

// planCheckpoint is persisted at artifacts/<plan-run>/plan.json so an
// interrupted or partially-failed plan can resume with `cx resume`, reusing
// completed unit records verbatim without regenerating non-deterministic LLM
// outputs.
type planCheckpoint struct {
	PlanID          string        `json:"plan_id"`
	FocusID         string        `json:"focus_id"`
	PlanAuthor      string        `json:"plan_author"`
	Units           []packet.Unit `json:"units"`
	SharedDecisions []string      `json:"shared_decisions,omitempty"`
	Completed       []string      `json:"completed,omitempty"`
}

// planCheckpointMu serializes checkpoint writes: units of a plan finish on
// different goroutines, and two plain writes of the same file at once left
// plan.json cut short or interleaved, which a resume could not read.
var planCheckpointMu sync.Mutex

func savePlanCheckpoint(dir string, cp planCheckpoint) {
	b, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return
	}
	planCheckpointMu.Lock()
	defer planCheckpointMu.Unlock()
	tmp := filepath.Join(dir, "plan.json.tmp")
	if os.WriteFile(tmp, append(b, '\n'), 0o644) == nil {
		_ = os.Rename(tmp, filepath.Join(dir, "plan.json"))
	}
}

func loadPlanCheckpoint(dir string) (planCheckpoint, error) {
	b, err := os.ReadFile(filepath.Join(dir, "plan.json"))
	if err != nil {
		return planCheckpoint{}, err
	}
	var cp planCheckpoint
	if err := json.Unmarshal(b, &cp); err != nil {
		return planCheckpoint{}, err
	}
	return cp, nil
}

// A plan is a task list the engine runs in waves: every unit whose
// dependencies are done runs in the next wave, at most maxParallelUnits
// at a time, and a dependent unit gets the records of the units it waits
// for, verbatim and checked, in its packet. The plan's shared decisions
// (names, interfaces, formats, scope) go into every unit's packet, since
// the original request alone does not stop parallel units from deciding
// the same thing differently. The engine refuses, by rule, plans that buy
// nothing: one unit, or units that would only run one after another
// (sequential work, where every measured multi-agent topology lost to a
// single agent). Those are answered in one call. (LLMCompiler's
// dependency graph; LCM's delegation guard; Cognition on shared context.)

// maxParallelUnits caps how many units of one wave run at once: the
// convergent practice of three to five, with coordination turns growing
// faster than the team.
const maxParallelUnits = 5

// unitSpec is one validated plan unit.
type unitSpec = packet.Unit

// planUnits validates a plan's units: ids and roles; dependencies only on
// units listed earlier, so the plan is acyclic; coder units changing the
// same repository chained in plan order, so they do not branch from the
// same base and conflict. It says why the plan should be answered in one
// call instead ("" to run it), and notes what it changed.
func planUnits(in []packet.Unit, max int) (units []packet.Unit, refuse string, notes []string) {
	valid := map[string]bool{}
	for _, r := range packet.UnitRoles {
		valid[r] = true
	}
	if max <= 0 {
		max = 8
	}
	seen := map[string]bool{}
	for i, u := range in {
		if len(units) == max {
			notes = append(notes, fmt.Sprintf("the plan has %d units; only the first %d run", len(in), max))
			break
		}
		if !valid[u.Role] {
			u.Role = config.RoleResponder
		}
		u.Repo = strings.TrimSpace(u.Repo)
		if u.Repo != "" {
			u.Role = config.RoleCoder
		}
		if u.ID == "" || seen[u.ID] {
			u.ID = fmt.Sprintf("u%d", i+1)
		}
		if u.Title == "" {
			u.Title = u.ID
		}
		var deps []string
		have := map[string]bool{}
		for _, d := range u.DependsOn {
			d = strings.TrimSpace(d)
			switch {
			case d == "" || have[d]:
			case !seen[d]:
				notes = append(notes, fmt.Sprintf("%s depends on %s, which is not an earlier unit; that dependency is dropped", u.ID, d))
			default:
				have[d] = true
				deps = append(deps, d)
			}
		}
		u.DependsOn = deps
		seen[u.ID] = true
		units = append(units, u)
	}
	// coder units of one repository run one after another, each on the
	// commit of the one before
	lastOn := map[string]int{}
	for i := range units {
		u := &units[i]
		if u.Repo == "" {
			continue
		}
		if j, ok := lastOn[u.Repo]; ok && !dependsOn(units, i, units[j].ID) {
			u.DependsOn = append(u.DependsOn, units[j].ID)
			notes = append(notes, fmt.Sprintf("%s and %s both change %s, so %s runs after %s, on its commit", units[j].ID, u.ID, u.Repo, u.ID, units[j].ID))
		}
		lastOn[u.Repo] = i
	}
	for _, u := range units {
		if u.Repo != "" {
			return units, "", notes // a repository change runs as a unit, whatever its shape
		}
	}
	if len(units) == 1 {
		return units, "a plan of one unit is one call", notes
	}
	chain := len(units[0].DependsOn) == 0
	for i := 1; i < len(units) && chain; i++ {
		chain = len(units[i].DependsOn) == 1 && units[i].DependsOn[0] == units[i-1].ID
	}
	if chain {
		return units, "every unit waits for the one before it, so nothing would run in parallel", notes
	}
	if topo := EvaluatePlanTopology(units); !topo.AllowFanOut {
		return units, topo.RefuseReason, notes
	}
	return units, "", notes
}

// dependsOn reports whether unit i waits for id, directly or through
// other units.
func dependsOn(units []packet.Unit, i int, id string) bool {
	byID := map[string]int{}
	for k, u := range units {
		byID[u.ID] = k
	}
	stack := append([]string{}, units[i].DependsOn...)
	seen := map[string]bool{}
	for len(stack) > 0 {
		d := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if d == id {
			return true
		}
		if k, ok := byID[d]; ok && !seen[d] {
			seen[d] = true
			stack = append(stack, units[k].DependsOn...)
		}
	}
	return false
}

// planWaves orders units into waves: a unit runs in the wave after the
// last of its dependencies. Dependencies point only to earlier units.
func planWaves(units []packet.Unit) [][]int {
	level := map[string]int{}
	var waves [][]int
	for i, u := range units {
		l := 0
		for _, d := range u.DependsOn {
			l = max(l, level[d]+1)
		}
		level[u.ID] = l
		for len(waves) <= l {
			waves = append(waves, nil)
		}
		waves[l] = append(waves[l], i)
	}
	return waves
}

// sharedDecisionsItem is the plan's shared decisions as a supplied record.
func sharedDecisionsItem(ds []string) *packet.Item {
	ds = nonBlank(ds)
	if len(ds) == 0 {
		return nil
	}
	return &packet.Item{Key: packet.KeyExtra, ID: "shared-decisions", Seq: 0,
		Header: "Decisions every unit of this plan follows (from the plan, not from the user)",
		Text:   "- " + strings.Join(ds, "\n- ")}
}

// fanOut runs a plan's units in waves, each on its role's provider, and
// assembles their records; a plan the engine refuses is answered in one
// call.
func (e *Engine) fanOut(ctx context.Context, focus vault.Turn, plan outcome) (outcome, error) {
	units, refuse, notes := planUnits(plan.H.Units, e.Cfg.MaxUnits)
	for _, n := range notes {
		e.logf("plan: %s", n)
	}
	if refuse != "" {
		return e.singleCall(ctx, focus, plan, refuse)
	}
	waves := planWaves(units)
	var names []string
	for _, u := range units {
		n := u.ID + "=" + u.Role
		if len(u.DependsOn) > 0 {
			n += " after " + strings.Join(u.DependsOn, "+")
		}
		names = append(names, n)
	}
	e.logf("plan from %s: %d units in %d wave(s) (%s)", plan.Author, len(units), len(waves), strings.Join(names, ", "))

	o := outcome{Runs: plan.Runs}
	dir := e.V.ArtifactDir(plan.RunID)
	cp := planCheckpoint{
		PlanID: plan.RunID, FocusID: focus.ID, PlanAuthor: plan.Author,
		Units: units, SharedDecisions: plan.H.SharedDecisions,
	}
	savePlanCheckpoint(dir, cp)
	return e.executePlanWaves(ctx, focus, plan.RunID, dir, cp, nil, o)
}

// ResumePlan resumes an interrupted or partially-failed plan from
// artifacts/<planID>/plan.json, reusing every already-completed unit record
// verbatim and running only incomplete or failed units before assembly.
func (e *Engine) ResumePlan(ctx context.Context, planID string) (outcome, error) {
	dir := filepath.Join(e.V.Dir, "artifacts", planID)
	cp, err := loadPlanCheckpoint(dir)
	if err != nil {
		return outcome{}, fmt.Errorf("resume plan %s: %w", planID, err)
	}
	focus, err := e.V.Turn(cp.FocusID)
	if err != nil {
		return outcome{}, fmt.Errorf("resume plan %s focus %s: %w", planID, cp.FocusID, err)
	}
	existing := map[string]unitRecord{}
	for _, u := range cp.Units {
		b, rerr := os.ReadFile(filepath.Join(dir, u.ID+".json"))
		if rerr != nil {
			continue
		}
		var rec unitRecord
		if json.Unmarshal(b, &rec) == nil && rec.Error == "" && rec.Status != "" && rec.Status != statusFailed {
			existing[u.ID] = rec
		}
	}
	e.logf("resuming plan %s: reusing %d of %d completed unit record(s)", planID, len(existing), len(cp.Units))
	o, err := e.executePlanWaves(ctx, focus, planID, dir, cp, existing, outcome{})
	if err != nil {
		return o, err
	}
	o.Focus = focus
	if e.shouldReview(o) {
		o = e.review(ctx, o)
	}
	return o, nil
}

func (e *Engine) executePlanWaves(ctx context.Context, focus vault.Turn, planID, dir string, cp planCheckpoint, existing map[string]unitRecord, o outcome) (outcome, error) {
	units := cp.Units
	shared := sharedDecisionsItem(cp.SharedDecisions)
	recs := make([]unitRecord, len(units))
	items := make([]packet.Item, len(units))
	index := map[string]int{}
	for i, u := range units {
		index[u.ID] = i
	}
	deps := make([][]int, len(units))
	for i, u := range units {
		for _, d := range u.DependsOn {
			if di, ok := index[d]; ok {
				deps[i] = append(deps[i], di)
			}
		}
	}

	// Barrier-free reactive DAG execution: each unit starts immediately once
	// its direct dependencies complete, without stalling behind slow unrelated
	// siblings at artificial superstep barriers.
	var stateMu sync.Mutex
	runDAG(ctx, len(units), deps, maxParallelUnits, func(i int) {
		u := units[i]
		if prev, ok := existing[u.ID]; ok {
			it := e.storeUnitRecord(dir, i, prev)
			stateMu.Lock()
			recs[i] = prev
			items[i] = it
			stateMu.Unlock()
			return
		}
		var extra []packet.Item
		if shared != nil {
			extra = append(extra, *shared)
		}
		stateMu.Lock()
		for _, d := range u.DependsOn {
			extra = append(extra, items[index[d]])
		}
		stateMu.Unlock()

		spec := callSpec{
			Kind: packet.KindUnit, Role: u.Role, Focus: focus, Parent: planID, UnitRecord: true, Extra: extra,
			Pinned:       e.newRecallIDs(u.Needs, focus, nil),
			Instructions: func(id string) string { return unitInstructions(id, u, len(units), shared != nil) },
		}
		var setupErr error
		if u.Repo != "" {
			sess, _ := e.V.Session()
			repo, ok := sess.FindRepo(u.Repo)
			if !ok {
				setupErr = fmt.Errorf("unit %s names repository %q, which is not attached", u.ID, u.Repo)
			} else {
				spec.Repo, spec.Title, spec.chain = &repo, u.Title, &repoChain{}
				stateMu.Lock()
				for _, d := range u.DependsOn {
					if up := recs[index[d]]; up.Error == "" {
						for _, ch := range up.Changes {
							if ch.Repo == u.Repo && ch.Commit != "" {
								spec.chain.base = ch.Commit
							}
						}
					}
				}
				stateMu.Unlock()
			}
		}

		var r outcome
		var runErr error
		if setupErr != nil {
			runErr = setupErr
		} else {
			r, runErr = e.runToCompletion(ctx, spec)
		}
		rec := e.unitRecordOf(planID, u.ID, u.Title, u.Role, focus, r, runErr)
		rec.DependsOn = u.DependsOn
		it := e.storeUnitRecord(dir, i, rec)

		stateMu.Lock()
		o.Runs = append(o.Runs, r.Runs...)
		recs[i] = rec
		items[i] = it
		if rec.Error == "" {
			cp.Completed = append(cp.Completed, u.ID)
		}
		cpSnap := cp
		cpSnap.Completed = append([]string(nil), cp.Completed...)
		stateMu.Unlock()
		savePlanCheckpoint(dir, cpSnap)
	})

	// Optional disagreement reconciliation round: when two units cite the same
	// stored text and state conflicting values, rerun each disagreeing unit
	// once with the other's verified quote pinned before assembly.
	if e.Cfg.DisagreementRerun {
		red := reduceRecords(recs)
		reconcileUnits := map[string][]string{}
		for _, g := range red.Groups {
			if len(g.Differ) == 0 {
				continue
			}
			for _, cid := range g.Claims {
				if uid, _, ok := strings.Cut(cid, "."); ok {
					reconcileUnits[uid] = append(reconcileUnits[uid], fmt.Sprintf("%s (%s): %s", g.Turn, strings.Join(g.At, ", "), strings.Join(g.Differ, "; ")))
				}
			}
		}
		for uid, conflicts := range reconcileUnits {
			idx, ok := index[uid]
			if !ok {
				continue
			}
			u := units[idx]
			e.logf("plan: rerunning unit %s to reconcile conflicting value claims", uid)
			note := packet.Item{
				Key: packet.KeyExtra, ID: "reconcile-" + uid, Seq: 85,
				Header: "Conflicting claims from sibling units citing the same text (reconcile against the verbatim source)",
				Text:   "- " + strings.Join(conflicts, "\n- "),
			}
			spec := callSpec{
				Kind: packet.KindUnit, Role: u.Role, Focus: focus, Parent: planID, UnitRecord: true,
				Extra:  []packet.Item{items[idx], note},
				Pinned: e.newRecallIDs(u.Needs, focus, nil),
				Instructions: func(id string) string {
					return unitInstructions(id, u, len(units), shared != nil) + "\n\nReconcile the disagreement noted under Supplied records against the exact verbatim text."
				},
			}
			if r2, err2 := e.runToCompletion(ctx, spec); err2 == nil {
				o.Runs = append(o.Runs, r2.Runs...)
				recs[idx] = e.unitRecordOf(planID, u.ID, u.Title, u.Role, focus, r2, nil)
				recs[idx].DependsOn = u.DependsOn
				items[idx] = e.storeUnitRecord(dir, idx, recs[idx])
			}
		}
	}

	// Cross-branch scratch integration check: when multiple coder units kept
	// branches on the same attached repository, merge them onto a throwaway
	// worktree and run verify once to catch semantic write skew.
	var integrationExtras []packet.Item
	branchesByRepo := map[string][]string{}
	baseByRepo := map[string]string{}
	for _, r := range recs {
		for _, ch := range r.Changes {
			if ch.Branch != "" {
				branchesByRepo[ch.Repo] = append(branchesByRepo[ch.Repo], ch.Branch)
				if baseByRepo[ch.Repo] == "" {
					baseByRepo[ch.Repo] = ch.Base
				}
			}
		}
	}
	if len(branchesByRepo) > 0 {
		sess, _ := e.V.Session()
		for rname, brs := range branchesByRepo {
			if len(brs) < 2 {
				continue
			}
			if repo, ok := sess.FindRepo(rname); ok {
				ic := ScratchMergeAndVerify(repo, baseByRepo[rname], brs)
				status := "PASSED"
				if ic.Error != "" {
					status = "FAILED: " + ic.Error
				}
				txt := fmt.Sprintf("Scratch integration of %d branches (%s) on %s: %s", len(brs), strings.Join(brs, ", "), rname, status)
				if ic.Output != "" {
					txt += "\n```\n" + ic.Output + "\n```"
				}
				e.logf("integration check (%s): %s", rname, status)
				integrationExtras = append(integrationExtras, packet.Item{
					Key: packet.KeyExtra, ID: "integration-" + rname, Seq: 89,
					Header: "Cross-branch scratch integration check for " + rname,
					Text:   txt,
				})
			}
		}
	}

	// Structured claim-key conflict & Verified-Quote Dominance audit: prevent
	// unverified/unsupported claims from overriding byte-verified quoted claims.
	if keyConflicts := AuditClaimKeyConflicts(recs); len(keyConflicts) > 0 {
		var lines []string
		for _, kc := range keyConflicts {
			lines = append(lines, "- "+kc.Summary)
		}
		integrationExtras = append(integrationExtras, packet.Item{
			Key: packet.KeyExtra, ID: "claim-key-conflicts", Seq: 88,
			Header: "Structured claim-key conflict audit (verified-quote dominance rules)",
			Text:   strings.Join(lines, "\n"),
		})
	}

	e.logChecks(recs)

	base := callSpec{Kind: packet.KindAssemble, Role: config.RoleAssembler, Focus: focus, Parent: planID,
		Extra:        integrationExtras,
		Instructions: func(id string) string { return assembleInstructions(id, len(units)) }}
	a, err := e.assemble(ctx, base, dir, planID, recs, items, shared)
	o.Runs = append(o.Runs, a.Runs...)
	if err != nil {
		e.logf("assembly failed (%v): returning the unit records as they are", err)
		o.Answer, o.Author, o.RunID, o.Status = losslessRecords(items), "units", planID, packet.StatusCompleted
		o.Supplied, o.Planned = a.Supplied, true
		return o, nil
	}
	var authors []string
	for _, r := range recs {
		if r.Error == "" && r.By != "" {
			authors = append(authors, r.By)
		}
	}
	a.Runs = o.Runs
	a.Author = a.Author + " (assembling " + strings.Join(uniq(authors), ", ") + ")"
	a.Planned = true
	return a, nil
}

// losslessRecords is the fallback answer when assembly fails: every unit
// record, in plan order.
func losslessRecords(items []packet.Item) string {
	var b strings.Builder
	for _, x := range items {
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", x.Header, x.Text)
	}
	return strings.TrimSpace(b.String())
}

// singleCall answers a message in one call after the engine refused its
// plan; the call is told why, and offered no plan.
func (e *Engine) singleCall(ctx context.Context, focus vault.Turn, plan outcome, why string) (outcome, error) {
	e.logf("plan from %s not run: %s; answering in one call", plan.Author, why)
	note := packet.Item{Key: packet.KeyExtra, ID: "plan-refused", Seq: 1, Header: "Why this is one call (from the engine)",
		Text: "A plan was proposed for this message and not run: " + why + ". Answer the message in this call; if the reply will not fit, return partial and it will be continued."}
	spec := callSpec{Kind: packet.KindRespond, Role: config.RoleResponder, Focus: focus, Parent: plan.RunID, Extra: []packet.Item{note},
		Instructions: func(id string) string { return respondText(id, false) }}
	o, err := e.runToCompletion(ctx, spec)
	o.Runs = append(append([]RunInfo{}, plan.Runs...), o.Runs...)
	return o, err
}
