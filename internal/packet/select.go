// Package packet builds the context for one model call. It selects records
// from the vault by following edges one hop out from the focus turn, sizes
// them against a provider profile, and renders the brief. It never shortens
// a record: a record either goes in verbatim or is listed as omitted by id,
// and the model can ask for it back (needs_recall).
package packet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// Selector keys.
const (
	KeySession   = "session"
	KeyFocus     = "focus"
	KeyRecall    = "recall"
	KeyDecision  = "decision"
	KeyExtra     = "extra" // task-supplied records (unit outputs for an assembler)
	KeyRecent    = "recent"
	KeyThread    = "thread"
	KeyFact      = "fact"
	KeyEntity    = "entity"
	KeyRelevant  = "relevant"
	KeyCatalog   = "catalog"
	KeyThreadIdx = "thread_index"
)

// Item is one record in a packet. Text is always the record verbatim (for
// turns, the exact stored body). Card is the one line used in the catalog
// and in the omitted list.
type Item struct {
	Key  string `json:"key"`
	ID   string `json:"id"`
	Tier int    `json:"tier"` // 0 never dropped; 1 dropped first, 4 last
	Seq  int    `json:"seq,omitempty"`
	// Sub orders the pieces of one turn (a part or an excerpt) among
	// themselves: one more than the byte offset of the piece in the body.
	Sub    int     `json:"sub,omitempty"`
	Header string  `json:"header"`
	Text   string  `json:"-"`
	Card   string  `json:"card,omitempty"`
	SHA    string  `json:"sha,omitempty"`
	Bytes  int     `json:"bytes"`
	Tokens int     `json:"tokens"`
	Score  float64 `json:"score,omitempty"`
	// Threads are the item's threads, used to group omitted turns.
	Threads []string `json:"threads,omitempty"`
	// Group is the title of the item's first thread, and Keywords its
	// search words (the card's, else its most frequent terms): together
	// they label a range of omitted turns.
	Group    string   `json:"-"`
	Keywords []string `json:"-"`
}

// protect ranks how hard an item is to drop: tier 0 is never dropped, then
// tier 4 outlives 3, 2 and 1.
func protect(tier int) int {
	if tier == 0 {
		return 100
	}
	return tier
}

// SelectOptions tunes selection.
type SelectOptions struct {
	Recent int // the last N turns before the focus
	// RecentEpoch aligns the start of the recent window to multiples of
	// this many turns, so the window keeps its first turn for several
	// messages and grows, instead of sliding: a sliding window changes the
	// prompt at its first turn on every message. Zero or less slides.
	RecentEpoch int
	// NoRanked leaves the ranked section out: recency alone, the baseline
	// ranking has to beat.
	NoRanked bool
	// AllTurns sends every turn before the focus verbatim and ranks
	// nothing (a history short enough to send whole). Recent still says
	// whose card mentions bring in entities and facts.
	AllTurns bool
	Relevant int // ranked whole turns, when RelevantTokens is off
	// RelevantTokens is the budget for ranked atoms: paragraphs of earlier
	// turns chosen for the focus, sent as excerpts. Zero or less ranks
	// whole turns instead.
	RelevantTokens int
	Pinned         []string // turn ids the model asked for (needs_recall), tier 0
	Extra          []Item   // records the task supplies, tier 0
	// NoHistory sends only the session, decisions, focus and extras (an
	// indexer or a chunk unit that must not wander).
	NoHistory bool
	// IncludeTurns lists turns that must go in at tier 0 besides the focus
	// (the indexer's batch).
	IncludeTurns []string
	// FocusItem replaces the focus turn's item (one part of an oversized
	// message, or its opening for the assembler).
	FocusItem *Item
	// Compile carries per-call ranking and layout options (lanes, weights,
	// working set, adaptive packing). Nil snapshots DefaultCompileOptions().
	Compile *CompileOptions
	// MapTokens is the budget for the map of the conversation: one line
	// per checked idea note, naming the paragraphs it points at. Zero or
	// less sends no map. A packet that holds every turn has none either.
	MapTokens int
	// PullTokens is how many tokens of paragraphs one recall request
	// brings back, told to the reader of the map.
	PullTokens int
}

