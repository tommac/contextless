package packet

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/vault"
)

// The value lane finds the one sentence that states a value about the
// question's topic, among turns full of the same topic words and of
// other numbers.
func TestValueLaneFindsTheStatementAmongTopicNoise(t *testing.T) {
	var turns []vault.Turn
	for i := 1; i <= 40; i++ {
		body := fmt.Sprintf("Someone should check which port the staging Postgres instance listens on after the restart. "+
			"The staging Postgres instance needs another look at ticket %d before the deploy. "+
			"The pool_size=20 setting on staging stays.", 4000+i)
		if i == 17 {
			body += " We settled on port 5433 for the staging Postgres instance."
		}
		turns = append(turns, vault.Turn{ID: vault.TurnID(i), Seq: i, Body: body})
	}
	got := RankTurns(turns, nil, "Which port does the staging Postgres instance use?", 0, 3)
	if len(got) == 0 || got[0].Turn.ID != "t000017" {
		var ids []string
		for _, r := range got {
			ids = append(ids, r.Turn.ID)
		}
		t.Fatalf("want t000017 first, got %v", ids)
	}
	// an exact phrase finds the newest statement of it first
	turns = append(turns, vault.Turn{ID: vault.TurnID(41), Seq: 41, Body: "Moved: we settled on port 5433 again after the rollback."})
	if got := RankTurns(turns, nil, "settled on port 5433", 0, 2); len(got) == 0 || got[0].Turn.ID != "t000041" {
		t.Fatalf("exact lane should put the newest statement first: %+v", got)
	}
	// nothing at or after beforeSeq is ranked
	for _, r := range RankTurns(turns, nil, "settled on port 5433", 20, 5) {
		if r.Turn.Seq >= 20 {
			t.Fatalf("ranked %s, which is not before t000020", r.Turn.ID)
		}
	}
}

// Digests label every omitted turn, grouped by thread and in blocks.
func TestDigestsNameEveryOmittedTurn(t *testing.T) {
	var items []Item
	for i := 1; i <= 150; i++ {
		it := Item{Key: KeyCatalog, ID: vault.TurnID(i), Tier: 1, Seq: i}
		if i%5 != 0 {
			it.Threads, it.Group, it.Keywords = []string{"th-kitchen"}, "Kitchen remodel", []string{"cabinets", "budget"}
		} else {
			it.Keywords = []string{"lisbon"}
		}
		items = append(items, it)
	}
	lines := Digests(items)
	kitchen, other := 0, 0
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, `thread "Kitchen remodel" (th-kitchen): `):
			kitchen++
			if !strings.HasSuffix(l, "· cabinets, budget") {
				t.Errorf("kitchen line lacks its keywords: %s", l)
			}
		case strings.HasPrefix(l, "other turns: "):
			other++
		default:
			t.Errorf("unexpected line %q", l)
		}
	}
	if kitchen != 3 || other != 1 {
		t.Fatalf("want 3 kitchen blocks and 1 other, got %d and %d:\n%s", kitchen, other, strings.Join(lines, "\n"))
	}
	named := map[int]bool{}
	re := regexp.MustCompile(`t(\d{6})(?:–t(\d{6}))?`)
	for _, l := range lines {
		for _, m := range re.FindAllStringSubmatch(l, -1) {
			a, _ := strconv.Atoi(m[1])
			b := a
			if m[2] != "" {
				b, _ = strconv.Atoi(m[2])
			}
			for s := a; s <= b; s++ {
				named[s] = true
			}
		}
	}
	for i := 1; i <= 150; i++ {
		if !named[i] {
			t.Fatalf("t%06d is not named by any digest line", i)
		}
	}
}
