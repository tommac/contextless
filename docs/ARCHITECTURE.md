# ContextlessAI architecture

ContextlessAI is a general chat assistant built on BookWriter v2's core idea:
**the model's context window is never the memory.** The conversation lives in
a vault of small records on disk. Every model call gets a fresh, sized packet
built from those records by deterministic code. Nothing is ever compacted,
and no record is ever summarized to make it fit.

It also takes MultiAIWorkspace's idea of per-provider lanes, so one request
can be split into units that run in parallel on claude, codex, kimi, gemini
(agy), grok and muse, each doing the kind of work it is measurably good at.

## 1. Principles (carried over from BookWriter, generalized)

1. **The packet is the context.** Each call runs in a packet directory the
   engine built for it. If a record is not in the packet, the model does not
   know it. Nothing is inherited from a previous session; every CLI runs with
   `--no-session-persistence` (or its equivalent) and compaction disabled.
2. **Size, do not compact.** The engine estimates every packet's token cost
   before it spawns a model and plans to 60 to 70 percent of that profile's
   window. If it does not fit it tries a larger window, then leaves records
   out by tier (listing each one it left out, by id), then splits the work
   into units. It never rewrites a record shorter.
3. **Verbatim is forever.** Every user and assistant turn is stored verbatim,
   append-only, one Markdown file per turn, committed to git. Any turn can be
   put back into any later packet word for word.
4. **Catalog cards are an index, not a replacement.** Each turn gets a
   one-line card (what it is about, its threads, its entities) tied to the
   sha of the turn body. Cards tell a model what exists so it can ask for it.
   A card never stands in for text that was selected: a turn ranked for the
   message goes in as exact excerpts (whole when all of it is picked), and
   its card stays in the catalog.
5. **Recall instead of memory.** A model that needs a turn it was not given
   answers `needs_recall` with the ids from the catalog or, when no catalog
   line points to what it needs, with search words. The engine searches
   every earlier turn, rebuilds the packet with the turns asked for and the
   paragraphs found pinned verbatim, tells the run what each search found and which ids do not
   exist, and runs again. A long turn can be recalled one part at a time.
6. **The engine never asks a model what to do next.** Control flow is code
   over schema-validated handoffs. A model can propose a plan (units with
   roles); code decides how to run it.
7. **Files are the truth for the conversation; SQLite is the truth for the
   machine.** Records are Markdown with YAML frontmatter in a git repository.
   `.cx/state.db` holds the index, full-text search, runs, and provider
   health. Models never write either directly; they return a handoff and the
   engine applies it.

## 2. The vault (one per conversation)

```
<workspace>/                         git repository
  session.md                         L0  purpose, standing instructions, preferences
  threads/th-*.md                    L1  topics / workstreams, open or resolved
  entities/en-*.md                   L2  people, projects, files, concepts
  facts/fa-*.md                      L2  stated facts with their source turn
  decisions/de-*.md                  L2  decisions with their source turn, supersedes
  cards/t000123.md                   L3  catalog card for one turn (sha-pinned)
  turns/t000123.md                   L5  the verbatim turn (immutable body)
  artifacts/<run-id>/...                 files produced by work units
  artifacts/<plan-run>/u1.json, u1.md    each plan unit's checked record (section 3d)
  artifacts/<plan-run>/assembly.json     how the assembler accounted for the units' claims
  .cx/state.db                       index, FTS, runs, provider health (gitignored)
<runs-root>/<workspace>/<run-id>/    outside the vault
  ws/brief.md ws/.cx/handoff.json    the packet exactly as the model saw it
  manifest.json                      every record sent: id, sha, tokens, tier, or "dropped"
  stdout.log stderr.log events.jsonl usage.json argv.json
```

Edges are stored on one side and indexed both ways:
`turn.threads`, `turn.mentions`, `fact.source`, `fact.entities`,
`decision.source`, `decision.threads`, `decision.supersedes`.

## 3. The life of one user message

1. **Append.** The user's message becomes `turns/tNNNNNN.md`, verbatim, and is
   indexed for full-text search at once.