// Select builds the candidate items for a focus turn. Order of the result
// is not meaningful; Render orders by key and sequence.
func Select(v *vault.Vault, s *store.Store, focus vault.Turn, opt SelectOptions) ([]Item, error) {
	byID := map[string]Item{}
	add := func(it Item) {
		it.Bytes = len(it.Text) + len(it.Header) + 8
		if old, ok := byID[it.ID]; ok && protect(old.Tier) >= protect(it.Tier) {
			if it.Score > old.Score {
				old.Score = it.Score
				byID[it.ID] = old
			}
			return
		}
		byID[it.ID] = it
	}

	sess, err := v.Session()
	if err != nil {
		return nil, err
	}
	add(Item{Key: KeySession, ID: "session", Tier: 0, Header: "Session", Text: renderSession(sess)})

	copt := DefaultCompileOptions()
	if opt.Compile != nil {
		copt = *opt.Compile
	}

	// Lazy metadata-first loading: when SQLite turn_meta and value_postings
	// cover every turn before focus and we are building a selected atom packet,
	// avoid reading all N turn files from disk. Load only the turns that
	// actually participate in recent, pinned, thread, or ranked selection,
	// verifying each loaded file's SHA-256 on read.
	var (
		turns    []vault.Turn
		metas    []store.TurnMeta
		lazyMode bool
	)
	if s != nil && !opt.AllTurns && !opt.NoHistory && opt.RelevantTokens > 0 && s.HasValuePostings() && focus.Seq > 1 {
		if ms, merr := s.TurnMetaBefore(focus.Seq); merr == nil && len(ms) == focus.Seq-1 && ms[len(ms)-1].Seq == focus.Seq-1 {
			metas = ms
			lazyMode = true
		}
	}

	turnByID := map[string]vault.Turn{}
	metaByID := map[string]store.TurnMeta{}
	if lazyMode {
		for _, m := range metas {
			metaByID[m.ID] = m
		}
		cards := v.Cards()
		need := map[string]bool{}
		addRef := func(id string) {
			if ref, ok := vault.ParseRef(id); ok {
				if _, exists := metaByID[ref.Turn]; exists {
					need[ref.Turn] = true
				}
			}
		}
		for _, id := range opt.Pinned {
			addRef(id)
		}
		for _, id := range opt.IncludeTurns {
			addRef(id)
		}
		// Recent turns
		start := len(metas) - opt.Recent
		if opt.RecentEpoch > 0 && start > 0 {
			start -= start % opt.RecentEpoch
		}
		if start < 0 {
			start = 0
		}
		for i := start; i < len(metas); i++ {
			need[metas[i].ID] = true
		}
		// Threads from last 2 turns + thread FTS
		thSet := map[string]bool{}
		for i := len(metas) - 1; i >= 0 && i >= len(metas)-2; i-- {
			if c, ok := cards[metas[i].ID]; ok && c.BodySHA == metas[i].BodySHA {
				for _, th := range c.Threads {
					thSet[th] = true
				}
			}
		}
		if thHits, err := s.SearchE(focus.Body, 3, vault.TypeThread); err == nil {
			for _, h := range thHits {
				thSet[h.ID] = true
			}
		}
		for th := range thSet {
			for _, tid := range s.In(th, "threads") {
				if _, ok := metaByID[tid]; ok {
					need[tid] = true
				}
			}
		}
		// Ranked candidates from FTS, phrase prefilter, value postings, working set,
		// and optional episode/session routing.
		candLimit := max(2*max(atomCandidates, opt.RelevantTokens/40), 48)
		if !opt.NoRanked {
			candIDs, _ := CandidateTurnIDs(s, focus.Body, focus.Seq, candLimit, copt)
			for _, tid := range candIDs {
				if _, ok := metaByID[tid]; ok {
					need[tid] = true
				}
			}
			if copt.SessionRoute {
				epIDs := RouteEpisodeTurns(nil, metas, cards, focus.Body, focus.Seq, 3, 12)
				copt.EpisodeTurns = epIDs
				for _, tid := range epIDs {
					if _, ok := metaByID[tid]; ok {
						need[tid] = true
					}
				}
			}
		}
		// ReplyTo headers and optional immediate reply turns (AutoReplyWindow) of needed turns.
		initIDs := make([]string, 0, len(need))
		for id := range need {
			initIDs = append(initIDs, id)
		}
		for _, id := range initIDs {
			if m, ok := metaByID[id]; ok {
				if m.ReplyTo != "" {
					if _, exists := metaByID[m.ReplyTo]; exists {
						need[m.ReplyTo] = true
					}
				}
				if copt.AutoReplyWindow && m.Seq+1 < focus.Seq {
					nextID := vault.TurnID(m.Seq + 1)
					if _, exists := metaByID[nextID]; exists {
						need[nextID] = true
					}
				}
			}
		}
		ids := make([]string, 0, len(need))
		for id := range need {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		loaded, lerr := v.TurnsByIDs(ids)
		if lerr != nil {
			return nil, lerr
		}
		turns = make([]vault.Turn, 0, len(loaded))
		for _, id := range ids {
			t := loaded[id]
			turnByID[id] = t
			turns = append(turns, t)
		}
		sort.Slice(turns, func(i, j int) bool { return turns[i].Seq < turns[j].Seq })
	} else {
		turns, err = v.Turns()
		if err != nil {
			return nil, err
		}
		for _, t := range turns {
			turnByID[t.ID] = t
		}
	}
	// source names a ledger record's turn with its date, so "next
	// Tuesday" in a quote can be read against the day it was said
	source := func(id, quote string) string {
		at := id
		if t, ok := turnByID[id]; ok {
			at += " · " + t.Created.Format("2006-01-02")
		} else if m, ok := metaByID[id]; ok && !m.Created.IsZero() {
			at += " · " + m.Created.Format("2006-01-02")
		}
		if quote == "" {
			return "(source " + at + ")"
		}
		return fmt.Sprintf("(source %s: \"%s\")", at, quote)
	}

	decisions := v.Decisions()
	allDecisions := map[string]vault.Decision{}
	for _, d := range decisions {
		allDecisions[d.ID] = d
	}
	for _, d := range decisions {
		if d.Status != "" && d.Status != "active" {
			continue
		}
		txt := d.Statement + "\n" + source(d.Source, d.Quote)
		for _, old := range d.Supersedes {
			if od, ok := allDecisions[old]; ok {
				txt += fmt.Sprintf("\nReplaces %s: %s %s", od.ID, od.Statement, source(od.Source, ""))
			}
		}
		add(Item{Key: KeyDecision, ID: d.ID, Tier: 0, Header: d.ID, Text: txt, Card: d.Statement})
	}
	for _, x := range opt.Extra {
		x.Tier = 0
		if x.Key == "" {
			x.Key = KeyExtra
		}
		add(x)
	}

	// map: the checked idea notes whose evidence lies before the focus, as
	// pointers to paragraphs. Evidence counts only while its turn still has
	// the body the quote was found in.
	if opt.MapTokens > 0 && !opt.AllTurns && !opt.NoHistory && focus.Seq > 1 {
		if notes := v.Notes(); len(notes) > 0 {
			current := func(turn, sha string) bool {
				if m, ok := metaByID[turn]; ok {
					return m.BodySHA == sha
				}
				t, ok := turnByID[turn]
				return ok && t.BodySHA == sha
			}
			items, _ := BuildMap(notes, current, focus.Seq, focus.Body, opt.MapTokens, opt.PullTokens)
			for _, it := range items {
				add(it)
			}
		}
	}

	cards := v.Cards()
	threadTitle := map[string]string{}
	for _, th := range v.Threads() {
		threadTitle[th.ID] = th.Title
	}
	cardOf := func(t vault.Turn) (vault.Card, bool) {
		c, ok := cards[t.ID]
		if !ok || c.BodySHA != t.BodySHA {
			return vault.Card{}, false // stale or missing: never used
		}
		return c, true
	}
	turnItem := func(t vault.Turn, key string, tier int) Item {
		it := Item{Key: key, ID: t.ID, Tier: tier, Seq: t.Seq, Header: TurnHeader(t), Text: t.Body, SHA: t.BodySHA}
		if c, ok := cardOf(t); ok {
			it.Card = CardLine(t, c, threadTitle)
			it.Threads = c.Threads
			it.Keywords = c.Keywords
			if len(c.Threads) > 0 {
				it.Group = threadTitle[c.Threads[0]]
			}
		} else {
			it.Card = PreviewLine(t)
		}
		it.Card += SizeNote(t)
		return it
	}
	partItem := func(t vault.Turn, k int) (Item, bool) {
		parts := vault.Parts(t.Body)
		if k < 1 || k > len(parts) {
			return Item{}, false
		}
		it := turnItem(t, KeyRecall, 0)
		it.ID = vault.PartID(t.ID, k)
		it.Header = fmt.Sprintf("%s · part %d of %d (%s)", TurnHeader(t), k, len(parts), it.ID)
		it.Text = parts[k-1]
		it.SHA = vault.SHA(parts[k-1])
		it.Sub = 1 + partStart(parts, k)
		return it, true
	}
	// atomItem is atoms a to b of a turn as one excerpt: the exact span,
	// headed by the turn, its place in it and the message it answers
	atomItem := func(t vault.Turn, a, b int, key string, tier int, score float64) (Item, bool) {
		atoms := atomsOf(t)
		if a < 1 || b < a || b > len(atoms) {
			return Item{}, false
		}
		it := turnItem(t, key, tier)
		it.Score = score
		if a == 1 && b == len(atoms) {
			return it, true // every atom: the turn itself
		}
		var replyTo *vault.Turn
		if q, ok := turnByID[t.ReplyTo]; ok {
			replyTo = &q
		}
		it.ID = vault.AtomRangeID(t.ID, a, b)
		it.Header = AtomHeader(t, a, b, len(atoms), replyTo)
		// a long code block is cut into several atoms: say so when the
		// excerpt is not the whole block, since its fence is not closed
		switch in, out := vault.InFence(t.Body, atoms[a-1].Start), vault.InFence(t.Body, atoms[b-1].End); {
		case in && out:
			it.Header += " · begins and ends inside a long code block"
		case in:
			it.Header += " · begins inside a long code block"
		case out:
			it.Header += " · ends inside a long code block"
		}
		it.Text = vault.SpanText(t.Body[atoms[a-1].Start:atoms[b-1].End])
		it.SHA = vault.SHA(it.Text)
		it.Sub = 1 + atoms[a-1].Start
		it.Card = fmt.Sprintf("%s (excerpt of %s): %s", it.ID, t.ID, oneLine(it.Text, 100))
		return it, true
	}

	if opt.FocusItem != nil {
		fi := *opt.FocusItem
		fi.Key, fi.Tier = KeyFocus, 0
		add(fi)
	} else {
		add(turnItem(focus, KeyFocus, 0))
	}
	// what the model asked for: whole turns, parts of long turns, and
	// atoms. Atoms asked for in several ranges are joined into runs, and
	// an atom inside a recalled part is not sent a second time. held
	// records every atom a recalled excerpt or part already holds.
	held := map[string]map[int]bool{}
	asked := map[string]map[int]bool{}
	var askedOrder []string
	for _, id := range append(append([]string{}, opt.Pinned...), opt.IncludeTurns...) {
		ref, ok := vault.ParseRef(id)
		t, found := turnByID[ref.Turn]
		if !ok || !found {
			continue
		}
		switch {
		case ref.Atom > 0:
			if n := len(atomsOf(t)); ref.AtomEnd > n {
				continue
			}
			if asked[t.ID] == nil {
				asked[t.ID] = map[int]bool{}
				askedOrder = append(askedOrder, t.ID)
			}
			for k := ref.Atom; k <= ref.AtomEnd; k++ {
				asked[t.ID][k] = true
			}
		case ref.Part > 0:
			if it, ok := partItem(t, ref.Part); ok {
				add(it)
				parts := vault.Parts(t.Body)
				from := partStart(parts, ref.Part)
				to := from + len(parts[ref.Part-1])
				for _, a := range atomsOf(t) {
					if a.Start >= from && a.End <= to {
						if held[t.ID] == nil {
							held[t.ID] = map[int]bool{}
						}
						held[t.ID][a.K] = true
					}
				}
			}
		default:
			add(turnItem(t, KeyRecall, 0))
		}
	}
	for _, id := range askedOrder {
		t := turnByID[id]
		if _, whole := byID[id]; whole {
			continue
		}
		for k := range asked[id] {
			if held[id][k] {
				delete(asked[id], k)
			}
		}
		for _, x := range JoinAtoms(asked[id]) {
			if it, ok := atomItem(t, x[0], x[1], KeyRecall, 0, 0); ok {
				add(it)
			}
		}
		for k := range asked[id] {
			if held[id] == nil {
				held[id] = map[int]bool{}
			}
			held[id][k] = true
		}
	}
	if opt.NoHistory {
		return values(dedupeParts(byID)), nil
	}

	// recent: the last N turns before the focus
	var before []vault.Turn
	if lazyMode {
		start := len(metas) - opt.Recent
		if opt.RecentEpoch > 0 && start > 0 {
			start -= start % opt.RecentEpoch
		}
		if start < 0 {
			start = 0
		}
		for i := len(metas) - 1; i >= start; i-- {
			if t, ok := turnByID[metas[i].ID]; ok {
				add(turnItem(t, KeyRecent, 4))
			}
		}
		for _, m := range metas {
			if t, ok := turnByID[m.ID]; ok {
				before = append(before, t)
			}
		}
	} else {
		for _, t := range turns {
			if t.Seq < focus.Seq {
				before = append(before, t)
			}
		}
		start := len(before) - opt.Recent
		if opt.RecentEpoch > 0 && start > 0 {
			start -= start % opt.RecentEpoch
		}
		for i := len(before) - 1; i >= 0 && (opt.AllTurns || i >= start); i-- {
			add(turnItem(before[i], KeyRecent, 4))
		}
	}

	// threads: those of the recent turns' cards plus threads matching the
	// focus text; every turn in them goes in verbatim at tier 2 (the bulk
	// of a topic is left out before the turns ranked for this message)
	threads := map[string]bool{}
	if lazyMode {
		for i := len(metas) - 1; i >= 0 && i >= len(metas)-2; i-- {
			if t, ok := turnByID[metas[i].ID]; ok {
				if c, ok := cardOf(t); ok {
					for _, th := range c.Threads {
						threads[th] = true
					}
				}
			}
		}
	} else {
		for i := len(before) - 1; i >= 0 && i >= len(before)-2; i-- {
			if c, ok := cardOf(before[i]); ok {
				for _, th := range c.Threads {
					threads[th] = true
				}
			}
		}
	}
	for _, h := range s.Search(focus.Body, 3, vault.TypeThread) {
		threads[h.ID] = true
	}
	for th := range threads {
		for _, id := range s.In(th, "threads") {
			if t, ok := turnByID[id]; ok && t.Seq < focus.Seq {
				add(turnItem(t, KeyThread, 2))
			}
		}
	}

	// entities and facts: mentions of recent turns plus full-text matches
	ents := map[string]bool{}
	if lazyMode {
		for i := len(metas) - 1; i >= 0 && i >= len(metas)-opt.Recent; i-- {
			if t, ok := turnByID[metas[i].ID]; ok {
				if c, ok := cardOf(t); ok {
					for _, e := range c.Mentions {
						ents[e] = true
					}
				}
			}
		}
	} else {
		for i := len(before) - 1; i >= 0 && i >= len(before)-opt.Recent; i-- {
			if c, ok := cardOf(before[i]); ok {
				for _, e := range c.Mentions {
					ents[e] = true
				}
			}
		}
	}
	for _, h := range s.Search(focus.Body, 6, vault.TypeEntity) {
		ents[h.ID] = true
	}
	for _, e := range v.Entities() {
		if !ents[e.ID] {
			continue
		}
		txt := e.Name
		if e.Kind != "" {
			txt += " (" + e.Kind + ")"
		}
		if len(e.Aliases) > 0 {
			txt += "; also: " + strings.Join(e.Aliases, ", ")
		}
		if e.Body != "" {
			txt += "\n" + e.Body
		}
		add(Item{Key: KeyEntity, ID: e.ID, Tier: 3, Header: e.ID, Text: txt, Card: e.Name})
	}
	factHits := map[string]bool{}
	for _, h := range s.Search(focus.Body, 8, vault.TypeFact) {
		factHits[h.ID] = true
	}
	facts := v.Facts()
	allFacts := map[string]vault.Fact{}
	for _, f := range facts {
		allFacts[f.ID] = f
	}
	for _, f := range facts {
		if f.Status != "" && f.Status != "active" {
			continue
		}
		linked := factHits[f.ID]
		for _, e := range f.Entities {
			linked = linked || ents[e]
		}
		for _, old := range f.Supersedes {
			linked = linked || factHits[old]
		}
		if !linked {
			continue
		}
		// a fact that replaced another carries the value it replaced, so
		// "what was it before?" has its answer and the old one is never
		// presented as current (LongMemEval's knowledge updates)
		txt := f.Statement + "\n" + source(f.Source, f.Quote)
		for _, old := range f.Supersedes {
			if of, ok := allFacts[old]; ok {
				txt += fmt.Sprintf("\nReplaces %s: %s %s", of.ID, of.Statement, source(of.Source, of.Quote))
			}
		}
		add(Item{Key: KeyFact, ID: f.ID, Tier: 3, Header: f.ID, Text: txt, Card: f.Statement})
	}

	// relevant: what is ranked for this message (exact, full-text and
	// value lanes fused), verbatim, tier 3: it outlives the bulk of its
	// thread and is left out least relevant first. With a token budget the
	// ranked units are atoms, sent as excerpts; a turn that is already in
	// whole is not excerpted, and a thread turn with a ranked atom is kept
	// as a ranked turn.
	if opt.AllTurns || opt.NoRanked {
		// every earlier turn is in whole, or this is the recency baseline
	} else if opt.RelevantTokens > 0 {
		if copt.SessionRoute && len(copt.EpisodeTurns) == 0 {
			copt.EpisodeTurns = RouteEpisodeTurns(turns, metas, cards, focus.Body, focus.Seq, 3, 12)
		}
		have := func(turn string, k int) bool {
			if _, ok := byID[turn]; ok {
				return true
			}
			return held[turn][k]
		}
		var runs []AtomRun
		var hit map[string]float64
		if copt.AdaptivePack {
			runs, hit = PickAtomsAdaptive(turns, s, focus.Body, focus.Seq, opt.RelevantTokens, have, copt)
		} else {
			runs, hit = pickAtomsWith(turns, s, focus.Body, focus.Seq, opt.RelevantTokens, have, copt)
		}
		for id, score := range hit {
			if byID[id].Key == KeyThread {
				it := turnItem(turnByID[id], KeyRelevant, 3)
				it.Score = score
				add(it)
			}
		}
		for _, r := range runs {
			if it, ok := atomItem(r.Turn, r.A, r.B, KeyRelevant, 3, r.Score); ok {
				add(it)
			}
		}
		// AutoReplyWindow: when a top-ranked selected turn ends with an
		// unanswered question, co-select its immediate reply (seq+1) within
		// a bounded reply allowance so speaker attribution and split Q&A
		// exchanges are preserved without spending an extra recall round.
		if copt.AutoReplyWindow {
			replyBudget := min(max(opt.RelevantTokens/4, 300), 1200)
			replyUsed := 0
			checked := map[string]bool{}
			// Rank selected runs highest score first.
			ordered := append([]AtomRun(nil), runs...)
			sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Score > ordered[j].Score })
			for _, r := range ordered {
				if checked[r.Turn.ID] {
					continue
				}
				checked[r.Turn.ID] = true
				if !NeedsReplyWindow(r.Turn) || r.Turn.Seq+1 >= focus.Seq {
					continue
				}
				nextID := vault.TurnID(r.Turn.Seq + 1)
				if _, already := byID[nextID]; already {
					continue
				}
				nextT, ok := turnByID[nextID]
				if !ok {
					if loaded, lerr := v.Turn(nextID); lerr == nil {
						nextT = loaded
						turnByID[nextID] = loaded
						ok = true
					}
				}
				if !ok {
					continue
				}
				toks := int(float64(len(nextT.Body)+64)*0.30) + 1
				if replyUsed+toks > replyBudget {
					continue
				}
				replyUsed += toks
				it := turnItem(nextT, KeyRelevant, 3)
				it.Score = r.Score * 0.99
				add(it)
			}
		}
	} else {
		for _, r := range RankTurns(turns, s, focus.Body, focus.Seq, opt.Relevant) {
			it := turnItem(r.Turn, KeyRelevant, 3)
			it.Score = r.Score
			add(it)
		}
	}

	// catalog: a card for every other earlier turn, tier 1. In lazy mode,
	// construct catalog lines directly from TurnMeta without reading files.
	if lazyMode {
		for _, m := range metas {
			if _, ok := byID[m.ID]; ok {
				continue
			}
			it := Item{Key: KeyCatalog, ID: m.ID, Tier: 1, Seq: m.Seq, SHA: m.BodySHA}
			if c, ok := cards[m.ID]; ok && c.BodySHA == m.BodySHA {
				it.Card = CardLine(vault.Turn{ID: m.ID, Role: m.Role}, c, threadTitle)
				it.Threads = c.Threads
				it.Keywords = c.Keywords
				if len(c.Threads) > 0 {
					it.Group = threadTitle[c.Threads[0]]
				}
			} else {
				it.Card = m.Preview
				it.Keywords = m.TopTerms
			}
			it.Card += sizeNoteFrom(m.ID, m.Bytes, m.PartCount)
			it.Text = it.Card
			add(it)
		}
	} else {
		for _, t := range before {
			if _, ok := byID[t.ID]; ok {
				continue
			}
			it := turnItem(t, KeyCatalog, 1)
			it.Text = it.Card
			it.Header = ""
			if len(it.Keywords) == 0 {
				it.Keywords = TopTerms(t.Body, 4)
			}
			add(it)
		}
	}
	return values(dedupeParts(byID)), nil
}

