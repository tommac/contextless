package engine

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

// Probes measure the selection layer without a model, after LANTERN's
// recovery probes: each probe is a question whose answer is stated in a
// known earlier turn. Instead of an LLM judge, recovery is read off the
// packet the engine would build for that question: is the source turn in
// it verbatim, named by its own catalog line, only inside an id range, or
// missing? And would a recall search for the question find it?

// Probe is one question with a known source turn.
type Probe struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Question string `json:"question"`
	Source   string `json:"source"`
	// Answer must occur in the source turn (case and whitespace aside).
	Answer string `json:"answer"`
	// Also lists other turns that state the same fact. In a real
	// conversation a fact is often repeated; any of them in the packet
	// verbatim answers the question.
	Also []string `json:"also,omitempty"`
	// Answers are other exact wordings of the fact. An Also turn counts
	// when it holds Answer or one of these; the source must hold Answer.
	Answers []string `json:"answers,omitempty"`
}

func (p Probe) accepted() []string { return append([]string{p.Answer}, p.Answers...) }

// Where a probe's source turn sits in the packet.
const (
	WhereVerbatim = "verbatim" // the turn (or the part holding the answer) is in the packet
	WhereCard     = "card"     // named by its own catalog or omitted line
	WhereRange    = "range"    // named only inside an id range
	WhereMissing  = "missing"  // not named at all (a coverage bug)
	WhereInvalid  = "invalid"  // the probe's answer is not in its source turn
)

// ProbeResult is one probe's outcome.
type ProbeResult struct {
	Probe
	Where string `json:"where"`
	// Search is true when the top five of a recall search for the
	// question put back text of the source that holds the answer.
	Search bool `json:"search"`
	// Any is true when the source or one of the Also turns is in the
	// packet verbatim with the answer in it.
	Any     bool   `json:"any"`
	Profile string `json:"profile"`
	Planned int    `json:"planned"`
	// Brief is the packet's brief in tokens, Ranked the part of it that
	// was ranked for the question (turns or excerpts), and Whole says the
	// whole history was sent, so nothing was ranked.
	Brief  int  `json:"brief"`
	Ranked int  `json:"ranked"`
	Whole  bool `json:"whole"`
	// BuildMs is how long the packet took to build. TokensToAnswer is the
	// ranked text read, in rank order, before the answer is reached (0 if
	// it never is). AlsoBad are Also turns holding no accepted answer,
	// which can never be credited: the probe file needs fixing.
	BuildMs        int64    `json:"build_ms"`
	TokensToAnswer int      `json:"tokens_to_answer"`
	AlsoBad        []string `json:"also_bad,omitempty"`
	// Rank is where the source sits in each ranking lane and in the fused
	// ranking of every earlier turn, to tell why a probe missed.
	Rank packet.Positions `json:"rank"`
}

// DryPacket builds the respond packet for msg as if it were the next user
// message, without appending it or running a model.
func (e *Engine) DryPacket(msg string) (packet.Sized, error) {
	focus := e.nextTurn(msg)
	cands, _ := e.R.Candidates(config.RoleResponder, nil)
	task := packet.Task{RunID: "run-dryrun", Kind: packet.KindRespond, Role: config.RoleResponder,
		Instructions: "(respond instructions)", Schema: packet.RespondSchema(true)}
	return e.buildPacket(task, focus, packet.SelectOptions{}, cands, packet.OutputTokens[packet.KindRespond])
}

