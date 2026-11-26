// Package engine runs a conversation: it appends each message verbatim,
// answers it through sized packets on the routed providers, follows the
// handoff (recall, continuation, plan fan-out, assembly), and indexes new
// turns in the background. Control flow is code over handoffs; no model
// decides what the engine does next.
package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/route"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// Engine is one open workspace.
type Engine struct {
	Cfg     config.Config
	Name    string
	V       *vault.Vault
	S       *store.Store
	R       *route.Router
	RunsDir string
	// Log receives one progress line per event; nil is silent.
	Log func(string)
	// OnEvent receives live provider stream events (e.g., streaming text deltas
	// for interactive chat or UI); nil disables streaming callbacks.
	OnEvent func(providers.Event)

	// Abandoned counts runs found still "running" from a dead process when
	// the workspace was opened.
	Abandoned int

	health     *store.Store
	lock       *os.File
	indexMu    sync.Mutex
	stableWSMu sync.Mutex
	bg         sync.WaitGroup
}

// Create makes a new workspace and opens it.
func Create(cfg config.Config, name, title, purpose string) (*Engine, error) {
	dir := cfg.WorkspaceDir(name)
	if title == "" {
		title = filepath.Base(dir)
	}
	if _, err := vault.Create(dir, title, purpose); err != nil {
		return nil, err
	}
	return Open(cfg, name)
}

// ErrInUse means another process holds the workspace.
var ErrInUse = errors.New("workspace is in use by another cx process")

// Open opens an existing workspace for writing. It takes the workspace
// lock (one writer per workspace, MultiAIWorkspace's dispatcher lock),
// marks runs a dead process left running as abandoned, and rebuilds the
// index when it is empty.
func Open(cfg config.Config, name string) (*Engine, error) { return open(cfg, name, false) }

// OpenReadOnly opens a workspace to read while another process may hold
// it (show, runs, packet dry runs).
func OpenReadOnly(cfg config.Config, name string) (*Engine, error) { return open(cfg, name, true) }

func open(cfg config.Config, name string, readOnly bool) (*Engine, error) {
	dir := cfg.WorkspaceDir(name)
	v, err := vault.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	var lock *os.File
	if !readOnly {
		lock, err = os.OpenFile(filepath.Join(dir, ".cx", "engine.lock"), os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return nil, err
		}
		if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			holder, _ := os.ReadFile(lock.Name())
			lock.Close()
			return nil, fmt.Errorf("%w (%s)", ErrInUse, strings.TrimSpace(string(holder)))
		}
		_ = lock.Truncate(0)
		_, _ = fmt.Fprintf(lock, "pid %d since %s", os.Getpid(), time.Now().Format("15:04:05"))
	}
	s, err := store.Open(dir)
	if err != nil {
		closeLock(lock)
		return nil, err
	}
	h, err := store.OpenFile(filepath.Join(cfg.Root, "machine.db"))
	if err != nil {
		s.Close()
		closeLock(lock)
		return nil, err
	}
	r, err := route.New(cfg, h)
	if err != nil {
		s.Close()
		h.Close()
		closeLock(lock)
		return nil, err
	}
	e := &Engine{Cfg: cfg, Name: name, V: v, S: s, R: r, RunsDir: cfg.RunsDir(dir), health: h, lock: lock}
	if readOnly {
		// The atom index is a cache. A failure to bring it up to date
		// (a turn file that fails its check, a database that cannot be
		// written) must not stop commands that never rank, such as runs
		// and stats; commands that read turns report the error
		// themselves.
		if err := e.ensureAtoms(false); err != nil {
			e.logf("paragraph index not updated: %v", err)
		}
		return e, nil
	}
	e.Abandoned = s.AbandonRunning()
	var n int
	_ = s.DB.QueryRow(`SELECT count(*) FROM records`).Scan(&n)
	if n == 0 {
		if err := e.Reindex(); err != nil {
			e.Close()
			return nil, err
		}
	}
	if err := e.ensureAtoms(true); err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}

