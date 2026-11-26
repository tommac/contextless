package engine

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// probeTopic is one thread of the synthetic conversation. Its filler
// shares vocabulary with the probe questions, so a probe is only
// recovered when retrieval tells the planted turn apart from the noise.
type probeTopic struct {
	id, title string
	summary   string
	keywords  []string
	sentences []string
	words     []string
}

var probeTopics = []probeTopic{
	{id: "th-db", title: "Database migration", summary: "Talked through the database migration.",
		keywords: []string{"postgres", "migration", "replica", "deploy"},
		sentences: []string{
			"The staging Postgres instance needs another look at %s before the deploy.",
			"We could move the %s to the replica once the migration settles.",
			"The connection pool on staging keeps hitting %s under load.",
			"Someone should check which port the %s listens on after the restart.",
			"The deploy checklist still mentions %s, which may be stale.",
			"max_connections on the primary is fine, but the %s replica is unclear.",
			"Backups run nightly, and the %s job is the slow one.",
		},
		words: []string{"vacuum settings", "index bloat", "the read path", "pgbouncer", "the schema diff", "lock waits", "the billing tables", "WAL archiving", "ticket 4821", "version 16.4", "the 02:00 backup job", "the pool_size=20 setting"}},
	{id: "th-kitchen", title: "Kitchen remodel", summary: "Discussed the kitchen remodel.",
		keywords: []string{"kitchen", "cabinets", "budget", "electrician"},
		sentences: []string{
			"The cabinets in the showroom were %s, and the wood samples looked different at home.",
			"The electrician wants to revisit %s before the inspection.",
			"The budget has room for %s if the countertop comes in lower.",
			"The contractor suggested %s for the backsplash wall.",
			"We still have to pick the wood for %s.",
			"The kitchen layout puts %s next to the window.",
		},
		words: []string{"the pantry", "the island", "under-cabinet lighting", "the range hood", "the sink base", "drawer pulls", "the tile edge", "the dishwasher panel", "the 36-inch range", "invoice 1187", "the $2,300 quote", "the May 3 delivery"}},
	{id: "th-travel", title: "Lisbon trip", summary: "Planned the Lisbon trip.",
		keywords: []string{"lisbon", "flight", "hotel", "itinerary"},
		sentences: []string{
			"The flight options to Lisbon all route through %s.",
			"The hotel near %s has better reviews but a later check-in.",
			"We could leave a day early and spend it in %s.",
			"The itinerary still needs %s on the second day.",
			"The return flight from Lisbon depends on %s.",
		},
		words: []string{"Alfama", "Sintra", "Madrid", "the river walk", "a food tour", "Belem", "the airport train", "Cascais", "terminal 1", "the 18:05 train", "seat 14C", "the March 2 deadline"}},
	{id: "th-team", title: "Team process", summary: "Talked about team process.",
		keywords: []string{"standup", "hiring", "review", "schedule"},
		sentences: []string{
			"The standup ran long again because of %s.",
			"Hiring for the platform role is waiting on %s.",
			"The review schedule clashes with %s on Thursdays.",
			"We agreed to time-box %s in the standup.",
			"The interview loop needs someone for %s.",
		},
		words: []string{"incident follow-ups", "the roadmap", "on-call handoff", "the design review", "budget planning", "the offsite", "a system design round", "release notes", "sprint 42", "the 3pm sync", "ticket 977", "the Q3 plan"}},
	{id: "th-code", title: "Catalog service code", summary: "Worked on the catalog service code.",
		keywords: []string{"catalog", "sku", "helper", "refactor"},
		sentences: []string{
			"The catalog service has a helper for %s that nobody owns.",
			"SKU parsing breaks when %s shows up in the feed.",
			"The refactor moves %s into its own package.",
			"Tests for the SKU code skip %s entirely.",
			"The catalog importer calls the helper twice for %s.",
		},
		words: []string{"price rounding", "unicode names", "bundle items", "the legacy feed", "image URLs", "variant codes", "tax classes", "stock levels", "PR 2210", "the parseFeedRow path", "catalog/feed.go", "the v2 schema"}},
}

