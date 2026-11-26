package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// IndexAsync indexes unindexed turns in the background. Selection never
// waits for it: until a turn has a card, it is found by recency and full
// text, and its catalog line is its opening words.
func (e *Engine) IndexAsync(ctx context.Context) {
	e.bg.Add(1)
	go func() {
		defer e.bg.Done()
		if n, err := e.IndexPending(ctx); err != nil {
			e.logf("index: %v (the turns stay searchable; indexing retries after the next message)", err)
		} else if n > 0 {
			e.logf("indexed %d turns", n)
		}
	}()
}

// IndexPending runs one index call over the oldest turns without a valid
// card and applies its ledger operations. It returns how many turns got
// cards.
func (e *Engine) IndexPending(ctx context.Context) (int, error) {
	e.indexMu.Lock()
	defer e.indexMu.Unlock()
	turns, err := e.V.Turns()
	if err != nil {
		return 0, err
	}
	cards := e.V.Cards()
	var batch []vault.Turn
	for _, t := range turns {
		if c, ok := cards[t.ID]; ok && c.BodySHA == t.BodySHA {
			continue
		}
		batch = append(batch, t)
		if len(batch) == e.Cfg.IndexBatchMax {
			break
		}
	}
	if len(batch) == 0 {
		return 0, nil
	}
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ID
	}
	last := batch[len(batch)-1]
	spec := callSpec{
		Kind: packet.KindIndex, Role: config.RoleIndexer, Focus: last, NoHistory: true,
		IncludeTurns: ids[:len(ids)-1], FocusLabel: "Turns to index (the last of the batch)",
		Extra:        e.ledgerListings(batch),
		Instructions: func(id string) string { return indexInstructions(id, ids, e.Cfg.Atlas == config.AtlasOn) },
	}
	res, err := e.call(ctx, spec)
	if err != nil {
		return 0, err
	}
	ih, err := packet.ParseIndexHandoff(res.Raw)
	if err != nil {
		return 0, err
	}
	n, dropped := e.applyIndex(batch, ih, res.RunID)
	if dropped > 0 {
		e.logf("index: %d proposed facts or decisions discarded (quote not found verbatim in its source turn)", dropped)
	}
	if len(ih.Ideas) > 0 {
		e.logf("index: idea notes: %s", e.applyIdeas(batch, ih.Ideas, res.RunID))
	}
	e.V.Commit(fmt.Sprintf("Index %s (%s)", packet.Ranges(seqs(batch)), res.Profile.ID))
	return n, nil
}

func seqs(ts []vault.Turn) []int {
	out := make([]int, len(ts))
	for i, t := range ts {
		out[i] = t.Seq
	}
	return out
}

// ledgerListings gives the indexer the existing threads and entities so it
// reuses their ids.
func (e *Engine) ledgerListings(batch []vault.Turn) []packet.Item {
	var th, en strings.Builder
	for _, t := range e.V.Threads() {
		fmt.Fprintf(&th, "- %s: %s (%s)", t.ID, t.Title, t.Status)
		if t.Goal != "" {
			fmt.Fprintf(&th, ". %s", t.Goal)
		}
		th.WriteString("\n")
	}
	for _, x := range e.V.Entities() {
		fmt.Fprintf(&en, "- %s: %s", x.ID, x.Name)
		if x.Kind != "" {
			fmt.Fprintf(&en, " (%s)", x.Kind)
		}
		if len(x.Aliases) > 0 {
			fmt.Fprintf(&en, "; also %s", strings.Join(x.Aliases, ", "))
		}
		en.WriteString("\n")
	}
	none := "(none yet)"
	tt, et := strings.TrimSpace(th.String()), strings.TrimSpace(en.String())
	if tt == "" {
		tt = none
	}
	if et == "" {
		et = none
	}
	items := []packet.Item{
		{Key: packet.KeyExtra, ID: "existing-threads", Seq: 1, Header: "Existing threads", Text: tt},
		{Key: packet.KeyExtra, ID: "existing-entities", Seq: 2, Header: "Existing entities", Text: et},
	}
	if f := e.factsInForce(batch, 40); f != "" {
		items = append(items, packet.Item{Key: packet.KeyExtra, ID: "facts-in-force", Seq: 3,
			Header: "Facts in force that these turns may update (supersede by id)", Text: f})
	}
	if e.Cfg.Atlas == config.AtlasOn {
		items = append(items, packet.Item{Key: packet.KeyExtra, ID: "existing-note-topics", Seq: 4,
			Header: "Existing note topics (with how many notes each has)", Text: e.noteTopicsListing(80)})
	}
	return items
}

