# ContextlessAI roadmap

## Stage 1 (done, f8b3bd0)

Verbatim vault, sized packets with escalation and drop tiers, recall,
continuations, plan fan-out across providers, oversized-message splitting,
background indexing, machine-wide lanes, cooldowns and failover, `cx` CLI.

## Stage 2 (built)

Each item below names what it proves or adds, and how it is checked.

1. **Long-conversation scale test.** The promise is that nothing is lost when
   history outgrows every window. A test builds a conversation of several
   hundred turns, far past the largest window, and checks that every packet
   fits its cap, that the recent turns and the focus are always verbatim, that
   omitted turns are listed by id, and that a recalled old turn comes back
   byte for byte. It also times selection, so a slow vault shows up early.
2. **Failed attempts teach the next attempt** (BookWriter `lastCritiques`,
   MAW proposal 2.6). When a run fails on one provider, the next provider's
   brief says what failed and why, as data.
3. **Crash recovery** (BookWriter `recover.go`). Runs left `running` by a dead
   process are marked `abandoned` on open. A user turn left unanswered can be
   answered with `cx retry` or `/retry`, without appending it again.
4. **`cx stats`** (MAW proposal 2.5: per-agent statistics from events already
   kept). Per profile and kind: runs, success rate, median wall time, planned
   and peak tokens, failure classes. This is how routing moves from
   reputation to measurement.
5. **Independent review gate** (BookWriter judge independence, MAW
   maker/checker). A reviewer on a provider that wrote none of the answer
   checks it against the packet; problems must quote the answer verbatim.
   On `revise`, the writer's role gets one revision run with the problems as
   a supplied record. Default: review assembled plan answers only; config
   `review: off | plans | always`.
6. **Repository code lane** (MultiAIWorkspace worker model, reduced). A
   workspace can attach git repositories (`cx attach`), each with an optional
   verify command and path allowlist. A plan unit with role `coder` and a
   `repo` runs in a fresh `git worktree` of that repo on branch `cx/<run>`,
   inside the packet directory. Agents never run git: the engine diffs,
   rejects any change outside the allowlist, commits, runs the verify
   command, reports branch, diffstat and verify result, and removes the
   worktree. The user's checkout and current branch are never touched;
   merging is the user's call.

## Stage 2b (built): lessons from the 2026 lossless-memory papers

Four 2026 papers do what ContextlessAI does in part: LCM (Voltropy:
immutable store, summary DAG, `lcm_grep`/`lcm_expand`), ARC (Fujitsu:
append-only store, a view rebuilt before every call, recall by address),
LANTERN (Cisco: archive every turn, hybrid retrieval, recovery probes) and
CogCanvas (verbatim-grounded decisions, facts and reminders in a graph).
What was taken, and how it is checked:

1. **Recovery probes** (LANTERN's evaluation, without its LLM judge).
   `cx probe` and `TestProbesRecoverPlantedFacts` report per fact type
   where a probe's source turn sits in the packet. Measured first, so every
   change below was kept only if the probes or a test showed it working.
2. **Ranked turns outlive the bulk of their threads.** A full-text match in
   a large thread used to take the thread's tier and be dropped oldest
   first. The probes showed it: 3 of 14 planted facts verbatim. With ranked
   turns at tier 3 above thread turns at tier 2: 11 of 14.
3. **Fused ranking with a value lane** (LANTERN's RRF; its config-value
   blind spot). Exact, full-text and value lanes, each lane's winner kept:
   13 of 14 verbatim, and 12 of 14 found by a search for the question.
4. **Recall by search** (LCM `lcm_grep`). `recall_queries` in the handoff;
   the engine searches every earlier turn and reports what each search
   found. Bad ids get a note naming the valid range (ARC).
5. **Parts of long turns** (ARC chunked recall). `tNNNNNN#pK`, deterministic
   paragraph cuts of 24 KB; catalog lines give size and parts; a search in a
   long turn returned the matching parts only (now the matching paragraphs,
   see Stage 3).
6. **Recall that does not fit is refused with a note** (ARC recall budget).
   Before, the overflow came back as "message too big" and the engine split
   the user's short message into parts.
7. **Digests for omitted turns** (LCM's summary-of-summaries, built by code
   as ARC insists). One line per block of up to 50 dropped catalog turns:
   thread title, id ranges, top keywords.
8. **Facts supersede facts, sources carry dates** (CogCanvas's stated gaps;
   LongMemEval's knowledge-update questions). The indexer sees the facts in
   force its turns may change; the packet shows the current value with the
   one it replaced.

Considered and not taken yet:

- **An embedding lane** (LANTERN, CogCanvas). It needs a local embedding
  model, which is a download to ask about first. LCM leaves embeddings out
  on purpose, and LANTERN's own misses were mostly coverage (361 of 420),
  not ranking (59 of 420).
- **An LLM reranker.** CogCanvas's largest ablation (-11.5 points without
  it), but LANTERN measured +1.9 points, not significant, for one more model
  call per message on subscriptions that already run out.
- **LCM's `llm_map`** (one schema-checked call per item over data larger
  than any window). The chunked path and plan units cover today's cases.
