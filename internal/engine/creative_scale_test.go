package engine

import (
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

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
		probe := r.Repair
		if probe.Design != fixedDesignID || probe.Requests != 24 || probe.ReserveBefore != 8 {
			t.Fatalf("unexpected repair probe: %+v", probe)
		}
		if r.Edit == "constraint-tightened" {
			if probe.Physical != 4 || probe.ServedBefore != 8 || probe.ReplacementsBefore != 0 || probe.Admitted != (r.Missed == 0) {
				t.Fatalf("incorrect original workload or admitted design: %+v", r)
			}
			if probe.Admitted {
				if probe.ReserveAfter != 4 || probe.ServedAfter != 24 || probe.ReplacementsAfter != 20 {
					t.Fatalf("admitted repair failed to start replacements: %+v", r)
				}
			} else if probe.ReserveAfter != 8 || probe.ServedAfter != 8 || probe.ReplacementsAfter != 0 {
				t.Fatalf("missed revision nevertheless repaired: %+v", r)
			}
		} else if probe.Physical != 8 || probe.ReserveAfter != 8 || probe.ServedBefore != 24 || probe.ServedAfter != 24 ||
			probe.ReplacementsBefore != 16 || probe.ReplacementsAfter != 16 {
			t.Fatalf("control changed working reservation behavior: %+v", r)
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

func TestScaleRepairRequiresActualPacketAdmission(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := &Engine{V: &vault.Vault{Dir: dir}, S: s}
	edits, err := scaleFixtures(e)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.Exec("UPDATE receipts SET question=? WHERE id=?", strings.Repeat("x", scaleReaderBytes), fixedDesignID); err != nil {
		t.Fatal(err)
	}
	r, err := measureScale(e, 0, edits[0], "source-linked")
	if err != nil {
		t.Fatal(err)
	}
	if r.Candidates != 1 || r.Packed != 0 || r.Repair.Admitted || r.Repair.ServedAfter != r.Repair.ServedBefore {
		t.Fatalf("selected design was repaired without fitting in the packet: %+v", r)
	}
}