// DryPull is DryPacket followed by one recall request: the packet for msg,
// then the packet the next run would get after asking for ids (and
// queries). No model runs and nothing is stored, so a logged request can
// be replayed to see exactly what the live path would have put in front
// of a model. added are the ids put back; notes say what was refused.
func (e *Engine) DryPull(msg string, ids, queries []string) (first, second packet.Sized, added, notes []string, err error) {
	focus := e.nextTurn(msg)
	cands, _ := e.R.Candidates(config.RoleResponder, nil)
	task := packet.Task{RunID: "run-dryrun", Kind: packet.KindRespond, Role: config.RoleResponder,
		Instructions: "(respond instructions)", Schema: packet.RespondSchema(true)}
	out := packet.OutputTokens[packet.KindRespond]
	if first, err = e.buildPacket(task, focus, packet.SelectOptions{}, cands, out); err != nil {
		return first, second, nil, nil, err
	}
	added, notes = e.recallRequest(ids, queries, focus, nil)
	second, err = e.buildPacket(task, focus, packet.SelectOptions{Pinned: added, Extra: recallNotes(notes)}, cands, out)
	return first, second, added, notes, err
}

// compileOptions are the ranking and layout options of the configuration.
func (e *Engine) compileOptions() packet.CompileOptions {
	copt := packet.DefaultCompileOptions()
	copt.AdaptivePack = e.Cfg.AdaptivePack
	copt.SessionRoute = e.Cfg.SessionRoute
	copt.AutoReplyWindow = e.Cfg.AutoReplyWindow
	if len(e.Cfg.LaneWeights) > 0 {
		if w, ok := e.Cfg.LaneWeights["exact"]; ok && w > 0 {
			copt.Weights.Exact = w
		}
		if w, ok := e.Cfg.LaneWeights["full_text"]; ok && w > 0 {
			copt.Weights.FullText = w
		}
		if w, ok := e.Cfg.LaneWeights["value"]; ok && w > 0 {
			copt.Weights.Value = w
		}
		if w, ok := e.Cfg.LaneWeights["working_set"]; ok && w > 0 {
			copt.Weights.WorkingSet = w
		}
	}
	if e.S != nil && e.Cfg.WorkingSetWindow > 0 {
		if ws, err := e.S.WorkingSetAtoms(e.Cfg.WorkingSetWindow, 8); err == nil {
			copt.WorkingSet = ws
		}
	}
	return copt
}

// indexStatus inspects snapshot coverage and SQLite search lane health up to
// targetSeq (typically focus.Seq-1) without reading turn bodies.
func (e *Engine) indexStatus(targetSeq int, query string, copt packet.CompileOptions) packet.IndexStatus {
	st := packet.IndexStatus{SnapshotSeq: targetSeq, Complete: true}
	if targetSeq <= 0 {
		return st
	}
	if e.S == nil {
		st.Complete = false
		st.LaneErrors = append(st.LaneErrors, "sqlite store unavailable")
		return st
	}
	seq, root, err := e.S.SnapshotRoot(targetSeq)
	if err != nil {
		st.Complete = false
		st.LaneErrors = append(st.LaneErrors, "snapshot_root: "+err.Error())
		return st
	}
	st.IndexedUpTo = seq
	st.SnapshotRoot = root
	if seq < targetSeq {
		st.Complete = false
	}
	if strings.TrimSpace(query) != "" {
		_, laneErrs := packet.CandidateTurnIDs(e.S, query, targetSeq+1, 16, copt)
		if len(laneErrs) > 0 {
			st.Complete = false
			st.LaneErrors = append(st.LaneErrors, laneErrs...)
		}
	}
	return st
}