- **Reminder and todo records** (CogCanvas). Threads with open or resolved
  status cover part of it.

## Next, from the held-out notes test (CONTEXT-COMPILER.md section 2d)

- Run the blind answer test on the 150 held-out questions. Offline, checked idea
  notes reached more needed passages than search, which does not yet show better
  answers.
- Test an overview plus a model's own request for paragraphs, with note lines
  and with code-built turn lines. Three pilot questions favoured it; that is a
  lead, not a result.
- Only then decide whether notes become a ranking lane in live packets. They
  are not used there now.

## The answer test ran (2026-11-25), and what follows from it

Results and limits are in CONTEXT-COMPILER.md section 2e. In short: the lead
the fixed rule tested (a code-built line per turn plus the model's own request
for paragraphs) does not beat search of the same size. The same run's secondary
comparisons favour model-written notes read as a map: 0.726 of needed passages
stated against 0.546 for search of the same size (adjusted p 0.032), with no
difference detected from the whole conversation (0.839) at one eighth of the
tokens. An exact packet (0.997) beat the whole conversation, so selection has
room above "send everything".

Built from it (2026-11-25): notes as checked vault records, the map of the
conversation in live packets, requests bounded by tokens, the `notes` recall
op, `cx notes`, and an indexer that writes notes when `atlas: on`. Also fixed:
the handoff schemas did not list `recall_ops`, so a provider held to its
schema could never use a typed recall verb.

**Confirmed the same day (CONTEXT-COMPILER.md section 2f).** A second test on
the 120 held-out questions the first run left unused, with its rule written
before any packet or answer existed:

- The earlier finding held, at half its size: note lines plus requested
  paragraphs 0.762 against 0.672 for search of the same size (+0.090, adjusted
  p 0.005).
- The engine's own packet with the map scored 0.882 against 0.706 for the same
  engine with `atlas: off`, one recall request each (+0.175; 66 questions won,
  4 lost). On the 29 questions that also have a whole-conversation answer it
  was level with the whole conversation (0.810) at under a quarter of the
  tokens.
- By the rule fixed beforehand, `atlas` stays on by default where notes exist.

All 150 held-out questions are now spent. Every test below needs new questions
(better: a new conversation), model calls, and so a yes; each gets its pass
rule written before it runs.

Next, in order:

1. **Notes from the indexer.** Both tests used notes from separate writers
   that read long stretches at once. Index a conversation with `atlas: on`,
   then compare: quotes found, share of paragraphs pointed at, and the answer
   test's score. Until this passes, the map's result belongs to notes the
   engine does not yet write by itself.
2. **Does a model ask without being told to?** Both tests offered the request
   as a step. Run the engine's real respond call and count how often the model
   asks, and what it scores when it does not.