// dedupeParts drops a part or an excerpt of a turn that is in the packet
// whole: the text is sent once. The whole turn takes over the piece's
// protection when the piece was harder to drop (a recalled excerpt of a
// turn that is also among the thread turns).
func dedupeParts(byID map[string]Item) map[string]Item {
	for id, it := range byID {
		ref, ok := vault.ParseRef(id)
		if !ok || ref.Whole() {
			continue
		}
		whole, ok := byID[ref.Turn]
		if !ok || whole.Key == KeyCatalog {
			continue
		}
		if protect(it.Tier) > protect(whole.Tier) {
			whole.Tier = it.Tier
			byID[ref.Turn] = whole
		}
		delete(byID, id)
	}
	return byID
}

// partStart is the byte offset of part k (1-based) in its turn's body.
func partStart(parts []string, k int) int {
	off := 0
	for _, p := range parts[:k-1] {
		off += len(p)
	}
	return off
}

// SizeNote marks a catalog line of a long turn with its size, and with its
// parts when it has more than one, so a model can recall only the part it
// needs (ARC's stub carries the length for the same reason).
func SizeNote(t vault.Turn) string {
	n := len(t.Body)
	if n < 8*1024 {
		return ""
	}
	return sizeNoteFrom(t.ID, n, len(vault.Parts(t.Body)))
}