// buildPacket selects and sizes one packet. opt carries the call's own
// records (pins, extras, a focus item); the history options come from the
// configuration.
//
// While the earlier turns add up to at most WholeHistoryTokens, all of
// them are tried first on the role's first profile: on real conversations
// that fit, the whole transcript held every probe's source, more than any
// selection (docs/ARCHITECTURE.md 3c; answer quality is not yet measured). If it does not fit there with nothing left
// out, the packet is selected and sized as usual. The whole-history
// attempt itself never moves a call to a larger provider; a selected
// packet is still sized by the usual rule, which tries a larger window
// before it leaves a record out. Plan units and their assembler always
// get a selected packet: a plan would otherwise send the whole transcript
// once per unit.
func (e *Engine) buildPacket(task packet.Task, focus vault.Turn, opt packet.SelectOptions, cands []providers.Profile, output int) (packet.Sized, error) {
	opt.Recent, opt.Relevant, opt.RelevantTokens = e.Cfg.RecentTurns, e.Cfg.RelevantTurns, e.Cfg.RelevantTokens
	opt.RecentEpoch, opt.NoRanked = e.Cfg.RecentEpoch, e.Cfg.NoRanked
	if opt.Compile == nil {
		copt := e.compileOptions()
		opt.Compile = &copt
	}
	if !opt.NoHistory {
		opt.MapTokens, opt.PullTokens = e.atlasMapTokens(cands), e.pullTokens()
	}
	wholeKind := task.Kind != packet.KindUnit && task.Kind != packet.KindAssemble
	if e.Cfg.FullHistory && !opt.NoHistory && wholeKind {
		return e.fullHistoryPacket(task, focus, opt, cands, output)
	}
	if !opt.NoHistory && wholeKind && e.Cfg.WholeHistoryTokens > 0 && len(cands) > 0 {
		if n, ok := e.historyWithin(focus, cands[0], e.Cfg.WholeHistoryTokens); ok && n > opt.Recent {
			whole := opt
			whole.AllTurns = true
			whole.Extra = append(append([]packet.Item{}, opt.Extra...), wholeHistoryNote())
			if items, err := packet.Select(e.V, e.S, focus, whole); err == nil {
				if sized, err := packet.Size(task, items, cands[:1], output); err == nil && len(sized.Dropped) == 0 {
					sized.WholeHistory = true
					return sized, nil
				}
			}
		}
	}
	items, err := packet.Select(e.V, e.S, focus, opt)
	if err != nil {
		return packet.Sized{}, err
	}
	sized, err := packet.Size(task, items, cands, output)
	if err != nil {
		return sized, err
	}
	// Attach snapshot coverage & degraded-lane integrity metadata so an index
	// lag or SQLite lane error is surfaced transparently instead of looking
	// like absence in the archive.
	needRender := false
	if !opt.NoHistory && focus.Seq > 1 {
		idxSt := e.indexStatus(focus.Seq-1, focus.Body, *opt.Compile)
		sized.Index = idxSt
		if len(idxSt.LaneErrors) > 0 {
			task.IndexWarning = fmt.Sprintf("Search index degraded at snapshot #%d (%s): %s. Empty search results do not prove absence.",
				idxSt.SnapshotSeq, short(idxSt.SnapshotRoot), strings.Join(idxSt.LaneErrors, "; "))
			needRender = true
		}
	} else {
		sized.Index = packet.IndexStatus{Complete: true}
	}
	// External paged omission catalog: when omitted Tier-1 catalog turns exceed
	// threshold, offload membership to a content-addressed .cxi file and emit
	// a ~215-byte snapshot-pinned handle in brief.md instead of huge digest lists.
	if thresh := e.Cfg.OmittedCatalogThreshold; thresh > 0 && e.S != nil && !e.S.ReadOnly {
		omittedSeqs := map[int]bool{}
		maxSeq := 0
		for _, it := range sized.Dropped {
			if it.Tier == 1 && it.Seq > 0 {
				omittedSeqs[it.Seq] = true
				if it.Seq > maxSeq {
					maxSeq = it.Seq
				}
			}
		}
		if len(omittedSeqs) >= thresh {
			if lastSeq, root, rerr := e.S.SnapshotRoot(focus.Seq - 1); rerr == nil && lastSeq >= maxSeq {
				catDir := e.V.Dir + "/.cx/catalogs"
				if h, herr := packet.BuildOmittedCatalog(catDir, root, lastSeq, func(seq int) bool { return omittedSeqs[seq] }); herr == nil {
					_ = os.WriteFile(catDir+"/"+h.Index+".json", []byte(h.JSON()+"\n"), 0o644)
					task.OmittedHandle = h.JSON()
					sized.OmittedHandle = task.OmittedHandle
					needRender = true
				}
			}
		}
	}
	if needRender {
		task.NativeSchema = sized.Profile.NativeSchema
		sized.Brief = packet.Render(task, sized.Items, sized.Dropped)
		sized.Budget.Brief = packet.Tokens(sized.Profile, len(sized.Brief))
		sized.Budget.Planned = sized.Profile.SysOverhead + sized.Budget.Brief + sized.Budget.Output
	}
	return sized, nil
}

