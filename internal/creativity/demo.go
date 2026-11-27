package creativity

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
)

//go:embed testdata/replay.json
var defaultWorkload []byte

type segment struct {
	Pages  []string `json:"pages"`
	Scan   int      `json:"scan"`
	Repeat int      `json:"repeat"`
	Value  int      `json:"value"`
	Cost   int      `json:"cost"`
}

type demoInput struct {
	Capacity   int       `json:"capacity"`
	Before     []segment `json:"before"`
	After      []segment `json:"after"`
	TwoTouch   *Policy   `json:"two_touch"`
	ValueFirst *Policy   `json:"value_first"`
}

type DemoRow struct {
	Workload string `json:"workload"`
	Policy   string `json:"policy"`
	Score    int    `json:"score"`
	Hits     int    `json:"hits"`
	Requests int    `json:"requests"`
}

func expand(parts []segment) ([]request, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("a workload must have at least one segment")
	}
	keys := map[string]int{}
	var trace []request
	next := 1
	for i, s := range parts {
		if (len(s.Pages) == 0) == (s.Scan == 0) || s.Scan < 0 || s.Scan > 100000 || s.Repeat < 0 || s.Repeat > 100000 || s.Value < 0 || s.Value > 100 || s.Cost < 0 || s.Cost > 100 {
			return nil, fmt.Errorf("segment %d requires exactly one of pages or scan, with bounded positive counts and weights", i+1)
		}
		n := max(len(s.Pages), s.Scan)
		repeats := max(1, s.Repeat)
		if n > (100000-len(trace))/repeats {
			return nil, fmt.Errorf("workload exceeds 100000 requests")
		}
		for range repeats {
			for k := range n {
				id := next
				if s.Scan == 0 {
					page := s.Pages[k]
					if page == "" {
						return nil, fmt.Errorf("empty page name in segment %d", i+1)
					}
					id = keys[page]
					if id == 0 {
						id = next
						next++
						keys[page] = id
					}
				} else {
					next++
				}
				trace = append(trace, request{key: id, value: max(1, s.Value), cost: max(1, s.Cost)})
			}
		}
	}
	return trace, nil
}

// WriteWorkload exports the built-in input for editing outside the source tree.
func WriteWorkload(out io.Writer) error {
	_, err := out.Write(defaultWorkload)
	return err
}

func DemoRows(path string) ([]DemoRow, error) {
	var rd io.Reader = bytes.NewReader(defaultWorkload)
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		rd = f
	}
	var d demoInput
	if err := decodeOne(rd, &d); err != nil {
		return nil, err
	}
	if d.Capacity < 1 || d.Capacity > 128 || d.TwoTouch == nil || d.ValueFirst == nil {
		return nil, fmt.Errorf("provide capacity 1..128 and both policy objects")
	}
	if err := d.TwoTouch.validate(); err != nil {
		return nil, err
	}
	if err := d.ValueFirst.validate(); err != nil {
		return nil, err
	}
	var rows []DemoRow
	for _, phase := range []struct {
		name  string
		parts []segment
	}{{"before", d.Before}, {"after", d.After}} {
		trace, err := expand(phase.parts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", phase.name, err)
		}
		for _, entry := range []struct {
			name, mode string
			p          Policy
		}{
			{"two-touch", "policy", *d.TwoTouch}, {"value-first", "policy", *d.ValueFirst},
			{"lru", "lru", Policy{}}, {"2q", "2q", Policy{}},
		} {
			r := runCache(trace, d.Capacity, entry.p, entry.mode)
			hits := 0
			for _, hit := range r.actions {
				hits += hit
			}
			rows = append(rows, DemoRow{phase.name, entry.name, r.score, hits, len(trace)})
		}
	}
	return rows, nil
}

func RunDemo(out io.Writer, path string) error {
	rows, err := DemoRows(path)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintln(out, "Cache replay (saved fetch cost × passage value; larger is better)"); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(out, "%-8s %-14s %6s   %s\n", "Trace", "Policy", "Score", "Hits/reads"); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err = fmt.Fprintf(out, "%-8s %-14s %6d   %d/%d\n", r.Workload, r.Policy, r.Score, r.Hits, r.Requests); err != nil {
			return err
		}
	}
	return nil
}
