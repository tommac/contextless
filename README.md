# ContextlessAI

**ContextlessAI (`cx`)** is an external-memory context compiler and quote-verified multi-agent runtime written in Go. It replaces lossy, autoregressive context-window compaction with an **immutable, Git-backed verbatim archive** and **stateless, deterministic context compilation** across heterogeneous frontier model lanes (`claude`, `codex`, `kimi`, `gemini`/`agy`, `grok`, `muse`).

Full system specifications, experimental logs, and design audits live in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [docs/CONTEXT-COMPILER.md](docs/CONTEXT-COMPILER.md), and [docs/ROADMAP.md](docs/ROADMAP.md).

### What Is Shown, and What Is Not

This is one author's research system. Read every claim below at the level of evidence in this table.

| Claim | Evidence | Level |
|---|---|---|
| A model that reads a map of checked notes and asks for paragraphs answers better than one given the same tokens of search results | Two blind-judged answer tests (30 and 120 held-out questions); the second was confirmatory, with its rule fixed first (Section 4.1, 4.1b) | Measured on answers, two private conversations |
| The engine's packet with the map answers better than the same engine without it | Confirmatory test, 120 questions: 0.882 against 0.706 (Section 4.1b) | Measured on answers; packet sizes differ, see limits |
| A small exact packet beats sending the whole conversation | Both tests: +0.161 and +0.172 (described, not the tested contrast) | Observed twice, not a declared test |
| Paragraph atoms put the answer in the packet about as often as whole turns, in smaller packets | 150 probe runs, no model (Section 4.3) | Retrieval only |
| Session routing and reply adjacency raise source coverage on LongMemEval-S and LoCoMo | Replay scripts on public data, no model (Section 4.2) | Retrieval only; scripts, not the Go engine's port |
| Quote checking, independent review, worktree isolation, sizing and omission accounting behave as specified | The test suite, with scripted providers (Section 7.4) | Behaviour tested; effect on answer quality not measured |
| This is better than a harness that compacts | None yet | Not measured: no head-to-head run exists |

The answer tests used private conversations: their questions, packets and transcripts are not in this repository, so Sections 4.1 and 4.1b report numbers and methods that an outside reader cannot yet rerun. Everything in Sections 4.2 to 4.4 is reproducible from public data or synthetic fixtures (Section 7.1).

---

## In Plain English

Most AI chat and coding tools slow down or lose accuracy on long projects because of **compaction**: once a conversation fills the model's context window, the tool quietly summarizes older messages, forgetting exact numbers, config flags, quotes, and which earlier decisions were later reversed.

**ContextlessAI (`cx`) never summarizes away your history.** Every message, fact, and decision is saved word-for-word in a local Git vault. Before every AI reply, deterministic Go code builds a fresh, right-sized packet containing only the exact verbatim paragraphs, active decisions, and pointer maps relevant to your question, and if the AI needs an older detail it wasn't given, it asks for it by ID and gets the exact original text back. It can also split complex tasks across **Claude, Codex/OpenAI, Kimi, Gemini, Grok, and Muse**, verify every cited quote against your vault, have a separate model review the answer, and run coding changes in isolated Git worktrees without touching your branch.