type plantedFact struct {
	at, topic          int
	sentence           string
	typ, question, ans string
	specificCard       bool
}

var planted = []plantedFact{
	{at: 31, topic: 0, sentence: "We settled on port 5433 for the staging Postgres instance.", typ: "config", question: "Which port does the staging Postgres instance use?", ans: "5433"},
	{at: 77, topic: 0, sentence: "Set max_connections to 240 on the replica and leave the primary alone.", typ: "config", question: "What did we set max_connections to on the replica?", ans: "240", specificCard: true},
	{at: 140, topic: 0, sentence: "Before every deploy run `make migrate-up ENV=staging` from the repo root.", typ: "command", question: "What command do we run before every deploy?", ans: "make migrate-up ENV=staging"},
	{at: 52, topic: 1, sentence: "Let's go with the walnut cabinets, not the white oak.", typ: "decision", question: "Which wood did we pick for the cabinets?", ans: "walnut", specificCard: true},
	{at: 95, topic: 1, sentence: "The electrician's name is Priya Natarajan, and she can start in May.", typ: "entity", question: "What's the electrician's name?", ans: "Priya Natarajan"},
	{at: 210, topic: 1, sentence: "The kitchen budget cap is $48,500 including the appliances.", typ: "number", question: "What is the kitchen budget cap?", ans: "$48,500"},
	{at: 120, topic: 2, sentence: "The flight to Lisbon leaves on March 14 at 07:40 from gate B.", typ: "date", question: "When does the flight to Lisbon leave?", ans: "March 14"},
	{at: 260, topic: 2, sentence: "We booked Hotel Avenida Palace for four nights.", typ: "entity", question: "Which hotel did we book in Lisbon?", ans: "Hotel Avenida Palace", specificCard: true},
	{at: 44, topic: 3, sentence: "The standup is at 9:15 every weekday.", typ: "superseded", question: "What time was the standup originally?", ans: "9:15"},
	{at: 520, topic: 3, sentence: "Move the standup to 10:30 from now on.", typ: "update", question: "What time is the standup now?", ans: "10:30", specificCard: true},
	{at: 300, topic: 3, sentence: "We decided the platform role needs a system design round with Marcus Oyelaran.", typ: "decision", question: "Who runs the system design round for the platform role?", ans: "Marcus Oyelaran"},
	{at: 66, topic: 4, sentence: "The helper that cleans SKUs is normalizeSkuCodes in catalog/sku.go.", typ: "code", question: "Where is the helper that cleans SKUs?", ans: "normalizeSkuCodes"},
	{at: 180, topic: 4, sentence: "Price rounding uses banker's rounding to two places, per the finance team.", typ: "decision", question: "How does the catalog round prices?", ans: "banker's rounding"},
	{at: 400, topic: 4, sentence: "The feed timeout is 45 seconds; raise it only with the vendor's sign-off.", typ: "config", question: "What is the feed timeout?", ans: "45 seconds"},
}

