// Command cx is the ContextlessAI command line: a chat whose memory is a
// vault of verbatim records, never a compacted context window.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/engine"
	"github.com/tommac/contextless/internal/mcp"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/ui"
	"github.com/tommac/contextless/internal/vault"
)

const usage = `cx: a chat whose memory is verbatim records on disk, never a compacted context.

usage: cx [-config file] [-q] <command> [args]

  new <name> [-title T] [-purpose P]   create a workspace (a git vault)
  chat <name>                          interactive chat (""" starts and ends a multi-line message)
  say <name> <message...>              send one message ("-" reads it from stdin)
  retry <name>                         answer the last message again if it went unanswered
  show <name> [-n N]                   print the transcript verbatim
  packet <name> [-brief] [-json] [-pull id,id] [-ask words;words] [-adaptive] [-session-route] [-auto-reply] <message...>
                                       dry run: show what would be sent, and the brief with -brief; -pull
                                       shows the packet after a recall request for those ids, -json lists
                                       the records for a script
  explain <name> [-json] [-adaptive] [-session-route] [-auto-reply] [-drop id,id] <message...>
                                       for a message: its query class, which lane found what, the reusable prefix,
                                       and size estimates per profile; -drop removes ids from the built packet (no model)
  query <name> [-kind decisions|facts|turns] [-op count|list|latest|presence|absent] [-filter TEXT] [-status active|superseded|all] [-as-of SEQ] [-strict] [-resume ID] [-json]
                                       count or list ledger records by code; resumable; withheld when unsure (no model)
  views <name> [-key KEY] [-filter TEXT] [-as-of SEQ] [-json]
                                       experimental: key=value views over the turn sequence (known defects: docs/ROADMAP.md)
  fork <src> <dst> [-as-of SEQ] [-title T]
                                       experimental: fork a workspace, or branch it at turn SEQ (known defects: docs/ROADMAP.md)
  pack <name> [-block 4096] [-verify]  build/verify a seekable 4 KiB block-compressed lossless body pack (.cx/bodies.cxp)
  watch <name> [-as-of SEQ] [-absent KEY] [-offline ANSWER] <question...>
                                       answer, and keep a receipt pinned to the current turn number
  stale <name> [-all]                  list watched receipts marked stale (or all receipts with -all)
  receipt <name> <receipt-id>          verify original evidence anchors and show verbatim delta turns (no model)
  refresh <name> <receipt-id>          update a stale receipt from the earlier receipt plus the turns added since
  tune <name> <probes.jsonl>           offline 60/40 train/held-out tuning of lane weights & adaptive pack (no model)
  resume <name> <plan-id>              resume an interrupted multi-unit plan from its checkpoint
  merge <name> <repo> <branch>         re-verify a cx/* branch on current HEAD before merging it
  serve <name> [-mcp]                  run a stdio JSON-RPC 2.0 MCP server backed by this workspace
  ui <name> [-addr 127.0.0.1:8787]     serve a local web page for inspecting packets
  obsidian export <name> <out-dir>     export workspace to Obsidian Markdown (preserves human notes)
  obsidian brief <board.canvas> [-vault DIR]
                                       compile an Obsidian Canvas board into a deterministic verbatim brief
  demo                                 walk the engine through a scripted conversation (no model)
  probe <name> <probes.jsonl> [-lanes exact,fulltext,value] [-granularity atoms|turns] [-budget N] [-whole N] [-arm A] [-adaptive]
                                       recovery probes: for each question with a known source turn, is that
                                       turn in the packet verbatim, named, or reachable by search? (no model)
                                       -granularity ranks atoms (paragraphs) or whole turns, -budget is the
                                       atom budget in tokens, -whole the whole-history limit (-1 is off), -arm a
                                       baseline: current, full, bm25, recent or turns
  compare <name> <messages> [-arms single,plan] [-matched PROFILE] [-out DIR] [-keep] [-yes]
                                       measure plans against single calls: each message (a line, or JSON lines
                                       with "message") is answered once per arm, each on a scratch copy of the
                                       workspace; writes blind answers to score, a key and costs. -matched adds
                                       a single call on PROFILE (a budget-matched arm). Spends model calls, so
                                       it only starts with -yes
  cache <name> [-last N] [-budget N] [-whole N] [-recent N] [-epoch N] [-read P] [-write P]
                                       replay past user messages through the packet builder and report how
                                       much of each brief repeats the one before, the part a provider's
                                       prompt cache can serve (no model)
  floor <name>=<probes.jsonl>... [-budgets 2000,4000,...] [-replay N] [-out DIR]
                                       the offline floor: answer in packet, tokens to the answer, input and
                                       build time for every context arm on the same probes (no model)
  zoom <name>=<probes.jsonl>... [-totals 4000,8000,16000,32000] [-notes DIR] [-out DIR]
                                       test ways to reach passages within one token total with a deterministic
                                       reader: a code-built zoom ladder, flat catalogs, search and its
                                       variants, and with -notes, model-written idea notes checked by code
                                       (no model runs here; none of this is in live packets)
                                       -dump DIR writes every arm's packet for an answer test with a model
  pull <dump dir> <run> <packet id> <id>...
                                       answer test only: writes the paragraphs or turns named by id, word for
                                       word and within the dump's budget, to a file and prints its path; one
                                       request per packet and run, logged
  gaps <session.jsonl|name>...         time between user messages, against prompt-cache lifetimes (no model)
  index <name>                         index unindexed turns now (normally automatic)
  reindex <name>                       rebuild the search index from the vault files
  attach <name> [<repo> <path> [-verify CMD] [-allow GLOBS]]
                                       attach a git repo coder units may change on their own branch; no repo lists them
  import <name> <file>                 import a transcript (JSON, JSONL, a Claude Code session .jsonl, or
                                       "## User"/"## Assistant" Markdown)
  notes <name>                         the map of the conversation: how many checked idea notes, how much
                                       of the conversation they point at, what the map costs, turns with none
  notes <name> import [-topics FILE] <notes.jsonl>...
                                       store notes written elsewhere (one JSON object per line: title, kind,
                                       topics, evidence [{turn, quote}]); a note is kept only through quotes
                                       found word for word in the turns they cite (no model)
  notes <name> list [-topic T] [-find WORDS] [-n N]
                                       print map lines: all, one topic's, or the best matches of some words
  runs <name> [-n N]                   recent runs: role, provider, tokens, outcome
  stats <name>                         per provider and kind: success rate, time, tokens, failures
  list                                 list workspaces
  providers                            routing table, lanes, cooldowns
  init-config                          print a commented config file

config: ~/.contextless/config.yaml unless -config is given.
`

