# ContextlessAI

ContextlessAI (`cx`) is a command-line program, written in Go, for long conversations with language models. It keeps every message word for word in a local Git repository and never summarizes old messages to make room. Before each model call, Go code builds a packet from the stored text: the paragraphs that match the question, the decisions still in force, and a list of what was left out. A model that needs something that was left out asks for it by id and gets the original text.

It can also split a request across several vendors' command-line tools (Claude, Codex, Kimi, Gemini, Grok and Muse), check every quoted claim against the stored text, have a model from a different vendor review the answer, and make code changes in a separate Git worktree.

This is one author's research system, not a product. The design is in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) and [docs/CONTEXT-COMPILER.md](docs/CONTEXT-COMPILER.md). Plans and known defects are in [docs/ROADMAP.md](docs/ROADMAP.md).

## What is shown, and what is not

Read every claim in this file at the level of evidence given here.

| Claim | Evidence | Level |
|---|---|---|
| A model that reads a map of checked notes and asks for paragraphs answers better than one given the same tokens of search results | Two blind-judged answer tests (30 and 120 held-out questions). The second was confirmatory, with its rule fixed first | Measured on answers, two private conversations |
| The engine's packet with the map answers better than the same engine without it | Confirmatory test, 120 questions: 0.882 against 0.706 | Measured on answers. The packets differ in size; see the limits |
| A small exact packet beats sending the whole conversation | Both tests: +0.161 and +0.172. Described, not the contrast the rule tested | Observed twice, not a declared test |
| Paragraph-sized pieces put the answer in the packet about as often as whole turns, in smaller packets | 150 probe runs, no model | Retrieval only |
| Session routing and reply adjacency raise source coverage on LongMemEval-S and LoCoMo | Replay scripts on public data, no model | Retrieval only. The scripts are Python, not the Go engine's port |
| Quote checking, independent review, worktree isolation, sizing and omission accounting behave as specified | The test suite, with scripted stand-ins for the model tools | Behaviour tested. Effect on answer quality not measured |
| This is better than a tool that compacts | None yet | Not measured. No head-to-head run exists |

The answer tests used two private conversations. Their questions, packets and transcripts are not in this repository, so an outside reader cannot rerun them yet. The retrieval measurements and the replays use public data or synthetic fixtures and can be rerun.

## How this repository was written

The code and most of the documents were written with AI coding agents, directed and reviewed by the author, in November 2026. That is a reason to check the claims. This is what there is to check them with:

