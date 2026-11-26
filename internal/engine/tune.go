package engine

import (
	"fmt"
	"strings"

	"github.com/tommac/contextless/internal/packet"
)

// FailureClass implements the 4-way UTILMEM failure attribution taxonomy so
// evaluations distinguish retrieval failures from navigation and reader errors.
type FailureClass string

const (
	FailureCorrect            FailureClass = "correct"
	FailureRetrievalMiss      FailureClass = "retrieval_miss"
	FailureNavigationMiss     FailureClass = "navigation_miss"
	FailureUtilizationFailure FailureClass = "utilization_failure"
	FailureStalenessOrAttrib  FailureClass = "staleness_or_attribution"
)

// ClassifyAnswerFailure attributes an answer's outcome against a known probe
// and the exact packet the reader saw.
func ClassifyAnswerFailure(p Probe, sized packet.Sized, answer string, staleValues ...string) FailureClass {
	if holdsAny(answer, p.accepted()) {
		return FailureCorrect
	}
	for _, sv := range staleValues {
		if sv != "" && holdsAny(answer, []string{sv}) {
			return FailureStalenessOrAttrib
		}
	}
	where := placeSource(sized, p.Source, p.accepted())
	switch where {
	case WhereVerbatim:
		return FailureUtilizationFailure
	case WhereCard, WhereRange:
		return FailureNavigationMiss
	default:
		return FailureRetrievalMiss
	}
}

// TuneCandidate is one evaluated compiler weight + packing configuration.
type TuneCandidate struct {
	Weights      packet.LaneWeights `json:"weights"`
	AdaptivePack bool               `json:"adaptive_pack"`
	TrainHits    int                `json:"train_hits"`
	TrainTotal   int                `json:"train_total"`
	HeldOutHits  int                `json:"held_out_hits"`
	HeldOutTotal int                `json:"held_out_total"`
}

// TuneReport summarizes an offline weight & packing optimization run with a
// strict train / held-out split so settings that overfit are automatically
// rejected.
type TuneReport struct {
	Baseline    TuneCandidate `json:"baseline"`
	Best        TuneCandidate `json:"best"`
	Adopted     bool          `json:"adopted"`
	Reason      string        `json:"reason"`
	Evaluated   int           `json:"evaluated"`
	SignalStats map[string]int `json:"signal_stats,omitempty"`
}

