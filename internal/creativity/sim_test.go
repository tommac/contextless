package creativity

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func requests(keys ...int) []request {
	var trace []request
	for _, key := range keys {
		trace = append(trace, request{key: key, value: 1, cost: 1})
	}
	return trace
}

func TestFullTwoQueueUsesGhostPromotion(t *testing.T) {
	trace := requests(1, 2, 3, 4, 1, 5, 1, 1)
	got := runCache(trace, 4, Policy{}, "2q")
	if !slices.Equal(got.actions, []int{0, 0, 0, 0, 1, 0, 0, 1}) || got.score != 2 || got.resident != 4 || got.history > 2 {
		t.Fatal(got)
	}
	if lru := runCache(trace, 4, Policy{}, "lru"); lru.score != 3 {
		t.Fatal(lru)
	}
}

func TestBoundedAdmissionEvictsMissingHistory(t *testing.T) {
	p := Policy{Admit: 2, History: 1}
	r := runCache(requests(1, 2, 3, 1, 1, 1), 2, p, "policy")
	if !slices.Equal(r.actions, []int{0, 0, 0, 0, 0, 1}) || r.history != 2 || r.resident != 1 {
		t.Fatal(r)
	}
}

func TestSchedulerAccountingAndDeadline(t *testing.T) {
	epochs := [][]job{{{request: request{0, 4, 2}, id: 1, expiry: 0}, {request: request{1, 9, 1}, id: 2, expiry: 0}, {request: request{2, 7, 1}, id: 3, expiry: 0}}}
	fifo := runSchedule(epochs, 2, Policy{}, "fifo")
	best := runSchedule(epochs, 2, Policy{}, "value-cost")
	if fifo.score != 4 || best.score != 16 || best.credits != 2 || fifo.credits != 2 {
		t.Fatalf("FIFO=%+v greedy=%+v", fifo, best)
	}
}

func TestBaselineBehaviorAndPredictions(t *testing.T) {
	for _, example := range []struct{ family, baseline string }{{"cache", "lru"}, {"schedule", "fifo"}} {
		for scenario := range budgets {
			q := Evaluation{Family: example.family, Scenario: scenario, Seed: 1447, Policy: &Policy{}, Test: Prediction{"burst", example.baseline, "tie"}}
			a, err := Evaluate(q)
			if err != nil {
				t.Fatal(err)
			}
			if !a.Feasible || a.Distinct || a.Success || a.BeatsBaseline || !a.TestConfirmed || a.TestDistinguishes || a.Score != a.Baselines[example.baseline] {
				t.Fatal(a)
			}
			b, err := Evaluate(q)
			if err != nil || !reflect.DeepEqual(a, b) {
				t.Fatal("non-reproducible result")
			}
		}
	}
}

func TestJSONAndDemo(t *testing.T) {
	var out bytes.Buffer
	bad := []string{
		`{"family":"cache","policy":{"bonus":2}}`,
		`{"family":"cache","policy":{"recency":5}}`,
		`{"family":"schedule","scenario":8,"policy":{}}`,
		`{"family":"cache","policy":{},"test":{"workload":"scan","baseline":"fifo","prediction":"win"}}`,
		`{} {}`,
	}
	for _, input := range bad {
		if err := EvaluateJSON(strings.NewReader(input), &out); err == nil {
			t.Errorf("accepted %s", input)
		}
	}
	if err := EvaluateJSON(strings.NewReader(`{"family":"cache","scenario":0,"seed":42,"policy":{},"test":{"baseline":"lru","prediction":"tie","workload":"reuse"}}`), &out); err != nil {
		t.Fatal(err)
	}
	var r Result
	if err := json.Unmarshal(out.Bytes(), &r); err != nil || !r.Feasible || !r.TestConfirmed {
		t.Fatal(r, err)
	}
	rows, err := DemoRows("")
	if err != nil || len(rows) != 8 {
		t.Fatal(rows, err)
	}
	wantScores := []int{33, 9, 27, 27, 33, 75, 27, 27}
	for i, row := range rows {
		if row.Score != wantScores[i] || row.Requests != 63 {
			t.Fatal("recorded comparison changed", rows)
		}
	}
	out.Reset()
	if err := WriteWorkload(&out); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "workload.json")
	if err := os.WriteFile(path, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	edited, err := DemoRows(path)
	if err != nil || !reflect.DeepEqual(edited, rows) {
		t.Fatal(edited, err)
	}
}

func TestRejectsUnboundedOrAmbiguousWorkload(t *testing.T) {
	for _, parts := range [][]segment{nil, {{}}, {{Scan: 2, Pages: []string{"a"}}}, {{Scan: 1000, Repeat: 1000}}, {{Pages: []string{""}}}} {
		if _, err := expand(parts); err == nil {
			t.Fatal("accepted", parts)
		}
	}
}
