package engine

import (
	"fmt"
	"os"
	"strings"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// DemoResult holds what the model-free demo produced: the stored turns, an
// explain plan with and without a dropped turn, receipts before and after
// a turn that changes their answer (one live, one read at an earlier turn
// number), and a typed recall.
type DemoResult struct {
	TurnsCreated      int
	ExplainPlan       packet.ExplainPlan
	AblatedPlan       packet.ExplainPlan
	LiveReceiptBefore store.Receipt
	LiveReceiptAfter  store.Receipt
	HistoricalReceipt store.Receipt
	StaleAlerts       []StaleAlert
	LiveDiff          ReceiptDiff
	HistoricalDiff    ReceiptDiff
	OpenBlockNotes    []string
	OpenBlockIDs      []string
	ExchangeNotes     []string
	GrepNotes         []string
	TimelineNotes     []string
	TimelineText      string
	KeyHistoryRows    []store.KeyStateRow
	KeyHistoryAt6     []store.KeyStateRow
}

// RunDemo runs the model-free demo inside a temporary
// directory (cleaned up automatically on return).
func RunDemo() (DemoResult, error) {
	var res DemoResult
	tmp, err := os.MkdirTemp("", "cx-demo-*")
	if err != nil {
		return res, err
	}
	defer os.RemoveAll(tmp)

	cfg := config.Default()
	cfg.Root = tmp
	cfg.Sandbox = false
	cfg.WholeHistoryTokens = -1 // force atom selection so Explain shows ranking lanes even on 6 turns
	cfg.AdaptivePack = true

	e, err := Create(cfg, "demo", "ContextlessAI demo", "Walk through receipts, their invalidation, explain and typed recall")
	if err != nil {
		return res, err
	}
	defer e.Close()

	turns := []struct{ role, author, body string }{
		{vault.RoleUser, "user", "Let's finalize the payments service migration.\n\nChecklist:\n- Move primary Postgres to us-east-1\n- Set pool_size=20 for initial canary\n- Enable TLS 1.3 on port 5432\n- Verify failover drills before GA"},
		{vault.RoleAssistant, "demo", "Here is the initial Go configuration for the canary deployment. We document all fields explicitly so every downstream service reads the same region, connection pool limit, TLS mode, and port defaults during rollout.\n\n```go\ntype DBConfig struct {\n\tRegion   string\n\tPoolSize int\n\tPort     int\n}\n\nfunc DefaultCanary() DBConfig {\n\treturn DBConfig{Region: \"us-east-1\", PoolSize: 20, Port: 5432}\n}\n```\n\nWe will hold pool_size=20 until load tests complete and the latency report is signed off by SRE and Payments."},
		{vault.RoleUser, "user", "Has anyone approved the Q4 security budget yet?"},
		{vault.RoleAssistant, "demo", "No, nobody has approved the Q4 security budget yet. Finance review is still pending."},
		{vault.RoleUser, "user", "Also note Redis cache TTL is 300s in staging."},
		{vault.RoleAssistant, "demo", "Noted: staging Redis TTL is 300s while database canary stays at pool_size=20."},
	}
	var saved []vault.Turn
	for _, spec := range turns {
		t, err := e.V.AppendTurn(spec.role, spec.author, "", "", spec.body)
		if err != nil {
			return res, err
		}
		saved = append(saved, t)
	}
	// Seed a superseded decision + replacement so `timeline` demonstrates chronological evolution.
	d1 := vault.Decision{
		ID:        vault.NewID("dec"),
		Statement: "Canary database uses pool_size=20 in us-east-1",
		Status:    "active", Source: saved[0].ID, Quote: "Set pool_size=20 for initial canary",
	}
	_ = e.V.PutDecision(d1)
	_ = e.Reindex()
	res.TurnsCreated = len(saved)

	// 1. Explain, with and without a dropped turn
	q := "What is the database pool_size for the canary?"
	plan, err := e.Explain(q, nil)
	if err != nil {
		return res, err
	}
	res.ExplainPlan = plan
	ablated, err := e.Explain(q, []string{saved[0].ID})
	if err != nil {
		return res, err
	}
	res.AblatedPlan = ablated

	// 2. Bind receipts at snapshot #6:
	//    (a) Live watched receipt with verified anchor + absence predicate
	//    (b) Historical cutoff receipt pinned to seq=6
	atoms1 := packet.AtomsOf(saved[0])
	ev1 := store.ReceiptEvidence{
		Turn: saved[0].ID, BodySHA: saved[0].BodySHA,
		Start: atoms1[len(atoms1)-1].Start, End: atoms1[len(atoms1)-1].End,
		Seg: vault.AtomSeg, AtomRef: atoms1[len(atoms1)-1].ID(),
		Quote: "Set pool_size=20 for initial canary",
	}
	liveRcpt, err := e.CreateReceipt(
		"What is the canary pool_size and has anyone approved the Q4 security budget?",
		saved[3].ID,
		"Canary pool_size=20; nobody has approved the Q4 security budget yet.",
		0, []store.ReceiptEvidence{ev1},
		[]store.ReceiptPredicate{{Kind: "absence", Key: "security budget approved", Value: "no approval at seq 6"}},
		"Waiting on Q4 security budget sign-off from Finance", "",
	)
	if err != nil {
		return res, err
	}
	res.LiveReceiptBefore = liveRcpt

	histRcpt, err := e.CreateReceipt(
		"As of turn 6, what was the canary database pool_size?",
		saved[1].ID,
		"At turn 6, the canary database pool_size was 20.",
		6, []store.ReceiptEvidence{ev1}, nil, "", "",
	)
	if err != nil {
		return res, err
	}
	res.HistoricalReceipt = histRcpt

	// 3. New contradicting evidence arrives (Turn 7) + decision supersedes d1!
	t7, err := e.V.AppendTurn(vault.RoleUser, "user", "", "",
		"Update: Finance approved the Q4 security budget ($140k) today, and we raised database pool_size=64 after load testing.")
	if err != nil {
		return res, err
	}
	d1.Status = "superseded"
	_ = e.V.PutDecision(d1)
	_ = e.V.PutDecision(vault.Decision{
		ID:        vault.NewID("dec"),
		Statement: "Canary database raised to pool_size=64 after load testing",
		Status:    "active", Source: t7.ID, Quote: "raised database pool_size=64 after load testing",
		Supersedes: []string{d1.ID},
	})
	_ = e.Reindex()
	res.TurnsCreated = 7

	alerts, err := e.CheckStaleReceipts(t7, nil, []string{d1.ID})
	if err != nil {
		return res, err
	}
	res.StaleAlerts = alerts
	res.LiveReceiptAfter, _, _ = e.S.GetReceipt(liveRcpt.ID)
	res.HistoricalReceipt, _, _ = e.S.GetReceipt(histRcpt.ID)

	// 4. Deterministic verification & delta diff (zero model calls)
	res.LiveDiff, err = e.DiffReceiptSince(liveRcpt.ID)
	if err != nil {
		return res, err
	}
	res.HistoricalDiff, err = e.DiffReceiptSince(histRcpt.ID)
	if err != nil {
		return res, err
	}

	// 5. Typed Recall Verbs (open_block, exchange, grep, timeline)
	focus := vault.Turn{Seq: 8}
	// Find an atom inside the fenced Go code block in turn 2
	codeAtom := vault.AtomID(saved[1].ID, 2)
	obIDs, _, obNotes := e.ExecuteRecallOps([]packet.RecallOp{{Op: "open_block", Target: codeAtom}}, focus, nil)
	res.OpenBlockIDs, res.OpenBlockNotes = obIDs, obNotes
	_, _, res.ExchangeNotes = e.ExecuteRecallOps([]packet.RecallOp{{Op: "exchange", Target: saved[2].ID, Radius: 1}}, focus, nil)
	_, _, res.GrepNotes = e.ExecuteRecallOps([]packet.RecallOp{{Op: "grep", Target: "pool_size="}}, focus, nil)
	_, tlExtras, tlNotes := e.ExecuteRecallOps([]packet.RecallOp{{Op: "timeline", Target: "pool_size"}}, focus, nil)
	res.TimelineNotes = tlNotes
	if len(tlExtras) > 0 {
		res.TimelineText = tlExtras[0].Text
	}
	res.KeyHistoryRows, _ = e.S.QueryKeyHistory("pool_size", 0)
	res.KeyHistoryAt6, _ = e.S.QueryKeyHistory("pool_size", 6)

	return res, nil
}

// FormatDemo renders a DemoResult as text.
func FormatDemo(r DemoResult) string {
	var b strings.Builder
	b.WriteString("=== ContextlessAI demo (no model) ===\n\n")
	fmt.Fprintf(&b, "1. Explain a packet (cx explain)\n")
	fmt.Fprintf(&b, "   Query:   %q (class: %s, adaptive=%v)\n", r.ExplainPlan.Query, r.ExplainPlan.Class, r.ExplainPlan.AdaptivePack)
	fmt.Fprintf(&b, "   Budget:  %d / %d tokens (brief %d tok, cache prefix %d%%)\n",
		r.ExplainPlan.Budget.Planned, r.ExplainPlan.Budget.Cap, r.ExplainPlan.Budget.Brief, r.ExplainPlan.CacheSharePct)
	if r.ExplainPlan.SubmodularCert != nil {
		fmt.Fprintf(&b, "   Coverage score: %.1f%% of an upper bound on a stand-in objective (utility %.3f / bound %.3f)\n",
			r.ExplainPlan.SubmodularCert.OptimalityPct, r.ExplainPlan.SubmodularCert.Utility, r.ExplainPlan.SubmodularCert.UpperBoundOPT)
	}
	fmt.Fprintf(&b, "   Lanes:   exact=%d full_text=%d value=%d working_set=%d\n",
		r.ExplainPlan.LaneCounts["exact"], r.ExplainPlan.LaneCounts["full_text"], r.ExplainPlan.LaneCounts["value"], r.ExplainPlan.LaneCounts["working_set"])
	fmt.Fprintf(&b, "   With -drop t000001 (removed after selection): omitted count %d -> %d, brief %d -> %d tok\n\n",
		r.ExplainPlan.OmittedCount, r.AblatedPlan.OmittedCount, r.ExplainPlan.Budget.Brief, r.AblatedPlan.Budget.Brief)

	fmt.Fprintf(&b, "2. Receipts and their invalidation (snapshot #6 -> #7)\n")
	fmt.Fprintf(&b, "   Live receipt %s:\n", r.LiveReceiptAfter.ID)
	fmt.Fprintf(&b, "     Before t000007: status=%s (snapshot #%d root %s)\n", r.LiveReceiptBefore.Status, r.LiveReceiptBefore.SnapshotSeq, short(r.LiveReceiptBefore.SnapshotRoot))
	fmt.Fprintf(&b, "     After  t000007: status=%s · trigger=%s\n", r.LiveReceiptAfter.Status, r.LiveReceiptAfter.StaleReason)
	fmt.Fprintf(&b, "     Original quotes SHA-verified=%v · snapshot_match=%v · delta turns=%d\n",
		r.LiveDiff.OriginalVerified, r.LiveDiff.SnapshotMatch, len(r.LiveDiff.DeltaTurns))
	for _, dt := range r.LiveDiff.DeltaTurns {
		fmt.Fprintf(&b, "     Delta %s: %s\n", dt.ID, oneLine(dt.Body, 110))
	}
	fmt.Fprintf(&b, "   Historical receipt %s (pinned as-of seq 6):\n", r.HistoricalReceipt.ID)
	fmt.Fprintf(&b, "     After  t000007: status=%s (immune to later turns) · original_verified=%v · snapshot_match=%v\n\n",
		r.HistoricalReceipt.Status, r.HistoricalDiff.OriginalVerified, r.HistoricalDiff.SnapshotMatch)

	fmt.Fprintf(&b, "3. Typed Recall Verbs (Zero-LLM)\n")
	for _, line := range r.OpenBlockNotes {
		fmt.Fprintf(&b, "   %s\n", line)
	}
	for _, line := range r.ExchangeNotes {
		fmt.Fprintf(&b, "   %s\n", line)
	}
	for _, line := range r.GrepNotes {
		fmt.Fprintf(&b, "   %s\n", line)
	}
	for _, line := range r.TimelineNotes {
		fmt.Fprintf(&b, "   %s\n", line)
	}
	if r.TimelineText != "" {
		b.WriteString("\n" + r.TimelineText + "\n")
	}
	if len(r.KeyHistoryRows) > 0 {
		b.WriteString("\n4. Key-value view over the turn sequence (experimental; pool_size)\n")
		for _, kv := range r.KeyHistoryRows {
			state := "ACTIVE"
			span := fmt.Sprintf("t%06d..current", kv.ValidFromSeq)
			if kv.ValidUntilSeq > 0 {
				state = "SUPERSEDED"
				span = fmt.Sprintf("t%06d..t%06d", kv.ValidFromSeq, kv.ValidUntilSeq)
			}
			fmt.Fprintf(&b, "   [%s] %s = %s (%s)\n", state, kv.Key, kv.RawValue, span)
		}
		if len(r.KeyHistoryAt6) > 0 {
			last := r.KeyHistoryAt6[len(r.KeyHistoryAt6)-1]
			fmt.Fprintf(&b, "   Time-travel query @ seq 6: %s = %s (superseded_by=%d)\n", last.Key, last.RawValue, last.SupersededBySeq)
		}
	}
	return b.String()
}
