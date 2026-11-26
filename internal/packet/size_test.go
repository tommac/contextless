package packet

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/providers"
)

func prof(id string, window int) providers.Profile {
	return providers.Profile{ID: id, Provider: id, Model: "m", Window: window, EffectiveWindow: window,
		SysOverhead: 1000, LiveUsage: providers.LivePerTurn, Enabled: true}.WithDefaults()
}

func history(n, bytesEach int) []Item {
	items := []Item{
		{Key: KeySession, ID: "session", Tier: 0, Header: "Session", Text: "Title: t"},
		{Key: KeyFocus, ID: fmt.Sprintf("t%06d", n+1), Seq: n + 1, Tier: 0, Header: "focus", Text: "the question"},
	}
	for i := 1; i <= n; i++ {
		key, tier := KeyCatalog, 1
		switch {
		case i > n-4:
			key, tier = KeyRecent, 4
		case i%5 == 0:
			key, tier = KeyThread, 3
		case i%7 == 0:
			key, tier = KeyRelevant, 2
		}
		body := fmt.Sprintf("turn %d ", i) + strings.Repeat("x", bytesEach)
		items = append(items, Item{Key: key, ID: fmt.Sprintf("t%06d", i), Seq: i, Tier: tier,
			Header: fmt.Sprintf("t%06d · user", i), Text: body, Card: fmt.Sprintf("t%06d user: card", i)})
	}
	return items
}

func TestEscalatesBeforeDropping(t *testing.T) {
	items := history(40, 3000) // about 36k tokens
	s, err := Size(Task{RunID: "r", Kind: KindRespond, Role: "responder"}, items, []providers.Profile{prof("small", 40_000), prof("big", 400_000)}, 8000)
	if err != nil {
		t.Fatal(err)
	}
	if s.Profile.ID != "big" || !s.Budget.Escalated || len(s.Dropped) != 0 {
		t.Fatalf("want escalation to big with nothing dropped, got %s escalated=%v dropped=%d", s.Profile.ID, s.Budget.Escalated, len(s.Dropped))
	}
}

func TestDropsByTierAndNeverAltersRecords(t *testing.T) {
	items := history(200, 3000) // far over a 60k window
	orig := map[string]string{}
	for _, it := range items {
		orig[it.ID] = it.Text
	}
	s, err := Size(Task{RunID: "r", Kind: KindRespond, Role: "responder"}, items, []providers.Profile{prof("small", 60_000)}, 8000)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Budget.Fits() || len(s.Dropped) == 0 {
		t.Fatalf("budget %+v dropped %d", s.Budget, len(s.Dropped))
	}
	maxDropped := 0
	for _, d := range s.Dropped {
		if d.Tier == 0 {
			t.Fatalf("tier 0 record %s dropped", d.ID)
		}
		if d.Tier > maxDropped {
			maxDropped = d.Tier
		}
	}
	for _, k := range s.Items {
		if k.Text != orig[k.ID] {
			t.Fatalf("record %s was altered", k.ID)
		}
		if k.Tier != 0 && k.Tier < maxDropped {
			t.Fatalf("kept tier %d record %s while dropping tier %d", k.Tier, k.ID, maxDropped)
		}
		if !strings.Contains(s.Brief, k.Text) {
			t.Fatalf("kept record %s not in the brief verbatim", k.ID)
		}
	}
	if !strings.Contains(s.Brief, "## Omitted for budget") {
		t.Fatal("brief does not list what was omitted")
	}
}

func TestTooBigWhenProtectedRecordsDoNotFit(t *testing.T) {
	items := []Item{{Key: KeyFocus, ID: "t000001", Tier: 0, Text: strings.Repeat("y", 400_000)}}
	_, err := Size(Task{RunID: "r", Kind: KindRespond}, items, []providers.Profile{prof("small", 60_000)}, 8000)
	if err == nil || !strings.Contains(err.Error(), "do not fit") {
		t.Fatalf("want ErrTooBig, got %v", err)
	}
}

func TestExtractJSONFindsTerminalHandoff(t *testing.T) {
	text := "Sure.\n```json\n{\"status\":\"started\"}\n```\nthen\n```json\n{\"run_id\":\"r\",\"status\":\"completed\",\"answer\":\"hi {there}\"}\n```"
	raw, err := ExtractJSON([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	h, err := ParseHandoff(raw)
	if err != nil || h.Answer != "hi {there}" {
		t.Fatalf("%+v %v", h, err)
	}
}