2. **Select.** The packet is built by following edges one hop out from the new
   turn, in this order (tier = drop order, 0 is never dropped):

   | Selector | What | Tier |
   |---|---|---|
   | `session` | `session.md` | 0 |
   | `focus` | the new user turn, verbatim | 0 |
   | `recall` | turns the model asked for by id | 0 |
   | `decisions` | active decisions | 0 |
   | `recent` | the last K turns, verbatim | 4 (oldest first) |
   | `relevant` | paragraphs of earlier turns ranked for the focus (section 3c), verbatim excerpts within a token budget | 3 (least relevant first) |
   | `facts` / `entities` | linked to the focus's threads and mentions | 3 (after the ranked turns) |
   | `thread` | the other turns of the focus's threads, verbatim | 2 (oldest first) |
   | `catalog` | cards for every turn not sent; a long turn's card gives its size and parts | 1 |

3. **Size.** Tokens are estimated per profile (bytes times measured
   tokens-per-byte). The packet must fit the routed profile's plan cap. If not:
   try a larger profile allowed for the role; then drop tier 1, 2, 3, 4
   records and name every one under "Omitted for budget": a dropped turn
   by its card, a dropped catalog card in a digest line per block of up to
   50 turns that gives the block's thread, its id ranges and its most common
   keywords (built by code, no model); then, if the focus itself is too big,
   split it into chunk units.
4. **Route.** The role (`responder` by default) maps to a list of profiles in
   order. A provider in cooldown (quota, rate limit) is skipped. Lanes cap
   concurrent runs per provider.
5. **Run.** The CLI is spawned headless in the packet directory, model and
   effort pinned, compaction disabled, usage streamed, with the watchdogs from
   BookWriter (kill at 85 percent of the window, taint on any compaction
   event).
6. **Collect.** The handoff decides what happens next:

   | Status | Engine action |
   |---|---|
   | `completed` | store the answer as an assistant turn |
   | `needs_recall` | pin the requested turns, paragraphs or parts (`recall_ids`) and what the searches found (`recall_queries`), add a note of what each search found and which ids do not exist, run again (at most 2 rounds). Turns that do not fit together are refused with a note; the overflow is never taken for an oversized message |
   | `partial` | store what is done; a continuation run gets `units_remaining` and `next_step` |
   | `plan` | check the plan, run its units in waves across lanes; each returns a record the engine checks; compare the records in code; an `assembler` run merges them and accounts for every claim (section 3d) |
   | `blocked` | the answer is a question back to the user |

   A tainted run (compaction seen, peak over 90 percent) is never used; it is
   retried on a larger profile.
7. **Index** (in the background, off the critical path). An `indexer` run
   reads the new turns and returns ledger operations: thread assignment, new
   entities, facts and decisions, and a card per turn. The engine checks each
   quote against the verbatim turn before applying it and commits. The
   indexer is shown the facts in force that the new turns' words match; a
   fact that changes a value (a new port, date, price, time) supersedes the
   old one. Packets show the current fact with the one it replaced, and every
   fact and decision with the date of its source turn.

## 3a. After the answer

- **Review** (`review: plans` by default). A reviewer on a provider that
  wrote none of the answer gets the same packet plus the answer. Problems
  must quote the answer verbatim or name something missing; others are
  discarded. For an answer assembled from units, the reviewer and the
  reviser also get the unit records. On `revise`, the writer's role gets one revision run with the
  surviving problems as a supplied record. One round: a check, not a loop.
- **Failures** carry forward. When a run fails on one provider, the next
  attempt's brief lists what failed, by profile and exit class.
- **Recovery.** One process writes a workspace at a time (a lock file under
  `.cx/`). On open, runs a dead process left `running` are marked
  `abandoned`; a user turn left unanswered is answered by `cx retry`.

## 3b. The repository lane

A workspace can attach git repositories (`session.md` `repos:`), each with
a verify command and a path allowlist. A coder unit naming one runs like a
MultiAIWorkspace worker, reduced to one unit:

