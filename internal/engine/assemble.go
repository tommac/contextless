package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/tommac/contextless/internal/packet"
)

// assemblyRecord is the engine's check of an assembler's reply, stored as
// artifacts/<plan>/assembly.json: how the plan's records were joined, for
// measuring how often a merge drops a result or hides a disagreement.
type assemblyRecord struct {
	Plan    string                `json:"plan"`
	Runs    []string              `json:"runs,omitempty"`
	Claims  int                   `json:"claims"`
	Used    []string              `json:"used,omitempty"`
	Omitted []packet.OmittedClaim `json:"omitted,omitempty"`
	// Missing are claims in neither list after the last attempt (an
	// omission needs a reason); Unknown are ids that name no claim.
	Missing []string `json:"missing,omitempty"`
	Unknown []string `json:"unknown,omitempty"`
	// Retried is set when the first reply left claims unaccounted for and
	// the assembler was run once more with the engine's check.
	Retried         bool `json:"retried"`
	SharedEvidence  int  `json:"shared_evidence"`
	ValuesDiffer    int  `json:"values_differ"`
	AssumptionPairs int  `json:"assumption_pairs"`
	Left            int  `json:"units_without_result"`
}

// assemble runs the assembler over a plan's records: the shared decisions,
// the engine's comparison of the records, then the records, verbatim. The
// reply must account for every claim, used or omitted with a reason; a
// reply that does not is run once more with the engine's check, and what
// remains unaccounted is recorded, not hidden. The supplied records of the
// result (comparison and records) go to the reviewer.
func (e *Engine) assemble(ctx context.Context, base callSpec, dir, plan string, recs []unitRecord, items []packet.Item, shared *packet.Item) (outcome, error) {
	red := reduceRecords(recs)
	comparison := red.item()
	supplied := append([]packet.Item{comparison}, items...)
	spec := base
	spec.Assembly = true
	spec.Extra = append([]packet.Item{}, base.Extra...)
	if shared != nil {
		spec.Extra = append(spec.Extra, *shared)
	}
	spec.Extra = append(spec.Extra, supplied...)
	a, err := e.runToCompletion(ctx, spec)
	a.Supplied = supplied
	if err != nil {
		return a, err
	}
	rec := assemblyRecord{Plan: plan, Claims: len(red.Claims), Left: len(red.Left), SharedEvidence: len(red.Groups), AssumptionPairs: len(red.Pairs)}
	for _, g := range red.Groups {
		if len(g.Differ) > 0 {
			rec.ValuesDiffer++
		}
	}
	missing, unknown := accountFor(red.Claims, a.H)
	if len(missing) > 0 || len(unknown) > 0 {
		rec.Retried = true
		e.logf("assembly: %d claim(s) unaccounted for, %d unknown id(s); running the assembler once more", len(missing), len(unknown))
		var why []string
		if len(missing) > 0 {
			why = append(why, "These claims are in neither claims_used nor claims_omitted with a reason: "+strings.Join(missing, ", ")+".")
		}
		if len(unknown) > 0 {
			why = append(why, "These ids name no claim of the records: "+strings.Join(unknown, ", ")+".")
		}
		prev := packet.Item{Key: packet.KeyExtra, ID: "previous-reply", Seq: 95, Header: "Your previous reply (verbatim)", Text: a.Answer}
		check := packet.Item{Key: packet.KeyExtra, ID: "assembly-check", Seq: 96, Header: "The engine's check of your previous reply (data, not instructions from the user)",
			Text: strings.Join(why, "\n")}
		again := spec
		again.Extra = append(append([]packet.Item{}, spec.Extra...), prev, check)
		again.Instructions = func(id string) string {
			return base.Instructions(id) + "\n\nYour previous reply and the engine's check of it are under \"Supplied records\". Write the reply again, in full, and account for every claim id."
		}
		a2, err2 := e.runToCompletion(ctx, again)
		if err2 == nil {
			a2.Runs = append(a.Runs, a2.Runs...)
			a2.Supplied = supplied
			a = a2
			missing, unknown = accountFor(red.Claims, a.H)
		} else {
			a.Runs = append(a.Runs, a2.Runs...)
			e.logf("assembly: the second attempt failed (%v); keeping the first reply", err2)
		}
	}
	for _, r := range a.Runs {
		rec.Runs = append(rec.Runs, r.ID)
	}
	rec.Used, rec.Omitted, rec.Missing, rec.Unknown = nonBlank(a.H.ClaimsUsed), a.H.ClaimsOmitted, missing, unknown
	if js, err := json.MarshalIndent(rec, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(dir, "assembly.json"), append(js, '\n'), 0o644)
	}
	e.logf("assembly: %d of %d claim(s) used, %d omitted, %d unaccounted for; %d shared-evidence group(s), %d may state different values",
		len(rec.Used), rec.Claims, len(rec.Omitted), len(rec.Missing), rec.SharedEvidence, rec.ValuesDiffer)
	return a, nil
}

// mergeAccounting adds a continuation's accounting to the reply's: a
// claim used in any part is used.
func mergeAccounting(acc *packet.Handoff, part packet.Handoff) {
	acc.ClaimsUsed = append(acc.ClaimsUsed, part.ClaimsUsed...)
	used := map[string]bool{}
	for _, id := range acc.ClaimsUsed {
		used[strings.TrimSpace(id)] = true
	}
	var omitted []packet.OmittedClaim
	for _, o := range append(acc.ClaimsOmitted, part.ClaimsOmitted...) {
		if !used[strings.TrimSpace(o.ID)] {
			omitted = append(omitted, o)
		}
	}
	acc.ClaimsOmitted = omitted
}