func sizeNoteFrom(id string, n, parts int) string {
	if n < 8*1024 {
		return ""
	}
	note := fmt.Sprintf(" · %d KB", (n+512)/1024)
	if parts > 1 {
		note += fmt.Sprintf(" in %d parts (%s to %s)", parts, vault.PartID(id, 1), vault.PartID(id, parts))
	}
	return note
}

// TopTerms are a turn's most frequent words, for labeling an omitted turn
// that has no card keywords yet. Deterministic and model-free.
func TopTerms(body string, n int) []string {
	count := map[string]int{}
	first := map[string]int{}
	for i, tk := range tokens(body) {
		w := strings.ToLower(tk.text)
		if len(w) < 4 || valueStop[w] || termStop[w] {
			continue
		}
		if _, ok := first[w]; !ok {
			first[w] = i
		}
		count[w]++
	}
	words := make([]string, 0, len(count))
	for w := range count {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		if count[words[i]] != count[words[j]] {
			return count[words[i]] > count[words[j]]
		}
		return first[words[i]] < first[words[j]]
	})
	if len(words) > n {
		words = words[:n]
	}
	return words
}

var termStop = map[string]bool{
	"that": true, "this": true, "with": true, "from": true, "have": true, "will": true, "your": true,
	"their": true, "there": true, "then": true, "than": true, "they": true, "them": true, "were": true,
	"been": true, "being": true, "into": true, "onto": true, "over": true, "under": true, "also": true,
	"just": true, "like": true, "some": true, "more": true, "most": true, "very": true, "only": true,
	"each": true, "other": true, "such": true, "here": true, "need": true, "needs": true, "want": true,
	"make": true, "made": true, "does": true, "doing": true, "done": true, "because": true, "before": true,
	"after": true, "again": true, "once": true, "same": true, "these": true, "those": true, "what": true,
	"which": true, "when": true, "where": true, "while": true, "would": true, "should": true, "could": true,
	"might": true, "must": true, "shall": true, "about": true, "above": true, "below": true, "between": true,
}