func closeLock(f *os.File) {
	if f != nil {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}
}

// Close waits for background indexing, closes the stores, and releases
// the workspace lock.
func (e *Engine) Close() error {
	e.bg.Wait()
	e.health.Close()
	err := e.S.Close()
	closeLock(e.lock)
	return err
}

// Wait blocks until background indexing is done.
func (e *Engine) Wait() { e.bg.Wait() }

func (e *Engine) logf(format string, args ...any) {
	if e.Log != nil {
		e.Log(fmt.Sprintf(format, args...))
	}
}

// Reply is the result of one Ask.
type Reply struct {
	User      vault.Turn
	Assistant vault.Turn
	Status    string
	Runs      []RunInfo
}

// Ask appends the user's message, answers it, appends the answer, commits,
// and starts indexing both turns in the background.
func (e *Engine) Ask(ctx context.Context, text string) (Reply, error) {
	if strings.TrimSpace(text) == "" {
		return Reply{}, errors.New("empty message")
	}
	// a message left unanswered by a dead process was never indexed
	if p, ok := e.Pending(); ok && !e.S.HasRecord(p.ID) {
		e.indexTurn(p, nil)
	}
	user, err := e.V.AppendTurn(vault.RoleUser, "", "", "", text)
	if err != nil {
		return Reply{}, err
	}
	return e.answer(ctx, user)
}

// Pending returns the last turn when it is a user turn with no answer
// (the process died or every provider failed). Reads only the newest turn
// file on disk rather than scanning the entire vault.
func (e *Engine) Pending() (vault.Turn, bool) {
	last, ok, err := e.V.LastTurn()
	if err != nil || !ok {
		return vault.Turn{}, false
	}
	return last, last.Role == vault.RoleUser
}

// Retry answers the pending user turn without appending it again.
func (e *Engine) Retry(ctx context.Context) (Reply, error) {
	p, ok := e.Pending()
	if !ok {
		return Reply{}, errors.New("nothing to retry: the last turn is answered")
	}
	return e.answer(ctx, p)
}

func (e *Engine) answer(ctx context.Context, user vault.Turn) (Reply, error) {
	o, err := e.respond(ctx, user)
	// The message is indexed only after it is answered: its own rows would
	// change the full-text statistics of every word it asks about, and the
	// packet would differ from the one cx packet and cx probe build.
	e.indexTurn(user, nil)
	if err != nil {
		e.V.Commit(fmt.Sprintf("%s user (unanswered)", user.ID))
		return Reply{User: user, Runs: o.Runs}, err
	}
	asst, err := e.V.AppendTurn(vault.RoleAssistant, o.Author, o.RunID, user.ID, o.Answer)
	if err != nil {
		return Reply{User: user, Runs: o.Runs}, err
	}
	e.indexTurn(asst, nil)
	// Reactive invalidation: check whether the new user or assistant turn
	// affects any live Query Receipts.
	for _, nt := range []vault.Turn{user, asst} {
		if alerts, aerr := e.CheckStaleReceipts(nt, nil, nil); aerr == nil {
			for _, al := range alerts {
				e.logf("receipt %s now stale (%s): %s", al.ReceiptID, al.TriggerTurn, al.Reason)
			}
		}
	}
	e.V.Commit(fmt.Sprintf("%s user, %s assistant (%s)", user.ID, asst.ID, o.Author))
	e.IndexAsync(context.WithoutCancel(ctx))
	return Reply{User: user, Assistant: asst, Status: o.Status, Runs: o.Runs}, nil
}

// outcome is the end of a chain of runs that produced one answer.
type outcome struct {
	Answer string
	Author string
	RunID  string
	Status string
	H      packet.Handoff
	Runs   []RunInfo
	// Planned is true when units or parts were assembled.
	Planned bool
	Focus   vault.Turn
	// Pinned are the ids the answering run recalled: its reviewer and
	// reviser get the same records.
	Pinned []string
	// Files are the vault paths of the files the runs saved, and Changes
	// what coder runs did to their repositories, in run order.
	Files   []string
	Changes []repoChange
	// Supplied are the unit records an assembled answer was written
	// from; its reviewer and reviser read them too.
	Supplied []packet.Item
	// record is set for a unit whose handoff is a unit record: its files
	// are listed in the record, not appended to its answer.
	record bool
}