// wholeHistoryNote tells a model its packet holds the whole conversation.
func wholeHistoryNote() packet.Item {
	return packet.Item{Key: packet.KeyExtra, ID: "whole-history", Seq: 0, Header: "Coverage of this packet",
		Text: "Every earlier turn of the conversation is in this packet in full, in order. There is no catalog and nothing to recall: what is not in a record below was not said."}
}

// fullHistoryPacket is the full-history arm: every earlier turn, on the
// first profile that holds it, or the largest with the oldest left out.
func (e *Engine) fullHistoryPacket(task packet.Task, focus vault.Turn, opt packet.SelectOptions, cands []providers.Profile, output int) (packet.Sized, error) {
	opt.AllTurns = true
	noted := opt
	noted.Extra = append(append([]packet.Item{}, opt.Extra...), wholeHistoryNote())
	items, err := packet.Select(e.V, e.S, focus, noted)
	if err != nil {
		return packet.Sized{}, err
	}
	sized, err := packet.Size(task, items, cands, output)
	if err == nil && len(sized.Dropped) == 0 {
		sized.WholeHistory = true
		return sized, nil
	}
	if items, err = packet.Select(e.V, e.S, focus, opt); err != nil {
		return packet.Sized{}, err
	}
	return packet.Size(task, items, cands, output)
}

// historyWithin counts the turns before focus and reports whether they
// add up to at most limit tokens for profile p. Uses SQLite turn_meta when
// up to date (zero turn-file reads), falling back to reading backward from
// the newest turn until the limit is passed.
func (e *Engine) historyWithin(focus vault.Turn, p providers.Profile, limit int) (n int, ok bool) {
	if e.S != nil && focus.Seq > 1 {
		if metas, err := e.S.TurnMetaBefore(focus.Seq); err == nil && len(metas) == focus.Seq-1 {
			bytes := 0
			for i := len(metas) - 1; i >= 0; i-- {
				n++
				bytes += metas[i].Bytes + 64
				if packet.Tokens(p, bytes) > limit {
					return n, false
				}
			}
			return n, true
		}
	}
	bytes := 0
	for seq := focus.Seq - 1; seq >= 1; seq-- {
		t, err := e.V.Turn(vault.TurnID(seq))
		if err != nil {
			return 0, false
		}
		n++
		bytes += len(t.Body) + 64
		if packet.Tokens(p, bytes) > limit {
			return n, false
		}
	}
	return n, true
}

func (e *Engine) nextTurn(msg string) vault.Turn {
	last, err := e.V.LastSeq()
	seq := 1
	if err == nil && last > 0 {
		seq = last + 1
	}
	return vault.Turn{ID: vault.TurnID(seq), Seq: seq, Role: vault.RoleUser, Created: time.Now(), Body: msg, BodySHA: vault.SHA(msg)}
}