3. **The same tokens of search.** The map packet is twice the size of the
   map-off packet. Add an arm that fills the map's tokens with ranked
   paragraphs, so size and selection are separated at 58,000 tokens as they
   were at 35,000.
4. **One-source questions.** Check the map costs nothing where search already
   works. Left out of both runs.
5. **A line for what no note covers.** In the first run 34 of the 56 passages
   the pull arm missed had no note. Arm: the map plus a code-built line for
   each turn no note points at.
6. **Selection when everything fits.** The exact packet beat the whole
   conversation by 0.16 and 0.17 in the two runs, and the whole conversation
   fell from 0.946 to 0.719 between a 176,000 and a 295,000 token
   conversation. Test a map plus request against the whole history for
   conversations short enough to send whole, which the engine does today.
7. **Cost in practice.** Whether the CLIs reuse the map as a cached prefix is
   unproven; measure billed input over a real session with the map on.
8. **Entity links as a ranking lane.** The indexer already keeps entities
   with aliases, and each card lists the entities its turn mentions, but
   ranking never follows those links: a question that names an entity gets
   the entity's description, not the paragraphs that mention it under another
   name. The lane: question to entity (by name or alias) to the turns that
   mention it, fused like the other lanes, off by default. Two repairs belong
   with it. A proposed entity is matched to an existing one by name only, so
   a name already recorded as an alias makes a duplicate; and aliases are
   fixed when the entity is created. Left out of the release on purpose. The
   first real probes missed no name (ARCHITECTURE.md 3c: the misses were
   paraphrase and ranking), so the room may be small, and an entity carries
   no quote from the turns, so a merge would need one before ranking leans on
   it. Measure the room first, with no model at query time: on an indexed
   conversation, the share of probe misses whose source turn mentions an
   entity the question names.

## Public release (prepared 2026-11-25)

Done: Apache 2.0 license with NOTICE and CITATION.cff, THIRD_PARTY.md for
benchmark-derived files and modules, SECURITY.md and a README safety section,
CONTRIBUTING.md, a CI workflow (build, vet, test, demo on Linux and macOS), a
refusal that explains itself when no sandbox exists and a warning whenever a
CLI runs unconfined, experimental labels on `cx views` and `cx fork -as-of`.
A search of every tracked file and every blob in the history against the two
private conversations, their question and probe files and their notes found
no passage of theirs: the only overlaps are the project's own sentences that
the conversations quoted.

Left, in order:

1. **Publish from a fresh history.** `scripts/export-public.sh <dir> "<name>"
   <no-reply address>` writes the current commit as one commit and refuses if
   a home path, an email address or a key-shaped string is in the tree. Done
   2026-11-26: published as `github.com/tommac/contextless`, which is where
   work continues. The earlier private history carries a personal email
   address on every commit and is not published.
2. **Set the module path** in `go.mod` to the public repository's path. Done
   2026-11-26: `github.com/tommac/contextless`. `go install` from that path
   still fails: ten file names in `idea-vault/` contain an apostrophe, which
   the Go module format refuses. Rename them and their links, or give
   `idea-vault/` its own `go.mod` so it is left out of the module.
3. **See CI pass on Linux.** Done 2026-11-26, after two test fixes: a test
   repository with no git identity, and a reopened engine the harness did not
   close. The sandbox tests skip on Linux.
4. **A container recipe for Linux**, tested, so "run it in a container" is one
   command instead of advice.