// respond answers a user turn: recall rounds, continuations, a plan fanned
// out across lanes, or splitting a message too large for any window.
func (e *Engine) respond(ctx context.Context, focus vault.Turn) (outcome, error) {
	return e.respondAs(ctx, focus, "")
}

// The arms of a measurement of plans against single calls (cx compare).
// "" is the configured behaviour.
const (
	ArmSingle = "single" // one call; no plan is offered
	ArmPlan   = "plan"   // the responder is asked for a plan whenever the request can be split
)

func (e *Engine) respondAs(ctx context.Context, focus vault.Turn, arm string) (outcome, error) {
	o, err := e.respondRaw(ctx, focus, arm)
	if err != nil {
		return o, err
	}
	o.Focus = focus
	if e.shouldReview(o) {
		o = e.review(ctx, o)
	}
	return o, nil
}

func (e *Engine) respondRaw(ctx context.Context, focus vault.Turn, arm string) (outcome, error) {
	maxUnits := e.Cfg.MaxUnits
	spec := callSpec{
		Kind: packet.KindRespond, Role: config.RoleResponder, Focus: focus, AllowPlan: arm != ArmSingle,
		Instructions: func(id string) string { return respondTextMax(id, true, maxUnits) },
	}
	switch arm {
	case ArmSingle:
		spec.Instructions = func(id string) string { return respondTextMax(id, false, maxUnits) }
	case ArmPlan:
		spec.Instructions = func(id string) string {
			return respondTextMax(id, true, maxUnits) + "\n\nThis run is one arm of a measurement of plans against single calls: return a plan whenever the request can be split, and answer it yourself only when it cannot."
		}
	}
	o, err := e.runToCompletion(ctx, spec)
	if errors.Is(err, packet.ErrTooBig) {
		e.logf("message %s is larger than any window: splitting it into parts", focus.ID)
		o, err = e.chunked(ctx, focus)
		o.Planned = true
		return o, err
	}
	if err != nil {
		return o, err
	}
	if o.H.Status == packet.StatusPlan {
		return e.fanOut(ctx, focus, o)
	}
	return o, nil
}