- `go test ./...` runs the engine against scripted stand-ins for the model tools. It makes no model call and needs no key. CI runs it on Linux and macOS.
- Each number under [Measurements](#measurements) names the script or report it came from.
- For both answer tests the questions, the arms and the pass rule were written down before the run. The first test's rule failed, and that result is reported first.
- The code added on 2026-11-24 was reviewed twice, independently, and each review had to reproduce the defects it reported. The fixes have tests. The defects still open are under [Known defects](#known-defects).

Some folders hold reading material written with model assistance. They are sources and data, not findings: `research_notes/`, the two literature reports in `reports/`, the notes in `idea-vault/`, and the design proposals in the Obsidian study, which are that study's data. The [repository map](#repository-map) says which is which.

## Run it without a model

You need macOS or Linux and Go 1.26 or later. The replay scripts need Python 3.9 or later. The binary is pure Go, with no cgo.

```sh
git clone https://github.com/tommac/contextless
cd contextless
make demo                   # builds bin/cx, then walks the engine through a scripted conversation
go test ./... -count=1      # the whole suite: no model calls, no keys
```

To install the binary alone, without the repository:

```sh
go install github.com/tommac/contextless/cmd/cx@latest
```

The replays below also start no model. The last one downloads three public Go API files.

```sh
# A scripted conversation in which a meeting is moved, with quote checks
python3 examples/replay/run.py

# One million events: a count that is resumed after corrections, and withheld when a record cannot be read
python3 docs/experiments/coverage/bench.py --records 1000000

# One million omitted turns named through a paged catalog
go run ./docs/experiments/packet_scale -catalog

# Which stored answers change between Go 1.24, 1.25 and 1.26 (1,128 queries)
api=$(mktemp -d)
for v in 24 25 26; do
  curl -fsSL "https://raw.githubusercontent.com/golang/go/go1.26.0/api/go1.$v.txt" -o "$api/go1.$v.txt"
done
python3 examples/answer_watch/replay.py "$api"
```

The LongMemEval-S, LoCoMo, ContextBench and value-postings measurements are rerun from the upstream datasets. The steps are in [docs/experiments/longmem/README.md](docs/experiments/longmem/README.md) and [reports/2026-11-20-large-memory/README.md](reports/2026-11-20-large-memory/README.md).

## Before you run it with real models

The demo, the tests and the replays above start no model. The commands that call a model (`chat`, `say`, `retry`, `index`, `compare -yes`, and anything driven through `cx serve -mcp` or `cx ui`) are different:

- **Permission prompts are bypassed.** A run with no terminal cannot answer a prompt, so each adapter passes the vendor tool's own bypass switch (`--permission-mode bypassPermissions` for `claude`, `approval_policy="never"` for `codex`, `--approval-mode yolo` for `gemini`, and the equivalents for the others). Inside its run a model can read, write and execute whatever the process is allowed to.
- **macOS: a sandbox contains each run, on by default.** The Seatbelt profile lets a run write only to its packet and run directories, read the system and the tool's own configuration, and reach the network only through the hosts its vendor needs. The `researcher` role is given the open network.
- **Linux: there is no sandbox.** `cx` refuses to start a vendor tool and says why. It runs only after you set `sandbox: false`, and then the tools are unconfined: a model can touch anything your user account can. `cx` prints a warning each time it runs this way. Use a container, a virtual machine or a throwaway account.
- **Windows is not supported.** The code does not build there.
- **Your conversation is sent to the vendors you enable.** Every packet holds text from the workspace word for word. Restrict vendors with `enable:` in the configuration, and keep anything you would not send to a vendor out of the workspace.
- **Code changes stay in a worktree.** A coder edits a fresh `git worktree`, a change outside the path allowlist is rejected, and nothing is merged for you. The verify command you configure runs outside the sandbox.

[SECURITY.md](SECURITY.md) has the same list, with the dashboard and the MCP server, and says how to report a vulnerability.

## Using it

```sh
bin/cx init-config > ~/.contextless/config.yaml   # optional: look at or change the model lists

# A new workspace (a Git repository under ~/.contextless/workspaces/), then a chat
bin/cx new arch-review -purpose "Storage engine design and benchmarks"
bin/cx chat arch-review

# Import a Claude Code session file, a JSONL file or a Markdown transcript, word for word
bin/cx import arch-review ~/transcripts/session.jsonl

# See the packet a question would get, and why each part is in it (no model call)
bin/cx packet arch-review -brief "What did we decide about WAL group commit latency?"
bin/cx explain arch-review "What did we decide about WAL group commit latency?"

# Import notes, list them, and try a request for paragraphs by hand
bin/cx notes arch-review import notes.jsonl
bin/cx notes arch-review
bin/cx packet arch-review -pull t000012#a3,t000040#a1-2 -ask "cache policy;eviction" "How did the cache policy change?"

# Let coder runs change a repository, inside an allowlist and behind a test command
bin/cx attach arch-review storage ~/code/storage -verify "go test ./..." -allow "internal/**,cmd/**"

# Serve the workspace to other tools over MCP, or open the local inspection page
bin/cx serve arch-review -mcp
bin/cx ui arch-review -addr 127.0.0.1:8787
```

| Area | Commands |
|---|---|
| Conversation | `new`, `chat`, `say`, `retry`, `show`, `import`, `list`, `fork` (`-as-of` is experimental) |
| Packets and notes | `packet` (`-brief`, `-adaptive`, `-pull`, `-ask`, `-json`), `explain`, `notes`, `index`, `reindex`, `pack` |
| Stored queries | `query` (`-strict`, `-as-of`, `-resume`), `views` (experimental), `watch`, `stale`, `receipt`, `refresh` |
| Measurement | `probe`, `floor`, `zoom`, `pull`, `compare` (`-yes`), `cache`, `gaps`, `tune`, `runs`, `stats`, `demo` |
| Repositories and other tools | `attach`, `merge`, `resume`, `providers`, `serve -mcp`, `ui`, `obsidian export\|brief` |

## How it works

### What is stored

```text
~/.contextless/workspaces/<name>/        a Git repository
├── session.md          standing instructions and attached repositories
├── turns/tNNNNNN.md    every message, word for word, never rewritten
├── cards/tNNNNNN.md    a one-line catalog entry per turn, pinned to the turn's hash
├── decisions/de-*.md   decisions in force and the ones they replaced, each with a quote
├── facts/fa-*.md       stated facts, each with a quote, and which fact replaced which
├── notes/no-*.md       notes for the map, each kept only through quotes found in the turns
└── .cx/state.db        SQLite search index, rebuilt from the files above
```

The engine keeps five rules:

1. A turn is stored word for word and is never shortened or rewritten.
2. A packet is sized before the model starts, to 60 to 70 percent of the model's window. If it does not fit, the engine moves to a model with a larger window (200k, then 500k, then 1M tokens) before it drops anything. It drops in a fixed order: the catalog, then thread turns, then ranked paragraphs and facts, then the oldest recent turns. Session notes, the question, recalled text and active decisions are never dropped.
3. Everything left out is named: by its card, by a digest line for each block of 50 turns, or through a paged catalog file.
4. Records written by a model (cards, threads, facts, notes) are pointers and search keys only. Evidence is always sent as exact bytes of a turn, and a quote written by the indexer is checked against the turn before it is saved.
5. The vendor tools run with session persistence and compaction turned off. A run that reports compaction, or passes 85 to 90 percent of its window, is thrown away and retried on a larger model.

### Paragraph addresses

Each turn is cut into paragraph-sized pieces, called atoms, at blank lines outside code blocks. Short pieces are joined up to 200 bytes. Pieces over 2,400 bytes are split at blank lines, line breaks or sentence ends. A piece has an address such as `t000012#a3`, a range is `t000012#a3-5`, and joining a turn's pieces gives the turn back byte for byte. Turns over 24 KB also have coarser parts, `t000012#p2`.

When the whole earlier history is under 60,000 tokens and fits the model, nothing is selected: the history is sent whole, in order.

### Ranking

Candidates come from three searches. None uses embeddings or a model call.

- **exact**: phrase match, newest first, so the latest value of a setting comes before the ones it replaced.
- **full text**: SQLite FTS5 with Porter stemming and BM25, over the paragraph together with its turn's card.
- **value**: numbers, dates, paths and identifiers within four tokens of a query word in the same sentence. For paragraphs only its top two hits are used, because number-near-topic matches are common across thousands of short paragraphs.

The three lists are combined by reciprocal rank fusion: a paragraph's score is the sum, over the lists it appears in, of `weight / (60 + rank)`. Each list's top hit is kept. Selected paragraphs from the same turn are merged when adjacent, carry a header (turn id, span, speaker, and the start of the prompt being answered), and are packed up to 8,000 tokens by default.

Four extras are off by default: session routing (`-session-route`), reply adjacency (`-auto-reply`), a smallest window that covers several query terms, and a two-hop walk over rare identifiers. Their only measurements are packet coverage from the replay scripts. None has an answer test, and the Go versions of session routing and reply adjacency have not been benchmarked.

### The map of notes

Search does badly on a question whose evidence is spread over many turns, such as "how did the cache design change, and why?". For those the packet can carry a map, which the code calls the Atlas:

1. A note is a short statement with a kind, topic tags and quotes. The engine keeps a note only through the quotes that code finds in the cited turn, ignoring whitespace, and records the byte range of each.
2. When notes exist, the packet lists them after the active decisions, one line per note with its paragraph address, up to 40,000 tokens. If the map does not fit, the lines closest to the question are shown and the rest are counted per topic.
3. The model is told that the lines are pointers, not evidence. It asks for the paragraphs it wants, and the engine sends them word for word, up to 8,000 tokens a round.

The results are under [Measurements](#measurements).

### Order of a packet

Vendors bill a cached prompt prefix at a fraction of the input price, so cutting tokens and cutting cost are different things. A packet is therefore ordered from what changes least to what changes most: session notes, decisions and facts, the map and the catalog, then ranked paragraphs, recent turns and the question. Run ids and timestamps go last. `cx cache` and `cx gaps` estimate prefix reuse and cost from recorded runs. Whether the vendor tools reuse the prefix in practice has not been measured, so the cached figures are an upper bound on the saving.

### Several models, checked claims and review

A request that splits into parts runs as a plan. Each part goes to a role, and each role has an ordered list of models:

| Role | Default order |
|---|---|
| `responder` | `claude` (Opus 5.5), `codex` (GPT-6.1-Sol), `kimi` (K3.1-Pro), `agy` (Gemini 4 Argon) |
| `coder` | `codex`, `claude` (Opus), `kimi` |
| `researcher` | `grok` (4.7, with network), `codex` |
| `long_reader` | `kimi` (1M window), `agy` (1M window), `codex` |
| `editor` | `muse`, `claude` (Sonnet, then Opus) |
| `reviewer` | `codex`, `claude` (Opus), `kimi` (Highspeed), `muse`, minus whoever wrote the draft |
| `assembler` | `claude` (Opus), `codex`, `kimi` |
| `indexer` | `codex` (medium effort), `claude` (Haiku), `kimi` (Highspeed) |

A part does not hand back free prose. It returns a record: its claims, each with evidence ids, a quote and a stated level of support, plus its assumptions, open questions and conflicts. Before the parts are combined, code looks up every quote in the stored text, records where it was found, marks a claim down when its quote is missing or not found, and flags parts that contradict each other. Where a checked quote meets an unchecked statement, the quote wins. The combined answer has to account for every claim.

An answer built from several parts, or every answer under `review: always`, is then read by a model from a vendor that wrote none of it. A criticism is dropped unless it quotes the draft exactly or names a missing requirement. What survives leads to at most one revision.

A lock shared by every `cx` process on the machine limits concurrent runs per vendor. A vendor that reports a rate limit is rested, and the next model in the list is tried with a note of what failed.

Whether plans or review improve answers has not been measured. `cx compare` exists for that and has not been run on real messages.

### Code changes

`cx attach <workspace> <repo> <path> -verify "<command>" -allow "<globs>"` lets coder runs change a repository:

1. Each coder works in a new `git worktree` on a branch named `cx/<run>`. The agent may not run `git`.
2. After the run the engine diffs against the base commit, rejects the run if any changed path is outside the allowlist or under `.git`, commits, and runs the verify command outside the sandbox.
3. Your working tree is not touched. `cx merge` runs the verify command again on the current `HEAD` before merging. It merges as you, so your git identity has to be set.

### Other parts

These were added late, none has a retrieval or answer measurement, and several have defects listed below.

- `cx explain` shows, for a question, which search proposed which paragraph, how the token budget was split, and how much of the prefix would be reused.
- `cx query`, `watch`, `stale`, `refresh`, `views` and `fork` store a question with the condition its answer depends on, so that a later change can reopen it, and answer simple count and history questions by code.
- `cx serve <workspace> -mcp` is an MCP server over stdio with 12 tools. `cx ui` is a local page for inspecting packets, limited to its own address and origin. `cx obsidian` exports a workspace to an Obsidian vault and builds a packet from a canvas. There are also body packs (`cx pack`), weight tuning on probes (`cx tune`), and opt-in HTTP adapters for three vendor APIs.
- `apps/macos` is a small SwiftUI window over the `cx` binary.

## Measurements

Two kinds of number appear below and they are kept apart. Packet coverage asks whether the needed text reached the prompt. Answer quality asks whether a model then stated it.

### Answer test, 30 questions (2026-11-25)

The questions, the arms and the pass rule were fixed in a private runbook on 2026-11-18, a week before the run. This was not registered with a third party ([docs/CONTEXT-COMPILER.md section 2e](docs/CONTEXT-COMPILER.md)).

There were 30 held-out questions, each needing passages from several places in one of two long conversations: 18 on a conversation never used for tuning, and 12 newly written. Claude Opus 5.5 at medium effort wrote the answers. Two judges, the answering model and a second model, gave 965 passage verdicts with the arm labels hidden and the order counterbalanced. They agreed on 0.936 of verdicts (Cohen's kappa 0.889). Of 255 model runs, none was dropped.

| Arm | What the reader was given | Share of needed passages stated | Mean input tokens |
|---|---|---:|---:|
| `none` | The question only (negative control) | 0.000 | 130 |
| `search` | Ranked paragraphs, default budget | 0.314 | 7,961 |
| `searchbig` | Ranked paragraphs, as many tokens as the map arms | 0.546 | 34,776 |
| `catpull` | A code-built line per turn, then the paragraphs the model asked for | 0.545 | 23,111 |
| `pull` | A map of checked notes, then the paragraphs the model asked for | 0.726 | 31,001 |
| `full` | The whole conversation (15 of the questions) | 0.839 | 247,856 |
| `oracle` | Only the exact needed paragraphs (positive control) | 0.997 | 1,209 |

- **The one hypothesis the rule tested failed.** A code-built line per turn plus the model's own request (`catpull`) did not beat search of the same size (`searchbig`): difference -0.001, 95% interval -0.125 to +0.122; 12 won, 12 lost, 6 tied.
- **Checked notes read as a map did better, in a comparison the rule had not named.** `pull` beat `searchbig` by +0.180 (interval +0.080 to +0.281; 18 won, 5 lost, 7 tied; adjusted p 0.032) and `catpull` by +0.181 (21 won, 7 lost; adjusted p 0.032). It was within noise of the whole conversation at one eighth of the tokens (-0.070, p 0.29, on 15 questions). Because the rule did not test this, it was a lead until the second run, which found half the effect.
- **The exact paragraphs beat the whole conversation.** 1,209 tokens of the needed text scored +0.161 over 247,856 tokens of everything (interval +0.059 to +0.286; 8 won, 0 lost). Sending the whole conversation is not the ceiling. The run does not show why.
- **A pointer was worth about a third of the passage.** A needed passage that was in the packet word for word was stated in 0.92 to 1.00 of verdicts (0.83 when the whole conversation was sent). One that a note line pointed to, without its paragraph, was stated in 0.33 to 0.38. One neither present nor pointed to was stated in 0.09 to 0.28.

### Confirming test, the other 120 questions (2026-11-25)

The rule was written before any packet or answer of this run existed ([docs/CONTEXT-COMPILER.md section 2f](docs/CONTEXT-COMPILER.md)). It named two tests, adjusted together for multiple comparisons (Holm). The first repeats `pull` against `searchbig` with the same binary, packets and prompts. The second sets the engine's own packet with the map against the same engine with `atlas: off`, each allowed one request for more text. The answering model and judges were the same as before. There were 810 model runs with no errors, and Cohen's kappa was 0.840.

| Arm | What the reader was given | n | Share of needed passages stated | Mean input tokens |
|---|---|---:|---:|---:|
| `none` | The question only (negative control) | 30 | 0.000 | 130 |
| `searchbig` | Ranked paragraphs, as many tokens as the note lines | 120 | 0.672 | 34,768 |
| `livenomap` | The engine's packet, `atlas: off`, one request | 119 | 0.706 | 28,984 |
| `pull` | Note lines, then requested paragraphs | 120 | 0.762 | 31,415 |
| `full` | The whole conversation | 30 | 0.810 | 247,854 |
| `livemap` | The engine's packet with the map, one request | 119 | 0.882 | 58,517 |
| `oracle` | Only the exact needed paragraphs (positive control) | 30 | 0.982 | 1,214 |

- **The earlier lead held, at half its size.** `pull` beat `searchbig` by +0.090 (interval +0.037 to +0.141; 57 won, 30 lost; adjusted p 0.005).
- **The map helped the engine.** `livemap` beat `livenomap` by +0.175 (interval +0.137 to +0.215; 66 won, 4 lost; adjusted p below 0.0001). On the 29 questions that also have a whole-conversation answer it was level with `full` (+0.035, interval -0.039 to +0.118) at under a quarter of the tokens.
- **Where the gain came from.** Before any request, both engine packets held 0.327 of the needed passages word for word. After one request the packet with the map held 0.840 and the one without held 0.614.
- **Limits.** The map packet is twice the size of the map-off packet, and no arm fills those tokens with search results. The notes came from separate writers, not the engine's indexer. The request was offered to the model as a step. Both conversations are one author's. All 150 held-out questions are now spent.

### Retrieval only, on public data

These rows count whether source text reached an 8,000-byte packet. No model ran. They come from the Python replay scripts in `docs/experiments/`, not from the Go engine ([reports/2026-11-20-large-memory/](reports/2026-11-20-large-memory/README.md), [fixed-window-results.md](reports/2026-11-20-large-memory/fixed-window-results.md)).

| Data | What was counted | Baseline ranking | With routing | Notes |
|---|---|---:|---:|---|
| LongMemEval-S, 470 answerable questions | Every annotated source message in the packet | 343 (73.0%) | 380 (80.9%) | 37 won, 0 lost, from session routing plus growing excerpts around the query. In 108 of the 127 baseline misses a needed message was outside the top 24 of both rankers, so finding candidates, not ranking them, is the limit there |
| LongMemEval-S, 157 questions with a literal answer | The reference answer string in the packet | 137 (87.3%) | 139 (88.5%) | 6 won, 4 lost. Covering 7.9 points more source messages did not put 7.9 points more answers in the packet |
| LoCoMo-10, 1,533 eligible questions | Every annotated source turn in the packet | 890 (58.1%) | 958 (62.5%) | 84 won, 16 lost, a gain in each of the 10 dialogues. The literal answer was kept for 361 against 330 of 478 |
| ContextBench, 8 histories of 12,800 log rows, 192 questions | Exact counts and id lookups | 192 by scanning 203.8 MB | 192 by reading 5,738 bytes | The index was built in one pass over 11.4 MB using the published record grammar. In the warm replay the scan took about 75 ms and the index about 120 ms, so the saving is in bytes read, not time |

A speaker-attribution check on LoCoMo is in [docs/experiments/locomo/](docs/experiments/locomo/README.md).

### Paragraphs against whole turns

150 probe runs (120 distinct questions, 60 written without sight of the source wording) over three developer sessions of 140 to 299 turns, with no index and no model ([docs/CONTEXT-COMPILER.md sections 2 to 2c](docs/CONTEXT-COMPILER.md)):

- Walking down the ranked list until the answer appears takes a median of 267 to 339 tokens with paragraphs and 1,718 to 2,077 with whole turns, about six times fewer.
- Inside a fixed budget that advantage shrinks to a tie. Paragraphs at 8,000 tokens put the answer in the packet for 107 of 120 questions (median packet 25,900 tokens); six whole turns did so for 103 of 120 (31,000 tokens; paired sign test p 0.48). The packets were 6 to 44 percent smaller. Four times the paragraph budget added 6 answers (113 of 120, p 0.03). The whole transcript, at a median of 297,500 tokens, had all 120.
- On these 120 probes the exact and value searches changed no packet outcome over stemmed full-text search alone. On a synthetic 800-turn test built so that many turns share vocabulary (`TestProbesRecoverPlantedFacts`), the tier order plus fusion with the value search raised planted facts found word for word from 3 of 14 to 13 of 14.
- Value postings computed at index time gave the same scores as computing them per query on all 568,887 query and paragraph pairs from LongMemEval-S, and cut lookup time from 6.283 s to 15.2 ms after an 8.759 s build over 500 histories.

### Stored answers and storage

- **Which stored answers to reopen** ([reports/2026-11-20-answer-updates.md](reports/2026-11-20-answer-updates.md)). Replaying the Go 1.24, 1.25 and 1.26 API files with 1,128 queries: watching the bytes an answer had cited missed all 204 answers that changed; reopening by package reopened 198 that had not; recomputing everything redid 924 unchanged answers. Storing the query's condition (presence, absence or a scoped count, with the snapshot it was read at) reopened exactly the 204.
- **A million events** ([docs/experiments/coverage/](docs/experiments/coverage/bench.py)). Reading 1,000,000 events (49.9 MB) took 4.95 s. Keeping a count current after 10,000 unrelated appends took 0.80 ms and matched a full scan through backdated corrections. When a record of unknown scope could not be read, the count was withheld.
- **Naming a million omitted turns.** Written inline they take 880,001 bytes, about 264,000 tokens. Through paged catalog files of 4,096 ids, the handle in the packet is 213 to 217 bytes.
- **Compression that can still be cited.** Compressing 246,750 LongMemEval messages in 4 KiB blocks took 250.37 MB to 113.26 MB. 2,000 random 96-byte citation reads then expanded 8.40 MB, against 131.14 MB with 64 KiB blocks.
- **Recall latency.** Five recall searches over 2,048 turns took 797 ms, then 240 ms (both in the report), and about 4 ms in `BenchmarkRecallRequest` on synthetic turns, once candidates were prefiltered in SQLite and each turn was checked once.

### Notes linked on an Obsidian canvas (exploratory)

This study was rated by models, not people. It has 8 distinct problems, about 54 comparisons with no correction for their number, and rater agreement of 0.19 on originality in round 3. Read it as a lead and nothing more.

Across four rounds, 144 design proposals written by models were scored with their source hidden ([round 1](examples/obsidian/creativity-study/README.md), [round 2](examples/obsidian/creativity-study2/README.md), [round 3](examples/obsidian/creativity-study3/README.md), [round 4](examples/obsidian/creativity-study4/README.md)). Proposals written from notes that a person had linked on a canvas, with labels such as `constrains` and `falsifies`, were rated higher on feasibility and on how well their tests discriminate than proposals written from search or from no notes. Revising a draft with the linked notes raised its originality rating in 12 of 16 pairs. It did not fix weak test designs in the draft.

## Known defects

From the audit in [docs/ROADMAP.md](docs/ROADMAP.md). The security defects that audit found (path traversal through ids, a dashboard that answered any origin, a non-atomic plan checkpoint, quotes counted as verified when absent) are fixed, each with a test. These are open:

- **`cx views`** (experimental) is wrong in ordinary cases. Carding an older turn can reopen a value that was replaced. It does not read negation, so "do not set pool_size=999" is taken as setting it. It tracks one time axis, not two.
- **`cx fork -as-of`** (experimental) cuts turns, facts and decisions at the chosen point but copies threads and entities whole, and makes manually retracted records active again.
- **`cx explain`** overstates. `-drop` removes items from a built packet without selecting again, so it does not show cause. Its "percent of OPT" figure is measured against a stand-in objective, not the selector that built the packet. Its cost is a price table times token counts.
- **The plan gate** refuses some plans on a ratio taken from a paper and fitted to nothing here. It and the plan scheduler changed the default with no `cx compare` run.
- **Packets.** An excerpt cut from inside a code block can leave a Markdown fence open. A reviewer cannot ask for more text.
- **Receipts.** Asking a receipt's own question again marks it stale, and three of the condition kinds it derives are never evaluated. Resuming a count query compares three sampled records only.
- **Others.** `cx merge` says "verified" when the repository has no verify command. `cx tune` picks its winner by its held-out probes, so the held-out gain it reports is biased. The direct HTTP adapters ignore a stream's error event and a stop for length, so a cut-off answer counts as complete. The body pack is read by nothing.

## Limits of the evidence

- **Two conversations, one author.** Both answer tests draw on the same two long private conversations. Nothing here shows the result on another person's work, another domain or another language.
- **Not independently reproducible yet.** The answer tests' data is private. A public conversation with public questions is the missing piece.
- **Models judging models.** The answers and the first judge are the same model, and the second judge is a sibling. Agreement was high (kappa 0.889 and 0.840) and the controls behaved (0.000, and above 0.98), which bounds leniency but not shared blind spots. No person graded an answer.
- **One answering model.** All answers are from Claude Opus 5.5 at medium effort. Other models may use a map differently.
- **The notes were written apart.** Separate writers read long stretches at once. The engine's own indexer has not yet produced the notes a test ran on, and whether a model asks for paragraphs without being offered the step is untested.
- **Size and selection are not fully separated.** The map packet is about twice the size of the map-off packet. Search filled to 35,000 tokens scored lower than both, and the whole conversation lower than the map, but no arm holds 58,000 tokens of search results.
- **The held-out questions are spent.** All 150 have been answered, so they are development data for anything that follows.
- **No comparison with compaction.** The premise is argued, and the parts are measured against search and against the whole conversation, but no run compares this with a compacting tool on the same work.

## Related work

The numbers in this table are the cited papers' own and were not rerun here.

| Work | What it does | Relation to this project |
|---|---|---|
| [LCM](https://arxiv.org/abs/2605.04050) (Ehrlich and Blackman, 2026) | An immutable store with a graph of summaries and tools to search and expand it | Same immutable source and recall by search. Here a model's summary never stands in for the text |
| [ARC](https://arxiv.org/abs/2607.25066) (Dang et al., 2026) | An append-only store of tool observations, a view rebuilt before each call, recall by address under a budget | Same recall by address, and the refusal of a recall that does not fit |
| [LANTERN](https://arxiv.org/abs/2606.05182) (Subramani, 2026) | Archives every turn, hybrid retrieval with rank fusion, probe questions, and a failure analysis that found a blind spot for configuration values | The fusion constant, the value search and model-free probes (`cx probe`) follow it |
| [CogCanvas](https://arxiv.org/abs/2601.00821) (An, 2026) | Version 2 extracts decisions, facts and reminders grounded in quotes. Version 5, retitled "Fidelity Before Structure", reports that verbatim chunks beat extracted records | Its stated gaps prompted facts that replace facts here. Its later result agrees with keeping the text |
| [GrepMem](https://arxiv.org/abs/2611.25123) (Yang et al., 2026) and [Folio](https://arxiv.org/abs/2611.16627) (Ruan et al., 2026) | An append-only log read back by regular expression; bounded reading with handles to the source | Close in aim. This project adds ranked packets built by code, the checked map, and stored queries |
| [LEDGER](https://arxiv.org/abs/2611.20927) (Islam et al., 2026) and [Read Boundaries in Agent Memory](https://arxiv.org/abs/2611.19264) (2026) | Facts that change, each with its source and period; reads that mix revisions | Packets, catalogs and stored queries are pinned to a turn number, and a replaced fact keeps its link |
| [TallyMem](https://arxiv.org/abs/2610.19474) (Yang et al., 2026), [Evergreen](https://arxiv.org/abs/2604.26180) (Lee et al., 2026) and [DBToaster](https://arxiv.org/abs/1207.0137) (Ahmad et al., 2012) | Counting by code over a kept source log; checking claims about aggregates; keeping views current incrementally | `cx query` answers scoped counts by code and withholds an answer when the index lags or a record's scope is unknown |
| [Dense X](https://arxiv.org/abs/2312.06648) (Chen et al., 2023), [RAPTOR](https://arxiv.org/abs/2401.18059) (Sarthi et al., 2024) and [HippoRAG](https://arxiv.org/abs/2405.14831) (Gutiérrez et al., 2024) | Smaller retrieval units, a tree of summaries, a graph walked with personalized PageRank | Paragraphs here are exact byte ranges, not rewritten propositions. The upper levels (cards, the map) only point |
| [Don't Break the Cache](https://arxiv.org/abs/2601.06007) (Lumer et al., 2026) and [Token Reduction Is Not Cost Reduction](https://arxiv.org/abs/2607.12161) (Weinberger and Hozez, 2026) | Prompt caching cut cost by 27.8 to 81.4 percent across the models and strategies in the first paper's Table 2. In the second, compressing tool output cut tokens and raised the bill | The packet order described above, and cost reported with cache reads counted |

More sources, with their context, are in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) and [reports/2026-11-20-large-memory/README.md](reports/2026-11-20-large-memory/README.md).

## Tests

```sh
go test ./... -count=1
```

The suite runs the engine against scripted stand-ins for the vendor tools, as real processes, with no model call. It covers the cutting of turns into paragraphs and their byte addresses, stemmed search and fusion with the value search, quote checks on notes and requests bounded by tokens, a 1,200-turn history (about 1.3M tokens) in which every turn must be either sent or named, a 400 KB message split into parts, plans across vendors with quote checks, failover on rate limits, review by another vendor, the worktree allowlist, and the 800-turn planted-fact test.

## Repository map

| Path | What it is |
|---|---|
| `cmd/cx`, `internal/` | The `cx` binary and the engine |
| `docs/` | The architecture, the packet builder's design and measurements, the roadmap, and the experiment scripts |
| `reports/` | Dated experiment reports with their result files. The two files with titles in place of dates are literature reports written with model assistance |
| `examples/` | The model-free replays, and the Obsidian study with its model-written proposals |
| `idea-vault/` | One note per idea or result, kept with model assistance. Each evidence line quotes the file it cites, and `docs/experiments/vault_graph.py` checks every quote |
| `research_notes/` | Literature notes written with model assistance while designing the system. Their citations were not all checked against the sources. Treat them as leads, not as a survey |
| `apps/macos` | A small SwiftUI window over the `cx` binary |

## License, credit and citation

ContextlessAI was designed and developed by Tom Maciejewski.

The code and documents are released under the [Apache License 2.0](LICENSE): you may use, change and redistribute them, including commercially, provided you keep the copyright and [NOTICE](NOTICE) and say what you changed. Result files derived from third-party benchmarks keep their sources' terms, including a non-commercial condition for LoCoMo ([THIRD_PARTY.md](THIRD_PARTY.md)).

If you use the software or its results in research, please cite it. [CITATION.cff](CITATION.cff) has the entry, and GitHub shows it under "Cite this repository". Contributions are welcome under the same license ([CONTRIBUTING.md](CONTRIBUTING.md)).

## Lineage

The vendor adapters, the sandbox, the sizing rules and the routing tables were generalized from two earlier private projects by the same author, BookWriter v2 and MultiAIWorkspace.
