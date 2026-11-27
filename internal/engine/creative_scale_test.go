package engine

import "testing"

func TestCreativeScaleMeasuresIndependentEdits(t *testing.T) {
	report, err := runCreativeScale(t.TempDir(), []int{100, 120})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Results) != 18 || report.InputCeiling != scaleReaderBytes {
		t.Fatalf("unexpected report: %+v", report)
	}
	for _, r := range report.Results {
		if r.PacketBytes > scaleReaderBytes || r.Packed > r.Candidates ||
			r.Packed != r.Unnecessary+r.Required-r.Missed ||
			r.Candidates != r.CandidateExtras+r.Required-r.CandidateMisses {
			t.Fatalf("inconsistent candidate or packet accounting: %+v", r)
		}
		if r.Packed > 0 && (r.PacketBytes == 0 || r.HistoricalBytes == 0 || r.Opened != 1 || r.Passages < r.Packed) {
			t.Fatalf("missing exact evidence: %+v", r)
		}
		switch r.Method {
		case "source-linked":
			if r.Missed != 0 || r.Unnecessary != 0 || r.Candidates != r.Required {
				t.Fatalf("failed source lookup: %+v", r)
			}
		case "keyword-watcher":
			if r.Edit == "constraint-tightened" && r.Missed != 1 || r.Edit == "unrelated-edit" && r.Unnecessary == 0 {
				t.Fatalf("wrong lexical selector outcome: %+v", r)
			}
		case "review-all":
			if r.Candidates != r.Unrelated+3 || r.Candidates <= r.Packed || r.Missed != 0 || r.Unnecessary == 0 {
				t.Fatalf("all-design baseline was not ceiling limited: %+v", r)
			}
		}
	}
}
