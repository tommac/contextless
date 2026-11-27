// Package creativity replays bounded cache and read-scheduling policies on fixed
// workloads. Evaluator-only traces are produced outside the reader process.
package creativity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

type Policy struct {
	Admit     int `json:"admit,omitempty"`
	History   int `json:"history,omitempty"`
	Recency   int `json:"recency,omitempty"`
	Frequency int `json:"frequency,omitempty"`
	Value     int `json:"value,omitempty"`
	Cost      int `json:"cost,omitempty"`
	Age       int `json:"age,omitempty"`
}

type Prediction struct {
	Workload   string `json:"workload"`
	Baseline   string `json:"baseline"`
	Prediction string `json:"prediction"`
}

type Evaluation struct {
	Family   string     `json:"family"`
	Scenario int        `json:"scenario"`
	Seed     uint32     `json:"seed"`
	Policy   *Policy    `json:"policy"`
	Test     Prediction `json:"test"`
}

type Result struct {
	Score                int               `json:"score"`
	Baselines            map[string]int    `json:"baselines"`
	Feasible             bool              `json:"feasible"`
	BeatsBaseline        bool              `json:"beats_baseline"`
	Distinct             bool              `json:"distinct"`
	Success              bool              `json:"success"`
	Fingerprint          string            `json:"fingerprint"`
	BaselineFingerprints map[string]string `json:"baseline_fingerprints"`
	TestObserved         string            `json:"test_observed"`
	TestConfirmed        bool              `json:"test_confirmed"`
	TestDistinguishes    bool              `json:"test_distinguishes"`
	TestDifference       int               `json:"test_difference"`
	PeakResident         int               `json:"peak_resident,omitempty"`
	PeakHistory          int               `json:"peak_history,omitempty"`
	PeakCredits          int               `json:"peak_credits,omitempty"`
}

type trial struct {
	score int
	// Cache emits a hit/miss bit for each request. The scheduler emits a stable
	// arrival ID for each fulfilled read and an epoch delimiter.
	actions                    []int
	resident, history, credits int
}

var budgets = [...]int{4, 6, 8, 4, 6, 8, 5, 7}

func (p Policy) validate() error {
	if p.Admit < 0 || p.Admit > 3 || p.History < 0 || p.History > 3 {
		return fmt.Errorf("admit must be 1..3 (or 0 for default 1), history 0..3")
	}
	for _, w := range []int{p.Recency, p.Frequency, p.Value, p.Cost, p.Age} {
		if w < -4 || w > 4 {
			return fmt.Errorf("policy weights must be integers in [-4,4]")
		}
	}
	return nil
}

func (p Policy) rank(recency, frequency, value, cost, age int) int {
	return p.Recency*recency + p.Frequency*frequency + p.Value*value + p.Cost*cost + p.Age*age
}

func compare(a, b int) string {
	if a > b {
		return "win"
	}
	if a < b {
		return "loss"
	}
	return "tie"
}

func fingerprint(actions []int) string {
	data, _ := json.Marshal(actions)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// Evaluate compares the same candidate and baseline programs on identical
// events. Diagnostic and intervention traces never contribute to the main score.
func Evaluate(q Evaluation) (Result, error) {
	if q.Scenario < 0 || q.Scenario >= len(budgets) || q.Policy == nil {
		return Result{}, fmt.Errorf("provide a policy object and scenario 0..7")
	}
	if err := q.Policy.validate(); err != nil {
		return Result{}, err
	}
	var baselines []string
	switch q.Family {
	case "cache":
		baselines = []string{"lru", "2q"}
	case "schedule":
		baselines = []string{"fifo", "value-cost"}
	default:
		return Result{}, fmt.Errorf("family must be cache or schedule")
	}
	if q.Test.Baseline != baselines[0] && q.Test.Baseline != baselines[1] {
		return Result{}, fmt.Errorf("test baseline must be one of %v", baselines)
	}
	switch q.Test.Workload {
	case "scan", "burst", "reuse":
	default:
		return Result{}, fmt.Errorf("unknown test workload %q", q.Test.Workload)
	}
	switch q.Test.Prediction {
	case "win", "tie", "loss":
	default:
		return Result{}, fmt.Errorf("prediction must be win, tie or loss")
	}
	bound := budgets[q.Scenario]
	mainTrace := makeWorkload(q.Scenario, q.Seed, "main")
	// Common fixed probes make behavior signatures comparable between repeats.
	var probes []workload
	for _, name := range []string{"scan", "burst", "reuse"} {
		probes = append(probes, makeWorkload(q.Scenario, 0x6743c127, name))
	}
	testTrace := makeWorkload(q.Scenario, q.Seed^0x971e481b, q.Test.Workload)
	run := func(trace workload, mode string) trial {
		if q.Family == "cache" {
			return runCache(trace.requests, bound, *q.Policy, mode)
		}
		return runSchedule(trace.epochs, bound, *q.Policy, mode)
	}
	signature := func(mode string) string {
		var all []int
		for _, probe := range probes {
			all = append(all, -2)
			all = append(all, run(probe, mode).actions...)
		}
		return fingerprint(all)
	}
	got := run(mainTrace, "policy")
	r := Result{Score: got.score, PeakResident: got.resident, PeakHistory: got.history, PeakCredits: got.credits,
		Baselines: map[string]int{}, BaselineFingerprints: map[string]string{}, Fingerprint: signature("policy"),
		Feasible:      got.resident <= bound && got.history <= 3*bound && got.credits <= bound,
		BeatsBaseline: true, Distinct: true}
	for _, base := range baselines {
		r.Baselines[base] = run(mainTrace, base).score
		r.BaselineFingerprints[base] = signature(base)
		if r.Score <= r.Baselines[base] {
			r.BeatsBaseline = false
		}
		if r.Fingerprint == r.BaselineFingerprints[base] {
			r.Distinct = false
		}
	}
	r.Success = r.BeatsBaseline && r.Distinct && r.Feasible
	observed, control := run(testTrace, "policy"), run(testTrace, q.Test.Baseline)
	r.TestDifference = observed.score - control.score
	r.TestObserved = compare(observed.score, control.score)
	r.TestConfirmed = r.TestObserved == q.Test.Prediction
	r.TestDistinguishes = r.TestObserved != "tie"
	return r, nil
}

func decodeOne(rd io.Reader, dst any) error {
	decoder := json.NewDecoder(rd)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("expected exactly one JSON object")
		}
		return err
	}
	return nil
}

func EvaluateJSON(rd io.Reader, out io.Writer) error {
	var q Evaluation
	if err := decodeOne(rd, &q); err != nil {
		return err
	}
	r, err := Evaluate(q)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(r)
}