func main() {
	fs := flag.NewFlagSet("cx", flag.ExitOnError)
	cfgPath := fs.String("config", "", "config file")
	quiet := fs.Bool("q", false, "no progress lines")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	_ = fs.Parse(os.Args[1:])
	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(2)
	}
	path := *cfgPath
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".contextless", "config.yaml")
	}
	cfg, err := config.Load(path)
	if err != nil {
		die(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logf := func(s string) {
		if !*quiet {
			fmt.Fprintf(os.Stderr, "\x1b[2m%s %s\x1b[0m\n", time.Now().Format("15:04:05"), s)
		}
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "new":
		f := flag.NewFlagSet("new", flag.ExitOnError)
		title := f.String("title", "", "title")
		purpose := f.String("purpose", "", "what this conversation is for")
		name := firstArg(rest, "new <name>")
		_ = f.Parse(rest[1:])
		e, err := engine.Create(cfg, name, *title, *purpose)
		if err != nil {
			die(err)
		}
		defer e.Close()
		fmt.Printf("created %s\n", e.V.Dir)
	case "chat":
		e := open(cfg, firstArg(rest, "chat <name>"), logf)
		defer e.Close()
		chat(ctx, e)
	case "say":
		e := open(cfg, firstArg(rest, "say <name> <message>"), logf)
		defer e.Close()
		msg := strings.Join(rest[1:], " ")
		if msg == "-" || msg == "" {
			b, _ := io.ReadAll(os.Stdin)
			msg = string(b)
		}
		r, err := e.Ask(ctx, msg)
		if err != nil {
			die(err)
		}
		fmt.Println(r.Assistant.Body)
		logf(fmt.Sprintf("%s by %s; indexing in the background", r.Assistant.ID, r.Assistant.Author))
		e.Wait()
	case "retry":
		e := open(cfg, firstArg(rest, "retry <name>"), logf)
		defer e.Close()
		r, err := e.Retry(ctx)
		if err != nil {
			die(err)
		}
		fmt.Println(r.Assistant.Body)
		e.Wait()
	case "stats":
		e := openRO(cfg, firstArg(rest, "stats <name>"))
		defer e.Close()
		fmt.Printf("%-40s %-9s %-11s %5s %6s %8s %8s %8s  %s\n", "profile", "kind", "role", "runs", "ok", "median", "planned", "peak", "failures")
		for _, st := range e.S.Stats() {
			var fails []string
			for k, v := range st.Failures {
				fails = append(fails, fmt.Sprintf("%s %d", k, v))
			}
			sort.Strings(fails)
			fmt.Printf("%-40s %-9s %-11s %5d %5.0f%% %7.1fs %8s %8s  %s\n", st.Profile, st.Kind, st.Role, st.Runs,
				100*float64(st.OK)/float64(st.Runs), float64(st.MedianWall)/1000, k(st.AvgPlanned), k(st.MaxPeak), strings.Join(fails, ", "))
		}
	case "show":
		f := flag.NewFlagSet("show", flag.ExitOnError)
		n := f.Int("n", 0, "last N turns (0 = all)")
		name := firstArg(rest, "show <name>")
		_ = f.Parse(rest[1:])
		e := openRO(cfg, name)
		defer e.Close()
		show(e, *n)
	case "packet":
		f := flag.NewFlagSet("packet", flag.ExitOnError)
		brief := f.Bool("brief", false, "print the whole brief")
		adaptive := f.Bool("adaptive", false, "use source-diverse adaptive excerpt packing")
		sessionRoute := f.Bool("session-route", false, "enable two-stage episode/thread candidate routing")
		autoReply := f.Bool("auto-reply", false, "co-select immediate reply turn (seq+1) for question anchors")
		asJSON := f.Bool("json", false, "list the records as JSON")
		pull := f.String("pull", "", "comma-separated ids of a recall request to replay")
		ask := f.String("ask", "", "searches of the same recall request, separated by semicolons")
		name := firstArg(rest, "packet <name> [-brief] [-json] [-pull id,id] [-ask words;words] [-adaptive] [-session-route] [-auto-reply] <message>")
		_ = f.Parse(rest[1:])
		if *adaptive {
			cfg.AdaptivePack = true
		}
		if *sessionRoute {
			cfg.SessionRoute = true
		}
		if *autoReply {
			cfg.AutoReplyWindow = true
		}
		e := openRO(cfg, name)
		defer e.Close()
		dryRun(e, strings.Join(f.Args(), " "), *brief, *asJSON, splitList(*pull), splitSemis(*ask))
	case "notes":
		name := firstArg(rest, "notes <name> [import [-topics FILE] <notes.jsonl>... | list [-topic T] [-find WORDS] [-n N]]")
		sub := ""
		if len(rest) > 1 {
			sub = rest[1]
		}
		switch sub {
		case "":
			e := openRO(cfg, name)
			defer e.Close()
			rep, err := e.Atlas()
			if err != nil {
				die(err)
			}
			fmt.Println(rep)
		case "import":
			f := flag.NewFlagSet("notes import", flag.ExitOnError)
			topics := f.String("topics", "", "a topic map: JSON with map (topic to canonical name) and areas")
			_ = f.Parse(rest[2:])
			if f.NArg() == 0 {
				die(errors.New("usage: cx notes <name> import [-topics FILE] <notes.jsonl>..."))
			}
			notes, err := engine.LoadIdeaNoteFiles(f.Args())
			if err != nil {
				die(err)
			}
			var tm *engine.TopicMap
			if *topics != "" {
				m, ok := engine.LoadTopicMapFile(*topics)
				if !ok {
					die(fmt.Errorf("cannot read topic map %s", *topics))
				}
				tm = &m
			}
			e := open(cfg, name, logf)
			defer e.Close()
			res, err := e.ImportNotes(notes, tm)
			if err != nil {
				die(err)
			}
			fmt.Println(res)
			if rep, err := e.Atlas(); err == nil {
				fmt.Println(rep)
			}
		case "list":
			f := flag.NewFlagSet("notes list", flag.ExitOnError)
			topic := f.String("topic", "", "only the notes filed under this topic")
			find := f.String("find", "", "the best matches of these words")
			n := f.Int("n", 0, "at most this many lines (0 is all)")
			_ = f.Parse(rest[2:])
			e := openRO(cfg, name)
			defer e.Close()
			notes := e.V.Notes()
			if want := strings.TrimSpace(*topic + " " + *find); want != "" {
				target := *topic
				if target == "" {
					target = *find
				}
				notes = packet.MatchNotes(notes, target, len(notes))
			}
			if *n > 0 && len(notes) > *n {
				notes = notes[:*n]
			}
			for _, x := range notes {
				fmt.Println("- " + packet.MapLine(x))
			}
		default:
			die(fmt.Errorf("unknown notes command %q", sub))
		}
	case "explain":
		f := flag.NewFlagSet("explain", flag.ExitOnError)
		asJSON := f.Bool("json", false, "emit structured JSON instead of formatted text")
		adaptive := f.Bool("adaptive", false, "use source-diverse adaptive excerpt packing")
		sessionRoute := f.Bool("session-route", false, "enable two-stage episode/thread candidate routing")
		autoReply := f.Bool("auto-reply", false, "co-select immediate reply turn (seq+1) for question anchors")
		drop := f.String("drop", "", "comma-separated turn/atom IDs to remove from the built packet")
		name := firstArg(rest, "explain <name> [-json] [-adaptive] [-session-route] [-auto-reply] [-drop id,id] <message>")
		_ = f.Parse(rest[1:])
		if *adaptive {
			cfg.AdaptivePack = true
		}
		if *sessionRoute {
			cfg.SessionRoute = true
		}
		if *autoReply {
			cfg.AutoReplyWindow = true
		}
		msg := strings.TrimSpace(strings.Join(f.Args(), " "))
		if msg == "" {
			die(errors.New("usage: cx explain <name> [-json] [-adaptive] [-session-route] [-auto-reply] [-drop id,id] <message>"))
		}
		var dropIDs []string
		for _, id := range strings.Split(*drop, ",") {
			if id = strings.TrimSpace(id); id != "" {
				dropIDs = append(dropIDs, id)
			}
		}
		e := openRO(cfg, name)
		defer e.Close()
		plan, err := e.Explain(msg, dropIDs)
		if err != nil {
			die(err)
		}
		if *asJSON {
			b, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Print(packet.FormatExplain(plan))
		}
	case "query":
		f := flag.NewFlagSet("query", flag.ExitOnError)
		kind := f.String("kind", "decisions", "record kind: decisions, facts, or turns")
		op := f.String("op", "list", "operator: count, list, latest, presence, or absent")
		filter := f.String("filter", "", "optional topic or keyword filter")
		status := f.String("status", "active", "status filter: active, superseded, or all")
		asOf := f.Int("as-of", 0, "historical point-in-time turn seq cutoff (0 = live)")
		strict := f.Bool("strict", true, "withhold exact verdict when unindexed turns exist in scope")
		resumeID := f.String("resume", "", "resume a previously saved ledger query by ID")
		asJSON := f.Bool("json", false, "emit JSON output")
		name := firstArg(rest, "query <name> [-kind decisions|facts|turns] [-op count|list|latest|presence|absent] [-filter TEXT] [-as-of SEQ] [-resume ID]")
		_ = f.Parse(rest[1:])
		e := open(cfg, name, logf)
		defer e.Close()
		var (
			res engine.LedgerQueryResult
			err error
		)
		if strings.TrimSpace(*resumeID) != "" {
			res, err = e.ResumeLedgerQuery(strings.TrimSpace(*resumeID))
		} else {
			res, err = e.RunLedgerQuery(engine.LedgerQuerySpec{
				Kind: *kind, Op: *op, Filter: *filter, Status: *status, AsOfSeq: *asOf, Strict: *strict,
			})
		}
		if err != nil {
			die(err)
		}
		if *asJSON {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Print(engine.FormatLedgerQuery(res))
		}
	case "pack":
		f := flag.NewFlagSet("pack", flag.ExitOnError)
		blockBytes := f.Int("block", vault.DefaultPackBlock, "uncompressed block size in bytes")
		verify := f.Bool("verify", false, "also verify every compressed block and turn SHA-256 after packing")
		name := firstArg(rest, "pack <name> [-block 4096] [-verify]")
		_ = f.Parse(rest[1:])
		e := open(cfg, name, logf)
		defer e.Close()
		packPath := filepath.Join(e.V.Dir, ".cx", "bodies.cxp")
		st, err := vault.BuildBodyPack(e.V, packPath, *blockBytes)
		if err != nil {
			die(err)
		}
		fmt.Printf("packed %d turns into %s (%d blocks × %d B): %d raw -> %d packed bytes (%.1f%% of raw)\n",
			st.Turns, st.Path, st.BlockCount, st.BlockBytes, st.RawBytes, st.PackedBytes, st.Ratio*100)
		if *verify {
			pk, err := vault.OpenBodyPack(packPath)
			if err != nil {
				die(err)
			}
			defer pk.Close()
			if err := pk.VerifyAll(); err != nil {
				die(err)
			}
			fmt.Println("verified: every block SHA-256 and turn body SHA-256 matched verbatim")
		}
	case "views":
		f := flag.NewFlagSet("views", flag.ExitOnError)
		key := f.String("key", "", "show full bitemporal version history for one key")
		filter := f.String("filter", "", "substring filter when listing active keys")
		asOf := f.Int("as-of", 0, "point-in-time turn seq cutoff (0 = current)")
		asJSON := f.Bool("json", false, "emit JSON output")
		name := firstArg(rest, "views <name> [-key KEY] [-filter TEXT] [-as-of SEQ] [-json]")
		_ = f.Parse(rest[1:])
		e := openRO(cfg, name)
		defer e.Close()
		if strings.TrimSpace(*key) != "" {
			rows, err := e.S.QueryKeyHistory(*key, *asOf)
			if err != nil {
				die(err)
			}
			if *asJSON {
				b, _ := json.MarshalIndent(rows, "", "  ")
				fmt.Println(string(b))
				return
			}
			if len(rows) == 0 {
				fmt.Printf("no materialized assertions for key %q\n", *key)
				return
			}
			for _, r := range rows {
				state := "ACTIVE"
				span := fmt.Sprintf("t%06d..current", r.ValidFromSeq)
				if r.ValidUntilSeq > 0 {
					state = "SUPERSEDED"
					span = fmt.Sprintf("t%06d..t%06d (superseded at t%06d)", r.ValidFromSeq, r.ValidUntilSeq, r.SupersededBySeq)
				}
				fmt.Printf("%-10s %-20s = %-20s %-36s %s\n", state, r.Key, r.RawValue, span, oneLine(r.Quote, 80))
			}
			return
		}
		rows, err := e.S.ListActiveKeys(*filter, *asOf, 200)
		if err != nil {
			die(err)
		}
		if *asJSON {
			b, _ := json.MarshalIndent(rows, "", "  ")
			fmt.Println(string(b))
			return
		}
		if len(rows) == 0 {
			fmt.Println("no active materialized key assertions")
			return
		}
		for _, r := range rows {
			fmt.Printf("%-24s = %-22s (valid from t%06d)  %s\n", r.Key, r.RawValue, r.ValidFromSeq, oneLine(r.Quote, 80))
		}
	case "fork":
		if len(rest) < 2 {
			die(errors.New("usage: cx fork <src> <dst> [-as-of SEQ] [-title T]"))
		}
		f := flag.NewFlagSet("fork", flag.ExitOnError)
		asOf := f.Int("as-of", 0, "time-travel cutoff turn seq (0 copies full workspace)")
		title := f.String("title", "", "optional title for forked workspace")
		srcName, dstName := rest[0], rest[1]
		_ = f.Parse(rest[2:])
		res, err := engine.ForkWorkspace(cfg, srcName, dstName, engine.ForkOptions{AsOfSeq: *asOf, Title: *title})
		if err != nil {
			die(err)
		}
		fmt.Printf("forked %s -> %s (%s) at seq %d: %d turns, %d facts, %d decisions (snapshot %s)\n",
			res.Source, res.Target, res.Dir, res.AsOfSeq, res.TurnsCopied, res.FactsKept, res.DecisionsKept, res.SnapshotRoot)
	case "watch":
		f := flag.NewFlagSet("watch", flag.ExitOnError)
		asOf := f.Int("as-of", 0, "pin receipt to a historical turn seq (immune to future invalidation)")
		absent := f.String("absent", "", "comma-separated absence watch keys")
		offline := f.String("offline", "", "record this answer text directly without invoking a model")
		name := firstArg(rest, "watch <name> [-as-of SEQ] [-absent KEY] [-offline ANSWER] <question>")
		_ = f.Parse(rest[1:])
		question := strings.TrimSpace(strings.Join(f.Args(), " "))
		if question == "" {
			die(errors.New("usage: cx watch <name> [-as-of SEQ] [-absent KEY] [-offline ANSWER] <question>"))
		}
		var absKeys []string
		for _, k := range strings.Split(*absent, ",") {
			if k = strings.TrimSpace(k); k != "" {
				absKeys = append(absKeys, k)
			}
		}
		e := open(cfg, name, logf)
		defer e.Close()
		if *offline != "" {
			var extra []store.ReceiptPredicate
			for _, ak := range absKeys {
				extra = append(extra, store.ReceiptPredicate{Kind: "absence", Key: ak, Value: "absent at watch time"})
			}
			rcpt, err := e.CreateReceipt(question, "", *offline, *asOf, nil, extra, "", "")
			if err != nil {
				die(err)
			}
			fmt.Printf("receipt %s (status=%s, snapshot=#%d root=%s)\n", rcpt.ID, rcpt.Status, rcpt.SnapshotSeq, rcpt.SnapshotRoot)
			return
		}
		rcpt, rep, err := e.WatchQuestion(ctx, question, *asOf, absKeys)
		if err != nil {
			die(err)
		}
		fmt.Println(rep.Assistant.Body)
		fmt.Fprintf(os.Stderr, "\nreceipt %s · status=%s · snapshot=#%d root=%s\n", rcpt.ID, rcpt.Status, rcpt.SnapshotSeq, rcpt.SnapshotRoot)
		e.Wait()
	case "stale":
		f := flag.NewFlagSet("stale", flag.ExitOnError)
		all := f.Bool("all", false, "show all receipts, not only stale ones")
		name := firstArg(rest, "stale <name> [-all]")
		_ = f.Parse(rest[1:])
		e := openRO(cfg, name)
		defer e.Close()
		status := "stale"
		if *all {
			status = ""
		}
		rs, err := e.S.ListReceipts(status)
		if err != nil {
			die(err)
		}
		if len(rs) == 0 {
			if *all {
				fmt.Println("no receipts recorded")
			} else {
				fmt.Println("no stale receipts")
			}
			return
		}
		for _, r := range rs {
			mode := fmt.Sprintf("live#%d", r.SnapshotSeq)
			if r.AsOfSeq > 0 {
				mode = fmt.Sprintf("as-of#%d", r.AsOfSeq)
			}
			fmt.Printf("%-20s %-10s %-10s %s\n", r.ID, r.Status, mode, oneLine(r.Question, 80))
			if r.StaleReason != "" {
				fmt.Printf("  trigger: %s\n", r.StaleReason)
			}
		}
	case "receipt":
		if len(rest) < 2 {
			die(errors.New("usage: cx receipt <name> <receipt-id>"))
		}
		e := openRO(cfg, rest[0])
		defer e.Close()
		diff, err := e.DiffReceiptSince(rest[1])
		if err != nil {
			die(err)
		}
		r := diff.Receipt
		fmt.Printf("Receipt %s · status=%s · snapshot=#%d (root %s) · current=#%d\n", r.ID, r.Status, r.SnapshotSeq, r.SnapshotRoot, diff.CurrentSeq)
		if r.AsOfSeq > 0 {
			fmt.Printf("Historical cutoff: as-of seq %d\n", r.AsOfSeq)
		}
		fmt.Printf("Original evidence verified: %v · snapshot root match: %v\n", diff.OriginalVerified, diff.SnapshotMatch)
		fmt.Printf("\nQuestion: %s\nAnswer:   %s\n", r.Question, r.AnswerText)
		if r.StaleReason != "" {
			fmt.Printf("\nStale trigger: %s\n", r.StaleReason)
		}
		if len(diff.VerifiedQuotes) > 0 {
			fmt.Println("\nVerified original evidence:")
			for _, q := range diff.VerifiedQuotes {
				fmt.Printf("  - %s\n", q)
			}
		}
		if len(diff.VerificationNotes) > 0 {
			fmt.Println("\nVerification notes:")
			for _, n := range diff.VerificationNotes {
				fmt.Printf("  ! %s\n", n)
			}
		}
		if len(diff.DeltaTurns) > 0 {
			fmt.Printf("\nDelta turns since snapshot (%d):\n", len(diff.DeltaTurns))
			for _, dt := range diff.DeltaTurns {
				fmt.Printf("\n### %s\n%s\n", packet.TurnHeader(dt), dt.Body)
			}
		}
	case "refresh":
		if len(rest) < 2 {
			die(errors.New("usage: cx refresh <name> <receipt-id>"))
		}
		e := open(cfg, rest[0], logf)
		defer e.Close()
		updated, o, err := e.RefreshReceipt(ctx, rest[1])
		if err != nil {
			die(err)
		}
		fmt.Println(o.Answer)
		fmt.Fprintf(os.Stderr, "\nrefreshed %s -> %s (snapshot=#%d)\n", rest[1], updated.ID, updated.SnapshotSeq)
		e.Wait()
	case "tune":
		if len(rest) < 2 {
			die(errors.New("usage: cx tune <name> <probes.jsonl>"))
		}
		probes, err := readProbes(rest[1])
		if err != nil {
			die(err)
		}
		e := openRO(cfg, rest[0])
		defer e.Close()
		rep, err := e.Tune(probes)
		if err != nil {
			die(err)
		}
		fmt.Print(engine.FormatTuneReport(rep))
	case "resume":
		if len(rest) < 2 {
			die(errors.New("usage: cx resume <name> <plan-id>"))
		}
		e := open(cfg, rest[0], logf)
		defer e.Close()
		o, err := e.ResumePlan(ctx, rest[1])
		if err != nil {
			die(err)
		}
		fmt.Println(o.Answer)
		e.Wait()
	case "merge":
		if len(rest) < 3 {
			die(errors.New("usage: cx merge <name> <repo> <branch>"))
		}
		e := open(cfg, rest[0], logf)
		defer e.Close()
		sess, err := e.V.Session()
		if err != nil {
			die(err)
		}
		var target *vault.Repo
		for i := range sess.Repos {
			if sess.Repos[i].Name == rest[1] {
				target = &sess.Repos[i]
				break
			}
		}
		if target == nil {
			die(fmt.Errorf("repository %q is not attached to %s", rest[1], rest[0]))
		}
		out, err := engine.MergeBranch(*target, rest[2])
		if err != nil {
			if out != "" {
				fmt.Fprintln(os.Stderr, out)
			}
			die(err)
		}
		fmt.Printf("verified and merged %s into %s (%s)\n%s\n", rest[2], target.Name, target.Path, strings.TrimSpace(out))
	case "serve":
		f := flag.NewFlagSet("serve", flag.ExitOnError)
		mcpMode := f.Bool("mcp", true, "serve MCP JSON-RPC 2.0 over stdio")
		name := firstArg(rest, "serve <name> [-mcp]")
		_ = f.Parse(rest[1:])
		if !*mcpMode {
			die(errors.New("serve: only -mcp mode is currently supported"))
		}
		e := open(cfg, name, func(string) {})
		defer e.Close()
		if err := mcp.New(e).Serve(ctx, os.Stdin, os.Stdout); err != nil && !errors.Is(err, context.Canceled) {
			die(err)
		}
	case "ui":
		f := flag.NewFlagSet("ui", flag.ExitOnError)
		addr := f.String("addr", "127.0.0.1:8787", "listen address for the local dashboard")
		name := firstArg(rest, "ui <name> [-addr 127.0.0.1:8787]")
		_ = f.Parse(rest[1:])
		e := openRO(cfg, name)
		defer e.Close()
		if host, _, err := net.SplitHostPort(*addr); err == nil && host != "127.0.0.1" && host != "localhost" && host != "::1" {
			fmt.Fprintf(os.Stderr, "warning: %s is not a loopback address. The dashboard has no login and serves the conversation's text to anyone who can reach that address.\n", *addr)
		}
		fmt.Fprintf(os.Stderr, "cx dashboard serving %q at http://%s (Ctrl-C to stop)\n", name, *addr)
		if err := ui.Serve(ctx, e, *addr); err != nil {
			die(err)
		}
	case "obsidian":
		if len(rest) < 2 {
			die(errors.New("usage: cx obsidian export <name> <out-dir> | cx obsidian brief <board.canvas> [-vault DIR]"))
		}
		switch rest[0] {
		case "export":
			if len(rest) < 3 {
				die(errors.New("usage: cx obsidian export <name> <out-dir>"))
			}
			e := openRO(cfg, rest[1])
			defer e.Close()
			st, err := e.ExportObsidian(rest[2])
			if err != nil {
				die(err)
			}
			fmt.Printf("exported to %s: %d turns, %d threads, %d entities, %d facts, %d decisions, %d receipts (%d existing human note(s) preserved)\n",
				rest[2], st.Turns, st.Threads, st.Entities, st.Facts, st.Decisions, st.Receipts, st.Preserved)
		case "brief":
			f := flag.NewFlagSet("obsidian brief", flag.ExitOnError)
			vaultDir := f.String("vault", "", "vault root directory for relative file links (default: canvas directory)")
			board := rest[1]
			_ = f.Parse(rest[2:])
			out, err := engine.CompileObsidianCanvas(board, *vaultDir)
			if err != nil {
				die(err)
			}
			fmt.Print(out)
		default:
			die(fmt.Errorf("obsidian: unknown subcommand %q (export, brief)", rest[0]))
		}
	case "demo":
		res, err := engine.RunDemo()
		if err != nil {
			die(err)
		}
		fmt.Print(engine.FormatDemo(res))
	case "probe":
		f := flag.NewFlagSet("probe", flag.ExitOnError)
		lanes := f.String("lanes", "exact,fulltext,value", "ranking lanes to fuse, to compare them on the same probes")
		gran := f.String("granularity", "", "ranked unit: atoms or turns (default: the configuration's)")
		budget := f.Int("budget", 0, "atom budget in tokens (default: relevant_tokens)")
		whole := f.Int("whole", 0, "whole-history limit in tokens, -1 for off (default: whole_history_tokens)")
		arm := f.String("arm", "", "context arm: current, full, bm25, recent, turns")
		adaptive := f.Bool("adaptive", false, "use source-diverse adaptive excerpt packing")
		vt := f.Int("value-top", packet.AtomValueTop, "how many of the value lane's best atoms are fused (0 turns the lane off among atoms)")
		const probeUsage = "probe <name> <probes.jsonl> [-lanes exact,fulltext,value] [-granularity atoms|turns] [-budget N] [-whole N] [-adaptive]"
		name := firstArg(rest, probeUsage)
		if len(rest) < 2 {
			die(errors.New("usage: cx " + probeUsage))
		}
		_ = f.Parse(rest[2:])
		packet.RankLanes = packet.Lanes{}
		for _, l := range strings.Split(*lanes, ",") {
			switch strings.TrimSpace(l) {
			case "exact":
				packet.RankLanes.Exact = true
			case "fulltext":
				packet.RankLanes.FullText = true
			case "value":
				packet.RankLanes.Value = true
			default:
				die(fmt.Errorf("probe: unknown lane %q (exact, fulltext, value)", l))
			}
		}
		if *vt < 0 || *budget < 0 {
			die(errors.New("probe: -value-top and -budget cannot be negative (use -granularity turns to rank whole turns)"))
		}
		packet.AtomValueTop = *vt
		probes, err := readProbes(rest[1])
		if err != nil {
			die(err)
		}
		// the flags change the configuration before the workspace is
		// opened: opening builds the paragraph index only when atoms are on
		if *budget > 0 {
			cfg.RelevantTokens = *budget
		}
		switch *gran {
		case "":
		case "atoms":
			if cfg.RelevantTokens <= 0 {
				cfg.RelevantTokens = config.Default().RelevantTokens
			}
		case "turns":
			cfg.RelevantTokens = 0
		default:
			die(fmt.Errorf("probe: unknown granularity %q (atoms, turns)", *gran))
		}
		if *whole != 0 {
			cfg.WholeHistoryTokens = *whole
		}
		if *adaptive {
			cfg.AdaptivePack = true
		}
		if *arm != "" {
			lanesSet := packet.RankLanes
			if err := engine.ApplyContextArm(&cfg, *arm); err != nil {
				die(err)
			}
			if *arm == engine.ArmCurrent {
				packet.RankLanes = lanesSet
			}
		}
		e := openRO(cfg, name)
		defer e.Close()
		res, err := e.RunProbes(probes)
		if err != nil {
			die(err)
		}
		fmt.Println(e.ProbeMode())
		fmt.Print(engine.ProbeReport(res))
	case "index":
		e := open(cfg, firstArg(rest, "index <name>"), logf)
		defer e.Close()
		for {
			n, err := e.IndexPending(ctx)
			if err != nil {
				die(err)
			}
			if n == 0 {
				break
			}
			fmt.Printf("indexed %d turns\n", n)
		}
	case "reindex":
		e := open(cfg, firstArg(rest, "reindex <name>"), logf)
		defer e.Close()
		if err := e.Reindex(); err != nil {
			die(err)
		}
		fmt.Println("index rebuilt from the vault")
	case "attach":
		name := firstArg(rest, "attach <name> <repo> <path> [-verify CMD] [-allow GLOB,GLOB]")
		e := open(cfg, name, logf)
		defer e.Close()
		sess, err := e.V.Session()
		if err != nil {
			die(err)
		}
		if len(rest) < 3 {
			if len(sess.Repos) == 0 {
				fmt.Println("no repositories attached")
			}
			for _, r := range sess.Repos {
				fmt.Printf("%-16s %s  allow=%s  verify=%q  base=%s\n", r.Name, r.Path, strings.Join(r.Allow, ","), r.Verify, orStr(r.Base, "HEAD"))
			}
			return
		}
		f := flag.NewFlagSet("attach", flag.ExitOnError)
		verify := f.String("verify", "", "shell command run in the unit's worktree after a change")
		allow := f.String("allow", "", "comma-separated globs a unit may change (empty: anything but .git)")
		base := f.String("base", "", "branch or commit units start from (default HEAD)")
		_ = f.Parse(rest[3:])
		root, err := engine.RepoRoot(rest[2])
		if err != nil {
			die(fmt.Errorf("%s is not a git repository: %w", rest[2], err))
		}
		repo := vault.Repo{Name: rest[1], Path: root, Verify: *verify, Base: *base}
		for _, g := range strings.Split(*allow, ",") {
			if g = strings.TrimSpace(g); g != "" {
				repo.Allow = append(repo.Allow, g)
			}
		}
		var kept []vault.Repo
		for _, r := range sess.Repos {
			if r.Name != repo.Name {
				kept = append(kept, r)
			}
		}
		sess.Repos = append(kept, repo)
		if err := e.V.PutSession(sess); err != nil {
			die(err)
		}
		e.V.Commit("Attach repository " + repo.Name)
		fmt.Printf("attached %s at %s; coder units work on branches cx/<run> there and never merge\n", repo.Name, repo.Path)
	case "import":
		if len(rest) < 2 {
			die(errors.New("usage: cx import <name> <file>"))
		}
		e := open(cfg, rest[0], logf)
		defer e.Close()
		n, err := importFile(e, rest[1])
		if err != nil {
			die(err)
		}
		fmt.Printf("imported %d turns verbatim; run `cx index %s` to build their cards\n", n, rest[0])
	case "runs":
		f := flag.NewFlagSet("runs", flag.ExitOnError)
		n := f.Int("n", 20, "how many")
		name := firstArg(rest, "runs <name>")
		_ = f.Parse(rest[1:])
		e := openRO(cfg, name)
		defer e.Close()
		for _, r := range e.S.RecentRuns(*n) {
			fmt.Printf("%s  %-8s %-11s %-38s %-7s %-14s %6s/%-6s peak %-6s %5.1fs %s\n", r.Started.Local().Format("Jan 2 15:04"), r.Kind, r.Role, r.Profile,
				r.Status, r.HandoffStatus+r.ExitClass, k(r.PlannedTokens), k(r.CapTokens), k(r.PeakContext), float64(r.WallMs)/1000, oneLine(r.Error, 80))
		}
	case "floor":
		f := flag.NewFlagSet("floor", flag.ExitOnError)
		budgets := f.String("budgets", "2000,4000,8000,16000,32000", "atom budgets in tokens")
		replay := f.Int("replay", 30, "past user messages to rebuild per workspace")
		out := f.String("out", "", "also write floor.md and floor.json here")
		var specs []string
		for len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
			specs, rest = append(specs, rest[0]), rest[1:]
		}
		_ = f.Parse(rest)
		if len(specs) == 0 {
			die(errors.New("usage: cx floor <name>=<probes.jsonl>... [-budgets ...] [-replay N] [-out DIR]"))
		}
		var sets []engine.FloorSet
		for _, s := range specs {
			name, path, ok := strings.Cut(s, "=")
			if !ok {
				die(fmt.Errorf("floor: %q is not <name>=<probes.jsonl>", s))
			}
			probes, err := readProbes(path)
			if err != nil {
				die(err)
			}
			sets = append(sets, engine.FloorSet{Workspace: name, Name: name + "/" + filepath.Base(path), Probes: probes})
		}
		var bs []int
		for _, s := range strings.Split(*budgets, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || n <= 0 {
				die(fmt.Errorf("floor: bad budget %q", s))
			}
			bs = append(bs, n)
		}
		cells, err := engine.RunFloor(cfg, sets, bs, *replay, logf)
		if err != nil {
			die(err)
		}
		fmt.Print(engine.FloorReport(cells))
		if *out != "" {
			if err := engine.WriteFloor(*out, cells); err != nil {
				die(err)
			}
		}
	case "zoom":
		f := flag.NewFlagSet("zoom", flag.ExitOnError)
		totals := f.String("totals", "4000,8000,16000,32000", "total tokens per question")
		out := f.String("out", "", "also write zoom.md and zoom.json here")
		f.IntVar(&engine.FeedbackSeeds, "fb-seeds", engine.FeedbackSeeds, "feedback: hits whose words are borrowed")
		f.IntVar(&engine.FeedbackTerms, "fb-terms", engine.FeedbackTerms, "feedback: words added to the question")
		f.IntVar(&engine.LinkSeeds, "link-seeds", engine.LinkSeeds, "links: hits that pass score on")
		f.Float64Var(&engine.LinkLambda, "link-lambda", engine.LinkLambda, "links: how much score a link passes")
		f.IntVar(&engine.CoverPool, "cover-pool", engine.CoverPool, "cover: ranked paragraphs considered")
		f.Float64Var(&packet.CoverCostPower, "cover-cost", packet.CoverCostPower, "cover: gain is divided by a paragraph's cost to this power (1 is gain per token)")
		f.BoolVar(&packet.FeedbackMax, "fb-max", packet.FeedbackMax, "feedback: fuse the two rankings by a paragraph's better place, not the sum of both")
		f.Float64Var(&engine.CoverSpread, "cover-spread", engine.CoverSpread, "cover: bonus for an unrepresented stretch of the conversation")
		f.IntVar(&packet.NeighbourRadius, "nb-radius", packet.NeighbourRadius, "neighbours: paragraphs read on each side of a hit")
		f.IntVar(&packet.NeighbourTop, "nb-top", packet.NeighbourTop, "neighbours: how many of the best hits are expanded")
		only := f.String("only", "", "run only these arms, separated by semicolons (search and catalog+terms always run)")
		notesDir := f.String("notes", "", "directory of idea notes (<workspace>-*.jsonl), to test notes as keys to paragraphs")
		dumpDir := f.String("dump", "", "notes: also write every arm's packet for every probe here, with the passages a complete answer draws on, for an answer test with a model")
		dumpTotal := f.Int("dump-total", 8000, "notes: the budget of the dumped packets, in tokens")
		dumpOnly := f.Bool("dump-only", false, "notes: write the packets and skip the test itself")
		f.BoolVar(&engine.DumpWhole, "dump-whole", false, "notes: also write a packet holding the whole conversation for every probe (large)")
		dumpSmall := f.String("dump-small", "", "notes: further budgets, comma-separated, at which the search and notes packets are also written")
		rawTopics := f.Bool("raw-topics", false, "notes: ignore topics-<workspace>.json and use topic names as written")
		f.Float64Var(&engine.ViewShare, "view-share", engine.ViewShare, "notes: share of the budget the view arm spends on note lines")
		f.Float64Var(&engine.NoteWeight, "note-weight", engine.NoteWeight, "notes: weight of the note lane against search")
		f.Float64Var(&engine.TopicWeight, "topic-weight", engine.TopicWeight, "notes: weight of the topic lane")
		f.IntVar(&packet.TopicSeeds, "topic-seeds", packet.TopicSeeds, "notes: best notes whose topics are followed")
		f.Float64Var(&packet.TopicShare, "topic-share", packet.TopicShare, "notes: share of the best topic's weight a topic needs to be followed")
		var specs []string
		for len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
			specs, rest = append(specs, rest[0]), rest[1:]
		}
		_ = f.Parse(rest)
		if len(specs) == 0 {
			die(errors.New("usage: cx zoom <name>=<probes.jsonl>... [-totals ...] [-out DIR]"))
		}
		var sets []engine.ZoomSet
		for _, s := range specs {
			name, path, ok := strings.Cut(s, "=")
			if !ok {
				die(fmt.Errorf("zoom: %q is not <name>=<probes.jsonl>", s))
			}
			probes, err := engine.ReadZoomProbes(path)
			if err != nil {
				die(err)
			}
			set := engine.ZoomSet{Workspace: name, Name: name + "/" + filepath.Base(path), Probes: probes}
			if *notesDir != "" {
				if set.Notes, err = engine.LoadIdeaNotes(*notesDir, name); err != nil {
					die(err)
				}
				if len(set.Notes) == 0 {
					die(fmt.Errorf("zoom: no notes for %s in %s", name, *notesDir))
				}
				if tm, ok := engine.LoadTopicMap(*notesDir, name); ok && !*rawTopics {
					set.Topics = &tm
				}
			}
			sets = append(sets, set)
		}
		var ts []int
		for _, s := range strings.Split(*totals, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || n <= 0 {
				die(fmt.Errorf("zoom: bad total %q", s))
			}
			ts = append(ts, n)
		}
		for _, a := range strings.Split(*only, ";") {
			if a = strings.TrimSpace(a); a != "" {
				engine.ZoomOnly = append(engine.ZoomOnly, a)
			}
		}
		if *dumpOnly && *dumpDir == "" {
			die(errors.New("zoom: -dump-only needs -dump DIR"))
		}
		for _, x := range strings.Split(*dumpSmall, ",") {
			if x = strings.TrimSpace(x); x == "" {
				continue
			}
			n, err := strconv.Atoi(x)
			if err != nil || n <= 0 {
				die(fmt.Errorf("zoom: bad -dump-small budget %q", x))
			}
			engine.DumpSmall = append(engine.DumpSmall, n)
		}
		if !*dumpOnly {
			cells, info, err := engine.RunZoom(cfg, sets, ts, logf)
			if err != nil {
				die(err)
			}
			for _, line := range info {
				fmt.Println(line)
			}
			fmt.Println()
			fmt.Print(engine.ZoomReport(cells))
			if *out != "" {
				if err := engine.WriteZoom(*out, cells); err != nil {
					die(err)
				}
			}
		}
		if *dumpDir != "" {
			index, err := engine.DumpPackets(cfg, sets, *dumpTotal, *dumpDir, logf)
			if err != nil {
				die(err)
			}
			fmt.Print("\n" + engine.DumpSummary(index))
		}
	case "pull":
		if len(rest) < 4 {
			die(errors.New("usage: cx pull <dump dir> <run> <packet id> <paragraph or turn id>..."))
		}
		text, err := engine.PullExcerpts(rest[0], rest[1], rest[2], rest[3:])
		if err != nil {
			die(err)
		}
		// a pointer, not the text: tools that run commands may cut long output
		fmt.Printf("%d paragraphs were written to this file. Read all of it:\n%s\n", strings.Count(text, " paragraph=\""), engine.PullPath(rest[0], rest[1], rest[2]))
	case "gaps":
		if len(rest) == 0 {
			die(errors.New("usage: cx gaps <session.jsonl|name>..."))
		}
		var sessions [][]time.Time
		for _, a := range rest {
			if data, err := os.ReadFile(a); err == nil {
				sessions = append(sessions, userTimes(parseClaudeCode(strings.Split(string(data), "\n"))))
				continue
			}
			e := openRO(cfg, a)
			turns, err := e.V.Turns()
			e.Close()
			if err != nil {
				die(err)
			}
			var ts []time.Time
			for _, t := range turns {
				if t.Role == vault.RoleUser {
					ts = append(ts, t.Created)
				}
			}
			sessions = append(sessions, ts)
		}
		fmt.Print(engine.GapReport(sessions))
	case "cache":
		f := flag.NewFlagSet("cache", flag.ExitOnError)
		last := f.Int("last", 0, "replay only the last N user messages")
		budget := f.Int("budget", 0, "atom budget in tokens (default: relevant_tokens)")
		whole := f.Int("whole", 0, "whole-history limit in tokens, -1 for off (default: whole_history_tokens)")
		recent := f.Int("recent", 0, "recent turns (default: recent_turns)")
		epoch := f.Int("epoch", 0, "align the recent window to multiples of N turns, -1 slides it (default: recent_epoch)")
		spine := f.Bool("spine", packet.CatalogSpine, "list every turn before the recent ones in the catalog")
		arm := f.String("arm", "", "context arm: current, full, bm25, recent, turns")
		adaptive := f.Bool("adaptive", false, "use source-diverse adaptive excerpt packing")
		read := f.Float64("read", 0.1, "price of a cached input token, relative to a normal one")
		write := f.Float64("write", 1, "price of an uncached input token, relative to a normal one (1.25 where writing the cache costs extra)")
		name := firstArg(rest, "cache <name> [-last N]")
		_ = f.Parse(rest[1:])
		if *budget > 0 {
			cfg.RelevantTokens = *budget
		}
		if *whole != 0 {
			cfg.WholeHistoryTokens = *whole
		}
		if *adaptive {
			cfg.AdaptivePack = true
		}
		if *recent > 0 {
			cfg.RecentTurns = *recent
		}
		if *epoch != 0 {
			cfg.RecentEpoch = *epoch
		}
		packet.CatalogSpine = *spine
		if err := engine.ApplyContextArm(&cfg, *arm); err != nil {
			die(err)
		}
		e := openRO(cfg, name)
		defer e.Close()
		calls, err := e.CacheReplay(*last)
		if err != nil {
			die(err)
		}
		fmt.Print(engine.CacheReport(calls, *read, *write))
	case "compare":
		f := flag.NewFlagSet("compare", flag.ExitOnError)
		armsFlag := f.String("arms", "single,plan", "arms to run: single, plan, or a context arm (current, full, bm25, recent, turns), each a single call")
		matched := f.String("matched", "", "a profile id for a budget-matched single-call arm")
		outFlag := f.String("out", "", "output directory (default <root>/compare/<name>-<time>)")
		keep := f.Bool("keep", false, "keep the scratch workspaces")
		yes := f.Bool("yes", false, "run the trials (they spend model calls)")
		const compareUsage = "compare <name> <messages> [-arms single,plan] [-matched PROFILE] [-out DIR] [-keep] [-yes]"
		name := firstArg(rest, compareUsage)
		if len(rest) < 2 {
			die(errors.New("usage: cx " + compareUsage))
		}
		_ = f.Parse(rest[2:])
		if _, err := os.Stat(cfg.WorkspaceDir(name)); err != nil {
			die(err)
		}
		msgs, err := engine.ReadMessages(rest[1])
		if err != nil {
			die(err)
		}
		type arm struct{ label, mode, profile, context string }
		var arms []arm
		for _, a := range strings.Split(*armsFlag, ",") {
			switch a = strings.TrimSpace(a); a {
			case engine.ArmSingle, engine.ArmPlan:
				arms = append(arms, arm{label: a, mode: a})
			case "":
			default:
				if !engine.IsContextArm(a) {
					die(fmt.Errorf("compare: unknown arm %q (single, plan, %s)", a, strings.Join(engine.ContextArms, ", ")))
				}
				arms = append(arms, arm{label: a, mode: engine.ArmSingle, context: a})
			}
		}
		if *matched != "" {
			if !engine.ProfileExists(cfg, *matched) {
				die(fmt.Errorf("compare: no profile %q (cx providers lists them)", *matched))
			}
			arms = append(arms, arm{label: "matched " + *matched, mode: engine.ArmSingle, profile: *matched})
		}
		if len(msgs) == 0 || len(arms) == 0 {
			die(errors.New("compare: no messages or no arms"))
		}
		if !*yes {
			fmt.Printf("would answer %d message(s) in %d arm(s): %d trials, each on a scratch copy of %s, spending model calls.\nRun again with -yes to start.\n",
				len(msgs), len(arms), len(msgs)*len(arms), name)
			return
		}
		stamp := time.Now().Format("20060102-150405")
		out := *outFlag
		if out == "" {
			out = filepath.Join(cfg.Root, "compare", name+"-"+stamp)
		}
		var items []engine.CompareItem
		for i, msg := range msgs {
			item := engine.CompareItem{Message: msg}
			for _, a := range arms {
				scratch := fmt.Sprintf("compare-%s-%s-m%d-%s", filepath.Base(name), stamp, i+1, strings.ReplaceAll(strings.ReplaceAll(a.label, " ", "-"), "/", "_"))
				if err := engine.CopyWorkspace(cfg, name, scratch); err != nil {
					die(err)
				}
				acfg := cfg
				if a.profile != "" {
					acfg.Routing = map[string][]string{}
					for k, v := range cfg.Routing {
						acfg.Routing[k] = v
					}
					acfg.Routing[config.RoleResponder] = []string{a.profile}
				}
				if err := engine.ApplyContextArm(&acfg, a.context); err != nil {
					die(err)
				}
				logf(fmt.Sprintf("message %d of %d, arm %s", i+1, len(msgs), a.label))
				e := open(acfg, scratch, logf)
				t := e.Trial(ctx, msg, a.mode)
				e.Close()
				t.Arm = a.label
				item.Trials = append(item.Trials, t)
				if !*keep {
					_ = os.RemoveAll(cfg.WorkspaceDir(scratch))
				}
				if ctx.Err() != nil {
					die(ctx.Err())
				}
			}
			items = append(items, item)
			if err := engine.WriteCompare(out, items); err != nil {
				die(err)
			}
		}
		fmt.Print(engine.CompareSummary(items))
		fmt.Printf("\nanswers to score: %s\nkey (open after scoring): %s\nthe packets each run saw: %s\n",
			filepath.Join(out, "answers.md"), filepath.Join(out, "key.json"), filepath.Join(cfg.Root, "runs"))
	case "list":
		entries, _ := os.ReadDir(filepath.Join(cfg.Root, "workspaces"))
		for _, en := range entries {
			if v, err := vault.Open(filepath.Join(cfg.Root, "workspaces", en.Name())); err == nil {
				s, _ := v.Session()
				turns, _ := v.Turns()
				fmt.Printf("%-24s %4d turns  %s\n", en.Name(), len(turns), s.Title)
			}
		}
	case "providers":
		h, err := store.OpenFile(filepath.Join(cfg.Root, "machine.db"))
		if err != nil {
			die(err)
		}
		defer h.Close()
		r, err := routeNew(cfg, h)
		if err != nil {
			die(err)
		}
		fmt.Print(r.Table())
	case "init-config":
		fmt.Print(config.Sample)
	default:
		fs.Usage()
		os.Exit(2)
	}
}