// runToCompletion runs one call and follows needs_recall and partial.
// Recalled turns are pinned verbatim; a run that asked for more than fits
// with the rest of its packet is told so and asked again (ARC's recall
// budget), instead of the overflow being taken for an oversized message.
func (e *Engine) runToCompletion(ctx context.Context, spec callSpec) (outcome, error) {
	o := outcome{record: spec.UnitRecord}
	pinned := append([]string{}, spec.Pinned...)
	var notes, added []string
	var recallExtras []packet.Item
	overflows := 0
	limitNote := "\n\nRecall limit reached: answer now with what this packet holds, and say plainly what you could not check."
	base := spec.Instructions
	for round := 0; ; round++ {
		s := spec
		s.Pinned = pinned
		s.Extra = append(append(append([]packet.Item{}, spec.Extra...), recallExtras...), recallNotes(notes)...)
		if round >= e.Cfg.RecallRounds {
			s.Instructions = func(id string) string { return base(id) + limitNote }
		}
		res, err := e.call(ctx, s)
		if errors.Is(err, packet.ErrTooBig) && len(added) > 0 {
			e.logf("recall: %s do not fit with the rest of the packet; asking again", strings.Join(added, ", "))
			pinned = pinned[:len(pinned)-len(added)]
			notes = append(notes, fmt.Sprintf("These records could not all be put back together with the rest of this packet: %s (about %d tokens). Ask for fewer at once, for fewer paragraphs of a turn (tNNNNNN#aK-M), or for the parts of a long turn (tNNNNNN#pK, listed on its catalog line).",
				strings.Join(added, ", "), e.tokensOf(added)))
			added = nil
			if round > e.Cfg.RecallRounds {
				return o, err
			}
			// no model ran: the round is not spent, so the run can act
			// on the note (bounded, so a run cannot loop on overflows)
			if overflows < e.Cfg.RecallRounds {
				overflows++
				round--
			}
			continue
		}
		if err != nil {
			return o, err
		}
		o.Runs = append(o.Runs, res.Info)
		o.Pinned = append([]string{}, pinned...)
		o.Files = append(o.Files, res.Artifacts...)
		if res.Change != nil {
			o.Changes = append(o.Changes, *res.Change)
		}
		added = nil
		switch res.H.Status {
		case packet.StatusNeedsRecall:
			ids, why := e.recallRequest(res.H.RecallIDs, res.H.RecallQueries, spec.Focus, pinned)
			opIDs, opExtras, opNotes := e.ExecuteRecallOps(res.H.RecallOps, spec.Focus, append(append([]string{}, pinned...), ids...))
			ids = append(ids, opIDs...)
			recallExtras = append(recallExtras, opExtras...)
			notes = append(notes, append(why, opNotes...)...)
			// Log recalled atoms/turns as free page-fault telemetry.
			if len(ids) > 0 && e.S != nil && !e.S.ReadOnly {
				sigs := make([]store.AtomSignal, 0, len(ids))
				for _, rid := range ids {
					sigs = append(sigs, store.AtomSignal{FocusTurn: spec.Focus.ID, AtomID: rid, Signal: "recalled", Lane: "recall"})
				}
				_ = e.S.LogSignals(sigs)
			}
			if (len(ids) > 0 || len(opExtras) > 0) && round < e.Cfg.RecallRounds {
				e.logf("recall: %s asked for %s verbatim", res.Profile.ID, strings.Join(ids, ", "))
				pinned = append(pinned, ids...)
				added = ids
				continue
			}
			if strings.TrimSpace(res.H.Answer) != "" {
				return e.done(o, res, packet.StatusCompleted), nil
			}
			if round > e.Cfg.RecallRounds {
				return o, fmt.Errorf("%s kept asking for recall after %d rounds", res.Profile.ID, round)
			}
			continue
		case packet.StatusPartial:
			return e.continueFrom(ctx, spec, pinned, res, o)
		case packet.StatusPlan:
			if !spec.AllowPlan {
				if strings.TrimSpace(res.H.Answer) != "" {
					return e.done(o, res, packet.StatusCompleted), nil
				}
				return o, fmt.Errorf("%s returned a plan where none is allowed", res.Profile.ID)
			}
			return e.done(o, res, packet.StatusPlan), nil
		default:
			return e.done(o, res, res.H.Status), nil
		}
	}
}

func (e *Engine) done(o outcome, res callResult, status string) outcome {
	o.Answer = res.H.Answer
	if len(res.Artifacts) > 0 && !o.record {
		o.Answer += "\n\nFiles saved:\n- " + strings.Join(res.Artifacts, "\n- ")
	}
	o.Author = res.Profile.ID
	o.RunID = res.RunID
	o.Status = status
	o.H = res.H
	return o
}