5. **A public conversation with public questions**, so the answer tests can be
   rerun by someone else (see the confirming test's limits).
6. Fix or keep labelled the open defects below.

## Audit of the 2026-11-24 engine waves (2026-11-25)

Two independent reviews had to reproduce each defect in the code added on
2026-11-24 (MCP and UI servers, receipts, ledger queries, views, forks, the
plan gate, the body pack). None of those features has a retrieval or answer
measurement yet; several changed default behaviour.

Fixed, each with a test that fails without the fix:

- A ledger query's resume id was joined into a file path unchecked, and the id
  inside the loaded file chose where the result was written: a read and a
  write outside the vault, reachable from the MCP server and the dashboard.
- The MCP `cx://turn/<id>` resource and receipt evidence read any Markdown
  file with front matter through an id such as `../../x`. The vault now
  refuses an id that is not a turn id.
- The MCP fork tool took a path as its target and copied the conversation
  there. Over MCP a target is now a plain workspace name.
- The dashboard answered any Host and Origin, so another site's page could
  read the conversation or run a query through the browser (cross-site
  request, DNS rebinding). It now answers only under its own address, to its
  own origin, and takes a body only as JSON; `cx ui` warns when asked to
  listen beyond loopback.
- A ledger witness and a receipt's evidence counted as verified when the quote
  was not in the turn.
- Plan checkpoints written by units finishing at once left `plan.json`
  unreadable (157 of 400 forced rounds); the write is now atomic.
- The handoff schemas did not list `recall_ops`, so no provider held to its
  schema could use a typed recall verb.
- Two comments credited the Go episode routing with gains measured on a Python
  prototype.

Open, in rough order of harm:

- **Key views are wrong in ordinary cases.** Carding an older turn reopens a
  superseded value, so two values show as active. Extraction has no polarity
  ("do not use pool_size=999" becomes the value). Existing workspaces get no
  backfill, and the views answer "no recorded assertions" while reporting the
  index complete. The views have one time axis, not two.
- **Forks leak the future**: threads and entities are copied whatever the
  cutoff, and the status rebuild turns manually retracted records active.
- **The plan gate** refuses any plan of four or more units whose depth is one
  less than its size, a four-unit diamond included, on constants taken from a
  paper and fitted to nothing here. It and the new plan scheduler changed the
  default with no `cx compare` run. Decide: measure, or put it behind a key.
- **Resuming a count query** compares three sampled witnesses and reported a
  live record as removed.
- **Re-asking a receipt's own question marks it stale** (its words match its
  watched terms), and three of the predicate kinds it derives are never
  evaluated.
- **The body pack** is read by nothing, is not refreshed when a turn is
  appended, and returns re-sealed text as verified; a crafted directory panics
  its reader.
- **`cx explain` overstates**: `-drop` removes items from a built packet
  without re-selecting (it is not causal), the "of OPT" figure certifies an
  invented objective and not the packet beside it, cost is a price table times
  tokens with a fixed output size.
- **`cx tune`** picks its winner by held-out hits, so its held-out gain is
  biased; `cx merge` says "verified" when the repository has no verify command.
- **Direct API adapters** ignore a stream's error event and a stop for length,
  so a cut-off answer counts as complete.
- **The creativity study's headline** rests on 8 distinct problems rated by
  the model that wrote the answers, with about 54 uncorrected contrasts and an
  inter-rater agreement of 0.19 on originality in round 3. The counts are
  right; the README's conclusions are stronger than they support.

Measured leads in that work that are not in the default path yet: routed
plus query-grown packing (343 to 380 of 470 on LongMemEval-S with no losses,
in Python only), source-diverse packing alone (+12), and the finding that 108
of 127 misses have a needed message outside both rankers' top 24, so
candidate discovery and not ranking is the bottleneck there.

## Stage 3 (later)

