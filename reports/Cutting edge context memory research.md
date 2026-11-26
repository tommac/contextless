# Prove the verbatim compiler before extending it

The 2025 to 2026 literature supports ContextlessAI's foundation and does not support most of what people want to add to it. Immutable verbatim text, retrieval over that text, code that assembles the context, and evidence addressed by stable id are the patterns that keep surviving independent tests, while LLM-rewritten memory, token-level compression and default knowledge graphs repeatedly tie or lose to BM25 over raw text. That is encouraging and it is not evidence about this system: **no model has yet answered a question from a ContextlessAI packet**, so every local number measures whether text reached the packet, not whether an answer was right or what it cost. The next gain comes from measurement and from pricing context correctly, not from a richer memory structure. A strong PhD researcher would first repair the instrument (the usage records cannot currently produce billed tokens), then establish what the provider caches actually do, then run one paired pilot of whole history against plain BM25 against the current packet. Four contributions look genuinely new after checking the notes for prior art: context selection priced by prefix-cache state, verified citations and recall faults as ranking labels, atom-level causal tests of memory use, and a billed cost-quality frontier for verbatim memory. Atoms, recall as paging, append-only prefixes, levels of detail and model-free link graphs all have close prior art and should not be claimed as novel. Of the first three steps, the first needs no approval, and the second and third spend model calls and need the owner's go-ahead.

## Most published gains are smaller and weaker than their headlines