1. `git worktree add -b cx/<run> <packet>/repo <base>` from the repo's HEAD
   (or a named base), inside the packet directory.
2. The coder CLI runs agentically in the packet directory, sandboxed, with
   instructions to edit only `repo/` and never run git.
3. The engine stages everything and diffs against the base (so a commit the
   agent made anyway is still checked), rejects the unit if any path falls
   outside the allowlist (or is under `.git`), and commits once.
4. The verify command runs in the worktree, outside the sandbox.
5. The worktree is removed; the branch is kept only when it holds a change.
   The report names the branch, commit, files and verify result. Nothing is
   merged: that is the user's call.

A unit's recall rounds and continuations build on its own previous commit.

## 3c. Ranking, recall and measurement

The ranked unit is the **atom**: a paragraph-sized piece of a turn (about
60 to 720 tokens), cut by code at blank lines. A code block is not cut at
its blank lines; short blocks join the next, and a heading or label opens
the atom of its section; anything over 2,400 bytes, a long code block or
table included, is cut at a blank line, a line end or a sentence end. An
atom is an address into the immutable turn, `tNNNNNN#aK`, never a
rewrite; a turn's atoms tile its body. The ranked section of a packet is
filled with the best atoms until their text and headers reach
`relevant_tokens` (8,000 by default, counted at 0.30 tokens per byte).
Adjacent atoms of a turn are sent as one excerpt, headed by the turn, its
paragraph numbers, its recall id and the opening of the message it
answers; a turn picked in full goes in as the turn, and a header says when
an excerpt begins or ends inside a long code block. A paragraph whose
exact text was already picked from another turn is not picked again. A
turn that matched only through its card goes in whole if it fits the
budget, else it is left to its catalog line. A turn already in the packet
through the focus's threads that holds a ranked atom is kept whole as a
ranked turn (tier 3) and is not charged to the budget. The turn an
excerpt comes from stays in the catalog, and more of its paragraphs can
be recalled as `tNNNNNN#aK-M`; paragraphs recalled in several ranges are
joined, and no paragraph is sent twice. On real conversations, imported
without cards or threads, this found slightly more probe answers than six
whole turns in packets 6 to 44 percent smaller (CONTEXT-COMPILER.md,
section 2b). `relevant_tokens: -1` goes back to ranking whole turns, and
then no atom index is built.