func open(cfg config.Config, name string, logf func(string)) *engine.Engine {
	e, err := engine.Open(cfg, name)
	if err != nil {
		die(err)
	}
	e.Log = logf
	return e
}

func openRO(cfg config.Config, name string) *engine.Engine {
	e, err := engine.OpenReadOnly(cfg, name)
	if err != nil {
		die(err)
	}
	return e
}

func firstArg(args []string, use string) string {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		die(errors.New("usage: cx " + use))
	}
	return args[0]
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "cx:", err)
	os.Exit(1)
}

func chat(ctx context.Context, e *engine.Engine) {
	s, _ := e.V.Session()
	turns, _ := e.V.Turns()
	fmt.Printf("%s · %d turns on disk, all verbatim · /help for commands\n", s.Title, len(turns))
	if e.Abandoned > 0 {
		fmt.Printf("note: %d run(s) were cut off when a previous cx process exited\n", e.Abandoned)
	}
	if p, ok := e.Pending(); ok {
		fmt.Printf("note: %s has no answer yet; /retry answers it\n", p.ID)
	}
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 1<<20), 64<<20)
	for {
		fmt.Print("\n\x1b[1myou>\x1b[0m ")
		if !in.Scan() {
			fmt.Println()
			return
		}
		line := in.Text()
		if strings.TrimSpace(line) == `"""` {
			var b strings.Builder
			for in.Scan() {
				if strings.TrimSpace(in.Text()) == `"""` {
					break
				}
				b.WriteString(in.Text() + "\n")
			}
			line = b.String()
		}
		switch strings.TrimSpace(line) {
		case "":
			continue
		case "/quit", "/exit":
			return
		case "/help":
			fmt.Println(`/history [n]  show the last n turns verbatim
/explain <q>  show the packet a question would get (no model)
/stale        list watched receipts invalidated by recent turns
/retry        answer the last message again if it went unanswered
/runs         recent runs
/stats        success rate and time per provider
/providers    routing table and cooldowns
/quit         leave (background indexing finishes first)
"""           start or end a multi-line message`)
			continue
		case "/stale":
			rs, _ := e.S.ListReceipts("stale")
			if len(rs) == 0 {
				fmt.Println("no stale receipts")
			}
			for _, r := range rs {
				fmt.Printf("%-20s %s\n  trigger: %s\n", r.ID, oneLine(r.Question, 80), r.StaleReason)
			}
			continue
		case "/retry":
			start := time.Now()
			r, err := e.Retry(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			fmt.Printf("\n\x1b[1mai>\x1b[0m \x1b[2m(%s, %.0fs)\x1b[0m\n%s\n", r.Assistant.Author, time.Since(start).Seconds(), r.Assistant.Body)
			continue
		case "/stats":
			for _, st := range e.S.Stats() {
				fmt.Printf("%-40s %-9s %4d runs %3.0f%% ok, median %.1fs\n", st.Profile, st.Kind, st.Runs, 100*float64(st.OK)/float64(st.Runs), float64(st.MedianWall)/1000)
			}
			continue
		case "/runs":
			for _, r := range e.S.RecentRuns(10) {
				fmt.Printf("%-8s %-11s %-38s %-12s %s/%s tokens\n", r.Kind, r.Role, r.Profile, r.HandoffStatus+r.ExitClass, k(r.PlannedTokens), k(r.CapTokens))
			}
			continue
		case "/providers":
			fmt.Print(e.R.Table())
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "/explain ") {
			q := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "/explain "))
			if plan, err := e.Explain(q, nil); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			} else {
				fmt.Print(packet.FormatExplain(plan))
			}
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "/history") {
			n := 6
			fmt.Sscanf(strings.TrimSpace(line), "/history %d", &n)
			show(e, n)
			continue
		}
		start := time.Now()
		streamedBytes := 0
		prevOnEvent := e.OnEvent
		e.OnEvent = func(ev providers.Event) {
			if ev.Kind == providers.KindText && ev.Message != "" {
				streamedBytes += len(ev.Message)
				fmt.Fprintf(os.Stderr, "\r\x1b[2mstreaming… %d bytes (%.1fs)\x1b[0m\x1b[K", streamedBytes, time.Since(start).Seconds())
			}
		}
		r, err := e.Ask(ctx, line)
		e.OnEvent = prevOnEvent
		if streamedBytes > 0 {
			fmt.Fprint(os.Stderr, "\r\x1b[K")
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			if ctx.Err() != nil {
				return
			}
			continue
		}
		fmt.Printf("\n\x1b[1m%s\x1b[0m \x1b[2m(%s, %.0fs)\x1b[0m\n%s\n", "ai>", r.Assistant.Author, time.Since(start).Seconds(), r.Assistant.Body)
		if stale, serr := e.S.ListReceipts("stale"); serr == nil && len(stale) > 0 {
			fmt.Fprintf(os.Stderr, "\x1b[33mnote: %d watched receipt(s) are stale (/stale to inspect)\x1b[0m\n", len(stale))
		}
	}
}