func values(m map[string]Item) []Item {
	out := make([]Item, 0, len(m))
	for _, it := range m {
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func renderSession(s vault.Session) string {
	var b strings.Builder
	b.WriteString("Title: " + s.Title + "\n")
	if s.Purpose != "" {
		b.WriteString("Purpose: " + s.Purpose + "\n")
	}
	for _, in := range s.Instructions {
		b.WriteString("- " + in + "\n")
	}
	if len(s.Repos) > 0 {
		b.WriteString("Attached repositories (change one with a coder unit naming it in `repo`):\n")
		for _, r := range s.Repos {
			line := "- " + r.Name + ": " + r.Path
			if len(r.Allow) > 0 {
				line += "; may change only " + strings.Join(r.Allow, ", ")
			}
			if r.Verify != "" {
				line += "; verified with `" + r.Verify + "`"
			}
			b.WriteString(line + "\n")
		}
	}
	if s.Body != "" {
		b.WriteString("\n" + s.Body + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// TurnHeader is the header line of a verbatim turn.
func TurnHeader(t vault.Turn) string {
	who := t.Role
	if t.Author != "" {
		who += " (" + t.Author + ")"
	}
	return fmt.Sprintf("%s · %s · %s", t.ID, who, t.Created.Format("2006-01-02 15:04"))
}

// CardLine renders a catalog line from a valid card.
func CardLine(t vault.Turn, c vault.Card, threadTitle map[string]string) string {
	var th []string
	for _, id := range c.Threads {
		if title := threadTitle[id]; title != "" {
			th = append(th, title)
		}
	}
	line := fmt.Sprintf("%s %s: %s", t.ID, t.Role, oneLine(c.Summary, 240))
	if len(th) > 0 {
		line += " [" + strings.Join(th, "; ") + "]"
	}
	return line
}

// PreviewLine is the catalog line of a turn with no valid card yet: its
// opening words, marked as such. It points at the turn; it is not a summary.
func PreviewLine(t vault.Turn) string {
	return fmt.Sprintf("%s %s (not yet indexed, opens with): %s", t.ID, t.Role, oneLine(t.Body, 120))
}

func oneLine(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}