func norm(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

// RunProbes places each probe's source turn in its dry-run packet.
func (e *Engine) RunProbes(probes []Probe) ([]ProbeResult, error) {
	var out []ProbeResult
	next := e.nextTurn("").Seq
	turns, err := e.V.Turns()
	if err != nil {
		return nil, err
	}
	for _, p := range probes {
		r := ProbeResult{Probe: p, Where: WhereInvalid}
		src, err := e.V.Turn(p.Source)
		if err != nil || p.Answer == "" || !strings.Contains(norm(src.Body), norm(p.Answer)) {
			out = append(out, r)
			continue
		}
		start := time.Now()
		sized, err := e.DryPacket(p.Question)
		r.BuildMs = time.Since(start).Milliseconds()
		if err != nil && sized.Profile.ID == "" {
			return nil, err
		}
		r.Profile, r.Planned = sized.Profile.ID, sized.Budget.Planned
		r.Brief, r.Whole = sized.Budget.Brief, sized.WholeHistory
		for _, it := range sized.Items {
			if it.Key == packet.KeyRelevant {
				r.Ranked += it.Tokens
			}
		}
		r.Where = placeSource(sized, p.Source, []string{p.Answer})
		r.Any = r.Where == WhereVerbatim
		for _, id := range p.Also {
			if placeSource(sized, id, p.accepted()) == WhereVerbatim {
				r.Any = true
			}
			if t, err := e.V.Turn(id); err != nil || !holdsAny(t.Body, p.accepted()) {
				r.AlsoBad = append(r.AlsoBad, id)
			}
		}
		if !e.Cfg.NoRanked && !e.Cfg.FullHistory {
			r.TokensToAnswer = packet.TokensToAnswer(turns, e.S, p.Question, next, p.Source, norm(p.Answer), e.Cfg.RelevantTokens > 0)
		}
		if e.Cfg.RelevantTokens > 0 {
			r.Rank = packet.RankAtomPositions(turns, e.S, p.Question, next, p.Source, norm(p.Answer))
		} else {
			r.Rank = packet.RankPositions(turns, e.S, p.Question, next, p.Source)
		}
		for _, id := range e.searchTurns(turns, p.Question, next, 5) {
			if ref, ok := vault.ParseRef(id); ok && ref.Turn == p.Source && strings.Contains(norm(refText(src, ref)), norm(p.Answer)) {
				r.Search = true
			}
		}
		out = append(out, r)
	}
	return out, nil
}

// refText is the text a recall id of turn t puts back.
func refText(t vault.Turn, ref vault.Ref) string {
	switch {
	case ref.Atom > 0:
		txt, _ := vault.AtomsText(t.ID, t.Body, ref.Atom, ref.AtomEnd)
		return txt
	case ref.Part > 0:
		if parts := vault.Parts(t.Body); ref.Part <= len(parts) {
			return parts[ref.Part-1]
		}
		return ""
	}
	return t.Body
}

func holdsAny(text string, answers []string) bool {
	text = norm(text)
	for _, a := range answers {
		if a = norm(a); a != "" && strings.Contains(text, a) {
			return true
		}
	}
	return false
}

func placeSource(sized packet.Sized, source string, answers []string) string {
	turnKeys := map[string]bool{packet.KeyRecall: true, packet.KeyRelevant: true, packet.KeyThread: true, packet.KeyRecent: true}
	where := WhereMissing
	for _, it := range sized.Items {
		t, _, _ := vault.ParseRecallID(it.ID)
		if t != source {
			continue
		}
		if turnKeys[it.Key] && holdsAny(it.Text, answers) {
			return WhereVerbatim
		}
		if it.Key == packet.KeyCatalog {
			where = WhereCard
		}
	}
	if where != WhereMissing {
		return where
	}
	for _, it := range sized.Dropped {
		if t, _, _ := vault.ParseRecallID(it.ID); t == source {
			if it.Tier == 1 {
				return WhereRange
			}
			return WhereCard
		}
	}
	return WhereMissing
}

// ProbeReport renders per-type recovery rates and lists the probes whose
// source is not in the packet verbatim.
func ProbeReport(results []ProbeResult) string {
	type agg struct{ n, verbatim, card, rng, missing, invalid, search, any int }
	by := map[string]*agg{}
	all := &agg{}
	for _, r := range results {
		a := by[r.Type]
		if a == nil {
			a = &agg{}
			by[r.Type] = a
		}
		for _, x := range []*agg{a, all} {
			x.n++
			switch r.Where {
			case WhereVerbatim:
				x.verbatim++
			case WhereCard:
				x.card++
			case WhereRange:
				x.rng++
			case WhereMissing:
				x.missing++
			default:
				x.invalid++
			}
			if r.Search {
				x.search++
			}
			if r.Any {
				x.any++
			}
		}
	}
	pct := func(k, n int) string {
		if n == 0 {
			return "-"
		}
		return fmt.Sprintf("%d%%", (100*k+n/2)/n)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%-12s %4s %9s %6s %6s %8s %8s %9s %10s\n", "type", "n", "verbatim", "card", "range", "missing", "invalid", "search@5", "any-turn")
	types := make([]string, 0, len(by))
	for t := range by {
		types = append(types, t)
	}
	sort.Strings(types)
	row := func(name string, a *agg) {
		valid := a.n - a.invalid
		fmt.Fprintf(&b, "%-12s %4d %9s %6s %6s %8s %8d %9s %10s\n", name, a.n, pct(a.verbatim, valid), pct(a.card, valid),
			pct(a.rng, valid), pct(a.missing, valid), a.invalid, pct(a.search, valid), pct(a.any, valid))
	}
	for _, t := range types {
		row(t, by[t])
	}
	row("all", all)
	var miss []string
	for _, r := range results {
		if r.Where != WhereVerbatim {
			s := "found by search"
			if !r.Search {
				s = "not found by search"
			}
			if r.Any {
				s += "; another turn stating it is verbatim"
			}
			pos := func(n int) string {
				if n == 0 {
					return "-"
				}
				return fmt.Sprint(n)
			}
			s += fmt.Sprintf("; ranked %s (exact %s, full text %s, value %s)", pos(r.Rank.Fused), pos(r.Rank.Exact),
				pos(r.Rank.FullText), pos(r.Rank.Value))
			miss = append(miss, fmt.Sprintf("- %s (%s) source %s: %s; %s", r.ID, r.Type, r.Source, r.Where, s))
		}
	}
	if len(miss) > 0 {
		b.WriteString("\nnot verbatim in the packet:\n" + strings.Join(miss, "\n") + "\n")
	}
	badAlso := 0
	for _, r := range results {
		badAlso += len(r.AlsoBad)
	}
	if badAlso > 0 {
		fmt.Fprintf(&b, "\n%d also entries name a turn holding no accepted answer, so they can never count: add the turn's wording under \"answers\" or remove them\n", badAlso)
	}
	var briefs, ranked []int
	whole := 0
	for _, r := range results {
		if r.Where == WhereInvalid {
			continue
		}
		briefs = append(briefs, r.Brief)
		ranked = append(ranked, r.Ranked)
		if r.Whole {
			whole++
		}
	}
	if len(briefs) > 0 {
		fmt.Fprintf(&b, "\npackets: median brief %d tokens, of which ranked for the question %d", median(briefs), median(ranked))
		if whole > 0 {
			fmt.Fprintf(&b, "; the whole history was sent for %d of %d", whole, len(briefs))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func median(xs []int) int {
	if len(xs) == 0 {
		return 0
	}
	s := append([]int(nil), xs...)
	sort.Ints(s)
	return s[len(s)/2]
}

// ProbeMode says how the ranked section of a packet is built under the
// engine's configuration, for the head of a probe report.
func (e *Engine) ProbeMode() string {
	mode := fmt.Sprintf("ranked unit: whole turns (%d)", e.Cfg.RelevantTurns)
	if e.Cfg.RelevantTokens > 0 {
		mode = fmt.Sprintf("ranked unit: atoms (budget %d tokens)", e.Cfg.RelevantTokens)
	}
	if e.Cfg.WholeHistoryTokens > 0 {
		mode += fmt.Sprintf("; whole history sent up to %d tokens", e.Cfg.WholeHistoryTokens)
	}
	return mode
}

// searchTurns finds earlier text for a recall query, best first, as
// recall ids (LCM's grep over the full store), with the same fused
// ranking that fills a packet's ranked section. With atoms on, it returns
// the best atoms, atomsPerSearchHit for each of limit (adjacent ones of a
// turn joined, a one-atom turn as the turn); otherwise the best turns, a turn longer than vault.PartBytes as
// the parts that match (ARC's chunked recall).
func (e *Engine) searchTurns(turns []vault.Turn, query string, beforeSeq, limit int) []string {
	if e.Cfg.RelevantTokens > 0 {
		return atomRunIDs(packet.RankAtoms(turns, e.S, query, beforeSeq, limit*atomsPerSearchHit))
	}
	q := norm(query)
	var out []string
	for _, r := range packet.RankTurns(turns, e.S, query, beforeSeq, limit) {
		out = append(out, matchingParts(r.Turn, q)...)
	}
	return out
}

// atomsPerSearchHit is how many paragraphs a recall search puts back for
// each turn it would have put back whole. A paragraph holds less than a
// turn: at one for one, the text a search returned held the answer for
// 108 of 150 real probes against 126 with whole turns; at four it is 127,
// still at a fraction of the tokens (docs/CONTEXT-COMPILER.md, 2b).
const atomsPerSearchHit = 4

// atomRunIDs turns ranked atoms into recall ids: the atoms of a turn are
// joined into runs of adjacent ones, and the runs come in the order of
// their best-ranked atom. A turn that matched only through its card, or
// whose every atom was found, comes back as the turn.
func atomRunIDs(ranked []packet.RankedAtom) []string {
	type pick struct {
		turn  vault.Turn
		ks    map[int]int // atom number to rank
		whole bool
	}
	var order []*pick
	by := map[string]*pick{}
	for i, r := range ranked {
		p := by[r.Turn.ID]
		if p == nil {
			p = &pick{turn: r.Turn, ks: map[int]int{}}
			by[r.Turn.ID] = p
			order = append(order, p)
		}
		p.whole = p.whole || r.CardOnly
		p.ks[r.Atom.K] = i
	}
	type run struct {
		id   string
		best int
	}
	var runs []run
	for _, p := range order {
		n := len(vault.Atoms(p.turn.ID, p.turn.Body))
		if p.whole {
			best := len(ranked)
			for _, i := range p.ks {
				best = min(best, i)
			}
			runs = append(runs, run{p.turn.ID, best})
			continue
		}
		for _, x := range packet.JoinAtoms(p.ks) {
			best := len(ranked)
			for k := x[0]; k <= x[1]; k++ {
				best = min(best, p.ks[k])
			}
			id := vault.AtomRangeID(p.turn.ID, x[0], x[1])
			if x[0] == 1 && x[1] == n {
				id = p.turn.ID
			}
			runs = append(runs, run{id, best})
		}
	}
	sort.SliceStable(runs, func(i, j int) bool { return runs[i].best < runs[j].best })
	out := make([]string, len(runs))
	for i, x := range runs {
		out[i] = x.id
	}
	return out
}

// matchingParts returns the whole turn when it has one part, else the
// parts that hold the query (or, failing an exact match, the two parts
// with the most query words).
func matchingParts(t vault.Turn, q string) []string {
	parts := vault.Parts(t.Body)
	if len(parts) == 1 {
		return []string{t.ID}
	}
	var hits []string
	for k, p := range parts {
		if strings.Contains(norm(p), q) {
			hits = append(hits, vault.PartID(t.ID, k+1))
		}
	}
	if len(hits) > 0 {
		if len(hits) > 3 {
			hits = hits[:3]
		}
		return hits
	}
	words := strings.Fields(q)
	type pc struct{ k, n int }
	var counts []pc
	for k, p := range parts {
		lp := norm(p)
		n := 0
		for _, w := range words {
			if len(w) >= 3 {
				n += strings.Count(lp, w)
			}
		}
		counts = append(counts, pc{k, n})
	}
	sort.SliceStable(counts, func(i, j int) bool { return counts[i].n > counts[j].n })
	for _, c := range counts[:min(2, len(counts))] {
		if c.n > 0 {
			hits = append(hits, vault.PartID(t.ID, c.k+1))
		}
	}
	if len(hits) == 0 {
		hits = []string{vault.PartID(t.ID, 1)}
	}
	return hits
}