func show(e *engine.Engine, n int) {
	turns, err := e.V.Turns()
	if err != nil {
		die(err)
	}
	if n > 0 && len(turns) > n {
		turns = turns[len(turns)-n:]
	}
	for _, t := range turns {
		fmt.Printf("\x1b[1m── %s\x1b[0m\n%s\n\n", packet.TurnHeader(t), t.Body)
	}
}

// splitList splits a comma-separated flag value, dropping empty entries.
func splitList(s string) []string {
	var out []string
	for _, x := range strings.Split(s, ",") {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}

// splitSemis splits a list written with semicolons, dropping empty entries.
func splitSemis(s string) []string {
	var out []string
	for _, x := range strings.Split(s, ";") {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}

func dryRun(e *engine.Engine, msg string, full, asJSON bool, pull, ask []string) {
	if strings.TrimSpace(msg) == "" {
		die(errors.New("usage: cx packet <name> <message>"))
	}
	_, skipped := e.R.Candidates(config.RoleResponder, nil)
	var (
		sized        packet.Sized
		added, notes []string
		err          error
	)
	var first packet.Sized
	if len(pull)+len(ask) > 0 {
		first, sized, added, notes, err = e.DryPull(msg, pull, ask)
	} else {
		sized, err = e.DryPacket(msg)
	}
	if err != nil && !errors.Is(err, packet.ErrTooBig) {
		die(err)
	}
	if asJSON {
		type rec struct {
			Key    string `json:"key"`
			ID     string `json:"id"`
			Tokens int    `json:"tokens"`
		}
		// what a request put back, as the packet words it
		type back struct {
			ID     string `json:"id"`
			Header string `json:"header"`
			Text   string `json:"text"`
			Tokens int    `json:"tokens"`
		}
		out := struct {
			Profile     string   `json:"profile"`
			Planned     int      `json:"planned"`
			Cap         int      `json:"cap"`
			Brief       int      `json:"brief_tokens"`
			TooBig      bool     `json:"too_big"`
			Whole       bool     `json:"whole_history"`
			Records     []rec    `json:"records"`
			Omitted     []rec    `json:"omitted"`
			Pulled      []string `json:"pulled,omitempty"`
			RecallNotes []string `json:"recall_notes,omitempty"`
			Recalled    []back   `json:"recalled,omitempty"`
			Pushed      []rec    `json:"pushed_out,omitempty"`
			Text        string   `json:"brief,omitempty"`
		}{Profile: sized.Budget.Profile, Planned: sized.Budget.Planned, Cap: sized.Budget.Cap, Brief: sized.Budget.Brief,
			TooBig: err != nil, Whole: sized.WholeHistory, Pulled: added, RecallNotes: notes}
		now := map[string]bool{}
		for _, it := range sized.Items {
			out.Records = append(out.Records, rec{it.Key, it.ID, it.Tokens})
			now[it.Key+"|"+it.ID] = true
			if it.Key == packet.KeyRecall && len(pull)+len(ask) > 0 {
				out.Recalled = append(out.Recalled, back{it.ID, it.Header, it.Text, it.Tokens})
			}
		}
		// what the packet before the request held and this one does not
		for _, it := range first.Items {
			if !now[it.Key+"|"+it.ID] && !now[packet.KeyRecall+"|"+it.ID] {
				out.Pushed = append(out.Pushed, rec{it.Key, it.ID, it.Tokens})
			}
		}
		for _, it := range sized.Dropped {
			out.Omitted = append(out.Omitted, rec{it.Key, it.ID, it.Tokens})
		}
		if full {
			out.Text = sized.Brief
		}
		js, _ := json.Marshal(out)
		fmt.Println(string(js))
		return
	}
	for _, s := range skipped {
		fmt.Printf("skipped %s: %s\n", s.Profile, s.Reason)
	}
	if len(pull)+len(ask) > 0 {
		fmt.Printf("after a recall request: %d put back (%s)\n", len(added), strings.Join(added, ", "))
		for _, n := range notes {
			fmt.Printf("  note: %s\n", n)
		}
	}
	b := sized.Budget
	fmt.Printf("profile %s · planned %d of cap %d tokens (window %d, system %d, brief %d, output %d)%s\n",
		b.Profile, b.Planned, b.Cap, b.EffectiveWindow, b.SysOverhead, b.Brief, b.Output, map[bool]string{true: " · escalated", false: ""}[b.Escalated])
	if err != nil {
		fmt.Printf("too big: %v (the engine would split the message into parts)\n", err)
	}
	sort.Slice(sized.Items, func(i, j int) bool {
		if sized.Items[i].Key != sized.Items[j].Key {
			return sized.Items[i].Key < sized.Items[j].Key
		}
		return sized.Items[i].Seq < sized.Items[j].Seq
	})
	shown, folds, mapTok := packet.MapStatsOf(sized.Items)
	fmt.Printf("\n%d records verbatim:\n", len(sized.Items)-shown-folds)
	for _, it := range sized.Items {
		if it.Key == packet.KeyMap {
			continue
		}
		fmt.Printf("  %-12s tier %d  %-20s %6d tokens\n", it.Key, it.Tier, it.ID, it.Tokens)
	}
	if shown+folds > 0 {
		fmt.Printf("\nmap of the conversation: %d note lines, %d topics named only, about %d tokens\n", shown, folds, mapTok)
	}
	if len(sized.Dropped) > 0 {
		fmt.Printf("\n%d omitted for budget (recallable by id):\n", len(sized.Dropped))
		for _, it := range sized.Dropped {
			fmt.Printf("  %-12s tier %d  %-20s %6d tokens\n", it.Key, it.Tier, it.ID, it.Tokens)
		}
	}
	if full {
		fmt.Println("\n" + sized.Brief)
	}
}

// importFile appends turns from a transcript file verbatim.
func importFile(e *engine.Engine, path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	type msg = importMsg
	var msgs []msg
	trimmed := strings.TrimSpace(string(data))
	switch {
	case strings.HasPrefix(trimmed, "["):
		if err := json.Unmarshal([]byte(trimmed), &msgs); err != nil {
			return 0, err
		}
	case strings.HasPrefix(trimmed, "{") && isClaudeCode(strings.Split(trimmed, "\n")):
		msgs = parseClaudeCode(strings.Split(trimmed, "\n"))
	case strings.HasPrefix(trimmed, "{"):
		for _, line := range strings.Split(trimmed, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			var m msg
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				return 0, fmt.Errorf("jsonl: %w", err)
			}
			msgs = append(msgs, m)
		}
	default:
		var cur *msg
		for _, line := range strings.Split(string(data), "\n") {
			h := strings.ToLower(strings.TrimSpace(line))
			if h == "## user" || h == "## assistant" || h == "## human" || h == "## ai" {
				if cur != nil {
					msgs = append(msgs, *cur)
				}
				role := strings.TrimPrefix(h, "## ")
				cur = &msg{Role: role}
				continue
			}
			if cur != nil {
				cur.Content += line + "\n"
			}
		}
		if cur != nil {
			msgs = append(msgs, *cur)
		}
	}
	n, err := e.V.ImportTranscript(msgs)
	if err != nil {
		return n, err
	}
	if err := e.Reindex(); err != nil {
		return n, err
	}
	e.V.Commit(fmt.Sprintf("Import %d turns from %s", n, filepath.Base(path)))
	return n, nil
}

func orStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func k(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}

func oneLine(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}

// readProbes reads one engine.Probe per JSON line; blank lines and lines
// starting with # are skipped.
func readProbes(path string) ([]engine.Probe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []engine.Probe
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var p engine.Probe
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, i+1, err)
		}
		if p.ID == "" {
			p.ID = fmt.Sprintf("line%d", i+1)
		}
		out = append(out, p)
	}
	return out, nil
}