// Tune runs an offline coordinate sweep over lane weights and adaptive packing
// on probes using a deterministic 60/40 train / held-out split. No model calls
// are made. A candidate is recommended (Adopted=true) only when it strictly
// improves held-out verbatim recovery without regressing on the training set.
func (e *Engine) Tune(probes []Probe) (TuneReport, error) {
	if len(probes) < 2 {
		return TuneReport{}, fmt.Errorf("tune requires at least 2 probes (got %d)", len(probes))
	}
	// Deterministic stratified 60/40 split: every 5th and 2nd item go to held-out.
	var train, heldOut []Probe
	for i, p := range probes {
		if i%5 == 1 || i%5 == 4 {
			heldOut = append(heldOut, p)
		} else {
			train = append(train, p)
		}
	}
	if len(heldOut) == 0 {
		heldOut = probes[len(probes)-1:]
		train = probes[:len(probes)-1]
	}

	eval := func(w packet.LaneWeights, adaptive bool, set []Probe) (int, error) {
		oldAdaptive := e.Cfg.AdaptivePack
		oldWeights := e.Cfg.LaneWeights
		e.Cfg.AdaptivePack = adaptive
		e.Cfg.LaneWeights = map[string]float64{
			"exact":       w.Exact,
			"full_text":   w.FullText,
			"value":       w.Value,
			"working_set": w.WorkingSet,
		}
		defer func() {
			e.Cfg.AdaptivePack = oldAdaptive
			e.Cfg.LaneWeights = oldWeights
		}()
		res, err := e.RunProbes(set)
		if err != nil {
			return 0, err
		}
		hits := 0
		for _, r := range res {
			if r.Where == WhereVerbatim {
				hits++
			}
		}
		return hits, nil
	}

	baseW := packet.LaneWeights{Exact: 1, FullText: 1, Value: 1, WorkingSet: 0.5}
	baseTrain, err := eval(baseW, false, train)
	if err != nil {
		return TuneReport{}, err
	}
	baseHeld, err := eval(baseW, false, heldOut)
	if err != nil {
		return TuneReport{}, err
	}
	baseline := TuneCandidate{
		Weights: baseW, AdaptivePack: false,
		TrainHits: baseTrain, TrainTotal: len(train),
		HeldOutHits: baseHeld, HeldOutTotal: len(heldOut),
	}

	grid := []struct {
		w        packet.LaneWeights
		adaptive bool
	}{
		{packet.LaneWeights{Exact: 1.0, FullText: 1.0, Value: 1.0, WorkingSet: 0.5}, true},
		{packet.LaneWeights{Exact: 1.5, FullText: 1.0, Value: 0.5, WorkingSet: 0.5}, false},
		{packet.LaneWeights{Exact: 1.5, FullText: 1.0, Value: 0.5, WorkingSet: 0.5}, true},
		{packet.LaneWeights{Exact: 1.0, FullText: 1.5, Value: 0.75, WorkingSet: 0.5}, false},
		{packet.LaneWeights{Exact: 1.0, FullText: 1.5, Value: 0.75, WorkingSet: 0.5}, true},
		{packet.LaneWeights{Exact: 1.0, FullText: 1.0, Value: 1.5, WorkingSet: 0.5}, true},
	}

	best := baseline
	evaluated := 1
	for _, g := range grid {
		th, err := eval(g.w, g.adaptive, train)
		if err != nil {
			return TuneReport{}, err
		}
		evaluated++
		if th >= best.TrainHits {
			hh, err := eval(g.w, g.adaptive, heldOut)
			if err != nil {
				return TuneReport{}, err
			}
			cand := TuneCandidate{
				Weights: g.w, AdaptivePack: g.adaptive,
				TrainHits: th, TrainTotal: len(train),
				HeldOutHits: hh, HeldOutTotal: len(heldOut),
			}
			if cand.HeldOutHits > best.HeldOutHits || (cand.HeldOutHits == best.HeldOutHits && cand.TrainHits > best.TrainHits) {
				best = cand
			}
		}
	}

	adopted := best.HeldOutHits > baseline.HeldOutHits && best.TrainHits >= baseline.TrainHits
	reason := "no candidate beat baseline on held-out probes; keeping baseline defaults"
	if adopted {
		reason = fmt.Sprintf("candidate improved held-out recovery from %d/%d to %d/%d (train %d/%d -> %d/%d)",
			baseline.HeldOutHits, baseline.HeldOutTotal, best.HeldOutHits, best.HeldOutTotal,
			baseline.TrainHits, baseline.TrainTotal, best.TrainHits, best.TrainTotal)
	}
	var sigs map[string]int
	if e.S != nil {
		sigs, _ = e.S.SignalSummary()
	}
	return TuneReport{
		Baseline:    baseline,
		Best:        best,
		Adopted:     adopted,
		Reason:      reason,
		Evaluated:   evaluated,
		SignalStats: sigs,
	}, nil
}

// FormatTuneReport renders a TuneReport for CLI output.
func FormatTuneReport(r TuneReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Offline compiler tuning (%d configurations evaluated)\n", r.Evaluated)
	fmt.Fprintf(&b, "  baseline: train %d/%d  held-out %d/%d  adaptive=%v  weights=%+v\n",
		r.Baseline.TrainHits, r.Baseline.TrainTotal, r.Baseline.HeldOutHits, r.Baseline.HeldOutTotal,
		r.Baseline.AdaptivePack, r.Baseline.Weights)
	fmt.Fprintf(&b, "  best:     train %d/%d  held-out %d/%d  adaptive=%v  weights=%+v\n",
		r.Best.TrainHits, r.Best.TrainTotal, r.Best.HeldOutHits, r.Best.HeldOutTotal,
		r.Best.AdaptivePack, r.Best.Weights)
	fmt.Fprintf(&b, "  verdict:  adopted=%v (%s)\n", r.Adopted, r.Reason)
	if len(r.SignalStats) > 0 {
		fmt.Fprintf(&b, "  recorded signals: %v\n", r.SignalStats)
	}
	return b.String()
}