// continueFrom follows a partial run with continuation runs that receive
// the work so far verbatim (BookWriter's continuation: the parent's
// next_step and units_remaining, depth-bounded).
func (e *Engine) continueFrom(ctx context.Context, spec callSpec, pinned []string, res callResult, o outcome) (outcome, error) {
	parts := []string{res.H.Answer}
	authors := []string{res.Profile.ID}
	h := res.H
	acc := res.H // a unit's record covers every part of its answer
	var notes []string
	stalls := 0
	for depth := 1; h.Status == packet.StatusPartial && depth <= e.Cfg.ContinuationDepth; depth++ {
		work := packet.Item{Key: packet.KeyExtra, ID: "work-so-far", Header: "Work so far (verbatim)", Text: strings.Join(parts, "\n\n")}
		next, remaining := h.NextStep, h.UnitsRemaining
		cs := spec
		cs.Kind = packet.KindContinue
		cs.AllowPlan = false
		cs.Pinned = pinned
		cs.Extra = append(append(append([]packet.Item{}, spec.Extra...), work), recallNotes(notes)...)
		cs.Instructions = func(id string) string {
			instr := continueInstructions(id, next, remaining)
			if spec.UnitRecord {
				instr += "\n\nThis continues a unit, whose record is read as a whole: give claims, assumptions, open_questions and conflicts for the new part only."
			}
			if spec.Assembly {
				instr += "\n\nThis continues an assembled reply: in claims_used and claims_omitted, list the claims this part accounts for; claims accounted for in earlier parts stay accounted for."
			}
			return instr
		}
		e.logf("continuation %d: %s", depth, oneLine(next, 100))
		r2, err := e.call(ctx, cs)
		if err != nil {
			parts = append(parts, fmt.Sprintf("[Stopped: the continuation failed (%v). Remaining: %s]", err, strings.Join(remaining, "; ")))
			break
		}
		o.Runs = append(o.Runs, r2.Info)
		o.Files = append(o.Files, r2.Artifacts...)
		if r2.Change != nil {
			o.Changes = append(o.Changes, *r2.Change)
		}
		if r2.H.Status == packet.StatusNeedsRecall {
			ids, why := e.recallRequest(r2.H.RecallIDs, r2.H.RecallQueries, spec.Focus, pinned)
			pinned = append(pinned, ids...)
			notes = append(notes, why...)
			depth--
			if len(ids) == 0 {
				stalls++
			}
			if len(pinned) > 40 || stalls > 1 {
				break
			}
			continue
		}
		parts = append(parts, r2.H.Answer)
		authors = append(authors, r2.Profile.ID)
		acc.Claims = append(acc.Claims, r2.H.Claims...)
		acc.Assumptions = append(acc.Assumptions, r2.H.Assumptions...)
		acc.OpenQuestions = append(acc.OpenQuestions, r2.H.OpenQuestions...)
		acc.Conflicts = append(acc.Conflicts, r2.H.Conflicts...)
		if n := strings.TrimSpace(r2.H.Notes); n != "" {
			acc.Notes = strings.TrimSpace(acc.Notes + "\n\n" + n)
		}
		mergeAccounting(&acc, r2.H)
		h = r2.H
		res = r2
	}
	if h.Status == packet.StatusPartial {
		parts = append(parts, fmt.Sprintf("[Stopped after %d continuations. Remaining: %s]", e.Cfg.ContinuationDepth, strings.Join(h.UnitsRemaining, "; ")))
	}
	o.Answer = strings.Join(parts, "\n\n")
	o.Author = strings.Join(uniq(authors), " + ")
	o.RunID = res.RunID
	o.Status = packet.StatusCompleted
	o.H = h
	if spec.UnitRecord {
		o.H.Claims, o.H.Assumptions, o.H.OpenQuestions, o.H.Conflicts, o.H.Notes = acc.Claims, acc.Assumptions, acc.OpenQuestions, acc.Conflicts, acc.Notes
	}
	if spec.Assembly {
		o.H.ClaimsUsed, o.H.ClaimsOmitted = acc.ClaimsUsed, acc.ClaimsOmitted
	}
	return o, nil
}

// newRecallIDs keeps requested ids that are real earlier turns or parts
// not yet pinned, at most maxRecallIDs.
func (e *Engine) newRecallIDs(ids []string, focus vault.Turn, pinned []string) []string {
	add, _ := e.recallRequest(ids, nil, focus, pinned)
	return add
}