// buildProbeConversation writes n synthetic turns on five threads with
// the planted facts, a card per turn, and returns the probes.
func buildProbeConversation(t *testing.T, e *Engine, n int) []Probe {
	t.Helper()
	rng := rand.New(rand.NewPCG(7, 11))
	for _, tp := range probeTopics {
		if err := e.V.PutThread(vault.Thread{ID: tp.id, Title: tp.title, Status: "open", Created: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}
	plantAt := map[int]plantedFact{}
	for _, p := range planted {
		plantAt[p.at] = p
	}
	var probes []Probe
	for i := 1; i <= n; i++ {
		ti := rng.IntN(len(probeTopics))
		p, isPlant := plantAt[i]
		if isPlant {
			ti = p.topic
		}
		tp := probeTopics[ti]
		var sb strings.Builder
		for k := 0; k < 40; k++ {
			if isPlant && k == 20 {
				sb.WriteString(p.sentence + " ")
			}
			s := tp.sentences[rng.IntN(len(tp.sentences))]
			sb.WriteString(fmt.Sprintf(s, tp.words[rng.IntN(len(tp.words))]) + " ")
			if k%6 == 5 {
				sb.WriteString("\n\n")
			}
		}
		role := vault.RoleUser
		if i%2 == 0 {
			role = vault.RoleAssistant
		}
		tu, err := e.V.AppendTurn(role, "", "", "", sb.String())
		if err != nil {
			t.Fatal(err)
		}
		summary := tp.summary
		if isPlant && p.specificCard {
			summary = strings.TrimSuffix(tp.summary, ".") + "; " + strings.TrimSuffix(p.question, "?") + "."
		}
		if err := e.V.PutCard(vault.Card{Turn: tu.ID, BodySHA: tu.BodySHA, Summary: summary, Threads: []string{tp.id}, Keywords: tp.keywords}); err != nil {
			t.Fatal(err)
		}
		if isPlant {
			probes = append(probes, Probe{ID: fmt.Sprintf("p%03d", i), Type: p.typ, Question: p.question, Source: tu.ID, Answer: p.ans})
		}
	}
	if err := e.Reindex(); err != nil {
		t.Fatal(err)
	}
	return probes
}

// TestProbesRecoverPlantedFacts is the recovery measurement (LANTERN's
// method without the LLM judge) on a conversation about twice the largest
// window. Every probe's source must at least be named in the packet, and
// a recall search for the question must reach most of them.
func TestProbesRecoverPlantedFacts(t *testing.T) {
	h := newHarness(t, 200_000, nil)
	probes := buildProbeConversation(t, h.e, 800)
	turns, _ := h.e.V.Turns()
	total := 0
	for _, tu := range turns {
		total += len(tu.Body)
	}
	if tok := packet.Tokens(providers.Profile{}, total); tok < 2*260_000 {
		t.Fatalf("history is only %d tokens; the probe needs it past every window", tok)
	}
	// whole turns as the ranked unit first, as before atoms: the same
	// conversation with the value lane off, for comparison, then all lanes
	atomBudget := h.e.Cfg.RelevantTokens
	h.e.Cfg.RelevantTokens = 0
	packet.RankLanes = packet.Lanes{Exact: true, FullText: true}
	res, err := h.e.RunProbes(probes)
	packet.RankLanes = packet.Lanes{Exact: true, FullText: true, Value: true}
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("value lane off:\n%s", ProbeReport(res))
	start := time.Now()
	res, err = h.e.RunProbes(probes)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d probes over %d turns (%d tokens) in %s\n%s", len(res), len(turns), packet.Tokens(providers.Profile{}, total), time.Since(start).Round(time.Millisecond), ProbeReport(res))
	searched := 0
	for _, r := range res {
		if r.Where == WhereInvalid || r.Where == WhereMissing {
			t.Errorf("probe %s: source %s is %s", r.ID, r.Source, r.Where)
		}
		if r.Search {
			searched++
		}
	}
	if searched*10 < len(res)*8 {
		t.Errorf("a recall search for the question reaches only %d of %d sources", searched, len(res))
	}

	// atoms as the ranked unit (the default). A search counts only when
	// the paragraphs it returns hold the answer.
	h.e.Cfg.RelevantTokens = atomBudget
	res, err = h.e.RunProbes(probes)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("atoms, budget %d tokens:\n%s", atomBudget, ProbeReport(res))
	verbatim, searched := 0, 0
	for _, r := range res {
		if r.Where == WhereInvalid || r.Where == WhereMissing {
			t.Errorf("atoms: probe %s: source %s is %s", r.ID, r.Source, r.Where)
		}
		if r.Where == WhereVerbatim {
			verbatim++
		}
		if r.Search {
			searched++
		}
	}
	if verbatim < 12 || searched < 12 {
		t.Errorf("atoms: %d of %d sources verbatim (floor 12), %d found by search (floor 12)", verbatim, len(res), searched)
	}
}