* **Best for:** Multi-week engineering/architecture projects, audit-grade research where every claim needs a verified quote, Obsidian Canvas brainstorming, or plug-and-play long-term memory for other AI agents via MCP (`cx serve <ws> -mcp`).
* **Before you run it with real models:** `cx` starts each vendor's command-line tool with its permission prompts turned off. On macOS a sandbox contains each run; on Linux there is none. Read [Safety](#safety-read-before-running-with-real-models) first.
* **Try it in 10 seconds (no API keys required):** `make demo` (or jump to [Section 7: Quickstart](#7-quickstart--reproducible-artifacts)).

---

## Motivation & Formal Problem Statement

Standard long-horizon agent harnesses manage finite context windows of capacity $B$ via iterative lossy compaction:

$$s_{t+1} = \mathcal{C}_\theta(s_t, x_t)$$

where a model summarizes prior state $s_t$ and new observation $x_t$ into a compressed text state $s_{t+1}$. The design starts from three concerns about iterated summarization as a memory. The first two are the motivating argument, taken from the literature in Section 6 and not measured here; the third is measured in Section 4.1:

1. **Non-monotonicity & Provenance Erasure**: Compression discards rare lexical anchors (ports, SHA digests, flags, exact thresholds) whose downstream utility is unknown at write time, severing byte-level attribution to primary turns.
2. **Temporal Conflation**: Lossy summaries routinely blend superseded assertions with active decisions or fail to invalidate derived conclusions when upstream premises change ("context rot" and read-boundary skew).
3. **Attention Dilution at the Opposite Extreme**: Even when an entire history fits a $200\text{k}\text{--}1\text{M}$ window, feeding uncurated transcripts degrades recall on dispersed multi-fact queries relative to compact, high-precision evidence packets ($+0.161$ and $+0.172$ for an exact packet over full-transcript input in Sections 4.1 and 4.1b; two conversations).

**ContextlessAI** decouples *durable memory* from *inference working memory*. The conversation is an append-only log of immutable Markdown records $\mathcal{V}_t = (x_1, \dots, x_t)$ committed to Git, paired with rebuildable relational, full-text, and structural indices $\mathcal{I}(\mathcal{V}_t)$ in SQLite. Every model invocation is a **stateless operator** evaluated over a freshly compiled, budget-bounded packet:

$$y_{t+1}, \Delta\mathcal{L} \leftarrow \mathcal{M}_p\!\left(\Pi_{B_p}\!\left(q_t, \mathcal{V}_{\le t}, \mathcal{I}(\mathcal{V}_{\le t})\right)\right)$$

where $\Pi_{B_p}$ is deterministic Go code, never an LLM, and $\mathcal{M}_p$ executes with session persistence and provider-side compaction disabled.

---

## 1. Architectural Invariants & Execution Pipeline

ContextlessAI generalizes the verbatim packet architecture of *BookWriter v2* with the machine-wide multi-provider concurrency lanes of *MultiAIWorkspace*:

```text
Append-Only Git Vault (~/.contextless/workspaces/<ws>/)
├── session.md              [Tier 0] Standing instructions, attached repo specs
├── decisions/de-*.md       [Tier 0] Active & superseded decisions (quote-anchored)
├── facts/fa-*.md           [Tier 2/3] Extracted facts with supersession edges
├── notes/no-*.md           [Atlas]  Content-addressed, byte-verified pointer notes
├── cards/tNNNNNN.md        [Tier 1] SHA-pinned 1-line turn navigation cards
├── turns/tNNNNNN.md        [Source] Immutable verbatim turns (never rewritten)
└── .cx/state.db            [Cache]  SQLite FTS5 (Porter), value postings, runs, views
         │
         ▼  Deterministic Context Compiler (internal/packet)
┌──────────────────────────────────────────────────────────────────────────────┐
│ 1. Candidate Discovery : Exact phrase ∪ FTS5 BM25 (+card) ∪ Value postings   │
│ 2. Fusion & Granularity: Reciprocal Rank Fusion (k=60) over paragraph atoms  │
│ 3. Sizing & Escalation : Target 60–70% of window; escalate 200k -> 500k -> 1M│
│                          before tier-dropping (1..4) with explicit omission  │
│ 4. Cache-Stable Layout : Static rules -> Ledger/Atlas -> Spans -> Focus last │
└──────────────────────────────────────────────────────────────────────────────┘
         │
         ▼  Stateless Execution & Verification (internal/engine, internal/route)
┌──────────────────────────────────────────────────────────────────────────────┐
│ • Schema-checked JSON handoff: completed | needs_recall | partial | plan     │
│ • Demand paging (needs_recall): IDs, spans (t#aK-M), parts (t#pK), recall ops│
│ • Multi-unit plans: typed claim records -> byte-level quote check -> merge   │
│ • Maker–checker review: cross-provider reviewer disjoint from writer family  │
│ • Repo coding lane: ephemeral git worktree + allowlist diff + verify command │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Core System Guarantees

1. **Verbatim Immutability**: Every user and assistant message is persisted word-for-word as `turns/tNNNNNN.md`. No turn is ever shortened or overwritten.
2. **Size, Never Compact**: Packet token cost is estimated per provider profile prior to process spawn ($60\text{--}70\%$ window target). Overflow escalates to larger context profiles (`claude` $200\text{k} \to$ `codex` $500\text{k} \to$ `kimi`/`agy` $1\text{M}$) before dropping lower-priority tiers (`1: catalog -> 2: thread turns -> 3: ranked atoms/facts -> 4: oldest recent turns`). Tier `0` (`session`, `focus`, `recall`, active `decisions`) is invariant.
3. **Complete Omission Accountability**: Every record excluded for budget is enumerated: individually by card, in deterministic 50-turn keyword/thread digests, or via a snapshot-pinned paged `.cxi` catalog handle.
4. **Indices Navigate; Verbatim Text Answers**: Model-generated index artifacts (cards, threads, facts, Atlas idea notes) act strictly as pointers and search keys. Ranked evidence is always injected as exact byte slices of `turns/tNNNNNN.md`, and indexer-emitted quotes are checked against source bytes before commit.
5. **Compaction Tainting & Watchdogs**: CLI harnesses run headless with `--no-session-persistence` and compaction flags disabled. Any run emitting a compaction telemetry event or exceeding the $85\text{--}90\%$ window watchdog is tainted, discarded, and retried on a larger profile.

---

## 2. Context Compilation, Granularity & Retrieval

### 2.1 Lossless Markdown-Aware Atomization (`internal/vault/atoms.go`)

Rather than retrieving whole turns (which suffer severe token waste on long responses) or LLM-rewritten propositions (which hallucinate or drop context), ContextlessAI segments each turn body $x_i$ into a tiling of contiguous byte intervals called **atoms**:

- Split points occur at blank lines outside fenced code blocks; short blocks coalesce up to `AtomMinBytes = 200` ($\approx 60$ tokens), headings/labels open their section's atom, and oversized blocks (`> AtomMaxBytes = 2400`, $\approx 720$ tokens) split deterministically at blank lines, line breaks, or sentence boundaries (`AtomSeg = 2`).
- Each atom is an immutable coordinate `tNNNNNN#aK` (and contiguous ranges `tNNNNNN#aK-M`); concatenating a turn's atoms reconstructs $x_i$ byte-for-byte. Turns exceeding $24\text{ KB}$ also expose coarse part coordinates `tNNNNNN#pK`.
- **Whole-History Bypass**: When total prior history satisfies $\sum_{i < t} \text{tok}(x_i) \le B_{\text{whole}}$ ($60,000$ tokens default) and fits the primary responder profile, selection is bypassed and the complete transcript is emitted verbatim in chronological order.

### 2.2 Multi-Lane Hybrid Retrieval & Reciprocal Rank Fusion (`internal/packet/atoms.go`)

Candidate atoms prior to the focus sequence are scored across specialized lexical and structural lanes without requiring external embedding downloads or query-time LLM calls:

- **`exact`**: Substring/phrase containment ordered newest-first, prioritizing the latest state of mutated parameters over superseded mentions.
- **`full_text`**: Porter-stemmed SQLite FTS5 BM25 over `(atom_text || card_text)`, allowing queries phrased in turn-level topic vocabulary ("titles as APIs") to resolve individual paragraphs.
- **`value`**: Sentence-local proximity postings matching numeric literals, dates, paths, or code identifiers within $\le 4$ tokens of a query stem. At whole-turn granularity this runs as a full lane; at atom granularity only its top $2$ hits enter fusion (preventing ubiquitous number-near-topic matches across thousands of short paragraphs from crowding out BM25 hits; see Section 4.3).
- **Locality & Structural Extensions**: Temporal `working_set` retention for recently cited/recalled atoms, plus opt-in session/episode routing (`-session-route`), conversational Q&A adjacency (`-auto-reply`), Clarke–Cormack minimal multi-atom cover intervals (`cover_window`), and 2-hop Personalized PageRank over rare identifiers (`bridge`). These extensions are off by default. Their only measurements are packet coverage from replay scripts (Section 4.2); none has an answer test, and the Go ports of session routing and reply adjacency have not been benchmarked.

Lanes are combined via **Reciprocal Rank Fusion** ($k = 60$), preserving each active lane's rank-1 witness:

$$\text{RRF}(a) = \sum_{\ell \in \mathcal{L}} \frac{w_\ell}{60 + \text{rank}_\ell(a)}$$

Selected atoms from the same turn merge when adjacent, carry contextual headers (turn ID, paragraph span, speaker, and prompt prefix of the turn being answered), deduplicate byte-identical paragraphs across turns, and pack greedily up to `relevant_tokens` ($8,000$ default).

### 2.3 Atlas: Checked Pointer Maps & Token-Bounded Two-Stage Pull (`internal/packet/atlas.go`)

Lexical search struggles on global synthesis queries ("how did our cache eviction design evolve, and why?") whose evidence spans many non-overlapping turns. The **Atlas** subsystem implements a verified pointer-then-pull protocol:

1. **Quote-Pinned Note Records**: An Atlas note (`notes/<id>.md`) records a concise statement, kind, topic tags, and source quotes. The engine keeps a note only through quotes that code finds in the cited `turns/tNNNNNN.md` (whitespace aside), pinning `(body_sha, byte_start, byte_end, atom_range)` and content-addressing the note ID by `(title, spans)`.
2. **Append-Only Map Prefix**: When notes exist (`atlas: auto`, or indexer-maintained under `atlas: on`), the brief renders an append-only navigation map (`- [t000012#a3] (decision) ...`, capped at `atlas_map_tokens = 40,000`) immediately after active decisions. Overflowing maps rank relevant lines and summarize remaining notes as per-topic turn-range counts queryable via the `notes` recall op.
3. **Token-Bounded Pulls**: Instructed that map lines are pointers rather than ground text, the reader requests exact supporting spans (`needs_recall`), which the engine serves verbatim in priority order up to `atlas_pull_tokens = 8,000` per round; what a search in the same request finds is paid for from the same tokens. Results are in Section 4.1b.

### 2.4 Prompt-Cache Prefix Topology & Billed-Cost Model (`docs/CONTEXT-COMPILER.md` §5)

Because frontier APIs bill cached prefix reads at $0.1\times\text{--}0.25\times$ base input rates, **token reduction is not isomorphic to cost reduction** ([Lumer et al., 2026](https://arxiv.org/abs/2601.06007); [Weinberger & Hozez, 2026](https://arxiv.org/abs/2607.12161)). ContextlessAI orders every brief monotonically from static to volatile:

$$\underbrace{\text{Session} \;\Vert\; \text{Decisions/Facts} \;\Vert\; \text{Append-Only Map \& Catalog}}_{\text{Cache-Eligible Prefix}} \;\Vert\; \underbrace{\text{Ranked Excerpts} \;\Vert\; \text{Recent Window} \;\Vert\; \text{Focus \& Run Metadata}}_{\text{Per-Turn Dynamic Suffix}}$$

Per-turn volatile identifiers (run IDs, timestamps) are confined strictly to the tail. `cx cache`, `cx gaps`, and `cx explain` evaluate prefix reuse and estimate provider cost/latency from recorded telemetry. Whether the CLI harnesses actually reuse a stable prefix has not been shown, so the cached figures are a ceiling, not a bill:

$$C_p = p_{\text{in}} T_{\text{uncached}} + p_{\text{read}} T_{\text{cached}} + p_{\text{write}} T_{\text{write}} + p_{\text{out}} T_{\text{out}}, \qquad \widehat{L}_p = a_p + b_p T_{\text{uncached}} + c_p T_{\text{out}}$$

---

## 3. Multi-Agent Verification, Maker–Checker Review & Repo Sandboxing

### 3.1 Structured Claim Verification & Conflict Auditing (`internal/engine/units.go`, `plan.go`)

When a request decomposes into a multi-unit `plan`, worker units execute across provider-specific concurrency lanes (`responder`, `coder`, `researcher`, `long_reader`, `editor`, `reviewer`, `assembler`). Workers do not pass unconstrained prose to downstream synthesis; each returns a schema-validated record containing:

- `claims`: structured propositions `(id, claim, evidence_ids, verbatim_quote, declared_support)`.
- `status` $\in \{\text{completed}, \text{no\_information}, \text{failed}\}$, explicit `assumptions`, `open_questions`, and `conflicts`.

Before the `assembler` runs, deterministic Go code resolves every `verbatim_quote` against the immutable vault or unit artifacts, records a durable byte-span `EvidenceAnchor` (`check = verified`), downgrades unlocated or unquoted claims (`not_found`, `no_quote`, `no_evidence`), flags cross-unit assumption/key clashes (enforcing **Verified-Quote Dominance** where byte-verified quotes meet unverified assertions), and verifies in `assembly.json` that the final synthesis accounts for every unit claim. Degenerate single-unit or serial chain plans are collapsed back to direct execution.

### 3.2 Cross-Family Independent Review

To reduce the chance that a model approves its own family's mistakes, assembled multi-unit answers (`review: plans`, or all answers under `review: always`) are audited by a `reviewer` profile strictly **disjoint from every provider family that contributed to the draft**:

| Role | Default Provider Preference Cascade |
|---|---|
| `responder` | `claude` (Opus 5.5) $\to$ `codex` (GPT-6.1-Sol) $\to$ `kimi` (K3.1-Pro) $\to$ `agy` (Gemini 4 Argon) |
| `coder` | `codex` $\to$ `claude` (Opus) $\to$ `kimi` |
| `researcher` | `grok` (4.7, network-enabled) $\to$ `codex` |
| `long_reader` | `kimi` ($1\text{M}$) $\to$ `agy` ($1\text{M}$) $\to$ `codex` |
| `editor` | `muse` $\to$ `claude` (Sonnet $\to$ Opus) |
| `reviewer` | `codex` $\to$ `claude` (Opus) $\to$ `kimi` (Highspeed) $\to$ `muse` *(filtered to exclude draft authors)* |
| `assembler` | `claude` (Opus) $\to$ `codex` $\to$ `kimi` |
| `indexer` | `codex` (Medium) $\to$ `claude` (Haiku) $\to$ `kimi` (Highspeed) |

Critiques lacking an exact substring quote of the draft (or an explicit missing-requirement specification) are discarded by code; surviving defects trigger at most one bounded revision pass. Machine-wide SQLite locks cap per-provider concurrency across simultaneous `cx` processes, with `Retry-After` / exponential cooldown failover and failure-context carry-over. Whether review or multi-unit plans improve answers has not been measured (`cx compare` exists for this and has not been run on real messages).

### 3.3 Sandboxed Git Worktree Coding Lanes (`cx attach`, `cx merge`)

Attached repositories (`cx attach <ws> <repo> <path> -verify "<cmd>" -allow "<globs>"`) isolate agentic code modifications:

1. Each `coder` unit receives an ephemeral `git worktree` checked out at `cx/<run>` inside its sandboxed packet directory; agents are forbidden from invoking `git`.
2. Post-run, the engine stages all filesystem mutations, diffs against the base commit, **rejects the run if any touched path falls outside `-allow` or touches `.git`**, commits cleanly, and executes `-verify` outside the agent sandbox in its own process group.
3. Unmerged `cx/<run>` branches leave the user's working tree untouched; `cx merge` re-verifies the branch on current `HEAD` prior to fast-forward/merge integration.

---

## 4. Empirical Evaluation

Benchmarks on public or synthetic data, with their negative results, paired win/loss counts, and reproduction harnesses, are committed in `docs/experiments/`, `examples/`, and `reports/`. The two answer tests (4.1, 4.1b) ran on private conversations and are reported as numbers and methods only. We distinguish strictly between **packet coverage** (whether required source text enters the prompt) and **end-to-end answer quality** (whether a model states the required information).

### 4.1 Held-Out Multi-Passage Answer Evaluation (`2026-11-25`, Rule Fixed Before the Run)

Questions, arms and the pass rule were fixed in a private runbook on `2026-11-18`, a week before the run; this was not registered with a third party ([docs/CONTEXT-COMPILER.md §2e](docs/CONTEXT-COMPILER.md)). $N = 30$ held-out synthesis questions requiring dispersed passages across two long conversations ($18$ on a never-tuned conversation, $12$ newly authored). Answering model: Claude Opus 5.5 (medium effort). Two blind judges (the answering model and a second model) scored $965$ passage verdicts in counterbalanced order with arbitrary arm labels (**exact judge agreement $= 0.936$, Cohen's $\kappa = 0.889$**; $0$ dropped runs across $255$ invocations).

| Arm | Context Supplied to Reader | Mean Passage Score | Mean Input Tokens |
|---|---|---:|---:|
| `none` | Question only (negative control) | $0.000$ | $130$ |
| `search` | Ranked atoms, default budget | $0.314$ | $7,961$ |
| `searchbig` | Ranked atoms, budget-matched to map arms | $0.546$ | $34,776$ |
| `catpull` | Code-built 1-line/turn overview $\to$ model paragraph pull | $0.545$ | $23,111$ |
| `pull` (the arm the Atlas was built from) | Quote-checked note map $\to$ model paragraph pull | **$0.726$** | $31,001$ |
| `full` | Complete verbatim transcript ($n = 15$ subset) | $0.839$ | $247,856$ |
| `oracle` | Exact gold verbatim paragraphs only (positive control) | **$0.997$** | $1,209$ |

Key findings, including falsified hypotheses:

- **The one hypothesis the rule tested did not hold (`catpull` vs. `searchbig`)**: Heuristic code-generated 1-line turn summaries plus model pull did *not* outperform budget-matched lexical search ($\Delta = -0.001$, $95\%$ CI $[-0.125, +0.122]$, $12$ won / $12$ lost / $6$ tied).
- **Checked note pointers drive two-stage recovery (`pull` vs. `searchbig` & `catpull`)**: Contrasts the rule marked as described only (Holm-adjusted among themselves) showed `pull` outperforming `searchbig` by $+0.180$ ($95\%$ CI $[+0.080, +0.281]$, $18\text{W}/5\text{L}/7\text{T}$, exact sign $p = 0.011$, $p_{\text{Holm}} = 0.032$) and `catpull` by $+0.181$ ($21\text{W}/7\text{L}$, $p_{\text{Holm}} = 0.032$), while coming within noise of `full` at **$1/8\times$ the token cost** ($\Delta = -0.070$, $p = 0.29$ on $n = 15$). Because the rule did not test this contrast, it was a lead until the confirming run below, which found half this effect.
- **An exact packet beat the whole conversation (`oracle` vs. `full`)**: Supplying *only* the $1,209$ tokens of gold verbatim spans beat supplying the entire $247,856$-token conversation by **$+0.161$** ($95\%$ CI $[+0.059, +0.286]$, $8$ won, $0$ lost). Sending the entire transcript is not the upper bound on answer completeness. The run does not show why; attention dilution is one reading.
- **Empirical passage-to-answer transfer rate**: Needed passages present verbatim in bounded packets were stated in **$0.92\text{--}1.00$** of verdicts ($0.83$ in `full`); passages pointed to by a shown note line *without* pulling the underlying paragraph were stated in **$0.33\text{--}0.38$**; passages neither present nor pointed to scored **$0.09\text{--}0.28$**. In this test a pointer was worth about a third of a held passage.

### 4.1b Confirming Run on the Remaining 120 Held-Out Questions (`2026-11-25`)

The rule was written before any packet or answer of this run existed ([docs/CONTEXT-COMPILER.md §2f](docs/CONTEXT-COMPILER.md)). Two tests, Holm-adjusted across the two: the earlier `pull` against `searchbig` contrast repeated with the same binary, packets and prompts; and the engine's own packet with the Atlas map against the same engine with `atlas: off`, each allowed one recall request (ids, searches or both). Same answering model and judges as 4.1; 810 invocations, 0 errors, Cohen's kappa 0.840.

| Arm | Context Supplied to Reader | n | Mean Passage Score | Mean Input Tokens |
|---|---|---:|---:|---:|
| `none` | Question only (negative control) | 30 | 0.000 | 130 |
| `searchbig` | Ranked atoms, budget-matched to the note lines | 120 | 0.672 | 34,768 |
| `livenomap` | Live engine packet, `atlas: off`, one recall request | 119 | 0.706 | 28,984 |
| `pull` | Note lines, then requested paragraphs | 120 | 0.762 | 31,415 |
| `full` | Complete verbatim transcript | 30 | 0.810 | 247,854 |
| `livemap` (**Atlas**) | Live engine packet with the map, one recall request | 119 | **0.882** | 58,517 |
| `oracle` | Exact gold paragraphs only (positive control) | 30 | 0.982 | 1,214 |

- **The earlier lead is confirmed at half its size**: `pull` beats `searchbig` by +0.090 (95% CI +0.037 to +0.141; 57 won, 30 lost; adjusted p 0.005).
- **The map helps the engine**: `livemap` beats `livenomap` by +0.175 (95% CI +0.137 to +0.215; 66 won, 4 lost; adjusted p below 0.0001), and on the 29 questions with a whole-conversation answer it is level with `full` (+0.035, CI -0.039 to +0.118) at under a quarter of the tokens.
- **Why**: before any request both live packets held 0.327 of the needed passages word for word; after one request the packet with the map held 0.840 and the one without 0.614.
- **Limits**: the map packet is twice the size of the map-off packet and no arm fills those tokens with search results; the notes came from separate writers, not the indexer; the request was offered as a step; two conversations by one author. All 150 held-out questions are now spent.

### 4.2 Public Conversational & Log Benchmarks (`8,000`-Byte Strict Cap)

These rows measure whether source text reaches a packet, not whether a model answers. They come from the Python replay scripts in `docs/experiments/` on public data; the Go engine's ports of session routing and reply adjacency (`-session-route`, `-auto-reply`) have not been benchmarked. No model ran ([reports/2026-11-20-large-memory/](reports/2026-11-20-large-memory/README.md), [fixed-window-results.md](reports/2026-11-20-large-memory/fixed-window-results.md)):

| Benchmark / Workload | Metric & Budget Constraint | Baseline Ranking | Experimental Policy | Diagnostic / Pairing Notes |
|---|---|---:|---:|---|
| **LongMemEval-S** ($N = 470$ answerable) | All annotated source messages in $8,000\text{ B}$ packet | $343 / 470$ ($73.0\%$) | **$380 / 470$ ($80.9\%$)** | $+37\text{W} / 0\text{L}$ via session routing + query-centered excerpt growth; $108/127$ baseline misses lacked a source in top-24 union (candidate-recall bottleneck) |
| **LongMemEval-S** ($n = 157$ non-trivial literal) | Exact reference answer substring in $8,000\text{ B}$ packet | $137 / 157$ ($87.3\%$) | $139 / 157$ ($88.5\%$) | $+6\text{W} / 4\text{L}$: demonstrates why $+7.9\text{ pp}$ message coverage does not automatically imply $+7.9\text{ pp}$ answer gain |
| **LoCoMo-10** ($N = 1,533$ eligible, frozen) | All annotated source turns in $8,000\text{ B}$ packet | $890 / 1,533$ ($58.1\%$) | **$958 / 1,533$ ($62.5\%$)** | $+84\text{W} / 16\text{L}$, positive net gain across all $10$ dialogues; literal answer retention $361/478$ vs. $330/478$ |
| **LoCoMo Speaker Audit** ($267$ minimal pairs) | Adversarial participant-swap attribution | $259 / 267$ ($97.0\%$) | $\pm 1$ turn exchange: $4/4$ | Remaining $4$ non-label errors were question turns whose answers resided in the adjacent reply ($4/4$ recovered vs. $1/4$ for BM25) |
| **ContextBench** ($8 \times 12,800$ log rows) | Exact severity/service counts & ID lookups ($192$ Qs) | $192 / 192$ ($203.8\text{ MB}$ scan) | **$192 / 192$ ($5,738\text{ B}$ read)** | Deterministic grammar index ($11.4\text{ MB}$ single pass) matches exhaustive scan identically |

### 4.3 Atom vs. Whole-Turn Granularity & Offline Floor (`cx floor`)

Across $150$ probe runs ($120$ distinct questions, $60$ written blind without viewing source wording) over three unindexed developer sessions ($140\text{--}299$ turns; [docs/CONTEXT-COMPILER.md §2–2c](docs/CONTEXT-COMPILER.md)):

- **Rank-Order Efficiency vs. Fixed-Budget Compression**: Walking ranked units until the answer appears requires a median of **$267\text{--}339$ tokens** with Porter-stemmed atoms versus **$1,718\text{--}2,077$ tokens** with whole turns ($\approx 6\times$ reduction). Inside fixed-budget packets (`relevant_tokens = 8,000`), however, that $6\times$ advantage compresses to a statistical tie in briefs $6\text{--}44\%$ smaller: $107/120$ distinct answers recovered at $8\text{k}$ atoms (median brief $25.9\text{k}$ tokens) vs. $103/120$ for six whole turns ($31.0\text{k}$ tokens, paired sign $p = 0.48$). Quadrupling the atom budget to $32\text{k}$ adds only $+6$ answers ($113/120$, $p = 0.03$); full transcript ($297.5\text{k}$ median brief) recovers $120/120$.
- **Lane Ablation Honesty**: On those $120$ unindexed prose probes, `exact` and `value` lanes changed $0$ packet outcomes over stemmed FTS5 alone (whereas on the $800$-turn/$620\text{k}$-token adversarial vocabulary-sharing benchmark `TestProbesRecoverPlantedFacts`, tier restructuring + RRF value fusion lifted verbatim recovery from $3/14 \to 13/14$).
- **Indexed Value Postings**: Precomputing sentence-local $(\text{term}, \text{value})$ postings reproduced query-time value scores **identically on all $568,887$ query–atom pairs** across LongMemEval-S, cutting lookup time from $6.283\text{ s}$ to $15.2\text{ ms}$ (after an $8.759\text{ s}$ index build across $500$ histories).

### 4.4 Temporal Answer Maintenance & Storage Scaling

- **Condition-Pinned Answer Invalidation ([reports/2026-11-20-answer-updates.md](reports/2026-11-20-answer-updates.md))**: Replaying Go $1.24 \to 1.25 \to 1.26$ API ledgers ($1,128$ live + historical queries), watching *previously cited bytes* missed all $204$ changed answers ($0$ reopened), package-level invalidation caused $198$ false reopenings, and full recomputation re-evaluated $924$ unchanged answers. Persisting the **logical query condition** (presence/absence or scoped count + snapshot cutoff) reopened **exact $204 / 204$ true changes with $0$ misses and $0$ false positives**.
- **Million-Event Resumable Stream ([docs/experiments/coverage/](docs/experiments/coverage/bench.py))**: Ingesting $10^6$ original events ($49.9\text{ MB}$) completed in $4.95\text{ s}$; maintaining a revision-pinned count/witness cursor took **$0.80\text{ ms}$** after $10,000$ out-of-scope appends, matched raw scans across backdated corrections and tie-breaking, and conservatively returned **`Withheld`** when an uninterpretable record of unknown scope entered the log.
- **External Paged Omission Catalogs & Seekable 4 KiB Compression**:
  - Naming $10^6$ omitted turns inline consumes $880,001\text{ B}$ ($\approx 264\text{k}$ tokens); external 4,096-ID bitmap/delta pages shrink the in-packet handle to **$213\text{--}217\text{ B}$** ($8.6\text{--}133.6\text{ KB}$ on disk) with $O(1)$ page point-lookups.
  - Seekable 4 KiB block compression over $246,750$ LongMemEval messages reduced raw storage from **$250.37\text{ MB} \to 113.26\text{ MB}$** while expanding only $8.40\text{ MB}$ across $2,000$ random 96-byte citation reads (vs. $131.14\text{ MB}$ read amplification at 64 KiB blocks).
  - Zero-scan SQLite candidate prefiltering + shared turn verification reduced 5-query recall latency at $2,048$ turns from **$797\text{ ms} \to 240\text{ ms} \to 3.61\text{ ms}$** ($152\text{ MB} \to 55\text{ MB} \to 3.0\text{ MB}$ allocated). The first step is in the report; the last is `BenchmarkRecallRequest` on synthetic turns (about $4\text{ ms}$ when rerun on 2026-11-25).

### 4.5 Curated Context Graphs for Design Synthesis (Obsidian Canvas Study)

Exploratory, and rated by models, not people: $8$ distinct problems, AI raters (round 3 originality agreement $0.19$), and about $54$ uncorrected contrasts. Read the following as a lead.

Across four blind-scored rounds ($144$ anonymized AI design proposals over $8$ systems design problems: [round 1](examples/obsidian/creativity-study/README.md) · [round 2](examples/obsidian/creativity-study2/README.md) · [round 3](examples/obsidian/creativity-study3/README.md) · [round 4](examples/obsidian/creativity-study4/README.md)), proposals written from human-curated Obsidian Canvas graphs (note passages linked with typed edge labels such as `constrains`, `analogous mechanism`, `falsifies`) were rated higher on feasibility and on how well their tests discriminate than proposals written from ordinary search or no notes. Revising an ungrounded draft with labeled notes raised originality ratings in $12/16$ pairs ($0$ regressions) without fixing weak initial test designs.

---

## 5. Extended Tooling, Audit Status & Threats to Validity

Beyond the core vault, compiler, Atlas map, and multi-agent router, the binary ships research instrumentation and integration surfaces:

- **Context Query Planner (`cx explain`)**: Inspects query class, per-lane candidates, section budgets, cache-prefix overlap, How-Provenance semiring polynomials ($\mathbb{N}[X]$), and an auxiliary submodular fractional-knapsack LP upper bound.
- **Living Answer Receipts & Scoped Queries (`cx watch`, `cx stale`, `cx refresh`, `cx query`, `cx views`, `cx fork`)**: Snapshot-pinned query receipts with Delta Resume, deterministic ledger/key-history queries with `-as-of` cutoffs, and workspace state branching. `cx views` and `cx fork -as-of` are **experimental**: both have known correctness defects, listed below, and should not be relied on for an answer.
- **Ecosystem Surfaces**: 12-tool JSON-RPC 2.0 stdio **MCP server** (`cx serve <ws> -mcp`), loopback/origin-locked **Context X-Ray Web UI** (`cx ui <ws>`), bidirectional **Obsidian Vault & Canvas compiler** (`cx obsidian export|brief`), seekable `.cxp` **body packs** (`cx pack`), offline hyperparameter splits (`cx tune`), and opt-in direct HTTP SSE adapters (`anthropic-api`, `openai-api`, `gemini-api`).

### Known Limitations & Active Audit Items ([docs/ROADMAP.md](docs/ROADMAP.md))

In keeping with strict experimental accounting, users and contributors should note what is **verified in end-to-end production** versus **partially wired or unmeasured**:

1. **Atlas Live Autonomy Unmeasured**: In both `2026-11-25` runs the recall request was offered as a step, over notes written by separate writers. Whether a model issues `needs_recall` unprompted from a live packet (rather than answering from pointer lines), and whether notes written by the indexer under `atlas: on` do as well, are the next tests and need new held-out questions.
2. **Engine Wave Audit Items (`2026-11-25`)**: Security boundaries on MCP/UI/ledger paths (`../../` traversal, DNS rebinding/CORS origin checks, atomic `plan.json` checkpoints, schema `recall_ops` exposure, and mandatory quote containment on witnesses) were hardened with regression tests. Still open in `ROADMAP.md`:
   - `cx views` tracks turn-sequence validity (unitemporal rather than full bitemporal transaction$\times$valid time), lacks negation-polarity parsing (`"do not set pool_size=999"`), and can reopen older values on out-of-order re-carding.
   - `cx fork -as-of` filters turns, facts, and decisions by cutoff, but currently copies thread/entity files without temporal trimming and resets manually retracted statuses.
   - `cx explain -drop` performs post-selection removal rather than full re-selection ablation, and its `% of OPT` bound certifies an auxiliary coverage-plus-shingle submodular surrogate rather than the primary RRF selector.
   - The multi-unit topology gate (`ParallelismRatio < 1.35`) and reactive DAG scheduler use literature constants not yet benchmarked via `cx compare` on live workloads.
   - Mid-block code excerpts can leave trailing Markdown fences unclosed in `brief.md`, and `reviewer` runs cannot issue new `needs_recall` faults beyond what writers already fetched.
   - `cx merge` reports a branch as verified when the repo has no `-verify` command; `cx tune` picks weights on the probes it then reports on; the direct HTTP adapters do not surface stream errors; resuming a count query compares sampled witnesses only; the `.cxp` body pack is not on the read path.

### Threats to Validity

- **Two conversations, one author.** Both answer tests draw on the same two long private conversations. Nothing here shows the result on another person's work, another domain or another language.
- **Not independently reproducible yet.** The answer tests' data is private (Section 4). A public conversation with public questions is the missing piece.
- **Models judging models.** Answers and the first judge are the same model; the second judge is a sibling. Agreement was high (kappa 0.889 and 0.840) and the controls behaved (0.000 and above 0.98), which bounds leniency but not shared blind spots. No human graded an answer.
- **One answering model.** All answers are from Claude Opus 5.5 at medium effort. Other models may use a map differently.
- **The map's notes were written apart.** Separate writers read long stretches at once; the engine's own indexer has not yet produced the notes a test ran on. The request for paragraphs was offered as a step, not left to the model.
- **Size and selection are not fully separated.** The map packet is about twice the size of the map-off packet. Search filled to 35,000 tokens scored lower than both, and the whole conversation lower than the map, but no arm holds 58,000 tokens of search results.
- **The held-out questions are spent.** All 150 have now been answered, so they are development data for anything that follows.
- **No comparison with compaction.** The project's premise is argued, and its parts are measured against search and against the whole conversation, but no run compares it with a compacting harness on the same work.

---

## 6. Relation to Contemporary Literature (2023–2026)

| System / Paper | Core Mechanism | What ContextlessAI Adopts & Where It Diverges |
|---|---|---|
| **LCM** (Voltropy, 2026) & **ARC** (Fujitsu, 2026) | Immutable turn log, hierarchical summary DAG, `grep`/address recall, budgeted recall refusal | Adopts immutable source log, deterministic code digests, `grep`/part/span recall, and explicit overflow refusal; forbids unverified model summaries from standing in for primary text |
| **LANTERN** (Cisco, 2026) & **CogCanvas** (2026) | Hybrid RRF with value sensitivity, planted recovery probes, verbatim-grounded decision/fact graphs | Adopts $\text{RRF}_{k=60}$, sentence-local `value` lane (capped to top-2 on atoms), supersession chains, and model-free recovery probes (`cx probe`) |
| **GrepMem** ([arXiv:2611.25123](https://arxiv.org/abs/2611.25123)) & **Folio** ([arXiv:2611.16627](https://arxiv.org/abs/2611.16627)) | Append-only raw log + regex/handle progressive reopening ($66/75$ vs. $54/75$ for compaction on LOCA-bench) | Pairs progressive span/regex recall with typed multi-lane compilation, quote-checked Atlas pointer maps, and snapshot-scoped queries |
| **LEDGER** ([arXiv:2611.20927](https://arxiv.org/abs/2611.20927)) & **Read Boundaries** ([arXiv:2611.19264](https://arxiv.org/abs/2611.19264)) | Mutable fact histories and cross-revision read skew in agent memory | Pins packets, omission catalogs, and query receipts to explicit turn-sequence snapshots; tracks `supersedes` lineage |
| **TallyMem** ([arXiv:2610.19474v2](https://arxiv.org/abs/2610.19474v2)), **SemViews** ([arXiv:2610.12768](https://arxiv.org/abs/2610.12768)) & **DBToaster** | Deterministic aggregation & incremental view maintenance over source logs | Separates deterministic scoped count/presence operators (`cx query`) from open-ended prose retrieval, returning `withheld` under index lag or unknown scope |
| **Dense X** ([Chen et al., 2023](https://arxiv.org/abs/2312.06648)), **RAPTOR** ([Sarthi et al., 2024](https://arxiv.org/abs/2401.18059)) & **HippoRAG** ([Gutiérrez et al., 2024](https://arxiv.org/abs/2405.14831)) | Sub-document retrieval granularity, tree-level navigation, graph PPR | Uses lossless byte-offset atoms (`t#aK`) rather than rewritten propositions; upper levels (`cards`, `Atlas`) only navigate |
| **Don't Break the Cache** ([Lumer et al., 2026](https://arxiv.org/abs/2601.06007)) & **Token $\neq$ Cost Reduction** ([Weinberger & Hozez, 2026](https://arxiv.org/abs/2607.12161)) | Prompt-cache prefix economics ($27.8\text{--}81.4\%$ savings; mid-prompt pruning can *increase* spend) | Enforces stability-ordered brief prefixes and evaluates cache-adjusted dollar/quota cost via `cx cache` and `cx gaps` |

Figures in the first column are the cited works' own and were not rerun here. Entries without a link are recorded with their context in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) and [reports/2026-11-20-large-memory/README.md](reports/2026-11-20-large-memory/README.md).

---

## 7. Quickstart & Reproducible Artifacts

### Safety: Read Before Running With Real Models

The demo, the tests and every command in 7.1 start no model and are safe to run anywhere. The commands that call a model (`chat`, `say`, `retry`, `index`, `compare -yes`, and anything driven through `cx serve -mcp` or `cx ui`) are different:

- **Permission prompts are bypassed.** A headless run cannot answer a prompt, so each adapter passes the vendor CLI's own bypass switch (`--permission-mode bypassPermissions` for `claude`, `approval_policy="never"` for `codex`, `--approval-mode yolo` for `gemini`, and the equivalents for the others). Inside its run a model can read, write and execute whatever the process is allowed to.
- **macOS: a sandbox contains each run, on by default.** The Seatbelt profile lets a run write only to its packet and run directories, read the system and the CLI's own configuration, and reach the network only through the hosts its provider needs. The `researcher` role is given the open network.
- **Linux: there is no sandbox.** `cx` refuses to start a CLI and says why. It runs only after you set `sandbox: false`, and then the CLIs are unconfined: a model can touch anything your user account can. `cx` prints a warning each time it runs this way. Use a container, a virtual machine or a throwaway account.
- **Windows is not supported.** The code does not build there.
- **Your conversation is sent to the providers you enable.** Every packet holds verbatim text from the workspace. Restrict providers with `enable:` in the configuration, and keep anything you would not send to a vendor out of the workspace.
- **Code changes stay in a worktree.** A coder unit edits a fresh `git worktree`, a change outside the path allowlist is rejected, and nothing is merged for you. The verify command you configure runs outside the sandbox.

[SECURITY.md](SECURITY.md) has the same list with the dashboard and MCP server, and how to report a vulnerability.

### Prerequisites

- **macOS or Linux.** macOS is the tested platform and the only one with a sandbox.

- **Go 1.26+** (pure Go binary; zero CGo dependencies; uses `modernc.org/sqlite` and `gopkg.in/yaml.v3`)
- **Python 3.9+** (optional, for standalone public-benchmark & replay scripts)

### 7.1 Zero-API-Key Replays & Benchmarks

Every command below runs locally in temporary directories **without spending model calls or requiring API keys** (step 5 downloads three public Go API files):

```sh
# 1. Build the cx binary and run the built-in deterministic engine showcase
make demo

# 2. Run the scripted meeting-reschedule supersession & byte-verification replay
python3 examples/replay/run.py

# 3. Reproduce the 1,000,000-event revisioned query & unknown-scope withholding benchmark
python3 docs/experiments/coverage/bench.py --records 1000000

# 4. Reproduce the 1,000,000-turn external paged omission catalog benchmark
go run ./docs/experiments/packet_scale -catalog

# 5. Reproduce the Go 1.24 -> 1.25 -> 1.26 API ledger invalidation replay (1,128 queries)
api=$(mktemp -d)
for v in 24 25 26; do
  curl -fsSL "https://raw.githubusercontent.com/golang/go/go1.26.0/api/go1.$v.txt" -o "$api/go1.$v.txt"
done
python3 examples/answer_watch/replay.py "$api"
```

Instructions for reproducing the pinned **LongMemEval-S**, **LoCoMo-10**, **ContextBench**, and **Value Postings** experiments from upstream datasets are in [docs/experiments/longmem/README.md](docs/experiments/longmem/README.md) and [reports/2026-11-20-large-memory/README.md](reports/2026-11-20-large-memory/README.md).

### 7.2 Interactive & Programmatic Usage

```sh
# Build binary
go build -o bin/cx ./cmd/cx
bin/cx init-config > ~/.contextless/config.yaml   # optional: inspect/customize provider lanes

# Create a Git-backed workspace and chat interactively
bin/cx new arch-review -purpose "Distributed storage engine design & benchmarks"
bin/cx chat arch-review

# Import an existing Claude Code (.jsonl), JSONL, or Markdown transcript verbatim
bin/cx import arch-review ~/transcripts/session.jsonl

# Inspect the compiled packet, Context Query Planner, and Atlas pointer map (no model call)
bin/cx packet arch-review -brief "What did we decide about WAL group commit latency?"
bin/cx explain arch-review "What did we decide about WAL group commit latency?"

# Import & inspect quote-verified Atlas notes, or simulate a two-stage paragraph pull
bin/cx notes arch-review import notes.jsonl
bin/cx notes arch-review
bin/cx packet arch-review -pull t000012#a3,t000040#a1-2 -ask "cache policy;eviction" "How did the cache policy change?"

# Attach a real Git repository with path allowlist and mandatory test gate for coder units
bin/cx attach arch-review storage ~/code/storage -verify "go test ./..." -allow "internal/**,cmd/**"

# Serve as a 12-tool MCP server or open the local Context X-Ray web inspector
bin/cx serve arch-review -mcp
bin/cx ui arch-review -addr 127.0.0.1:8787
```

### 7.3 CLI Surface Summary

| Category | Commands |
|---|---|
| **Conversation & Vault** | `new`, `chat`, `say`, `retry`, `show`, `import`, `list`, `fork` (`-as-of` experimental) |
| **Compiler & Atlas Map** | `packet` (`-brief`, `-adaptive`, `-pull`, `-ask`, `-json`), `explain`, `notes`, `index`, `reindex`, `pack` |
| **Temporal & Ledger** | `query` (`-strict`, `-as-of`, `-resume`), `views` (`-key`, `-as-of`; experimental), `watch`, `stale`, `receipt`, `refresh` |
| **Evaluation & Telemetry** | `probe`, `floor`, `zoom`, `pull`, `compare` (`-yes`), `cache`, `gaps`, `tune`, `runs`, `stats`, `demo` |
| **Repos & Integrations** | `attach`, `merge`, `resume`, `providers`, `serve -mcp`, `ui`, `obsidian export\|brief` |

### 7.4 Verification Suite

```sh
go test ./internal/... -count=1
```

The automated suite drives every engine, compiler, storage, and security path against deterministic subprocess fakes (real OS processes, zero live LLM calls): lossless atom tiling and byte-span addressing, stemmed FTS5 + value-lane RRF, Atlas quote verification and token-bounded pulls, a $1,200$-turn ($\approx 1.3\text{M}$-token) scale invariant check ensuring every turn is either verbatim or cataloged, $400\text{ KB}$ oversized message chunking, multi-provider plan fan-out and quote auditing, quota cooldown failover with critique carry-over, cross-family reviewer independence, `git worktree` allowlist enforcement, and $800$-turn planted-fact recovery regression floors (`TestProbesRecoverPlantedFacts`).

---

## Repository Map

| Path | What it is |
|---|---|
| `cmd/cx`, `internal/` | The `cx` binary and engine |
| `docs/` | Architecture, the context compiler's design and measurements, the roadmap, and reproducible experiments |
| `reports/` | Dated experiment reports with their result files |
| `examples/` | Model-free replays, and the Obsidian studies |
| `idea-vault/` | One note per idea or result; each evidence line quotes the file it cites, and `docs/experiments/vault_graph.py` checks the quotes |
| `research_notes/` | Literature notes written with model assistance while designing the system. Their citations were not all checked against the sources; treat them as leads, not as a survey |

## License, Credit & Citation

ContextlessAI was designed and developed by **Tom Maciejewski**.

The code and documents are released under the [Apache License 2.0](LICENSE): you may use, change and redistribute them, including commercially, provided you keep the copyright and [NOTICE](NOTICE) and say what you changed. Result files derived from third-party benchmarks keep their sources' terms, including a non-commercial condition for LoCoMo ([THIRD_PARTY.md](THIRD_PARTY.md)).

If you use the software or its results in research, please cite it; [CITATION.cff](CITATION.cff) has the entry, and GitHub shows it under "Cite this repository". Contributions are welcome under the same license ([CONTRIBUTING.md](CONTRIBUTING.md)).

## Lineage

- **`internal/providers`**: Generalized from *BookWriter v2*: subprocess adapters across seven CLI harnesses plus opt-in SSE APIs, macOS Seatbelt sandboxing, token-window watchdogs, compaction taint detectors, and stream telemetry parsers.
- **`internal/packet`**: Generalizes *BookWriter v2*'s pre-spawn sizing, escalation ladder, and drop tiers to paragraph-addressed conversational records, multi-lane RRF, and cache-stable brief topologies.
- **`internal/route`**: Unifies *BookWriter v2*'s role-preference tables and cooldown ladders with *MultiAIWorkspace*'s machine-wide SQLite concurrency lanes and empirical run statistics.