While the earlier turns total at most `whole_history_tokens` (60,000 by
default, at the first profile's token density) and fit the role's first
profile, all of them are sent verbatim in order, nothing is ranked, and
the packet says it is complete. When the whole transcript fit, it held
every probe's source, more than any selection did; whether models answer
better from it has not been measured. Plan units and their assembler always get a selected
packet. The whole-history attempt never moves a call to a larger
provider: a history that does not fit the first profile whole is selected
instead. A selected packet is still sized by principle 2, so it can move
to a larger window before any record is left out.

The atom index is a cache in `state.db`, rebuilt by any process that
opens the workspace and finds turns missing from it (a workspace made
before atoms, or a change of the splitting rules); several processes can
do so at once. A writing open also re-indexes a turn whose atoms were
indexed without its card.

Ranking is `packet.RankAtoms` (and `packet.RankTurns` for whole turns),
the same ranking a recall search uses. Ranked lanes over the verbatim
text are fused with Reciprocal Rank Fusion (k = 60, LANTERN's choice),
and each lane's top hit always makes the cut:

- **exact**: the whole query occurs in the turn, newest first, so a value
  that changed is found before the value it replaced;
- **full text**: bm25 over every word (SQLite FTS5). The atom index is
  stemmed, so "rejected" meets "rejections", and each atom is searched
  with its turn's card text, so a question in the card's words reaches
  the turn's paragraphs;
- **value**: the turn states a value (a number, date, time, amount, code
  identifier, path or name) within four words of a query word in the same
  sentence; the best single statement decides, ties go to the full-text
  rank. LANTERN found config values, names and decisions the facts most
  often lost; their questions share only topic words with the answer.
  Among whole turns this is a full lane. Among atoms only its two best
  hits are fused: as a full lane it cost answers on real conversations,
  because a number near a topic word is common among thousands of short
  paragraphs.

A turn longer than 24 KB has addressable parts, `tNNNNNN#pK`, cut at
paragraph boundaries (deterministic, lossless). Its catalog line gives its
size and part range. A recall search returns the paragraphs that match,
not the turn or the part around them: four paragraphs for each turn it
would have returned whole, because one paragraph per hit held the answer
far less often than a whole turn (108 against 126 of 150 real probes;
127 at four).

**Probes** measure this layer without a model, after LANTERN's recovery
probes: each probe is a question with a known source turn and answer, and
`cx probe <ws> probes.jsonl` reports per fact type whether the source is in
the packet verbatim, named by its own line, only inside a digest range, or
missing, and whether a recall search for the question puts back text
that holds the answer. On the
synthetic 800-turn conversation in `probe_test.go` (about 620k tokens,
twice the largest test window, five threads whose filler shares the
questions' vocabulary and numbers), the source turn of 13 of 14 planted
facts is in the packet verbatim (3 of 14 before the ranking and tier
change), and a search for the question returns the answer for 12 with
whole turns as the ranked unit and 13 with atoms. The miss asks for a
value that was later changed; the fact ledger's supersession answers that
one.
These numbers come from synthetic text, so they are a regression floor,
not evidence about real conversations.

**On real transcripts** (2026-11-15). Three Claude Code sessions were
imported with `cx import` (user prompts and assistant text; tool output
left out), and 30 probes per conversation were written by a sub-agent that
read the turns but not the ranking code. `cx probe -lanes` compares lanes on
the same probes. Source turn in the packet verbatim:

| Conversation | Turns | History (est.) | Full text only | All lanes | 10 ranked slots |
|---|---|---|---|---|---|
| one BookWriter session | 140 | about 80k tokens | 26 of 30 | 27 of 30 | 28 of 30 |
| that session plus the project's 8 later sessions | 299 | about 170k | 26 of 30 | 26 of 30 | 28 of 30 |
| a math research session (Conway's 99-graph) | 201 | about 290k | 28 of 30 | 29 of 30 | 29 of 30 |

- Names and values were never missed. The five distinct misses were two
  decisions, two numbers and a branch name.
- Every miss is a ranking miss, not a coverage miss: the source was named
  in the packet by its catalog line (recallable by id) but ranked 7th to
  48th for six slots, and a search for the question's words did not find
  it in the top five either. The causes are paraphrase without stemming
  (the turn index used the default tokenizer, so "rejected" did not match
  "rejections"; the atom index is now stemmed), questions that restate the fact in other words ("twenty
  minutes" for 13:34Z to 13:54Z), lowercase names the value lane does not
  treat as values, a colon read as a sentence end that cut a number off
  from its topic, and topic words shared by 20 to 80 other turns.
- The value lane and the fusion weights were tuned on the 14 synthetic
  probes, so this is their first held-out test. The value lane added one
  probe on the single BookWriter session and one on the research session,
  nothing on the full BookWriter history, and never cost one. There, one
  source ranked 2nd in the value lane still finished 9th after fusion. The exact lane changed nothing,
  as expected for paraphrased questions.
- The default route and a claude-only route gave identical packets. The
  packet stays at 55k to 65k tokens because only recent and ranked turns
  go in verbatim, so sizing and escalation were not exercised.
- On the 140-turn session the whole history (about 80k tokens) would have
  fit claude's 140k plan cap; sending all of it would have put every
  source in the packet. Selection that leaves turns out when everything
  fits loses to a plain transcript.

Caveats: the workspaces were not indexed (no model was run), so the
thread, fact and decision selectors were not exercised and catalog cards
were preview lines. A model wrote the questions after reading the source,
and full text alone finds 87 to 93 percent, so most questions share rare
words with their answer; vaguer questions would score lower. With 30
probes, one probe is 3.3 points. Imported turns carry the import date, not
the date they were said.

## 3c-2. The map of the conversation

A workspace can hold idea notes (`notes/<id>.md`): one thing the conversation
says, in plain words, with the quotes that state it. A note is written by a
model (the indexer with `atlas: on`, or a separate pass imported with
`cx notes <name> import`) and trusted for nothing: it is stored only if code
finds a quote in the turn it cites, and it records the byte span, the
paragraphs that hold it and the hash of the body it was found in. Its id is a
hash of the title and spans.

A selected packet shows the notes as a map, one line each
(`- [t000012#a3] (decision) title`), after the decisions in force and before
anything chosen for the message. The section says a note points at text and
is not the text, and that the paragraphs an answer rests on are to be asked
for with `needs_recall`. Rules of the section:

- A note appears only through evidence before the message being answered, on
  turns whose body still has the hash that was checked.
- While all lines fit `atlas_map_tokens` (40,000, and never more than a
  quarter of the largest window's plan cap) the section is the same for every
  message and grows only at its end.
- Past that, the lines ranked for the message are shown and every other note
  is counted under its topic with its turn range; the recall op `notes` lists
  a topic's lines or the best matches of some words.
- While a map is shown, a recall request is bounded by tokens
  (`atlas_pull_tokens`, 8,000): ids are served in the order asked, one that
  does not fit is passed over for a smaller one, and the next packet names
  what was left out.
- A packet that holds the whole conversation has no map, and neither does an
  index run.

Why a map, and what is measured about it, is in CONTEXT-COMPILER.md section
2e. The handoff schemas now list `recall_ops`: before, a schema with no such
property and no additional properties allowed forbade the typed recall verbs
the instructions offered.

## 3d. Unit records

A plan unit, and each part of an oversized message, hands back a record
rather than free text (`packet.UnitSchema`):

- its output;
- the claims the output relies on, each with evidence (turn, paragraph
  or part ids, a file the unit wrote, or a web address), a quote copied
  from that evidence, and a support level: quoted, inferred or
  unsupported;
- the choices it made that another unit could have made differently;
- what it could not settle, and evidence that disagrees with it;
- notes.

A unit that finds nothing for its task says so with `no_information`
instead of inventing an answer.

The engine checks every claim before the assembler reads it:

- An evidence id must name stored text at or before the message being
  answered.
- A quote must occur in the text it cites, whitespace aside. A quote that
  does is pinned to the paragraphs that hold it (`t000123#a4`), also when
  the unit cited the whole turn. A quote found elsewhere in a cited turn
  is pinned where it is, with a note.
- A quoted claim whose quote is in none of its texts, or that gives no
  quote, counts as inferred. A claim with nothing the engine can resolve
  is unsupported. Web sources are marked as not checked, since the engine
  does not fetch. The check only ever lowers a claim.

A coder unit's record carries each run's repository change as fields
(base, branch, commit, files, verify result), and every file a unit wrote
carries its sha.

Records are stored with the plan as `artifacts/<plan-run>/<unit>.json`
and `.md` (the parts of a split message under `artifacts/<turn>-parts/`)
and committed with the vault. Each plan logs how many claims were checked
and how many quotes were found. The assembler reads every record verbatim,
with the engine's check beside each claim; a `no_information` unit is one
line with what it looked for. The assembler is told to rely on claims
whose quote was found and to say where units disagree or assumed
differently. The reviewer reads the same records, or the answer alone when
they do not fit its window, so it can check the join: a unit result the
answer drops, an unchecked claim stated as fact.

**Plans.** A plan is a task list. Each unit has an id, a role,
instructions, `needs` (turns to see verbatim) and `depends_on` (earlier
units whose records it needs); the plan's `shared_decisions` (names,
interfaces, formats, scope) go into every unit's packet, since the request
alone does not stop parallel units from deciding the same thing
differently. Units run in waves: every unit whose dependencies are done
runs in the next wave, at most five at once, and a dependent unit gets its
upstream records verbatim. The engine checks the plan by rule:

- a dependency on a later or unknown unit is dropped, so a plan is
  acyclic;
- coder units changing the same repository are chained in plan order,
  each starting from the previous one's commit;
- a plan of one unit, or of units that would only run one after another,
  is not run (sequential work is where every measured multi-agent
  topology lost to a single agent): the message is answered in one call,
  told why, and offered no plan. A coder unit changing a repository runs
  as a plan whatever its shape.

**Compare in code, then assemble.** Before the assembler writes, code
compares the records. It leaves out units that found nothing or failed,
lists claims of different units whose quotes lie in overlapping paragraphs
of one turn (marking pairs that share topic words but state different
values), pairs assumptions of different units that may be the same choice,
and collects open questions and conflicts. It decides nothing; it shows
where to look. The assembler reads this comparison, then the records, and
must account for every claim: `claims_used` for what its reply relies on,
`claims_omitted` with a reason for the rest. A reply that leaves claims
out of both lists, or names ids that do not exist, is run once more with
the engine's check. What is still unaccounted for is recorded, not hidden,
in `artifacts/<plan-run>/assembly.json`, with the counts of shared
evidence and possible disagreements, so how often a merge drops a result
can be measured. The reviewer gets the comparison too.

**Measuring.** `cx compare <ws> <messages>` answers each message once per
arm: a single call, a plan, and optionally a single call on a profile you
name (`-matched`, the budget-matched arm, since unmatched comparisons
flatter multi-agent systems). Every trial runs on its own scratch copy of
the workspace, so the arms see the same history and the conversation is
untouched. The answers are written blind (`answers.md`, letters in an
order set per message) with a key of tokens, time, runs and units
(`key.json`) and a cost summary by arm. It spends model calls, so it
starts only with `-yes`; nothing has been measured with it yet. The context
arms `current`, `full`, `bm25`, `recent` and `turns` compare what a packet's
history is made of, each as a single call. Every run records uncached input,
cache reads, cache writes, output and any reported price separately.

Not built yet: rerunning units that disagree with each other's evidence
pinned; resuming a plan from stored records; a check that merges several
coder branches on a scratch branch and runs verify once.

## 4. Roles and routing

Roles are seeded from BookWriter's measurements on a book-length manuscript
and are config, not code:

| Role | First choice | Why |
|---|---|---|
| `responder` | claude opus | conversation and prose |
| `coder` | codex gpt-6.1-sol | highest integration rate on structured edits and fixes |
| `indexer` | codex medium, then claude haiku | 61 of 61 summaries and 130 of 130 annotations integrated |
| `long_reader` | kimi k3.1-pro (1M), then agy gemini (1M) | whole-corpus reads |
| `researcher` | grok | the only networked role |
| `reviewer` | any provider except the writer's | independence rule |
| `editor` | muse, then claude sonnet | line editing |
| `assembler` | claude opus, then codex | merges unit records into one answer |

## 5. What came from where

| Piece | Source |
|---|---|
| Provider adapters, sandbox, watchdogs, compaction switches, usage parsing | BookWriter `internal/bw/providers` (copied whole) |
| Packet cost model, plan/kill/taint fractions, drop tiers, "Omitted for budget" | BookWriter `internal/bw/packet` (re-implemented for chat records) |
| Layered records with one-sided edges, sha-pinned summaries | BookWriter vault design (docs/v2/02, 03) |
| Handoff with `units_done`, `units_remaining`, `next_step`, partial runs, continuations | BookWriter handoff contract |
| Per-provider lanes, cooldowns as data, interleaving across providers | MultiAIWorkspace lanes and reservations |
| Reviewer never on the writer's provider | BookWriter judge independence, MAW maker/checker |
| Recall by search over the whole store | LCM (Voltropy, 2026) `lcm_grep` |
| Digest lines over blocks of omitted turns | LCM's summary DAG, built by code as in ARC |
| Parts of a long turn, notes for bad ids, refusing recall that does not fit | ARC (Fujitsu, 2026): chunked recall, nearest-match answers, recall budgets |
| Fused lanes with RRF, the value lane, recovery probes | LANTERN (Cisco, 2026) and its failure analysis |
| Facts that supersede facts, dated sources | CogCanvas (2026) and its stated gaps; LongMemEval's knowledge updates |