// factsInForce lists the active facts the batch's words match, so the
// indexer can mark the ones a new value replaces.
func (e *Engine) factsInForce(batch []vault.Turn, limit int) string {
	var text strings.Builder
	for _, t := range batch {
		text.WriteString(t.Body)
		text.WriteString("\n")
	}
	facts := map[string]vault.Fact{}
	for _, f := range e.V.Facts() {
		if f.Status == "" || f.Status == "active" {
			facts[f.ID] = f
		}
	}
	var b strings.Builder
	for _, h := range e.S.Search(text.String(), limit, vault.TypeFact) {
		if f, ok := facts[h.ID]; ok {
			fmt.Fprintf(&b, "- %s: %s (source %s)\n", f.ID, f.Statement, f.Source)
		}
	}
	return strings.TrimSpace(b.String())
}

// applyIndex validates and applies an index handoff. Ids must resolve;
// every fact and decision quote must appear verbatim (whitespace aside) in
// its source turn, or the operation is discarded; a bad operation never
// fails the batch (BookWriter's anchor lesson: drop the op, keep the run).
func (e *Engine) applyIndex(batch []vault.Turn, ih packet.IndexHandoff, runID string) (carded, dropped int) {
	e.V.Lock()
	byID := map[string]vault.Turn{}
	for _, t := range batch {
		byID[t.ID] = t
	}
	threads := map[string]vault.Thread{}
	for _, t := range e.V.Threads() {
		threads[t.ID] = t
	}
	entities := map[string]vault.Entity{}
	byName := map[string]string{}
	for _, x := range e.V.Entities() {
		entities[x.ID] = x
		byName[strings.ToLower(x.Name)] = x.ID
	}
	decisions := map[string]vault.Decision{}
	for _, d := range e.V.Decisions() {
		decisions[d.ID] = d
	}
	keys := map[string]string{}
	var changedThreads []vault.Thread
	var changedEntities []vault.Entity
	now := time.Now().UTC().Truncate(time.Second)
	for _, nt := range ih.NewThreads {
		if strings.TrimSpace(nt.Title) == "" || nt.Key == "" {
			continue
		}
		t := vault.Thread{ID: vault.NewID("th"), Title: strings.TrimSpace(nt.Title), Status: "open", Goal: nt.Goal, Created: now}
		if e.V.PutThread(t) == nil {
			keys[nt.Key] = t.ID
			threads[t.ID] = t
			changedThreads = append(changedThreads, t)
		}
	}
	for _, ne := range ih.NewEntities {
		name := strings.TrimSpace(ne.Name)
		if name == "" || ne.Key == "" {
			continue
		}
		if id, ok := byName[strings.ToLower(name)]; ok {
			keys[ne.Key] = id
			continue
		}
		x := vault.Entity{ID: vault.NewID("en"), Name: name, Kind: ne.Kind, Aliases: ne.Aliases, Body: ne.Description}
		if e.V.PutEntity(x) == nil {
			keys[ne.Key] = x.ID
			entities[x.ID] = x
			byName[strings.ToLower(name)] = x.ID
			changedEntities = append(changedEntities, x)
		}
	}
	resolve := func(ids []string, known func(string) bool) []string {
		var out []string
		seen := map[string]bool{}
		for _, id := range ids {
			if k, ok := keys[id]; ok {
				id = k
			}
			if known(id) && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
		return out
	}
	isThread := func(id string) bool { _, ok := threads[id]; return ok }
	isEntity := func(id string) bool { _, ok := entities[id]; return ok }

	var cards []vault.Card
	for _, c := range ih.Cards {
		t, ok := byID[c.Turn]
		if !ok || strings.TrimSpace(c.Summary) == "" {
			continue
		}
		card := vault.Card{Turn: t.ID, BodySHA: t.BodySHA, Summary: strings.TrimSpace(c.Summary),
			Threads: resolve(c.Threads, isThread), Mentions: resolve(c.Mentions, isEntity), Keywords: c.Keywords, Run: runID}
		if e.V.PutCard(card) == nil {
			cards = append(cards, card)
		}
	}
	oldFacts := map[string]vault.Fact{}
	for _, f := range e.V.Facts() {
		oldFacts[f.ID] = f
	}
	var facts []vault.Fact
	for _, f := range ih.Facts {
		t, ok := byID[f.Source]
		if !ok || !containsVerbatim(t.Body, f.Quote) || strings.TrimSpace(f.Statement) == "" {
			dropped++
			continue
		}
		x := vault.Fact{ID: vault.NewID("fa"), Statement: f.Statement, Source: t.ID, Quote: f.Quote,
			Entities: resolve(f.Entities, isEntity), Status: "active"}
		for _, old := range f.Supersedes {
			if of, ok := oldFacts[old]; ok && (of.Status == "" || of.Status == "active") {
				of.Status = "superseded"
				if e.V.PutFact(of) == nil {
					oldFacts[old] = of
					x.Supersedes = append(x.Supersedes, old)
					facts = append(facts, of)
				}
			}
		}
		if e.V.PutFact(x) == nil {
			facts = append(facts, x)
		}
	}
	var decs []vault.Decision
	for _, d := range ih.Decisions {
		t, ok := byID[d.Source]
		if !ok || !containsVerbatim(t.Body, d.Quote) || strings.TrimSpace(d.Statement) == "" {
			dropped++
			continue
		}
		x := vault.Decision{ID: vault.NewID("de"), Statement: d.Statement, Source: t.ID, Quote: d.Quote,
			Threads: resolve(d.Threads, isThread), Status: "active"}
		for _, old := range d.Supersedes {
			if od, ok := decisions[old]; ok && od.Status == "active" {
				od.Status = "superseded"
				if e.V.PutDecision(od) == nil {
					x.Supersedes = append(x.Supersedes, old)
					decs = append(decs, od)
				}
			}
		}
		if e.V.PutDecision(x) == nil {
			decs = append(decs, x)
		}
	}
	for _, ts := range ih.ThreadStatus {
		if t, ok := threads[keyOr(keys, ts.Thread)]; ok && (ts.Status == "open" || ts.Status == "resolved") && t.Status != ts.Status {
			t.Status = ts.Status
			if e.V.PutThread(t) == nil {
				changedThreads = append(changedThreads, t)
			}
		}
	}
	e.V.Unlock()

	for _, t := range changedThreads {
		e.indexThread(t)
	}
	for _, x := range changedEntities {
		e.indexEntity(x)
	}
	for _, c := range cards {
		c := c
		e.indexTurn(byID[c.Turn], &c)
	}
	var supFacts, supDecs []string
	for _, f := range facts {
		e.indexFact(f)
		supFacts = append(supFacts, f.Supersedes...)
	}
	for _, d := range decs {
		e.indexDecision(d)
		supDecs = append(supDecs, d.Supersedes...)
	}
	if (len(supFacts) > 0 || len(supDecs) > 0) && len(batch) > 0 {
		if alerts, aerr := e.CheckStaleReceipts(batch[len(batch)-1], supFacts, supDecs); aerr == nil {
			for _, al := range alerts {
				e.logf("receipt %s invalidated by ledger update: %s", al.ReceiptID, al.Reason)
			}
		}
	}
	return len(cards), dropped
}

// applyIdeas stores the idea notes of an index handoff. Evidence must cite
// a turn of the batch, the only turns the indexer had in front of it word
// for word; a note is kept through the quotes found there and discarded
// otherwise, like a fact with a quote that is not in its source.
func (e *Engine) applyIdeas(batch []vault.Turn, ideas []packet.IdeaOp, runID string) NotePut {
	byID := make(map[string]vault.Turn, len(batch))
	for _, t := range batch {
		byID[t.ID] = t
	}
	in := make([]IdeaNote, 0, len(ideas))
	for _, op := range ideas {
		n := IdeaNote{Title: op.Title, Kind: op.Kind, Topics: op.Topics}
		for _, ev := range op.Evidence {
			n.Evidence = append(n.Evidence, NoteEvidence{Turn: ev.Turn, Quote: ev.Quote})
		}
		in = append(in, n)
	}
	e.V.Lock()
	defer e.V.Unlock()
	res, _ := e.putNotes(in, func(id string) (vault.Turn, bool) { t, ok := byID[id]; return t, ok }, nil, runID)
	return res
}

func keyOr(keys map[string]string, id string) string {
	if k, ok := keys[id]; ok {
		return k
	}
	return id
}

// containsVerbatim reports whether quote occurs in body, comparing with
// runs of whitespace collapsed. An empty quote never matches.
func containsVerbatim(body, quote string) bool {
	q := strings.Join(strings.Fields(quote), " ")
	if len(q) < 3 {
		return false
	}
	return strings.Contains(strings.Join(strings.Fields(body), " "), q)
}

// --- the store index (a rebuildable cache of the vault) ---

func (e *Engine) indexTurn(t vault.Turn, c *vault.Card) {
	rec, edges, at, meta, postings, kvAssert, rare := buildTurnIndexRows(t, c, e.Cfg.RelevantTokens > 0)
	var atoms []store.AtomTurn
	if at != nil {
		atoms = []store.AtomTurn{*at}
	}
	if err := e.S.BatchWrite([]store.Rec{rec}, edges, atoms, []store.TurnMeta{meta}, postings); err != nil {
		e.logf("index: store write failed for %s: %v", t.ID, err)
	}
	if err := e.S.ApplyTurnViews(kvAssert, rare); err != nil {
		e.logf("index: materialized view update failed for %s: %v", t.ID, err)
	}
}

func buildTurnIndexRows(t vault.Turn, c *vault.Card, withAtoms bool) (store.Rec, []store.EdgeSet, *store.AtomTurn, store.TurnMeta, []store.ValuePosting, []store.KeyAssertion, []store.RareKeyRow) {
	text := t.Body
	var edges []store.EdgeSet
	if c != nil {
		text += "\n" + c.Summary + "\n" + strings.Join(c.Keywords, " ")
		edges = append(edges,
			store.EdgeSet{Src: t.ID, Field: "threads", Dsts: c.Threads},
			store.EdgeSet{Src: t.ID, Field: "mentions", Dsts: c.Mentions},
		)
	}
	rec := store.Rec{ID: t.ID, Type: vault.TypeTurn, Path: vault.TurnPath(t.ID), SHA: t.BodySHA, Bytes: len(t.Body), Seq: t.Seq, Text: text}
	atoms := packet.AtomsOf(t)
	var at *store.AtomTurn
	var postings []store.ValuePosting
	var kvAssert []store.KeyAssertion
	var rare []store.RareKeyRow
	for _, x := range atoms {
		txt := x.Text(t.Body)
		kvAssert = append(kvAssert, store.ExtractKeyAssertions(t.Seq, x.K, txt)...)
		rare = append(rare, store.ExtractRareTokens(t.Seq, x.K, txt)...)
	}
	if withAtoms {
		a := store.AtomTurn{Turn: t.ID, Seq: t.Seq, Card: cardText(c)}
		for _, x := range atoms {
			txt := x.Text(t.Body)
			a.Atoms = append(a.Atoms, store.AtomRec{ID: x.ID(), Text: txt})
			postings = append(postings, packet.ExtractValuePostings(t.Seq, x.K, txt)...)
		}
		at = &a
	}
	meta := store.TurnMeta{
		Seq:       t.Seq,
		ID:        t.ID,
		Role:      t.Role,
		Author:    t.Author,
		ReplyTo:   t.ReplyTo,
		Created:   t.Created,
		BodySHA:   t.BodySHA,
		Bytes:     len(t.Body),
		Preview:   packet.PreviewLine(t),
		TopTerms:  packet.TopTerms(t.Body, 4),
		AtomCount: len(atoms),
		PartCount: len(vault.Parts(t.Body)),
	}
	return rec, edges, at, meta, postings, kvAssert, rare
}

// indexAtoms indexes a turn's atoms for stemmed search, with the turn's
// card text when it has a card. A turn is immutable, so its atoms are
// written once, and once more when its card arrives. With atoms switched
// off nothing is written; ensureAtoms catches up when they are switched
// on.
func (e *Engine) indexAtoms(t vault.Turn, c *vault.Card) error {
	if e.Cfg.RelevantTokens <= 0 {
		return nil
	}
	return e.S.IndexAtomTurns([]store.AtomTurn{atomTurn(t, c)})
}

func atomTurn(t vault.Turn, c *vault.Card) store.AtomTurn {
	at := store.AtomTurn{Turn: t.ID, Seq: t.Seq, Card: cardText(c)}
	for _, a := range vault.Atoms(t.ID, t.Body) {
		at.Atoms = append(at.Atoms, store.AtomRec{ID: a.ID(), Text: a.Text(t.Body)})
	}
	return at
}

func cardText(c *vault.Card) string {
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.Summary + "\n" + strings.Join(c.Keywords, " "))
}