Scores from different papers cannot be compared. On LoCoMo alone, the same full-context baseline ranges from [56.92%](https://arxiv.org/html/2608.03463) to [72.90%](https://arxiv.org/html/2504.19413) to [92.62%](https://github.com/dial481/locomo-audit) depending on reader, judge and prompt, a spread larger than almost any claimed system gain. The tables below therefore separate results by how much they can be trusted. Absence claims ("no paper does X") mean the research notes found none, not that none exists.

### What this project has measured itself

| Result | Number | What it does not show |
|---|---|---|
| Atoms against six whole turns at the 8k budget | 107 against 103 of 120 distinct questions; 52 against 48 of 60 fresh ones; packets 6 to 44 percent smaller (`docs/CONTEXT-COMPILER.md` 2b) | A difference. The paired test gave p about 0.5. With 18 discordant questions in 120 (15 percent), my power calculation puts the smallest detectable gap near 10.5 points; this one is 3.3 |
| Atoms at a third of the tokens (4k) | 45 against 48 of 60 fresh | The Phase 1 bar was missed |
| Whole history when it fits | 30 of 30 sources against 29 for selection, 83,100 against 18,800 tokens | That a model answers better from it. This is presence only |
| `cx cache` replay with spine and epoch 4 | Median shared prefix 37, 45 and 44 percent | Any actual cache hit |

The replay supports one calculation the docs do not yet make. Using its figures, turning the spine and epoch on saves 28.6, 37.9 and 28.1 percent if every call hits and costs 6.4, 6.2 and 11.6 percent more if none do. **The change breaks even once 18, 14 and 29 percent of calls hit** (my arithmetic, at read price 0.1 and write price 1.0, excluding the CLI's own prefix). The live test therefore has a concrete target: it must distinguish a hit rate above roughly 14 to 29 percent from zero.

### Measured, peer reviewed

| Finding | Evidence | Meaning here |
|---|---|---|
| Focused evidence can beat full history | LongMemEval, GPT-4o: [0.870 with evidence sessions, 0.606 with full history](https://arxiv.org/html/2410.10813) | Whole history is not automatically the quality ceiling |
| Topical segments beat turns, and full history lost | SeCom: [88.81 at 820 tokens against 63.85 at 19,287](https://arxiv.org/html/2502.05589) | Supports small verbatim units, on older readers |
| Index value tracks how many questions need several units | [Graph methods +13.0 points over a grep agent on LoCoMo (60.7% structural), -15.4 on HotpotQA (17.7%)](https://aclanthology.org/2026.aacl-long.211.pdf) | Measure the local structural rate before building links |
| Some "graph" gain is chronology | Synapse: [removing graph structure halves temporal F1 (50.1 to 25.4), leaves multi-hop flat (35.7 to 35.2)](https://arxiv.org/html/2601.02744) | Adjacent-turn expansion is the required control |
| Convex score fusion beat RRF | [Bruch et al.](https://arxiv.org/abs/2210.11934), on document retrieval | RRF k = 60 is a default, not a result |
| Models answer wrongly rather than abstain on insufficient context | [Joren et al., ICLR 2025](https://arxiv.org/abs/2411.06037) | Presence metrics miss this failure |

### Measured, preprint or workshop

| Finding | Evidence | Meaning here |
|---|---|---|
| Fewer visible tokens need not cut the bill | [2,848 Claude Code runs: RTK-ML removed 38.4% of tool-output tokens, cost +6.8%; Headroom +48.4%](https://arxiv.org/html/2607.12161) | Report billed cost per success |
| Stable prefixes save most of the input cost | [Cost down 27.8% to 81.4% across models and strategies](https://arxiv.org/html/2601.06007v2) | The stable-first brief is right in principle |
| Masking old tool output matches summaries | [SWE-bench, n=500 per model; only two significant solve-rate changes, both losses on one model](https://arxiv.org/html/2508.21433v3) | No summarizer is needed for coding history |
| Compaction loses constraints | [Violations 0% to 30% after compaction](https://arxiv.org/abs/2606.22528); [17% of side constraints retained](https://arxiv.org/abs/2608.11242); [5% to 64% stable damage across 14 summarizers](https://arxiv.org/html/2610.31710v2) | Strongest external case for the verbatim rule |
| Verbatim files beat condensed memory | [PersonaMem 78.1 against 37.5 for curated; REALTALK 77.6 against 41.2 for condensed](https://arxiv.org/html/2607.26637v1) | Same |
| BM25 is a hard floor | [GroupMemBench: BM25 43.22, above Mem0 25.73 and GraphRAG 20.56](https://arxiv.org/html/2605.14498); [MemoryAgentBench: 41.5, level with HippoRAG-v2 41.6](https://arxiv.org/html/2507.05257) | Every lane must beat BM25 alone |
| Light structure over verbatim text can help | [EdgeMem, no LLM at build: flat BM25 49.74 to 61.01](https://arxiv.org/html/2609.05553v1) | The most relevant structural result, and close prior art |
| Selection pays when the budget is tight | [Reranking +17.4 points keeping 3 of 30, +1.5 keeping 20 of 30](https://arxiv.org/pdf/2609.34227) | Test selection at the real pool ratio |
| One small embedder is the best-supported addition | [BM25 81.3 to 90.5 Recall@10 on LongMemEval with Qwen3-Embedding-0.6B; a larger embedder alone fell to 59.2](https://arxiv.org/pdf/2610.13277) | Bigger is not safer. The table is from authors with their own method in it |
| Validity metadata beats ranking for stale values | [BM25 with validity: stale exposure 0.000; without: 0.802](https://arxiv.org/html/2610.10265), on a small synthetic family | Key assignment is the hard part: false merges removed current values |
| Evidence present does not mean answer right | [FullText recall 100, accuracy 77.40](https://arxiv.org/html/2608.03463); [73.3% of sampled failures were reachable but unsolved](https://arxiv.org/html/2606.17328) | Directly limits what `cx probe` can show |

### Vendor reported

| Claim | Status |
|---|---|
| [LCM: verbatim store plus summary DAG, 74.8 against 70.3 for Claude Code on OOLONG, behind at 8K and 16K](https://arxiv.org/html/2605.04050v1) | The nearest architecture. One split, no independent replication, and the authors say the agent is not guaranteed to retrieve |
| [Zep: 71.2 against 60.2 full context on LongMemEval-S at 1.6k tokens](https://arxiv.org/html/2501.13956) | No ablation isolating the graph |
| [Mem0 on LoCoMo: full context 72.90, above its own 66.88 and 68.44](https://arxiv.org/html/2504.19413) | Its own table favors full context on quality |
| [Mastra 94.87% on LongMemEval-S, 84.23% with gpt-4o](https://mastra.ai/research/observational-memory) | Reader alone moves it 10.6 points; configurations were run repeatedly |
| [Chroma Context Rot: every model did better with focused input on LongMemEval](https://www.trychroma.com/research/context-rot) | Many results appear only in charts |
| [LangChain: deterministic evidence graph, +27% accuracy over its wiki](https://www.langchain.com/blog/building-agent-memory-openwiki-graphs) | Question subset unstated |
| [Anthropic Contextual Retrieval: top-20 misses down 49%, 67% with reranking](https://www.anthropic.com/news/contextual-retrieval) | A model call per chunk at index time; a [smaller independent test](https://arxiv.org/html/2504.19754) found modest gains |

### Disputed or unreliable

| Item | Problem |
|---|---|
| Mem0 against Zep on LoCoMo | [Zep's 84% was conceded wrong, then 75.14%; Mem0 says 58.44%; closed without agreement](https://github.com/getzep/zep-papers/issues/5) |
| LoCoMo scores above about 93.6 | [An audit finds 6.4% of gold answers wrong, a 93.57% ceiling, and a judge accepting 62.81% of vague wrong answers](https://github.com/dial481/locomo-audit). Higher scores are not evidence of better memory |
| EverMemOS on LoCoMo | [92.32% claimed, 38.38% reported by a third party](https://github.com/dial481/locomo-audit) |
| MemPalace | [100% LoCoMo used top_k=50, effectively the whole conversation; 96.6% is recall, with QA about 67.2%](https://arxiv.org/pdf/2604.21284) |
| Grep against vectors | [Grep won all 10 inline pairs on a LongMemEval-S subset](https://arxiv.org/html/2605.15184), while [a grep agent scored 17.0 elsewhere](https://aclanthology.org/2026.aacl-long.211.pdf). Harness and result format differ |
| Google cached-read price | [One paper uses 0.25x](https://arxiv.org/pdf/2607.19214); [Vertex documents a 90% discount](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/context-cache/context-cache-overview) |
| "Reliability without Validity" | [Withdrawn by arXiv admin on 2026-11-03](https://arxiv.org/abs/2606.19544). Do not cite it |

### Merely proposed, with no measurement behind the idea

[Karpathy's LLM Wiki gist](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f) contains no measurements. [Conversation Tree Architecture](https://arxiv.org/html/2603.21278v1) has a prototype and no evaluation. LSM tiers and FM-indexes for agent memory turned up no implementing paper. Inside this project, levels of detail, personalized PageRank expansion, atom titles, working-set prefetch and learning from citations are all still proposals, and so is every design I add below.

## Ten popular ideas fail for this design

| Idea | What the evidence says | Decision |
|---|---|---|
| LLM summaries as memory | [Summaries tied or lost to masking](https://arxiv.org/html/2508.21433v3) and [drop constraints](https://arxiv.org/abs/2608.11242) | Keep out of the read path. A summary may only ever be a navigation aid |
| Token-level prompt compression | [LLMLingua 39.3 against 56.0 uncompressed on AppWorld](https://arxiv.org/pdf/2510.00615); [patch applicability fell from 27/40 to 15/40](https://arxiv.org/html/2607.12161) | Reject. It also breaks exact quote checks |
| Knowledge graph as the default store | [On accurate retrieval, BM25 scored 60.5 against 40.9 for GraphRAG and 37.5 for Zep](https://arxiv.org/html/2507.05257); [HippoRAG 2's stored form could answer about 10% of sampled AMABench questions](https://aclanthology.org/2026.aacl-long.211.pdf) | No extracted triples. Links only as an index over atoms |
| Model-written Zettelkasten notes or a compiled wiki | [A-MEM lost to the plain baseline on GPT-4o single-hop, 48.43 against 61.56](https://arxiv.org/html/2502.12110); [a wiki used about 21 times the tokens per query, and decomposition RAG recovered most of its advantage](https://arxiv.org/html/2605.18490) | Reject as storage |
| Backlinks and evergreen titles as a proven gain | No controlled measurement of either was found; the nearest test is [header-chain prefixes on one private knowledge base](https://arxiv.org/html/2608.00824). [Real heading paths beat LLM-written context prefixes](https://arxiv.org/html/2610.10170) | Test real structural keys first; model-written titles come last |
| Route down a hierarchy, then rank | [Section routing lost to flat search, -0.033 and -0.015, with first-stage recall as low as 0.17](https://arxiv.org/html/2610.10170) | Use structure as a score feature, never as a gate |
| AdaGReS-style "guaranteed" selection | [Its bound is cardinality based](https://arxiv.org/html/2512.25052), and [plain MMR beat it, 0.9222 against 0.7771](https://arxiv.org/html/2610.29234) | Use MMR as the baseline to beat |
| Learned sufficiency gates | [Unsupervised signals correlate about 0.2 to 0.3 with retrieval benefit](https://arxiv.org/html/2604.07985); [a CRAG replication labeled 88.3% of ARC questions Ambiguous](https://arxiv.org/html/2603.16169v1) | Use signals to allocate budget and word the brief, never to suppress recall |
| Branching raises answer quality | [+2.5% on synthetic scenarios with a GPT-4 judge, no token counts](https://arxiv.org/html/2512.13914v1), confounded with simply removing distractors | Treat as unproven; test a filter on the linear log first |
| A LoCoMo leaderboard position | [Labels and judge are audited as flawed](https://github.com/dial481/locomo-audit) and [conversations fit in context](https://www.getzep.com/blog/lies-damn-lies-statistics-is-mem0-really-sota-in-agent-memory/) | Never a headline metric |

Two opposite slogans also fail. "Less context is always better" is contradicted by [no measurable threshold effect on two 1M-context models](https://arxiv.org/abs/2610.31778), and "the full transcript always wins" is contradicted by the LongMemEval and SeCom rows above. On the agent side, the earlier parallel-agents evidence applies unchanged: [multi-agent coordination averaged -3.5% across 260 configurations](https://arxiv.org/html/2512.08296v2), so a shared writable memory for units has no support. Records passed by reference remain the right shape.

### Five statements in the project docs should change

`docs/ARCHITECTURE.md` 3c says the whole transcript "beat every selection". Only presence was measured, so that sentence should say so. `docs/CONTEXT-COMPILER.md` section 5 says the Token Reduction bill rose "because cache creation and reads dominate input cost"; the paper [attributes the rise to changed trajectories and does not demonstrate cache invalidation](https://arxiv.org/html/2607.12161). The same section quotes 41 to 80 percent cost savings and 13 to 31 percent faster first token, but [the paper's own tables run from 27.8% to 81.4% on cost and include a first-token regression](https://arxiv.org/html/2601.06007v2). Section 4 step 7 claims a known bound for greedy selection with a redundancy penalty, yet the bound needs [a monotone submodular objective, selection by gain per token, and a best-single-item fallback](https://ar5iv.labs.arxiv.org/html/2008.05391), which gives at least 0.405 of optimal; a subtractive penalty breaks monotonicity. Section 10 reports 46 against 48 where section 2b's table gives 45.

## Four contributions look new, and the attractive architecture ideas do not

The novelty judgments below come only from the five research notes. Each "not found" must be rechecked with a targeted search and by reading the listed unread papers before anyone claims priority.

| Candidate contribution | Closest prior art in the notes | What would be new | Novelty judgment |
|---|---|---|---|
| **Prefix-priced context selection**: choose which atoms to send given what the provider has cached | [ClawVM's utility-per-token assembly](https://arxiv.org/html/2604.10352), [Budget-Aware Routing](https://aclanthology.org/2026.findings-acl.2114.pdf), [TokenPilot](https://arxiv.org/html/2606.17016), [Pricing the Context Window](https://www.preprints.org/manuscript/202610.0614) | A joint optimizer with the prefix-closed constraint, validated against real bills and answer quality | **Likely new, uncertain.** The last paper prices cached tokens with a shadow price and could not be read (HTTP 403). It must be read first |
| **Verified citations and recall faults as ranking labels** | [Counterfactual learning to rank](https://www.cs.cornell.edu/people/tj/publications/joachims_etal_17a.pdf), [generative relevance feedback](https://aclanthology.org/2026.findings-eacl.298.pdf), and the warning that [up to 57% of citations can be unfaithful](https://arxiv.org/abs/2412.18004) | Measuring whether a quote-checked citation predicts causal use, then training on it | **Likely new.** No use of LLM citations or recall requests as labels was found. The strongest candidate |
| **Atom-level causal evaluation** with removal, sufficiency and value swap on a replayable log | [ContextCite](https://arxiv.org/html/2409.00729v1), [Compact-Bench paired replay](https://arxiv.org/html/2610.31710v2), [MemTrace](https://arxiv.org/html/2606.17328), [UTILMEM](https://arxiv.org/pdf/2608.30508), [When Retrieval Is Not Enough](https://neurips.cc/virtual/2026/169567) | A validated value-swap protocol for closed APIs, applied across memory systems | **Likely new as a protocol, moderately uncertain.** The last two were seen only as snippets |
| **Billed cost-quality frontier for verbatim memory** | [Total Recall at What Cost](https://arxiv.org/html/2608.11879), [Beyond the Context Window](https://arxiv.org/abs/2603.04814) | Real cache fields, latency, paired quality and a Pareto analysis | **Partly new.** No formal Pareto report was found, but the question is already being asked, so speed matters |
| Natural tangent-and-return evaluation | [Beyond Continuity](https://arxiv.org/html/2605.09268) concatenates conversations; [ContextBranch](https://arxiv.org/html/2512.13914v1) is synthetic | Tangents that really occurred and return to the main thread, judged by downstream answers | **Gap is real, publishability is weak**: the transcripts are private and [new benchmarks are crowded](https://neurips.cc/virtual/2026/loc/sydney/events/evaluations-datasets-2026) |
| Lexical submodular selector for transcripts | [Lin and Bilmes](https://aclanthology.org/N10-1134.pdf), [MMR comparisons](https://arxiv.org/pdf/2605.00336) | The application and a density-against-raw-gain test | **Incremental.** Worth building, not a headline |

**Not novel, whatever their engineering value:** verbatim addressable spans ([LCM](https://arxiv.org/html/2605.04050v1), [ARC](https://arxiv.org/abs/2607.25066), [EdgeMem](https://arxiv.org/html/2609.05553v1)); recall as paging ([Pichay](https://arxiv.org/html/2603.09023v1), [ClawVM](https://arxiv.org/html/2604.10352)); levels of detail with deterministic tie-breaks (ClawVM again); immutable session versions ([CMV](https://arxiv.org/html/2602.22402v1)); lexical evidence gates ([MemTxn](https://arxiv.org/html/2607.27834)); and model-free hubs ([EdgeMem](https://arxiv.org/html/2609.05553v1), [LinearRAG](https://arxiv.org/html/2510.10114v4)). A previous-value benchmark is also not open ground: one note reports none, but [MemTrace already probes current, historical and change-trajectory questions](https://arxiv.org/html/2606.17328).

### The sharpest open problem is a priced knapsack with a prefix constraint

An atom gets the cached price only if everything before it in the prompt is unchanged, so the cheap set is prefix-closed and ordinary knapsack selection does not apply. Formally, let P0 through PK be the reusable prefixes of the previous prompt, ending at cacheable boundaries. Choose one prefix and a suffix S to maximize coverage f(Pk ∪ S), paying the read price for Pk and the fresh price for S, under one billed budget. For a fixed k, the marginal objective is still monotone submodular, so gain-per-token greedy with the best-single-item fallback gets at least 0.405 of that k's best marginal value. Because f(Pk) is kept in full, taking the best of the K+1 runs keeps the 0.405 factor overall. **This is my derivation and has not been reviewed.** It stops holding with a second hard constraint such as the context window, or with a per-token distractor penalty, and both matter in practice: [longer input can hurt even when the evidence is present](https://arxiv.org/abs/2510.05381v1). The multi-call version is harder and more interesting, because writing a prefix now pays only if later calls arrive inside the lifetime. [A write at 1.25x with reads at 0.1x pays back after one read](https://platform.claude.com/docs/en/build-with-claude/prompt-caching), which makes it an online rent-or-buy decision driven by the gap between messages. The catalog spine and epoch alignment are one hand-picked point in this space.

### Determinism is the instrument that makes these contributions possible

A deterministic compiler over immutable text means two arms differ only by a known diff of atoms. Most memory systems change extraction, retrieval and wording together, so they cannot do this, and no study of byte-level context reproducibility was found. Reproducibility alone is not a paper. Its value is that it makes removal, swap and replay experiments clean. The quote-checked unit records add something rare: citations whose text a program has confirmed. That is where the multi-agent work feeds the memory research, since every plan run produces verified evidence links.

## A PhD researcher would fix the instrument, then test one claim at a time

A strong researcher would not start by building Phases 3 to 5. They would ask what result would change the design, and they would notice that the current harness cannot yet answer it. They would then write the analysis plan before seeing outcomes, keep a locked test set, and aim for an evaluation-led first paper, since [NeurIPS Evaluations & Datasets welcomes audits, reproductions and negative results and does not require beating prior work](https://blog.neurips.cc/2026/03/23/introducing-the-evaluations-datasets-track-at-neurips-2026/). A natural thesis is that **context should be compiled as a priced, reproducible, causally tested program rather than stored as a smarter memory**. Three papers fall out of it: causal evaluation of memory use, prefix-priced selection for a systems venue, and verified citations as feedback for an information retrieval venue.

### Reading the code changes the plan

I read the relevant code without changing it. Four facts block honest billing and context comparisons today. `Usage` in `internal/providers/types.go` has input, cached and output fields but no cache-write field. `adapter_claude.go` adds `cache_creation_input_tokens` into `InputTokens`, so a write at a premium looks like ordinary input, and `adapter_codex.go` ignores the [`cache_write_input_tokens` field Codex emits](https://github.com/openai/codex/blob/main/codex-rs/exec/src/exec_events.rs). `TrialResult` in `internal/engine/compare.go` totals only input and output, and its arms are single call, plan and a matched profile, with no full-history, BM25 or recency arm. So `cx compare` as built answers "plan or single call", not the question this report is about. Separately, each run gets its own directory (`runDir/ws` in `call.go`), Claude runs use `--no-session-persistence`, and the brief arrives as one stdin message, so the caller does not place cache breakpoints at all. That combination predicts misses on [Anthropic](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) and [GPT-5.6 and later](https://developers.openai.com/api/docs/guides/prompt-caching), where reuse stops at block or message boundaries, and [Claude Code's startup context includes the working directory](https://code.claude.com/docs/en/prompt-caching).

### The first three things to do

| Order | Action | Approval | Spend |
|---|---|---|---|
| 1 | **R1: repair the measurements and record the offline floor** | **None needed**: code and local replay only | No model calls, no download |
| 2 | **R2: run the cache truth table** | **Owner approval required** | About 300 short calls and 6M input tokens, much of it the CLI's own prefix |
| 3 | **R3: run the paired answer pilot** | **Owner approval required** | About 400 calls and 20M to 25M input tokens before caching |

Neither step 2 nor step 3 needs a download. R9 and R10 do, and **the disk is nearly full (free space moved between about 117 MB and 7 GB on 2026-11-18)**, so check space before fetching any dataset or model. Step 2 comes before step 3 because its result decides whether the pilot's whole-history arm can ever be cheap. The two can share the repaired harness.

### The prioritized agenda

Power figures are my own exact McNemar calculations (two-sided, alpha 0.05, 80% power), and they reproduce the table in the evaluation notes. "Discordant" is the share of items where two arms differ in correctness. Non-inferiority sample sizes use the normal approximation with a one-sided 5% test. Local results come from only three conversations, two of which overlap, so they cannot support general claims however many probes are added: questions from one conversation are not independent. Confirmation belongs to R9.

| Rank | Item | Model calls | Download | Owner approval |
|---|---|---|---|---|
| R1 | Repair instrument, offline floor | No | No | No |
| R2 | Cache truth table | Yes, about 300 | No | **Yes** |
| R3 | Whole history, BM25 and current packet pilot | Yes, about 400 | No | **Yes** |
| R4 | Excerpt readability | Yes, about 240 | No | **Yes** |
| R5 | Knapsack-correct lexical selector | No offline | No | No, until a live check |
| R6 | Prefix-priced selection | Offline first, then about 100 to 200 | No | **Yes for the live stage** |
| R7 | Causal use of memory | Yes, about 300 to 400 | No | **Yes** |
| R8 | Verified citations and faults as labels | Yes, about 400 | No | **Yes** |
| R9 | Locked external benchmark | Yes, thousands | Dataset | **Yes** |
| R10 | Local embedding or late-interaction lane | No for retrieval | Model | **Yes** |
| R11 | Value chains and time intent | No offline | No | No, until a live check |
| R12 | Link graph against adjacency | No offline | No | No, until a live check |
| R13 | Tangent filter against branching | Yes, about 300 | No | **Yes** |
| R14 | Plan against single call | Yes, about 100 to 200 | No | **Yes** |

#### R1. Repair the instrument and record the offline floor

| Field | Plan |
|---|---|
| Hypothesis | Offline conclusions change once probes rank the way live calls do, since [28 of 150 ranked sections currently differ](../docs/CONTEXT-COMPILER.md) |
| Experiment | Store uncached input, cache read, cache write, output and reported USD separately for each provider. Add context arms to the compare path. Index the message before ranking in `cx probe`. Allow several accepted answers per probe. Recover real message times, since imported turns carry the import date, and compute the share of gaps under 5 minutes and 1 hour. Have the owner write at least 100 new probes without viewing the source, and lock them |
| Baselines | Full-history presence, plain BM25 atoms (`-lanes fulltext`), recency only, six whole turns, current fused lanes |
| Metrics | Answer in packet at 2k, 4k, 8k, 16k and 32k; tokens to answer; replayed billed-equivalent tokens; packet build time p50 and p95. Billed tokens do not exist yet at this stage |
| Sample and power | 120 existing distinct questions plus 100 or more new ones. At 15% discordance, 120 detects 10.5 points and 200 detects 8. The old fresh probes are no longer held out for the value lane |
| Falsified if | Rankings and every arm-to-arm conclusion stay the same after the fixes. The repair is still needed for billing |
| Needs | No model calls, no download. Owner time for the probes |

#### R2. Establish what the caches actually do

| Field | Plan |
|---|---|
| Hypothesis | A shared start inside one stdin block gets no brief-level hit on Claude or GPT-5.6 and later, and per-run directories also prevent reuse. A separately delivered block in a stable directory is required |
| Experiment | Send A+B then A+C with A of about 7,000 tokens, above every minimum in the notes, including [Vertex's 6,144 for some Gemini 3 models](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/context-cache/context-cache-overview). Cross one block against separate blocks (where the CLI accepts them) and stable against per-run directory, on Claude, Codex and one Gemini route with API-key caching. Add an identical-prompt control, which must hit or the harness is broken. Keep the output tiny |
| Baselines | A drawn from a real whole-history brief and from a plain BM25 brief; the current brief order |
| Metrics | Cache read and write tokens, uncached input, billed-equivalent tokens, first-token and total latency, hit rate by condition |
| Sample and power | 20 pairs in the best condition and 10 in each other, per provider, about 300 calls. Zero hits in 20 gives a 95% upper bound of 13.9%, just below the lowest break-even of 14%. Zero in 10 gives 25.9% |
| Falsified if | Single-block A+B and A+C hits on Claude or GPT-5.6+ exceed the bound, or separate blocks in a stable directory still show no hits |
| Needs | **Model calls, approval required.** No download. If no CLI route hits, leave the spine and epoch off, and decide separately whether a direct API lane is worth building |

#### R3. Test whole history, BM25 and the current packet on answers

| Field | Plan |
|---|---|
| Hypothesis | The current 8k packet is within 8 points of whole history at under 40% of its billed input, and beats plain BM25 atoms enough to justify the extra lanes |
| Experiment | Same reader, same instructions, one call per arm, no recall in the first pass. Use one profile whose window holds the 290k session for every arm, such as the seeded 500k GPT-6.1 Sol profile; repeat the arms that fit on a second provider if the pilot is promising. Rerun 10% of items to measure reader noise |
| Baselines | Closed book, **full context**, oracle (gold turn only), **plain BM25 atoms at 8k**, recency only, current fused packet |
| Metrics | Correct answer by normalized match, with a blinded owner check of a sample; abstention; separate token classes; billed-equivalent tokens; cost per correct answer; first-token and total latency |
| Sample and power | 60 pilot questions (30 per distinct conversation), 360 calls plus repeats. At 15% discordance this detects only 14 points, so it is for discordance estimates and go or no-go. Non-inferiority at -8 points needs about 145 items at 15% discordance and 193 at 20% |
| Falsified if | Fused is worse than full context by more than 8 points, or BM25 ties fused at equal or lower cost, in which case the extra lanes do not earn their place in packets |
| Needs | **Model calls, approval required.** No download. Depends on R1 |

#### R4. Find out whether a model can read an atom

| Field | Plan |
|---|---|
| Hypothesis | A verbatim atom with its current header answers as well as the whole source turn when retrieval is taken out of the question. This is the project's stated main risk |
| Experiment | Place the known gold evidence directly, rendered four ways: bare atom, atom with header, atom with neighbors, whole turn. Compare with whole history and with the BM25 packet from R3 |
| Baselines | **Full context**, **plain BM25 packet**, whole gold turn |
| Metrics | Correct answer, wrong-referent errors, billed-equivalent tokens, latency, recall requests when recall is enabled |
| Sample and power | 60 items, 240 calls. Detects about 14 points at 15% discordance. A non-inferiority claim at -8 needs about 145 items |
| Falsified if | Header-only atoms fall more than 8 points below whole turns, after which neighbor expansion becomes a priced default |
| Needs | **Model calls, approval required.** No download |

#### R5. Replace fill-by-rank with a knapsack-correct selector

| Field | Plan |
|---|---|
| Hypothesis | Lexical coverage selected by gain per token, with concave saturation, an RRF modular term and a best-single-atom fallback, finds more answers at tight budgets than rank-order filling, with no model. No such study on transcripts was found |
| Experiment | Implement it behind a flag. Compare with raw-gain greedy, MMR using stem overlap, and token-cost exponents below 1. Log the data-dependent optimality bound per query |
| Baselines | **Plain BM25 fill**, current RRF fill, MMR, **full-history presence** |
| Metrics | Answer in packet at 2k, 4k and 8k; distinct query terms and values covered; replayed billed-equivalent tokens; selection p95 time |
| Sample and power | The locked 100+ probes plus the 120 existing, about 200 usable, detect 8 points at 15% discordance. Add multi-answer probes, since single-answer probes cannot reward coverage |
| Falsified if | It does not beat BM25 fill by at least 8 points at 2k or 4k on locked probes, or it loses at 8k |
| Needs | No model calls or download offline. A later confirming arm in R3's harness needs approval |

#### R6. Price selection by what is already cached

| Field | Plan |
|---|---|
| Hypothesis | Selecting over reuse boundaries lowers billed-equivalent input by at least 20% against cache-blind selection, at equal packet recovery |
| Experiment | Replay every user message with the K+1 boundary solver, using measured hit behavior from R2 and message gaps from R1. Sweep the epoch length. Run live only if R2 finds a working route. For plan waves, also test releasing sibling units after the first response begins, because [a cache entry is available only then](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) |
| Baselines | Cache-blind current packet, spine plus epoch as built, **plain BM25 packet**, **full history as an append-only prefix** where it fits |
| Metrics | All token classes, billed-equivalent tokens, cost per correct answer, latency, answer in packet, answer correctness in the live stage |
| Sample and power | Log-cost pairs: at a standard deviation of 0.5, 49 pairs detect a 20% change and 196 detect 10% |
| Falsified if | Savings are under 10% after real hit rates, or quality falls by more than the pre-set margin, or simple fixed epochs match the solver |
| Needs | Offline needs nothing. **The live stage needs model calls and approval.** First read Pricing the Context Window and TokenPilot in full |

#### R7. Test whether answers really depend on retrieved memory

| Field | Plan |
|---|---|
| Hypothesis | A sizable share of correct answers do not depend on the cited atom, and value swaps expose this where citations do not |
| Experiment | On items answered correctly, re-ask after removing the top or cited atom, with only that atom, and with the answer value changed in a scratch copy of the workspace. The vault itself stays untouched. Score with exact matching of the swapped value |
| Baselines | Closed book, **full context with the same swap**, **plain BM25 packet**, current packet |
| Metrics | Dependence rate, sufficiency rate, swap-follow rate, stale-answer rate, billed-equivalent tokens, latency |
| Sample and power | 100 correct items give about plus or minus 9 points on a rate near 70%, and 200 give about 6. Cluster by conversation |
| Falsified if | Nearly all answers follow the swap and fail on removal, in every arm, so the test separates nothing. A protocol that cannot distinguish arms is not a contribution |
| Needs | **Model calls, approval required.** No download |

#### R8. Learn only from signals proven to mean something

| Field | Plan |
|---|---|
| Hypothesis | Removing a quote-verified cited atom changes answers more often than removing a matched uncited atom. Recall faults identify atoms that ranking wrongly left out |
| Experiment | Stage one is the causal check. Stage two trains a handful of inspectable weights (lane weights, recency, cost exponent) from frozen logs, with pairwise constraints that place recalled atoms above low-value selected ones, then replays on locked probes. Normalize scores first, because [unnormalized BM25 defeated one learning loop](https://arxiv.org/pdf/2605.03675) |
| Baselines | **Plain BM25 ranking**, current hand-set weights, weights tuned directly on probes, **full-history presence** |
| Metrics | Change rate after removal, locked-probe recovery, fault rate per call, billed-equivalent tokens, latency |
| Sample and power | 200 matched pairs detect 9.5 points at 20% discordance. Keep parameters few: [two-stage counterfactual results needed 100K to 1M interactions](https://arxiv.org/html/2506.20854), far above one user's volume |
| Falsified if | Cited and uncited removal effects do not differ, in which case citations are not labels and section 7 of the compiler design should drop them |
| Needs | **Model calls, approval required** for stage one. Training itself needs none |

#### R9. Run one locked external benchmark

| Field | Plan |
|---|---|
| Hypothesis | The best compiler configuration beats budget-matched BM25 on billed cost per correct answer, and is non-inferior to full context on accuracy |
| Experiment | Add a LongMemEval importer with no model at ingest. Pin the file version, since [the histories changed in 2025](https://github.com/xiaowu0162/LongMemEval). Tune on 200 questions, lock 300, and freeze the plan with a hash before the single run. Use two readers from different providers. Choose a second dataset at registration; [LongMemEval-V2](https://arxiv.org/html/2605.12493v1) is closest to agent work |
| Baselines | Closed book, oracle, **full context**, **plain BM25 at matched tokens**, BM25 plus embedding if R10 passes, current compiler |
| Metrics | Calibrated accuracy by question type, evidence reached, evidence used, all token classes, cost per correct answer, p50 and p95 latency, Pareto position with intervals |
| Sample and power | 300 locked questions detect 6.5 points at 15% discordance (9 at alpha 0.0125 and 20%). Non-inferiority at -5 needs about 495 at 20% discordance, so use -8 here or a tighter claim only if discordance is low. [Calibrate the judge on about 200 human-labeled items](https://arxiv.org/html/2511.21140v3) |
| Falsified if | The compiler is dominated by BM25 at matched budget, or misses the non-inferiority margin against full context |
| Needs | **Dataset download and many model calls, approval required.** Full context alone is about 34.5M input tokens per reader. Check disk space first |

#### R10. Add one local semantic lane only if it earns it

| Field | Plan |
|---|---|
| Hypothesis | One small embedder, or token matching limited to a lexical shortlist, adds at least 5 points on paraphrased probes without hurting exact and value lookups |
| Experiment | Retrieval only at first. Test dense as a lane, shortlist MaxSim, and convex fusion against RRF. Pin the model hash and use exact search, so packets stay reproducible |
| Baselines | **Plain BM25**, fused lexical lanes, dense alone, **full-history presence** |
| Metrics | Answer in packet at the real token budget, index size, query p95 on this laptop, replayed billed-equivalent tokens |
| Sample and power | 200 paraphrase probes detect about 8 points at 15% discordance. LongMemEval's 500 questions add a public retrieval test with no LLM |
| Falsified if | The gain is under 5 points, or exact lookup falls, or p95 exceeds about 50 ms. [Published results are the authors' own](https://arxiv.org/pdf/2610.13277), and [hybrid sometimes adds nothing](https://arxiv.org/pdf/2608.12365) |
| Needs | **Model download, approval required**, and free disk. No LLM calls for retrieval |

#### R11. Handle changed values with chains, not a filter

| Field | Plan |
|---|---|
| Hypothesis | High-precision chains built from exact value and key matches, plus a rule-based router for current, previous and change questions, cut stale exposure without hiding history |
| Experiment | Build chains in code from the value index, favoring precision, because [false merges removed current values](https://arxiv.org/html/2610.10265). Label old values with their dates. Plant update sequences and also label real ones |
| Baselines | **Plain BM25**, current newest-first exact lane, validity filter alone, **full-history presence** |
| Metrics | Current-value recovery, predecessor recovery, stale exposure, false merges, replayed tokens, build and query latency |
| Sample and power | 100 chains with current, previous and change questions. Published effects are very large, so 100 suffices for detection; false-merge rate needs its own count |
| Falsified if | False merges remove any meaningful share of current values, or the plain newest-first lane matches it |
| Needs | No model calls or download offline. Answer-level confirmation needs approval |

#### R12. Earn the link graph against the adjacency control

| Field | Plan |
|---|---|
| Hypothesis | `same-value` and `answers` edges help questions spanning several atoms beyond what adjacent-turn expansion gives |
| Experiment | First measure the local structural rate. Then write multi-hop probes, because the present single-source probes cannot show a benefit. Treat each expansion as a priced candidate. Score only what fits the budget, since [neighbor credit once made trivial BM25 look like 99%](https://arxiv.org/pdf/2608.12365) |
| Baselines | **Plain BM25**, adjacent-atom expansion, question-answer pair expansion, personalized PageRank, **full-history presence** |
| Metrics | All-evidence-in-packet rate, tokens, replayed billed-equivalent tokens, expansion latency, regressions on single-hop probes |
| Sample and power | 100 multi-hop probes detect 13 points at 20% discordance |
| Falsified if | Adjacency ties the graph, or the local structural rate is low enough that few questions can benefit. Then Phase 3's graph is dropped |
| Needs | No model calls or download offline. Probe writing is owner or sub-agent work |

#### R13. Test tangent handling on tangents that really happened

| Field | Plan |
|---|---|
| Hypothesis | Filtering a linear log by thread and relevance matches explicit branching on answers, using fewer tokens than unfiltered history |
| Experiment | Label real tangents that return to the main thread in the owner's transcripts. Ask questions after the return. Compare inherited context rules, and restate the task near the query as an arm, since [restating it recovered much of the loss](https://arxiv.org/html/2506.00069v1) |
| Baselines | **Full context**, **plain BM25 packet**, current packet, simulated branch that excludes the tangent, filter on the linear log |
| Metrics | Correct answer, carry-over errors, billed-equivalent tokens, latency, segment quality judged downstream rather than by [boundary F1 alone](https://arxiv.org/html/2512.17083v2) |
| Sample and power | 100 return questions detect 13 points at 20% discordance |
| Falsified if | No arm differs from full context on quality, so tangents matter only for cost |
| Needs | **Model calls, approval required.** No download. The data stays private |

#### R14. Run `cx compare` for the question it was built for

| Field | Plan |
|---|---|
| Hypothesis | Plans help only on decomposable requests and do not beat a budget-matched single call overall, as [equal-budget studies](https://arxiv.org/html/2604.02460v1) suggest |
| Experiment | Run the existing blind comparison after the usage repair, on messages stratified by request class |
| Baselines | Single call on the current packet, budget-matched single call, single call on **full context** where it fits, single call on a **plain BM25 packet**, plan |
| Metrics | Blind owner preference, unit claims dropped, billed-equivalent tokens, wall time, runs |
| Sample and power | 30 messages. An exact sign test needs the preferred arm to win about 75% of untied messages (approximate), so only large effects are visible |
| Falsified if | The plan arm wins at matched budget across classes, which would overturn the expected result |
| Needs | **Model calls, approval required.** No download |

## Conclusion

The field has spent two years building ever more elaborate memories, and its best evidence now says the elaboration is often the problem. That moves the frontier for ContextlessAI. The interesting question is no longer how to organize remembered text, but which verbatim bytes are worth paying for on this call, given what the provider already holds and whether the model will actually use them.

The project's unusual asset is that it can ask that question cleanly: immutable text, reproducible packets and citations a program can verify make causal experiments cheap where other systems cannot isolate a change. If R2 shows no reachable cache and R3 shows BM25 matching the fused packet, the honest result is a simpler system, and that would still be worth publishing. If the priced selector and the citation labels hold up, ContextlessAI stops being another memory system and becomes the reference method for deciding what context is worth.