// chunked handles a message too large for any window: split at paragraph
// boundaries into parts that fit the long_reader's largest window, run the
// parts in parallel, and assemble. Every byte of the message is read by
// some run; nothing is summarized away.
func (e *Engine) chunked(ctx context.Context, focus vault.Turn) (outcome, error) {
	var o outcome
	cands, skipped := e.R.Candidates(config.RoleLongReader, nil)
	if len(cands) == 0 {
		return o, fmt.Errorf("no long_reader provider available: %v", skipped)
	}
	big := cands[0]
	for _, p := range cands[1:] {
		if p.EffectiveWindow > big.EffectiveWindow {
			big = p
		}
	}
	sess, _ := e.V.Session()
	fixed := 12000 + len(sess.Body) + len(sess.Purpose) // rules, the unit record guide and its schema
	size := packet.ChunkBytes(big, packet.OutputTokens[packet.KindUnit], fixed)
	if size < 2000 {
		return o, fmt.Errorf("no room for message parts on %s", big.ID)
	}
	parts := splitParagraphs(focus.Body, size)
	opening := prefix(focus.Body, 3000)
	openItem := packet.Item{Key: packet.KeyExtra, ID: focus.ID + "#opening", Header: fmt.Sprintf("Opening of message %s (verbatim, first %d of %d characters)", focus.ID, len([]rune(opening)), len([]rune(focus.Body))), Text: opening}
	e.logf("split %s (%d bytes) into %d parts of up to %d bytes for %s", focus.ID, len(focus.Body), len(parts), size, config.RoleLongReader)

	type result struct {
		o   outcome
		err error
	}
	results := make([]result, len(parts))
	var wg sync.WaitGroup
	offset := 0
	for i, p := range parts {
		item := packet.Item{ID: fmt.Sprintf("%s#part%d", focus.ID, i+1), Seq: focus.Seq,
			Header: fmt.Sprintf("%s part %d of %d (verbatim, bytes %d to %d)", focus.ID, i+1, len(parts), offset, offset+len(p)), Text: p}
		offset += len(p)
		var extra []packet.Item
		if i > 0 {
			extra = []packet.Item{openItem}
		}
		wg.Add(1)
		go func(i int, item packet.Item, extra []packet.Item) {
			defer wg.Done()
			spec := callSpec{
				Kind: packet.KindUnit, Role: config.RoleLongReader, Focus: focus, FocusItem: &item,
				FocusLabel: "The message to answer (one part)", Extra: extra, NoHistory: true, UnitRecord: true,
				Instructions: func(id string) string { return chunkInstructions(id, i+1, len(parts), focus.ID) },
			}
			r, err := e.runToCompletion(ctx, spec)
			results[i] = result{r, err}
		}(i, item, extra)
	}
	wg.Wait()
	// a part's record cites the message itself; parts are stored under
	// the message's id, since no planning run names them
	split := focus.ID + "-parts"
	recs := make([]unitRecord, len(results))
	for i, r := range results {
		o.Runs = append(o.Runs, r.o.Runs...)
		recs[i] = e.unitRecordOf(split, fmt.Sprintf("part%d", i+1), fmt.Sprintf("part %d of %d", i+1, len(parts)),
			config.RoleLongReader, focus, r.o, r.err)
	}
	e.logChecks(recs)
	dir := e.V.ArtifactDir(split)
	items := make([]packet.Item, len(recs))
	for i, r := range recs {
		items[i] = e.storeUnitRecord(dir, i, r)
	}
	openFocus := openItem
	base := callSpec{
		Kind: packet.KindAssemble, Role: config.RoleAssembler, Focus: focus, FocusItem: &openFocus,
		FocusLabel:   "The message to answer (opening only; every part was processed above)",
		Instructions: func(id string) string { return assembleInstructions(id, len(parts)) },
	}
	a, err := e.assemble(ctx, base, dir, split, recs, items, nil)
	if err != nil {
		o.Runs = append(o.Runs, a.Runs...)
		o.Answer, o.Author, o.Status = losslessRecords(items), "parts", packet.StatusCompleted
		o.Supplied = a.Supplied
		return o, nil
	}
	a.Runs = append(o.Runs, a.Runs...)
	return a, nil
}

// splitParagraphs is vault.SplitParagraphs: lossless paragraph pieces.
func splitParagraphs(text string, size int) []string { return vault.SplitParagraphs(text, size) }

func prefix(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func oneLine(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
