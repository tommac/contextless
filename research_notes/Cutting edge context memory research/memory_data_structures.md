# Data Structures for Long-Term LLM and Agent Memory (2025 to November 2026)

Evidence labels used below: **[Peer-reviewed]** venue confirmed; **[Preprint]** author-run and not confirmed as peer reviewed; **[Vendor]** evaluated by the company selling the system; **[Third-party]** run by authors with no stake in the named system, though they may be proposing a rival; **[Disputed]** contested in public; **[Claim]** stated but not measured in the material read. Model names are given as each source reports them. Scores from different papers are not comparable unless stated: reader, judge, question subset and retrieval budget all differ.

## Hierarchical trees and summary DAGs: structure, update cost, measured gains

### Takeaway
Trees and summary DAGs show real but modest and setup-dependent gains, and the designs that hold up keep raw text at the leaves. The closest match to an immutable-log engine is LCM (verbatim store plus derived summary DAG), but its evidence is vendor-run, and its summaries are LLM-generated apart from a deterministic truncation fallback.

### Cited Findings

**Summary DAG over an immutable log**
- **LCM** (Ehrlich and Blackman, Voltropy PBC; paper dated 14 Feb 2026, although the arXiv ID 2605.04050 implies a later posting; Opus 4.6 with Haiku 4.5 as auxiliary) **[Vendor] [Preprint]**. Structure: an immutable store holds every user message, assistant response and tool result verbatim; the active context is recent raw messages plus summary nodes; leaf summaries cover message spans, condensed summaries cover other summaries, and provenance links point to parents — [LCM](https://arxiv.org/html/2605.04050v1)
- LCM update path: no overhead below a soft token threshold, asynchronous compaction between soft and hard thresholds, blocking compaction above the hard threshold. Escalation is LLM summary at target T, LLM bullet summary at T/2, then deterministic truncation to 512 tokens; an output is accepted only if smaller than its input, which guarantees convergence. No token or dollar cost is reported — [LCM](https://arxiv.org/html/2605.04050v1)
- LCM results on OOLONG (`trec_coarse` split only): Volt 74.8 vs Claude Code v2.1.4 70.3 on average; gain over raw Opus 4.6 is +29.2 vs +24.7. Claude Code leads at 8K (+13.1 vs +11.2) and 16K (+26.3 vs +25.0); Volt leads from 32K up, including 256K (+18.5 vs +8.5), 512K (+42.4 vs +29.8) and 1M (+51.3 vs +47.0). The authors excluded tasks where the model recognized benchmark data, could not detect such contamination in the raw baseline, and say the agent is not guaranteed to navigate back to prior state — [LCM](https://arxiv.org/html/2605.04050v1)

**Temporal trees**
- **TiMem** (6 Jan 2026; gpt-4o-mini-2024-07-18 for consolidation, Qwen3-Embedding-0.6B) **[Preprint]**: five levels (segment, session, day, week, profile), with parent time intervals containing child intervals. L1 is created per turn; higher levels are created when their window closes. Recall uses two LLM calls per query. LoCoMo LLM-judge: 75.30 ± 0.16 vs MemOS 69.24, MemoryOS 60.79, Mem0 57.79, A-MEM 51.29, MemoryBank 39.77. LongMemEval-S (gpt-4o-mini answers): 76.88 ± 0.30 vs MemOS 68.68, Mem0 64.96. No consolidation token cost is reported, and no full-context or BM25 result appears in the extracted tables — [TiMem](https://arxiv.org/html/2601.02845v1)
- **MemForest** (v1 16 May 2026, v3 6 Sep 2026; Qwen3-4B, Qwen3-30B, Gemma-4-12B; deepseek-v4-flash judge) **[Preprint]**: time-ordered trees whose leaves are atomic facts or original dialogue units, with internal nodes summarizing contiguous intervals. Note the name clash: its "MemTree" is not the 2024 MemTree below. Insertion touches the new leaf plus its ancestor path, O(log_k N) per affected tree; refresh cost scales with distinct dirty nodes — [MemForest](https://arxiv.org/html/2605.23986)
- MemForest results (Qwen3-30B, pass@1): LongMemEval-S 81.80 (Planner) vs Zep Local 71.00, LightMem 70.20, EverMemOS 67.00, Mem0 50.20. LoCoMo categories 1–4: EverMemOS 84.22 narrowly beats MemForest 84.09; under Gemma-4-12B EverMemOS leads 88.57 to 82.66. On a 60-question LongMemEval subset at pass@8, flat top-10 scored 78.3 vs 85.0 (embedding) and 90.0 (LLM plus planner). The authors state their write-path benefit is latency, not total tokens — [MemForest](https://arxiv.org/html/2605.23986)

**Semantic trees and layered indexes**
- **RAPTOR** (31 Jan 2024; arXiv page lists no venue): bottom-up recursive clustering and summarization. Its abstract claims a 20-point absolute accuracy improvement on QuALITY when paired with GPT-4 **[Claim from abstract; table not checked]** — [RAPTOR](https://arxiv.org/abs/2401.18059)
- **MemTree** (v1 Oct 2024, v3 19 Mar 2025; GPT-4o, Llama-3.1-70B) **[Preprint]**: new content descends by embedding similarity against depth-adaptive thresholds; each ancestor on the path is rewritten by an LLM. The paper claims O(log N) insertion and measures 3.27 ± 2.38 LLM calls per insertion on MultiHop RAG vs 1 for MemoryStream — [MemTree](https://arxiv.org/html/2410.14052)
- MemTree results include cases where raw text wins: MSC (15 rounds) full history 95.6 vs MemTree 84.8; QuALITY full text 65.1 vs GraphRAG 62.8, MemTree 59.8, RAPTOR 59.0. On MSC-E (200 rounds), MemTree 82.5 beats MemoryStream 80.7 and full history 78.0. On MultiHop RAG, RAPTOR 81.0 narrowly beats MemTree 80.5 — [MemTree](https://arxiv.org/html/2410.14052)
- **H-MEM** (23 Jul 2025; Qwen, LLaMA 3.2 and DeepSeek-R1 small models; DeepSeek-R1-8B for memory extraction) **[Preprint]**: domain, category and trace layers act as an index, routing by stored child-position pointers to an episode layer. LoCoMo F1 with DeepSeek-R1-7B: 39.43 vs 31.57 for the best baseline. It reports under 100 ms vs over 400 ms for MemoryBank. Memory weights are changed by user feedback and decay, so this design is mutable — [H-MEM](https://arxiv.org/html/2507.22925)
- **EraRAG** (v2 4 Jul 2025; Llama-3.1-8B reader, BGE-M3) **[Preprint]**: stored random hyperplanes hash chunks consistently into buckets, buckets split or merge under size bounds, and only affected segments and their ancestors are re-summarized. It reports up to 57.6% fewer tokens and up to 77.5% less rebuild time than RAPTOR, and about 20 s for a single-chunk insert. Accuracy gains are small: QuALITY 60.25 vs RAPTOR 55.48, HotpotQA 55.39 vs 53.29 — [EraRAG](https://arxiv.org/html/2506.20963)

**Cross-method comparison**
- **"Memory in the LLM Era"** (v3 6 Aug 2026; formatted for PVLDB; Qwen3.5-9B and 27B, GPT-5.4-mini judge, all-MiniLM-L6-v2) **[Third-party for the 12 baselines; the authors also propose a method]**: selected results on LongMemEval F1 with 27B are MemoryOS 51.50, MemTree 49.62, MemOS 44.96, Mem0g 43.36, Mem0 42.23. On LoCoMo with 27B, MemOS 48.88, Zep 46.62, MemoryOS 44.36, Mem0 41.68, MemTree 39.95. Query latency ranges from MemTree 24.5 ms to Mem0g 3,783.1 ms. The authors conclude hierarchy helps, while keeping raw dialogue is essential to avoid semantic loss; MemTree and Zep scale poorly in token cost because of update overhead — [Wu et al.](https://arxiv.org/html/2604.01707v3)
- In that paper's ablation, swapping the proposed method's tree for a graph drops LoCoMo F1 from 43.40 to 40.60, and removing the tree gives 39.01 — [Wu et al.](https://arxiv.org/html/2604.01707v3)

### Inferences
- The strongest transferable pattern is **immutable leaves plus disposable derived nodes**. LCM, MemForest session trees, and the raw-dialogue lesson from Wu et al. all point that way.
- For a deterministic engine, a summary should be treated as a stored artifact keyed by source span, model, prompt and parameters. Regenerating it at context-build time would make the same request non-reproducible.
- Time-interval containment (TiMem, MemForest) suits append-only history: new turns mostly affect the rightmost path, and sealed windows need never change. This is my reasoning; neither paper tests it against an immutable log.
- A plain time or session hierarchy can be built without a model. The summaries are the expensive and non-deterministic part, so structure and summary should be separate layers.
- MemTree's ancestor rewrites and H-MEM's mutable weights do not fit immutability unless expressed as new versions or a separate overlay.

### Gaps
- No independent replication of LCM on OOLONG was found, and no result separates the value of the DAG from Volt's other harness differences.
- I found no head-to-head incremental-update benchmark across RAPTOR successors. adRAP surfaced in search but was not read.
- Most tree papers omit BM25 and full-context baselines on their main tables, so added value over simple retrieval is often unknown.

## Graphs, temporal knowledge graphs and hypergraphs: time, supersession, contradiction, results

### Takeaway
Temporal graphs offer the most developed model of supersession (bitemporal validity plus invalidation instead of deletion), but headline accuracy is heavily disputed and sensitive to setup. Graph gains are clearest on multi-hop and temporal questions, while triple-only stores lose information; the best-supported designs keep source text linked to graph elements.

### Cited Findings

**Temporal models**
- **Zep/Graphiti** (20 Jan 2025; gpt-4o-mini-2024-07-18 for graph construction, BGE-m3) **[Vendor]**: edges carry `t_valid`, `t_invalid`, `t'_created` and `t'_expired`. An LLM compares a new edge with related edges; on a temporally overlapping contradiction the older edge is invalidated at the newer edge's valid time, preserving history. Raw episodes remain linked to derived entities and facts — [Zep](https://arxiv.org/html/2501.13956)
- Zep's own results: LongMemEval-S gpt-4o 71.2 vs full context 60.2; gpt-4o-mini 63.8 vs 55.4; about 1.6k context tokens vs 115k. DMR margins are negligible (gpt-4o-mini 98.2 vs full conversation 98.0). Zep trails on single-session-assistant questions, by 17.7% with gpt-4o — [Zep](https://arxiv.org/html/2501.13956)
- **Mem0 and Mem0g** (28 Apr 2025; GPT-4o-mini) **[Vendor]**: base Mem0 uses LLM-selected ADD, UPDATE, DELETE and NOOP, so contradicted memories can be removed. Mem0g stores entity-relation graphs in Neo4j and marks obsolete relations invalid rather than deleting them — [Mem0](https://arxiv.org/html/2504.19413)
- Mem0's LoCoMo results: full context 72.90, Mem0g 68.44, Mem0 66.88, Zep 65.99, best RAG 60.97, LangMem 58.10, OpenAI 52.90, A-Mem 48.38. Mem0 used 1,764 tokens vs 26,031 for full context; Mem0g added p95 total latency (2.590 s vs 1.440 s) and 3,616 tokens — [Mem0](https://arxiv.org/html/2504.19413)
- **Bitemporal Neo4j store** (Niksarli and Baheti, 29 Jul 2026; Titan Embed v2) **[Preprint]**: an identity node is never changed, content sits in version nodes with valid-time and transaction-time intervals, and updates close the current version. On 60 sampled LongMemEval questions, overall R@10 was only 46.7%, with no external baseline; it indexes user messages only — [Bitemporal store](https://arxiv.org/html/2607.26520)
- **MemTxn** (30 Jul 2026; Qwen and Llama local models, GPT-5.4 family) **[Preprint]**: append-only versions plus an active-version pointer map and durable recovery journal. A lexical gate admits a write only if cited evidence appears in the source and value tokens appear in order, with no verifier model call. FactConsolidation F1 with GPT-5.4: 49.31 vs dense 32.25. The audit used template-generated corruptions, and the gate does not establish semantic truth — [MemTxn](https://arxiv.org/html/2607.27834)
- **T-Mem** (9 Oct 2026; GPT-5.5 and DeepSeek-V4-Flash) **[Preprint]**: builds a query-time forest over original messages with Update, Expand and Root edges, without changing stored conversations. GroupMemBench (GPT-5.5): T-Mem 63.09, HippoRAG2 54.50, StructMem 52.75, Hindsight 51.68, BM25 with expansion 49.26, embedding 48.32. Arranging the same BM25 messages as a forest rather than a flat list added 3.09 to 7.60 points. Edge labels come from per-query LLM calls — [T-Mem](https://arxiv.org/html/2610.12874)
- **AriGraph** (v2 9 Sep 2024; GPT-4) **[Preprint]**: semantic triplets plus episodic vertices holding full observations. An LLM selects outdated triplets, which are removed. TextWorld scores: 1.0 on Treasure Hunt vs full history 0.47, summary 0.33 and RAG 0.33; Cooking 1.0 vs 0.18, 0.52 and 0.36 — [AriGraph](https://arxiv.org/html/2407.04363v2)

**Static graph retrieval**
- **HippoRAG 2** (20 Feb 2025; Llama-3.3-70B, NV-Embed-v2) **[Preprint]**: phrase nodes, passage nodes, synonym edges and Personalized PageRank, with an LLM filtering retrieved triples. Average F1 59.8 vs dense 57.0, HippoRAG 53.1, GraphRAG 49.6, RAPTOR 48.8, BM25 47.7, LightRAG 6.6. The authors attribute much of the gain over dense retrieval to 2Wiki. No incremental update experiment is reported — [HippoRAG 2](https://arxiv.org/html/2502.14802v1)
- **LightRAG** (v3 28 Apr 2025): the abstract claims an incremental update algorithm and dual-level retrieval **[Claim; implementation not examined]** — [LightRAG](https://arxiv.org/abs/2410.05779). HippoRAG 2's run measured it far below the other methods (average F1 6.6) while using 68.5M input tokens to index MuSiQue, vs 9.2M for HippoRAG 2 and 115.5M for GraphRAG. This is a rival's run — [HippoRAG 2](https://arxiv.org/html/2502.14802v1)
- **RAG vs GraphRAG** (v3 4 Mar 2026; Llama-3.1 8B and 70B generators; GPT-4o-mini graph construction) **[Third-party]**: with 8B, NQ F1 is RAG 64.78 vs best graph 63.01; HotpotQA 60.04 vs HippoRAG2 63.01; MultiHop-RAG 67.02 vs 70.27; temporal 30.70 vs 50.60 for Community-Local; null-query accuracy favors RAG (96.01 vs 80.07, and 19.27 for Community-Global). Only about 65.8% of HotpotQA answer entities appeared in the knowledge graph — [Han et al.](https://arxiv.org/html/2502.11371v3)

**Hypergraphs**
- **HyperGraphRAG** (v1 27 Mar 2025, v3 21 Oct 2025; NeurIPS 2025; GPT-4o-mini) **[Peer-reviewed]**: n-ary facts as hyperedges — [HyperGraphRAG abstract](https://arxiv.org/abs/2503.21322). Across its five reported domain columns, overall F1 is 31.30 to 48.71 vs StandardRAG 27.43 to 43.20, a gain of 2.37 to 7.45 points. Construction is 3.084 s and $0.0063 per 1k tokens; query cost is $3.184 vs $1.016 per 1k queries for StandardRAG — [HyperGraphRAG paper](https://proceedings.neurips.cc/paper_files/paper/2025/file/df55ee6e59f8ac4a625219e11fe9ddba-Paper-Conference.pdf)
- **HyperMem** (v2 10 Apr 2026; GPT-4.1-mini answers, GPT-4o-mini judge, Qwen3 embedding and reranker; ingestion LLM unnamed) **[Preprint on arXiv; search results show an ACL 2026 Anthology entry that I did not open]**: topic, episode and fact nodes with weighted hyperedges. LoCoMo 92.73 vs HyperGraphRAG 86.49, MIRIX 85.38, HippoRAG 2 81.62, LightRAG 79.87, Mem0g 68.44. Token use is about 7.5× Mem0. Many memory-system rows come from prior papers, the paper also gives 92.66 elsewhere, and it describes no conflict-resolution mechanism — [HyperMem](https://arxiv.org/html/2604.08256); [ACL listing](https://preview.aclanthology.org/ingest-acl/2026.acl-long.1627/)
- **EdgeMem** (3 Sep 2026; gpt-4o-mini reader; gpt-4o strict judge; spaCy 3.8.11) **[Preprint]**: verbatim turns are vertices; month, entity or keyword, and session hubs are hyperedges. LoCoMo strict judge 61.01 vs CompassMem 58.70, CAM 54.94, MemoryOS 47.20, MemGAS 44.05, A-Mem 39.74, SeCom 38.12. Flat BM25 scored 49.74; session structure raised this to 58.18, co-occurrence to 60.43 and time to 61.01. No supersession model is described in the extracted material — [EdgeMem](https://arxiv.org/html/2609.05553v1)

**The Mem0/Zep dispute [Disputed]**
- Zep (6 May 2025) says Mem0 misconfigured Zep: both speakers mapped to the user role, timestamps placed in message text rather than `created_at`, and searches run sequentially. Its corrected figure is 75.14 ± 0.17 — [Zep blog](https://www.getzep.com/blog/lies-damn-lies-statistics-is-mem0-really-sota-in-agent-memory/)
- Mem0's CTO (8 May 2025) argued that Zep's original 84% wrongly included adversarial category 5 and reported 58.44 ± 0.20 after restricting scoring to categories 1–4 and aligning prompts. Zep acknowledged the scoring error, gave 75.14, rejected Mem0's revised method, and the issue was closed for inactivity on 19 May 2025 — [zep-papers issue #5](https://github.com/getzep/zep-papers/issues/5)
- A later run of self-hosted Graphiti ("Zep Local") in MemForest scored 76.75 on LoCoMo categories 1–4 with Qwen3-30B. That is a different reader and judge, so it does not settle the dispute — [MemForest](https://arxiv.org/html/2605.23986)

### Inferences
- The most transferable temporal design is **derived validity records over immutable turns**: a fact or edge can be closed by a later record while the source turn remains untouched. Zep's bitemporal fields map cleanly to this.
- MemTxn suggests a useful split for determinism: an LLM may propose a fact, but admission (quoted evidence exists in the source) and visibility (declared chronological rule) can be pure code.
- Supersession should be an explicit, inspectable relation between records. Automatic "latest wins" is a policy, not evidence of truth; MemTxn says this directly for its resolver.
- EdgeMem and T-Mem together suggest a graph need not replace text. It can serve as an index or ordering over verbatim turns, which fits this engine far better than Mem0-style fact replacement.
- Hyperedges have a simple relational form here: a hub ID plus member turn IDs. The expressive benefit of LLM-extracted n-ary facts is a separate question from this cheap structure.

### Gaps
- No neutral, fully reproduced Mem0-vs-Zep run with vendor-approved configurations was found.
- I did not find a benchmark that isolates bitemporal modelling from extraction quality and retrieval.
- Graphiti's repository and current behavior were not examined; findings rest on the January 2025 paper and third-party runs.
- HyperMem's ACL status is based on a search result, and the fetched arXiv page listed no venue.

## Systems ideas applied to context: LSM tiers, versioning, content addressing, paging, compressed indexes

### Takeaway
Paging, versioning and content-addressed storage now have concrete agent-specific prototypes, and some are deterministic. I found no paper that actually implements an LSM-tree or a succinct full-text index for agent memory, so those remain engineering analogies rather than tested designs.

### Cited Findings

**Paging and memory hierarchies**
- **Pichay** (Mason, 9 Mar 2026; Claude Code client, Claude Opus; Sonnet 4 and Haiku 4.5 for quality checks) **[Preprint]**: a transparent proxy evicts tool results older than 4 user turns and larger than 500 bytes, replacing them with retrieval handles; one fault pins the page for the session until its content hash changes. Corpus: 857 sessions, 54,170 API calls, 4.45B effective input tokens, 21.8% structural waste — [Pichay](https://arxiv.org/html/2603.09023v1)
- Pichay's results are mixed: offline replay gave a 0.0254% fault rate (354 faults in 1.393M evictions), but one 681-turn live session had 97% faults and thrashed. A collapse dropped the cache hit rate from 100% to 25% for one turn, costing about one recompute of 105K tokens. All data comes from one user — [Pichay](https://arxiv.org/html/2603.09023v1)
- **ClawVM** (11 Apr 2026; EuroMLSys '26) **[Peer-reviewed per the arXiv listing]**: typed pages have stable IDs, provenance and minimum-fidelity floors, at full, compressed, structured and pointer resolutions. Prompt assembly installs pinned minimums, then greedily upgrades by utility per token with deterministic tie-breaks — [ClawVM paper](https://arxiv.org/html/2604.10352); [ClawVM listing](https://arxiv.org/abs/2604.10352)
- ClawVM replay results: mean explicit faults 0 vs 1.5 (compaction hybrid), 10.8 (retrieval plus cache) and 67.8 (retrieval). Median policy overhead is 18 to 44 μs per turn. The main experiments are synthetic replays with very small budgets (120 to 500 tokens), the baselines were built by the authors, and there is no end-to-end task-quality evaluation with a live model — [ClawVM](https://arxiv.org/html/2604.10352)

**Versioned histories**
- **Contextual Memory Virtualisation** (Santoni, 25 Feb 2026; Claude Opus 4.6) **[Preprint]**: sessions form a DAG of immutable JSONL snapshots with branch and trim operations. Across 76 sessions from one user, trimming saved a mean 20% and median 12% of tokens. The median cache-miss penalty was $0.32 and median break-even 38 turns. Reasoning accuracy was not measured — [CMV](https://arxiv.org/html/2602.22402v1)
- **Git Context Controller** (Wu, 30 Jul 2025; Claude 3.5 Sonnet) **[Preprint]**: COMMIT, BRANCH, MERGE and CONTEXT over a `.GCC/` directory. It reports 48.00% on SWE-Bench-Lite vs 43.00% for CodeStory Aide, but compares with numbers from other reports and has no variance analysis — [GCC](https://arxiv.org/html/2508.00031v1)

**Content addressing and storage growth**
- **The Hidden Footprint** (13 Jul 2026; DeepSeek-V4-Flash, with Qwen replications) **[Preprint]**: a roughly 300-line content-addressed store moves strings and blobs over 1 KB into zstd-compressed, hash-referenced storage. Across 70 runs of seven frameworks, retained storage fell by 4.8× to 32.7×, and all 310 retained files passed restore checks — [Hidden Footprint](https://arxiv.org/html/2607.11149v1)
- In its repeated-observation workload, OpenAI Agents appended each message once (growth exponent 0.73, duplication 1.0), while LangGraph, AutoGen and LlamaIndex re-serialized history (exponents 1.74 to 1.95). Garbage collection, deletion semantics and cross-tenant privacy are out of scope — [Hidden Footprint](https://arxiv.org/html/2607.11149v1)

**Compaction without a summarizer**
- **ReCAP** (30 Sep 2026; Qwen3-Coder-30B and gpt-oss-20b) **[Preprint]**: a persistent graph of atomic history blocks stores attention-derived importance and dependency edges; selection needs no model call, though black-box use needs a proxy model pass. On Lost-in-Conversation with Qwen: 85.4 vs full history 65.6 and summarization 75.4. On SWE-Together: 28.5 vs full history 24.4 and summarization 28.7 — [ReCAP](https://arxiv.org/html/2609.40118v1)

**LSM and succinct indexes**
- A targeted search found LSM-tree storage-engine papers and adjacent agent-memory work, but no agent-memory paper implementing LSM tiers — [search-surfaced example: MemTxn](https://arxiv.org/html/2607.27834); [Hidden Footprint](https://arxiv.org/html/2607.11149v1)
- A targeted search for FM-index or suffix-array use in agent memory found only general compressed-index research, for example the Suffixient Array work on repetitive collections — [Suffixient Arrays](https://link.springer.com/article/10.1007/s00224-026-10287-6)

### Inferences
- ClawVM's assembly is the closest published template for deterministic context building: pinned minimums, precomputed representations, utility-per-token upgrades and stable tie-breaks. Its weak point is missing evidence on answer quality.
- Prompt-cache stability should be a first-class constraint. Pichay and CMV both show that structural edits can cost more than they save. A deterministic builder should prefer append-at-tail and infrequent, batched changes to earlier context.
- An LSM analogy is still attractive as engineering: immutable turn segments, sealed per-segment indexes and background merges match append-only history. This is an untested proposal for this domain, and merge policy must not change retrieval results unless the index version is part of the build input.
- Content addressing fits immutable turns directly: hashes give deduplication, integrity and stable references in summary provenance. Deletion and privacy need separate design, because hash-linked immutable stores make erasure difficult.
- A suffix or FM-index could add exact substring search with compact storage, especially for repetitive tool output. Whether it improves recall over BM25 plus regex is unknown.

### Gaps
- No direct comparison of paging policies on a shared agent workload was found.
- No evidence was found on LSM-style memory or succinct indexes for this use. Absence from these searches is not proof that no such paper exists.
- Paging results largely measure tokens and faults. Answer or task quality is thinly tested.

## Do complex structures beat simple ones?

### Takeaway
Often not by much, and sometimes not at all. Across independent and multi-system studies, raw text with BM25, dense retrieval, or file tools repeatedly matches or beats LLM-transformed memories; the better recent results add light structure over verbatim evidence rather than replacing it. No single ranking is safe because reader, judge, retrieval budget and benchmark can change the order.

### Cited Findings

**Simple retrieval matching or beating memory systems**
- **GroupMemBench** (v1 14 May 2026; gpt-4o-mini ingestion, GPT-5 reader and judge, text-embedding-3-large) **[Third-party]** reports that "a simple BM25 baseline matches or exceeds most agent memory systems" — [abstract](https://arxiv.org/abs/2605.14498). Average accuracy: Hindsight 46.01, BM25 43.22, HippoRAG 39.72, embeddings 38.04, A-Mem 35.10, MemGPT 28.15, Mem0 25.73, GraphRAG 20.56. The authors attribute this to ingestion erasing thread structure, speaker context and wording — [full paper](https://arxiv.org/html/2605.14498)
- **MemoryAgentBench** (v4 28 Jun 2026; several GPT, Claude and Gemini models; venue not confirmed on the fetched page) **[Third-party]**: overall BM25 41.5 vs HippoRAG-v2 41.6, MemoRAG 30.9, RAPTOR 27.0, GraphRAG 23.4; GPT-4o-mini with long context 42.2 and GPT-5-mini 60.6. On accurate retrieval, BM25 60.5 beats GraphRAG 40.9, Zep 37.5, RAPTOR 36.8 and Mem0 32.6, while HippoRAG-v2 reaches 65.1 — [Hu et al.](https://arxiv.org/html/2507.05257)
- **Filesystem-Based Memory** (29 Jul 2026; gpt-5.4 family, gpt-4.1 family for skill execution) **[Third-party]**: on 158 LoCoMo questions, closed-book 18.4, BM25 chunks 81.6, verbatim session files 84.2, foldered sessions 86.1, agent-curated 86.1, condensed reorganization 79.1. On PersonaMem 32k, verbatim 78.1, BM25 chunks 71.9, curated 37.5. On REALTALK, verbatim 77.6 vs condensed 41.2. Organization reduced search cost but did not improve answers for any agent measured. Repeated judging moved LoCoMo by about ±1.3 points, and vector or graph stores were not tested — [Zhou et al.](https://arxiv.org/html/2607.26637v1)
- **Mem0's own paper** has full context at 72.90 above its best memory result, Mem0g at 68.44, on LoCoMo **[Vendor]** — [Mem0](https://arxiv.org/html/2504.19413)
- **Letta** (12 Aug 2025; GPT-4o mini) **[Vendor]** reported 74.0 on LoCoMo with a file-based agent vs Mem0's reported 68.5. This is not a pure grep result: files were automatically embedded, and the agent had to call `search_files` first, with `grep` also available. Letta also could not reproduce how Mem0 loaded LoCoMo into MemGPT — [Letta](https://www.letta.com/blog/benchmarking-ai-agent-memory/)
- **Beyond Memory Leaderboards** (18 Jul 2026) **[Preprint; the authors include their own system, Theoria]**: on PAIM, Graphiti's lead came with 2.6M retrieved characters per query and disappeared after controlling for budget. On PTr, adding BM25 to dense retrieval was the largest intervention, and hybrid Simple RAG, Mem0 and Theoria tied within 0.03 points — [abstract](https://arxiv.org/abs/2607.16848)

**Evidence that added structure can help**
- EdgeMem added 11.27 strict-judge points over its flat BM25 configuration on LoCoMo, entirely with non-LLM structure (49.74 to 61.01) — [EdgeMem](https://arxiv.org/html/2609.05553v1)
- T-Mem beat BM25 with expansion by 13.83 points on GroupMemBench under GPT-5.5 (63.09 vs 49.26), at about 5.6× the tokens — [T-Mem](https://arxiv.org/html/2610.12874)
- On multi-hop and temporal subsets, graph methods can beat plain RAG, as in the MultiHop-RAG temporal result of 50.60 vs 30.70 — [Han et al.](https://arxiv.org/html/2502.11371v3)
- **AgentMemBench** (page dated 16 Jun 2026, ID 2608.00009; Qwen2.5-7B 4-bit as generator and judge; author-built adapters) **[Third-party, low power]**: dense key-value retrieval had macro R@5 0.792 vs summarization 0.556, in-context windowing 0.468 and spaCy entity graph 0.361. On LoCoMo, R@5 was 0.573 for dense, 0.005 for summaries and 0.001 for the entity graph. BM25, grep, full context, Mem0 and Zep were not tested — [AgentMemBench](https://arxiv.org/html/2608.00009v1)

**Evaluation reliability**
- **Auditing Long-Term Memory Evaluation** (v2 1 Oct 2026; Claude Opus via Claude Code, GPT-6.1 Sol and others; gpt-4o-2024-08-06 judge) **[Preprint; single author]**: on LongMemEval-S, re-judging unchanged answers flipped three labels. Headline Opus runs scored 479 and 475 of 500. Reader lanes ranged from 93 to 479, although lanes also differed in tools and context. A full evidence package beat baseline sessions 474 to 454, but this is post hoc, developed on all 500 questions, and not separated by component. The author says the results do not establish a transferable memory advantage — [Audit](https://arxiv.org/html/2609.38021v2)
- EdgeMem scored 61.01 under a strict gpt-4o judge and 77.86 under a lenient gpt-4o-mini judge on the same LoCoMo setup, a 16.85-point gap — [EdgeMem](https://arxiv.org/html/2609.05553v1)
- The same system varies sharply across papers. Mem0 on LoCoMo is 66.88 in its own paper, 57.79 in TiMem and 59.22 in MemForest; these use different readers and judges, and MemForest restricts to categories 1–4 — [Mem0](https://arxiv.org/html/2504.19413); [TiMem](https://arxiv.org/html/2601.02845v1); [MemForest](https://arxiv.org/html/2605.23986)
- BM25 moved from 43.22 in GroupMemBench to 49.26 for T-Mem's expanded BM25 on the same benchmark, with different reader versions and a changed retrieval method. Hindsight's margin over BM25 stayed similar (2.79 vs 2.42 points) — [GroupMemBench](https://arxiv.org/html/2605.14498); [T-Mem](https://arxiv.org/html/2610.12874)

### Inferences
- The most consistent pattern is **information preservation first, structure second**. Systems that discard wording, speaker or thread context tend to lose to raw retrieval, while structure layered over verbatim text can add value.
- Any new structure in the engine should have to beat, on the same reader, judge and context budget: (a) recency only, (b) BM25 over verbatim turns, (c) BM25 plus dense retrieval, and (d) full context where it fits.
- Evaluation should report retrieved evidence recall and tokens used alongside answer accuracy. Budget control changed conclusions in Beyond Memory Leaderboards.
- The Letta result shows that an agent with search tools did well, not that grep beats embeddings. It is only 1.1 points above Mem0's reported full-context number, and the two runs are not matched.
- A deterministic engine has an advantage here: fixed retrieval and packing let answer changes be traced to the context diff, reducing one major source of evaluation noise. The reader and judge still remain stochastic or variable.

### Gaps
- "Retrieval Beats Cheap Structured Memory" (preprints.org, August 2026) surfaced in search but returned HTTP 403, so its claims are not used.
- A matched comparison of pure BM25, grep agent, dense, hybrid and graph retrieval on LongMemEval-S with one reader and judge was not found.
- Coding-agent evidence is thin. Most of these benchmarks are conversational, and results may not transfer to long tool-heavy sessions.

## Ingest cost, latency and model-free construction

### Takeaway
Ingest cost differs by orders of magnitude, and useful structure can be built with no LLM: session and time links, entity and keyword hubs, BM25, and relation-free entity graphs all have measured results. The cost of LLM-derived memory is often deferred or hidden, because many papers report throughput or relative tokens instead of calls per stored item.

### Cited Findings

| System | Ingest model use | Reported cost or speed | Source |
|---|---|---|---|
| EdgeMem | None (spaCy, BM25, date parsing); LLM only as reader | LoCoMo: 0 build tokens, 1.8 min build; retrieval P95 10.73 ms. Baselines 0.46M to 11.77M build tokens and 17 min to 10.1 h | [EdgeMem](https://arxiv.org/html/2609.05553v1) |
| LinearRAG | spaCy NER, embeddings, no LLM relation extraction | 2Wiki: 249.78 s and 0 LLM tokens vs HippoRAG2 1,147.01 s with 4.98M prompt tokens; LightRAG 4,933.22 s with 35.52M prompt tokens | [LinearRAG](https://arxiv.org/html/2510.10114v4) |
| T-Mem | None at ingest; LLM per query | Whole evaluation: 31.58M tokens, $44.04 vs BM25 5.66M, $38.68; HippoRAG2 135.04M, $65.50; Hindsight 487.35M, $122.59 (estimates) | [T-Mem](https://arxiv.org/html/2610.12874) |
| Verbatim session files | None | No model cost to store, but agent search on PersonaMem cost about 3.9 to 4.0 cents per query vs 1.4 to 1.6 for reorganized or curated stores | [Zhou et al.](https://arxiv.org/html/2607.26637v1) |
| MemTree (2024) | LLM rewrite along insertion path | 3.27 ± 2.38 LLM calls per insertion vs 1 for MemoryStream | [MemTree](https://arxiv.org/html/2410.14052) |
| HippoRAG 2 | LLM NER and OpenIE | MuSiQue, 11,656 passages: 9.2M input and 3.0M output tokens; about 1.1 s per passage on 4× H100; under $22 by GPT-4o-mini batch. GraphRAG 115.5M input; LightRAG 68.5M; RAPTOR 1.7M | [HippoRAG 2](https://arxiv.org/html/2502.14802v1) |
| HyperGraphRAG | GPT-4o-mini extraction | 3.084 s and $0.0063 per 1k tokens; GraphRAG 9.272 s and $0.0058; HippoRAG2 2.758 s and $0.0056 | [HyperGraphRAG](https://proceedings.neurips.cc/paper_files/paper/2025/file/df55ee6e59f8ac4a625219e11fe9ddba-Paper-Conference.pdf) |
| HyperMem | LLM for boundaries, topics and facts | Relative tokens with Mem0 = 1: HyperMem 7.5×, fact-only 2.5×, HyperGraphRAG 26.3×, GraphRAG 35.3× | [HyperMem](https://arxiv.org/html/2604.08256) |
| MemForest | LLM extraction and summaries, parallelized | Qwen3-30B turns/s on LongMemEval-S: MemForest 2.841, Mem0 1.397, EverMemOS 0.471, MemoryOS 0.202, Zep Local 0.114 | [MemForest](https://arxiv.org/html/2605.23986) |
| Zep | Multi-stage LLM entity, fact, date and community work | No per-message call count in the material read; Mem0's paper claims Zep stores over 600k tokens per conversation vs 26k raw **[Vendor rival; Disputed setup]** | [Zep](https://arxiv.org/html/2501.13956); [Mem0](https://arxiv.org/html/2504.19413) |

- LinearRAG **[Preprint on the fetched page; search results list ICLR 2026]** retrieves through entity activation and Personalized PageRank. GPT-judged accuracy: HotpotQA 66.50 vs HippoRAG2 64.30 and naive RAG 58.60; 2Wiki 63.70 vs HippoRAG 59.90 and naive 43.00; MuSiQue 37.00 vs HippoRAG2 35.00 and naive 29.60 — [LinearRAG](https://arxiv.org/html/2510.10114v4); [ICLR listing](https://iclr.cc/virtual/2026/poster/10007603)
- EdgeMem's LoCoMo answer cost is about 1,000 tokens and 1.3 s per question, against baseline ranges of 1,061 to 41,348 tokens and 2.1 to 45.8 s. On LongMemEval-S its strict-judge accuracy is about level with MemGAS (60.00 vs 60.20) using 3,202 vs 8,829 reader tokens, but the baseline numbers are transcribed from another paper. The reader abstained on 28.0% of answerable LoCoMo questions — [EdgeMem](https://arxiv.org/html/2609.05553v1)
- Non-LLM construction is not automatically good: AgentMemBench's spaCy entity graph was its weakest retriever, and the author attributes this largely to noisy NER — [AgentMemBench](https://arxiv.org/html/2608.00009v1)
- TiMem reports P50/P95 answer latency of 2.35/4.91 s on LoCoMo, close to Mem0's 2.44/4.29 s, with LLM calls dominating — [TiMem](https://arxiv.org/html/2601.02845v1)
- Mem0's retrieval-side latency advantage over full context is large (p95 total 1.440 s vs 17.117 s), but this was measured with GPT-4o-mini on LoCoMo and is vendor-run — [Mem0](https://arxiv.org/html/2504.19413)

### Inferences
Suggested build order for an immutable-turn, deterministic engine, based on the evidence above:

1. **Verbatim append-only turn store with content hashes** and stable turn IDs. Supported by LCM's source-of-truth design and the storage results in Hidden Footprint.
2. **Deterministic metadata indexes:** session, sequence, timestamp, speaker, tool, file path and reply or thread relation. EdgeMem and GroupMemBench both indicate this context matters.
3. **BM25 or inverted index over verbatim turns,** with pinned tokenizer and index version. This is the required baseline and the cheapest strong retriever.
4. **Model-free hubs:** entity, keyword and identifier hubs, with rare-key limits as in EdgeMem. Code-oriented identifiers may suit this engine better than general NER, but that needs testing.
5. **Optional dense retrieval as a versioned derived index,** with embedding model and exact-search or tie-break rules pinned. AgentMemBench and Beyond Memory Leaderboards argue against dismissing it.
6. **Optional stored LLM artifacts:** summaries and extracted facts with provenance, validity records and lexical evidence checks. These should be generated off the read path and frozen before use.
7. **Deterministic packer:** fixed quotas, pinned minimums, utility-per-token ordering, stable tie-breaks, and cache-prefix awareness, following ClawVM's pattern.

Build determinism requires more than deterministic code: every derived index or artifact version must be an input to the context build. Approximate nearest-neighbor search, floating-point ties and changing NLP models are specific risks.

### Gaps
- Calls per stored item are rarely reported. Mem0, Zep and HyperMem do not give a directly usable figure in the material read.
- Dollar comparisons depend on model prices and on which stages each paper counts. T-Mem's figures are estimates.
- No study was found that measures byte-for-byte reproducibility of retrieved context across memory systems.
- Numbers were extracted from arXiv HTML, publisher pages and PDF text by automated page reading. I did not hand-check every table against rendered originals, so critical numbers should be rechecked before publication.
- Mem0's current benchmark page returned 404, so its newer vendor claims are excluded.