- **Context compiler** (design: [CONTEXT-COMPILER.md](CONTEXT-COMPILER.md)).
  Paragraph-sized atoms with stable ids, levels of detail, a cache-stable
  layout and budgeted selection.
  - **Phase 1 is built** (2026-11-15): atoms, a stemmed atom index,
    `#aK` recall ids, ranked atoms as excerpts under a token budget, the
    whole history when it is short, `cx probe -granularity`. On 150 real
    probe runs (120 distinct questions, 60 written fresh for the test, on
    workspaces with no cards or threads) the answer's source was in the
    packet for 136 with atoms at the 8k default against 130 with six
    whole turns, in packets 6 to 44 percent smaller. The difference is
    within noise: no worse, for less. At a third of the ranked tokens
    atoms were three probes behind, so the sixfold saving of the first
    experiment did not carry over to fixed-budget packets
    (CONTEXT-COMPILER.md 2b).
  - **Phase 1 was reviewed** (2026-11-15) by independent agents that
    had to reproduce each defect: 58 findings, none refuted. Fixed: the
    atom index built with holes when two processes opened a workspace at
    once; the probe's search column counted any paragraph of the source
    turn (atoms looked level with whole turns and were 18 probes behind;
    a search now returns four paragraphs per hit and is level); the
    ranked section ran over its budget in every packet; a body of
    non-text bytes hung the splitter; excerpts cut from code blocks went
    unmarked; recalled ranges were sent two and three times; a plan sent
    the whole history once per unit. Not fixed yet, each listed in
    CONTEXT-COMPILER.md 2b: an excerpt can leave a code fence open, and
    reviewers cannot recall.
  - Next, in order: an end-to-end run with a model to see whether
    excerpts read as well as whole turns (model calls, ask first); the
    cache-stable layout and cost model (phase 2); atom titles and
    questions written at index time, which the card result argues for.
- **Probes on real transcripts: first run done (2026-11-15).** `cx import`
  reads Claude Code session files. On three imported sessions (140, 299
  and 201 turns, 30 probes each, written by a sub-agent that never saw the
  ranking code), the source turn was in the packet verbatim for 27, 26 and
  29 of 30. Every miss was a ranking miss. The value lane added one probe
  on two of the three and never cost one (details and caveats in
  ARCHITECTURE.md 3c). Open, each to be measured on fresh probes, not on
  these:
  - Done in Phase 1 above: the whole history when it is short, and a
    stemmed index (for atoms).
  - Run with the index built (`cx index`, model calls, ask first), so
    threads, facts, decisions and real cards are exercised.
  - Probes written by Tom, and a LongMemEval or LoCoMo conversation.
  - Compare against the same CLI with its own compaction.
- **Plan units return checked records** (built 2026-11-17). The first
  change from a survey of parallel multi-agent systems (2023 to 2026),
  whose clearest measured result is that a typed record per unit carries
  most of the gain of splitting work. Each unit hands back its output and
  the claims it relies on, with evidence ids, a quote and a support level,
  plus assumptions, open questions, conflicts and a `no_information`
  status. The engine checks every quote against the vault, pins it to its
  paragraphs, and lowers a claim it cannot check. The assembler and the
  reviewer read the checked records (ARCHITECTURE.md 3d).
  - **Also built (2026-11-17):** plans as a task list with `depends_on`
    and `shared_decisions`, run in waves of at most five, refused by rule
    when one unit or a plain chain (answered in one call), with coder
    units on one repository chained; a comparison of the records in code
    before the assembler writes; an assembler that accounts for every
    claim, checked by the engine and recorded in `assembly.json`; and
    `cx compare`, blind answers of single, plan and budget-matched arms
    on scratch copies.
  - **Measurement repaired (2026-11-18):** runs record cache reads, writes
    and reported cost; real calls rank exactly like `cx probe`; probes accept
    other wordings; `cx compare` has context arms; `cx floor` and `cx gaps`
    give the offline baseline (CONTEXT-COMPILER.md 2c).
  - Next: run `cx compare` on real messages (model calls, ask first). It
    decides whether the plan path is worth more work. Then, only if it
    is: rerun disagreeing units with each other's evidence, resume plans
    from stored records, and a scratch merge-and-verify across coder
    branches.
- Expose verbatim history as read-only files to agentic roles, sized by the
  watchdog rather than the brief.
- Seeded-defect controls for the reviewer (MAW proposal 5: recall of
  verifiers is unmeasured everywhere).
- `cx merge` with re-verification on the moved base (MAW integration step).
- A small HTTP or MCP front end so other tools can talk to a workspace.
- Usage-based throttling from the runs table instead of learning limits
  from 429s.