// atomIndexVersion names the splitting rules and the row layout the atom
// index was built with.
func atomIndexVersion() string { return fmt.Sprintf("%d.%d", vault.AtomSeg, store.AtomLayout) }

// atomBatch is how many turns ensureAtoms writes per transaction.
const atomBatch = 200

// ensureAtoms brings the atom index up to date: every turn's atoms under
// the splitting rules in force. The index is a cache derived from the
// immutable turns, so this is safe to run from any process, a read-only
// one included. Which turns are missing is read from the file names; only
// those turn files are read. checkCards also re-indexes a turn whose
// atoms were indexed with other card text than the card on disk (a card
// written while atoms were off, or an index write that failed).
func (e *Engine) ensureAtoms(checkCards bool) error {
	if e.Cfg.RelevantTokens <= 0 || e.S.ReadOnly {
		return nil
	}
	// The unlocked read keeps the common case free of a write transaction;
	// ResetAtomsTo reads the version again under the write lock.
	if e.S.Meta("atom_seg") != atomIndexVersion() {
		if _, err := e.S.ResetAtomsTo(atomIndexVersion()); err != nil {
			return fmt.Errorf("atom index: %w", err)
		}
	}
	have := e.S.AtomTurns()
	ids := e.V.TurnIDs()
	var cards map[string]vault.Card
	if checkCards || len(have) < len(ids) {
		cards = e.V.Cards()
	}
	var todo []string
	for _, id := range ids {
		indexed, ok := have[id]
		if !ok {
			todo = append(todo, id)
			continue
		}
		if checkCards {
			c, has := cards[id]
			if (has && indexed != cardText(&c)) || (!has && indexed != "") {
				todo = append(todo, id)
			}
		}
	}
	if len(todo) > 500 {
		e.logf("building the paragraph index for %d turns (once per workspace)", len(todo))
	}
	var batch []store.AtomTurn
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		err := e.S.IndexAtomTurns(batch)
		batch = batch[:0]
		if err != nil {
			return fmt.Errorf("atom index: %w", err)
		}
		return nil
	}
	for _, id := range todo {
		t, err := e.V.Turn(id)
		if err != nil {
			return err
		}
		if c, ok := cards[id]; ok && c.BodySHA == t.BodySHA {
			batch = append(batch, atomTurn(t, &c))
		} else {
			batch = append(batch, atomTurn(t, nil))
		}
		if len(batch) == atomBatch {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	return flush()
}

func (e *Engine) indexThread(t vault.Thread) {
	_ = e.S.Upsert(store.Rec{ID: t.ID, Type: vault.TypeThread, Path: "threads/" + t.ID + ".md", SHA: vault.SHA(t.Title + t.Goal + t.Body),
		Text: t.Title + "\n" + t.Goal + "\n" + t.Body})
}

func (e *Engine) indexEntity(x vault.Entity) {
	_ = e.S.Upsert(store.Rec{ID: x.ID, Type: vault.TypeEntity, Path: "entities/" + x.ID + ".md", SHA: vault.SHA(x.Name + x.Body),
		Text: x.Name + "\n" + strings.Join(x.Aliases, " ") + "\n" + x.Kind + "\n" + x.Body})
}

func (e *Engine) indexFact(f vault.Fact) {
	_ = e.S.Upsert(store.Rec{ID: f.ID, Type: vault.TypeFact, Path: "facts/" + f.ID + ".md", SHA: vault.SHA(f.Statement), Text: f.Statement + "\n" + f.Quote})
	_ = e.S.SetEdges(f.ID, "entities", f.Entities)
	_ = e.S.SetEdges(f.ID, "source", []string{f.Source})
}

func (e *Engine) indexDecision(d vault.Decision) {
	_ = e.S.Upsert(store.Rec{ID: d.ID, Type: vault.TypeDecision, Path: "decisions/" + d.ID + ".md", SHA: vault.SHA(d.Statement + d.Status), Text: d.Statement + "\n" + d.Quote})
	_ = e.S.SetEdges(d.ID, "threads", d.Threads)
	_ = e.S.SetEdges(d.ID, "source", []string{d.Source})
}

// Reindex rebuilds the whole store index from the vault files using batched
// SQLite transactions (hundreds of turns per commit instead of per-turn commits).
func (e *Engine) Reindex() error {
	if err := e.S.Reset(); err != nil {
		return err
	}
	if err := e.S.SetMeta("atom_seg", atomIndexVersion()); err != nil {
		return err
	}
	turns, err := e.V.Turns()
	if err != nil {
		return err
	}
	cards := e.V.Cards()
	withAtoms := e.Cfg.RelevantTokens > 0
	const batchSize = 250
	var (
		recs     []store.Rec
		edges    []store.EdgeSet
		atoms    []store.AtomTurn
		metas    []store.TurnMeta
		postings []store.ValuePosting
		kvAssert []store.KeyAssertion
		rare     []store.RareKeyRow
	)
	flush := func() error {
		if len(recs) == 0 && len(edges) == 0 && len(atoms) == 0 && len(metas) == 0 && len(postings) == 0 {
			return nil
		}
		if err := e.S.BatchWrite(recs, edges, atoms, metas, postings); err != nil {
			return err
		}
		err := e.S.ApplyTurnViews(kvAssert, rare)
		recs, edges, atoms, metas, postings = recs[:0], edges[:0], atoms[:0], metas[:0], postings[:0]
		kvAssert, rare = kvAssert[:0], rare[:0]
		return err
	}
	for _, t := range turns {
		var cp *vault.Card
		if c, ok := cards[t.ID]; ok && c.BodySHA == t.BodySHA {
			c := c
			cp = &c
		}
		r, es, at, m, ps, kvs, rr := buildTurnIndexRows(t, cp, withAtoms)
		recs = append(recs, r)
		edges = append(edges, es...)
		if at != nil {
			atoms = append(atoms, *at)
		}
		metas = append(metas, m)
		postings = append(postings, ps...)
		kvAssert = append(kvAssert, kvs...)
		rare = append(rare, rr...)
		if len(recs) >= batchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	for _, t := range e.V.Threads() {
		recs = append(recs, store.Rec{ID: t.ID, Type: vault.TypeThread, Path: "threads/" + t.ID + ".md", SHA: vault.SHA(t.Title + t.Goal + t.Body),
			Text: t.Title + "\n" + t.Goal + "\n" + t.Body})
	}
	for _, x := range e.V.Entities() {
		recs = append(recs, store.Rec{ID: x.ID, Type: vault.TypeEntity, Path: "entities/" + x.ID + ".md", SHA: vault.SHA(x.Name + x.Body),
			Text: x.Name + "\n" + strings.Join(x.Aliases, " ") + "\n" + x.Kind + "\n" + x.Body})
	}
	for _, f := range e.V.Facts() {
		recs = append(recs, store.Rec{ID: f.ID, Type: vault.TypeFact, Path: "facts/" + f.ID + ".md", SHA: vault.SHA(f.Statement), Text: f.Statement + "\n" + f.Quote})
		edges = append(edges,
			store.EdgeSet{Src: f.ID, Field: "entities", Dsts: f.Entities},
			store.EdgeSet{Src: f.ID, Field: "source", Dsts: []string{f.Source}},
		)
	}
	for _, d := range e.V.Decisions() {
		recs = append(recs, store.Rec{ID: d.ID, Type: vault.TypeDecision, Path: "decisions/" + d.ID + ".md", SHA: vault.SHA(d.Statement + d.Status), Text: d.Statement + "\n" + d.Quote})
		edges = append(edges,
			store.EdgeSet{Src: d.ID, Field: "threads", Dsts: d.Threads},
			store.EdgeSet{Src: d.ID, Field: "source", Dsts: []string{d.Source}},
		)
	}
	if err := flush(); err != nil {
		return err
	}
	if withAtoms {
		if err := e.S.SetMeta("value_postings_ver", "1"); err != nil {
			return err
		}
	}
	return nil
}
