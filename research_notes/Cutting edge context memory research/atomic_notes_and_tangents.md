# Atomic and linked notes as AI memory, and handling tangents and branches in conversation

Research date: 2026-11-18. Priority window: 2024 through November 2026; older work is included only as a baseline and is dated.

Evidence labels used on every finding:
- **[PR]** published at a peer-reviewed venue
- **[PP]** preprint, not confirmed as peer reviewed
- **[V]** vendor or interested party reporting on its own category or product
- **[A]** anecdotal: blog, gist comment, or community report
- **[S]** seen only through a search-result summary; the primary text was not read, so verify before quoting

Unless marked [S], numbers were read from the primary page or PDF. Scores on LoCoMo and LongMemEval are not comparable across papers, because judges, models, question subsets and category handling differ.

## Zettelkasten and Obsidian structures applied to LLMs: what was measured versus anecdotal?

### Takeaway
Measured support exists for LLM-compiled, linked notes on cross-document synthesis and multi-hop questions, but it comes with large build or query costs and usually confounds links with other changes. No study isolates backlinks, block references or evergreen titles in a real vault, and most Obsidian, MCP and "LLM wiki" claims are anecdotal or vendor-reported.

### Cited Findings

**A-MEM and successors**
- [PR] A-MEM (arXiv Feb 2025, v11; [presented at NeurIPS 2025](https://nips.cc/media/neurips-2025/Slides/119020.pdf)) builds Zettelkasten-style notes with link generation and memory evolution. GPT-4o-mini F1 on LoCoMo: multi-hop 27.02, temporal 45.85, open-domain 12.14, single-hop 44.65, adversarial 50.03. Raw-LoCoMo baseline: 25.02, 18.41, 12.04, 40.36, 69.23. Context was 2,520 tokens versus 16,910 — [A-MEM paper, Table 1](https://arxiv.org/html/2502.12110).
- [PR] With GPT-4o, A-MEM's own table shows it **losing** to the plain baseline on single-hop (48.43 vs 61.56) and adversarial (36.35 vs 52.61), while winning multi-hop (32.86 vs 28.00), temporal (39.41 vs 9.09) and open-domain (17.10 vs 16.47) — [A-MEM, Table 1](https://arxiv.org/html/2502.12110).
- [PR] A-MEM ablation, GPT-4o-mini F1, in the order no links or evolution → links only → full system: multi-hop 9.65 → 21.35 → 27.02; temporal 24.55 → 31.24 → 45.85; single-hop 13.28 → 39.17 → 44.65 — [A-MEM, Table 3](https://arxiv.org/html/2502.12110).
- [PP] A-MEM's score moves sharply when others rerun it: F1 average 33.3 in [Synapse's table](https://arxiv.org/html/2601.02744) (Jan 2026), but F1 14.72 with a 40.81 judge score in [Nano-Memory's LoCoMo table](https://arxiv.org/html/2604.11628) (Apr 2026). The two papers use different setups.
- [PR] A cross-scenario evaluation at AACL 2026 ran twelve baselines across five scenario families. An unindexed grep-and-read agent (DCI-Lite) was "among the strongest" baselines, with mean rank 5.20 versus 9.15 for Mem0, 9.20 for the note-based SimpleMem, 5.35 for HippoRAG 2 and 5.35 for long context. The proposed AutoMEM hybrid ranked 2.75 — [Chen et al.](https://aclanthology.org/2026.aacl-long.211.pdf).

**LLM wiki pattern (Karpathy, 2026: the note exists)**
- [A] Karpathy's "LLM Wiki" gist was created 2026-04-04. It defines immutable raw sources, an LLM-maintained linked Markdown wiki, and a schema file, with ingest, query and lint operations plus `index.md` and `log.md`. It claims the index approach works at about 100 sources and hundreds of pages. **It contains no measurements** — [gist](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f).
- [A][S] Reach was reported at about 16 million views as of 2026-04-14 — [Implicator](https://implicator.ai/karpathys-llm-wiki-got-16-million-views-the-code-isnt-the-point-2/).
- [A] Gist comments (self-reported, unverified): @hallaai scored wiki-only 0.64, chunk RAG 0.72 and whole-document context 0.88 on 77 questions over financial filings and drawings, saying compilation dropped specific facts. @hyunsoo36 reported agents still had to read raw sources. @HagaiHen named near-duplicate pages as the default failure. @sassarini-marco reported about 72% of ingest tokens went to routing decisions — [gist comments](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f).
- [PP] The best controlled comparison I found is Cochran, preregistered (arXiv May 2026, v3 Oct 2026): 24 papers, 13 questions, Claude Opus 4.7 generating, LLM judges. On synthesis questions (n=4) the wiki beat single-round RAG on inter-paper mapping by +6.25 on a 10-point scale. RAG led groundedness on bias-check questions (n=3) by +2.00, though that sign reversed under the second judge — [arXiv 2605.18490](https://arxiv.org/html/2605.18490).
- [PP] Cochran's costs: wiki 1,651,357 tokens per query versus 78,093 for RAG (about 21×), 22.0 versus 3.3 minutes, and 904.9M versus 1.40M ingest tokens by the registered metric (about 136M–162M for the wiki after a post-hoc caching adjustment). A decomposition-RAG baseline recovered most of the wiki's advantage at about 491k tokens per query. No citation-support advantage for the wiki was established, and an initial pro-wiki claim-level result was withdrawn after a scoring bug — [arXiv 2605.18490](https://arxiv.org/html/2605.18490).
- [PP] "Retrieval as Reasoning" (May 2026) compiles passages into bidirectionally linked pages read through tools. F1: HotpotQA 0.839, MuSiQue 0.739, 2Wiki 0.911, versus LightRAG 0.819, 0.659, 0.847. It lost to HippoRAG 2 on single-document questions (76.0 vs 78.3) — [arXiv 2605.25480](https://arxiv.org/html/2605.25480).
- [V] LangChain OpenWiki (2026-10-28), synthetic insurance corpus of 111 documents. Simple lookups: search 0.84, wiki 0.83, with latency rising from 9 to 12 s. Connected questions: search 0.10, wiki 0.52. The baseline was keyword search only — [LangChain](https://www.langchain.com/blog/building-agent-memory-openwiki-graphs).
- [V][S] Agent Wikis reports 89% for a compiled wiki versus 63% for RAG on its own benchmark — [summary page](https://jc-aiwiki.pages.dev/karpathy-pattern/agentwikis-platform).

**Graphiti, Basic Memory, Khoj and vault tools**
- [PP][V] Zep/Graphiti (Jan 2025) stores a bi-temporal knowledge graph and retrieves with cosine, BM25 and breadth-first search. On LongMemEval-S with GPT-4o: 71.2% versus 60.2% full context, at 1.6k versus 115k context tokens. It has **no ablation** isolating graph traversal, and single-session-assistant questions fell 17.7% with GPT-4o — [Zep paper](https://arxiv.org/html/2501.13956).
- [V] Basic Memory stores notes as Markdown with observations and typed `[[wikilink]]` relations, indexed by SQLite or Postgres with hybrid full-text and vector search. Its README reports **no retrieval-quality benchmark** (4.6k stars when read, Nov 2026) — [repo](https://github.com/basicmachines-co/basic-memory).
- [V][S] Khoj self-reports raising gemini-1.5-flash from 26.3% to 63.5% on 200 FRAMES questions. This is a general research-agent test, not an evaluation of personal notes or Obsidian links — [Khoj blog](https://blog.khoj.dev/posts/evaluate-khoj-quality/).
- [A] A 99-note synthetic Obsidian vault test on a 0–3 scale (blog dated May 20, year not shown): pure wikilink traversal beat vector RAG on grounding (2.533 vs 2.133) but **lowered faithfulness** (2.000 vs 2.067), which the author attributes to query-agnostic links pulling in other domains. A similarity threshold plus surfacing of unlinked notes gave faithfulness 2.333, and the author credits the unlinked-note step more than the link filter — [DEV post](https://dev.to/nickyeolk/think-with-your-second-brain-a-proper-claude-code-harness-for-obsidian-2c0o).
- [A][S] A claim of 41.5% versus 94.6% on multi-hop questions in a 3,400-note vault has no published method and should not be relied on — [DailyAIWorld](https://dailyaiworld.com/mcp-directory/build-obsidian-vault-mcp-server-markdown-graph-traversal-backlinks).

**Titles (nearest evidence to "evergreen note titles")**
- [PP] Prefixing chunks with a document and header chain, on one private Markdown knowledge base (196,411 chunks, 1,600 queries, Aug 2026): MRR@5 rose from 0.374 to 0.463 overall, and 0.828 to 0.925 on answerable queries. Top-1 changed for 86.9% of queries. Caution: LLM annotator agreement fell from kappa 0.45 to 0.04 when prefixes were removed, so the authors warn the absolute scores can be artifacts — [arXiv 2608.00824](https://arxiv.org/html/2608.00824).

### Inferences
- The LLM wiki behaves like a precomputed synthesis cache: useful for repeated cross-source questions, risky for exact values, and not a replacement for raw sources.
- A-MEM's ablation shows its components matter inside A-MEM, but a threefold single-hop gain from adding links looks more like a weak no-link condition than a pure effect of linking. A strong hybrid-retrieval control would be more informative.
- Keeping source pointers alongside atomic notes is the safest reading of the evidence, because the failures reported are mostly dropped specifics and over-broad linking.

### Gaps
- I found no controlled measurement of backlinks alone, block references, or evergreen or declarative titles in an LLM memory system. The title-chain paper is the closest and tests hierarchy prefixes, not note-title style.
- I found no independent benchmark of Basic Memory or of Obsidian MCP servers on retrieval or answer quality.
- I found no long-duration test of an LLM wiki in real use, and no replication of Cochran's result on a second corpus or model.

## Are explicit links useful beyond search, and what happens when a model maintains them?

### Takeaway
Links help when answers must be assembled across units that do not share words, and they often add nothing or hurt on single-fact lookup and procedural logs. Time-adjacency and agent-controlled expansion have better support than generic semantic links, and model-maintained memory is currently weak at noticing that old notes have gone stale.

### Cited Findings

**Graph retrieval and personalized PageRank**
- [PP] HippoRAG 2 (Feb 2025), QA F1 average: 59.8 versus 57.0 for the NV-Embed-v2 dense baseline, 53.1 for HippoRAG 1, 49.6 for GraphRAG, 48.8 for RAPTOR and 6.6 for LightRAG. The biggest gain was 2Wiki (71.0 vs 61.5); others were small, such as HotpotQA 75.5 vs 75.3 and NarrativeQA 25.9 vs 25.7 — [HippoRAG 2, Table 2](https://arxiv.org/html/2502.14802).
- [PP] HippoRAG 2 recall@5 averaged 78.2 versus 73.4 for dense retrieval; 2Wiki went from 76.5 to 90.4 and HotpotQA from 94.5 to 96.3. The authors state that previous structure-augmented methods underperform the strongest embedding RAG — [HippoRAG 2](https://arxiv.org/html/2502.14802).
- [PP] HippoRAG 2 ablations (recall@5 average over three multi-hop sets, full system 87.1): removing passage nodes 81.0; removing the triple filter 86.4; replacing query-to-triple linking with named-entity-to-node 74.6 — [HippoRAG 2, Table 4](https://arxiv.org/html/2502.14802). What I read does not include a PageRank-off condition.
- [PR] At AACL 2026, indexed memory's value tracked the "structural rate", the share of questions no single retrieved unit can answer. On LoCoMo (60.7%) graph methods averaged +13.0 points over the grep agent. On HotpotQA (17.7%) they averaged −15.4, and on AMABench-ALF (10.3%) −14.2 — [Chen et al., Table 6](https://aclanthology.org/2026.aacl-long.211.pdf).
- [PR] In the same paper, adding a graph-based index view to the grep agent gave +22.5 points on LoCoMo (45.0 → 67.5) and +40.0 on LongMemEval-S (17.0 → 57.0), with negligible change on procedural logs. The ablation bundles the index with agent-controlled selection, as the authors note — [Chen et al.](https://aclanthology.org/2026.aacl-long.211.pdf).
- [PR] Build-time structure can destroy evidence: HippoRAG 2's stored form could answer about 10% of sampled AMABench questions in an oracle probe (the table's final column, which I read as overall), and its build cost needed about 32 shared questions to break even against the grep agent, versus about 2 for LightMem — [Chen et al., Table 4 and Figure 3](https://aclanthology.org/2026.aacl-long.211.pdf).

**Spreading activation**
- [PR] Synapse (arXiv Jan 2026; [Findings of ACL 2026](https://aclanthology.org/2026.findings-acl.1108/)), LoCoMo average F1: full 40.5, no graph structure 32.9, no activation dynamics 30.5, vectors only 25.2 — [Synapse, Table 3](https://arxiv.org/html/2601.02744).
- [PR] The Synapse loss is concentrated. Removing graph structure cut temporal F1 from 50.1 to 25.4, but multi-hop barely moved (35.7 → 35.2) and single-hop rose (48.9 → 49.9). Its edges include consecutive-episode temporal links, abstraction links, and semantic links above 0.92 similarity — [Synapse](https://arxiv.org/html/2601.02744).

**Agent traversal of notes and generated links**
- [PP] In the LLM-Wiki paper, removing wiki structure (flat chunks with no indices, pages or wikilinks) cost 6.1–7.0 F1 points, and removing progressive traversal cost 11.7–13.8. There is **no links-only ablation** — [arXiv 2605.25480](https://arxiv.org/html/2605.25480).
- [V] LangChain's deterministic evidence graph, built from existing citations with no extra model calls, reported +27% accuracy, −58% cost per correct answer and about double the citation consistency over its wiki alone; the question subset for the accuracy figure is not stated — [LangChain](https://www.langchain.com/blog/building-agent-memory-openwiki-graphs).
- [V] Mem0's graph variant barely moved the LoCoMo judge score (68.44 vs 66.88), helped open-domain and temporal questions, hurt single-hop (65.71 vs 67.13) and multi-hop (47.19 vs 51.15), and nearly doubled p95 latency (2.590 s vs 1.440 s). Full context scored 72.90 — [Mem0 paper](https://arxiv.org/html/2504.19413).
- [PP] Nano-Memory (Apr 2026), raw-history retrieval with turn-level scoring and query-time pruning, reported LoCoMo F1 22.66 versus 14.95 for HippoRAG 2 and 14.72 for A-MEM. It did not win every judged metric; MemGAS scored higher on LongMemEval-S (60.20 vs 57.20) — [arXiv 2604.11628](https://arxiv.org/html/2604.11628).

**Staleness and model-maintained notes**
- [PP] STALE (May 2026): 400 implicit-conflict scenarios and 1,200 queries. Best plain model was Gemini-3.1-pro at 55.2%. Memory frameworks on a GPT-4o-mini backbone: LightMem 17.8, Mem0 8.3, LiCoMemory 7.6, Zep 6.0, A-MEM 5.1, versus 8.7 for context-truncated GPT-4o-mini alone — [STALE, Table 2](https://arxiv.org/html/2605.06527).
- [PP] In STALE's write-time diagnostic, recalled entries included the old evidence in 60.5% of cases, but only 3.3% of those old entries were judged as needing an update. A write-time consolidation prototype, CUPMem, reached 68.0%, but relies on a predefined state schema — [STALE](https://arxiv.org/html/2605.06527).
- [A] A gist commenter (@NDOTO-G) reported a health check missing a changed fact, after which models repeated the wrong value; @liushengyu called retiring outdated knowledge the unsolved problem — [gist comments](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f).

### Inferences
- A practical default is search first, then bounded, query-filtered expansion. Treat links as candidates to be scored, not as context to include automatically.
- Synapse suggests some "graph" gains are really chronology. A baseline with adjacent-turn expansion should be included before crediting semantic links.
- Deterministic links from existing structure (citations, reply edges, timestamps) are cheaper and easier to audit than LLM-inferred links, though the supporting evidence here is mainly vendor-reported.
- Staleness looks like a write-time adjudication problem as much as a retrieval problem, since systems often retrieve the old note and still fail to revise it.

### Gaps
- I found no study measuring link rot rates, such as broken or semantically outdated links over time, in LLM-maintained vaults.
- I found no clean experiment removing only links while holding atomic pages, indices and the agent constant.
- STALE tests one conflict pair per instance, so repeated updates and gradual drift remain unmeasured.

## Files as agent memory: Markdown plus grep versus vector or graph memory

### Takeaway
Grep-and-read agents are strong baselines, but no method wins universally: the harness, result format, corpus wording and question structure can matter more than the retriever. The best-supported design gives the agent raw text plus an index it can choose to use.

### Cited Findings
- [PP] "Is Grep All You Need?" (2026-05-14), 116-question LongMemEval-S subset. With tool results returned inline, grep beat vector in **all 10** model-harness pairs, for example Opus 4.6 in the Chronos harness 93.1 vs 83.6 and GPT-5.4 in Codex CLI 93.1 vs 75.9. With file-based ("programmatic") results, vector won 5 of 10 — [arXiv 2605.15184](https://arxiv.org/html/2605.15184).
- [PP] Harness effects were as large as retriever effects: Haiku 4.5 inline grep scored 83.6 in Chronos versus 55.2 in Claude Code, and GPT-5.4 in Codex CLI went from 93.1 inline to 55.2 programmatic. The authors do not claim grep generally beats vector, and limit conclusions to long-memory conversational QA with literal spans — [arXiv 2605.15184](https://arxiv.org/html/2605.15184).
- [PR] Conflicting result: the AACL 2026 paper's grep agent scored **17.0** on LongMemEval-S, versus 57.0 for its hybrid and 48.0 for long context, while scoring 87.5 on HotpotQA — [Chen et al.](https://aclanthology.org/2026.aacl-long.211.pdf). Models, harnesses and question sets differ from the study above, so the two figures should not be averaged.
- [V] Letta (2025-08-12): an agent over one uploaded LoCoMo file scored 74.0% with GPT-4o mini, versus 68.5% reported for Mem0's graph variant. **This was not grep-only**: files were also embedded, and the agent had `search_files` and was required to start with it. Letta itself says the comparison is apples to oranges and that the result reflects agent design — [Letta](https://www.letta.com/blog/benchmarking-ai-agent-memory).
- [V] Oracle (dev post dated Apr 16, year not shown): on a larger paraphrase-heavy corpus the filesystem agent scored 29.7% versus 87.1% for a database agent; on a small keyword-friendly corpus quality was similar. It appears to be a single run, the filesystem agent is described as a simplified demo, and Oracle sells the winning approach — [Oracle Developers](https://dev.to/oracledevs/a-practical-guide-to-choosing-the-right-memory-substrate-for-your-ai-agents-33hj).
- [A] Secondary coverage (Leonis, 2026-10-28) repeats the Oracle benchmark and says an unlocked filesystem "silently lost 97% of writes" at 500 simultaneous writes (14 entries survived). The Oracle text I could read describes this only qualitatively, with figures in a chart — [Leonis](https://lsvp.com/stories/file-systems-for-agents).
- [V] Benchmark disputes remain unresolved. Zep's response (2025-05-06, later modified) gives 75.14% ± 0.17 on LoCoMo against 68.44% for Mem0's best, and argues LoCoMo conversations of about 16k–26k tokens fit in context, lack knowledge-update tests, and have data-quality problems — [Zep](https://www.getzep.com/blog/lies-damn-lies-statistics-is-mem0-really-sota-in-agent-memory/). The current Mem0 paper lists Zep at 65.99 — [Mem0](https://arxiv.org/html/2504.19413).
- [V] Mem0's own comparison page reports 91.6 on LoCoMo and 93.4 on LongMemEval for its April 2026 algorithm, at about 7k retrieved tokens, and notes the LongMemEval score lacks a matched full-context baseline. These are not comparable with its 2025 paper — [Mem0](https://mem0.ai/blog/zep-vs-mem0-which-ai-memory-layer-should-you-choose). Secondary sources circulate slightly different figures (92.5 and 94.4) — [Atlan](https://atlan.com/know/zep-vs-mem0/) [S].

### Inferences
- "Grep beats vectors" and "vectors beat grep" are both overgeneralizations. Literal, session-local questions favor grep; paraphrase and distant evidence favor an index.
- Letta's result is better read as evidence for iterative agent retrieval than for Markdown or grep specifically.
- Any internal comparison should fix the harness and the tool-result format, because those alone can move scores by tens of points.

### Gaps
- I found no large independent head-to-head of Markdown plus grep, vector memory and graph memory under one harness on real personal or project notes.
- The "Is Grep" paper reports different Claude Code grep figures across experiments (76.7 inline in Experiment 1, 94.0 at full history in Experiment 2). I did not find an explanation in what I read, so those tables need checking before citation.
- Concurrency and durability have little neutral measurement.

## Tangents: topic shifts, disentanglement, branching, contamination and measured effects

### Takeaway
Irrelevant prior context can measurably hurt answers, and removing it saves large amounts of context, but the quality gain from branching itself is so far small and tested only in synthetic LLM-judged settings. Stronger models detect clean topic pivots well, yet real conversational boundaries, thread reconstruction and context carryover remain unreliable.

### Cited Findings

**Contamination from unrelated context**
- [PP] Hankache et al. (2025-05-29): 198 GPQA questions after up to 64k tokens of prior context. Unrelated real chat cost about 3.5% to 13.1% of accuracy. Cross-domain Q&A with the task stated only at the top cost up to 73.2% (Gemini Flash), 54.6% (GPT-4o-mini) and 32.0% (GPT-4o) — [arXiv 2506.00069](https://arxiv.org/html/2506.00069v1).
- [PP] Repeating the task just before the query recovered much of that loss: Gemini Flash rose from 0.128 to 0.438, and the GPT-4o-mini drop shrank from about 56% to 8%. The authors attribute many failures to format deviation rather than reasoning — [arXiv 2506.00069](https://arxiv.org/html/2506.00069v1).
- [PR] IJCAI 2026, "semantic inertia": influence from a previous turn persists after explicit topic switches, is strongest early in generation, and decays over about 100–200 tokens. It appears mainly as style residue, with weaker and task-dependent effects on factual accuracy. Reset instructions worked consistently only for Gemini-2.0-flash and were counter-productive for Grok-4 — [Wang et al.](https://www.ijcai.org/proceedings/2026/0088.pdf).
- [PR] LongMemEval (ICLR 2025), GPT-4o: 0.870 with only evidence sessions versus 0.606 with full history (0.924 vs 0.640 with Chain-of-Note) — [LongMemEval](https://arxiv.org/html/2410.10813).
- [PR] SeCom (ICLR 2025): full history scored below focused retrieval. On Long-MT-Bench+, full history got GPT4Score 63.85 at 19,287 tokens, versus 88.81 for segment retrieval at 820 tokens — [SeCom, Table 1](https://arxiv.org/html/2502.05589).
- [PR][S] Laban et al. report an average 39% drop from single-turn to multi-turn performance when instructions are revealed gradually; this concerns underspecified multi-turn instructions, not tangents specifically — [arXiv 2505.06120](https://arxiv.org/html/2505.06120v1).

**Topic-switch benchmarks**
- [PP] "Beyond Continuity" (Netflix, 2026-05-10; a search listing shows an [ICLR 2026 workshop page](https://www.iclr.cc/virtual/2026/10014634)): pivots are built by concatenating conversations from TopiOCQA (625 sessions) and MSC (330 sessions). Pivot F1 on TopiOCQA without cues: o3 0.973, Claude 3.7 Sonnet 0.941, Gemini 2.5 Pro 0.871, GPT-4o 0.695, Mistral-7B 0.214 — [arXiv 2605.09268](https://arxiv.org/html/2605.09268).
- [PP] MSC is harder: o3 0.763, Claude 3.7 Sonnet 0.703. Over-carry of stale context at true pivots reached 1.000 for GPT-4o-mini and 0.991 for Mistral-7B. Pivot detection generally worsened later in a session. **Answer quality and tokens were not measured** — [arXiv 2605.09268](https://arxiv.org/html/2605.09268).

**Segmentation and disentanglement**
- [PR] SeCom's topical segments beat the best turn-level baseline on LoCoMo (71.57 vs 65.58, BM25) and Long-MT-Bench+ (88.81 vs 84.91, MPNet). Segmenter quality mattered: on LoCoMo with MPNet, GPT-4 segments scored 69.33 and RoBERTa 61.84. Zero-shot segmentation F1 varied from 0.888 (DialSeg711) to 0.596 (TIAGE) — [SeCom](https://arxiv.org/html/2502.05589).
- [PP] "When F1 Fails" (Dec 2025): on DialSeg711 one system had exact F1 0.434 but window-tolerant F1 0.767 with 2.53 times as many boundaries as gold. Threshold choice moved scores more than method choice, and periodic or random baselines could compete when density matched. Downstream memory quality was **not measured** — [arXiv 2512.17083](https://arxiv.org/html/2512.17083v2).
- [PR] LREC 2026: out-of-the-box GPT-4o performed poorly at conversation disentanglement. Fine-tuned GPT-4o Mini reached 74.4% link F1 but 45.3% F1 for perfectly matched conversations, and fell short of models using complex structural information — [Pal et al.](https://lrec.elra.info/lrec2026-main-229).
- [PR] CASCON 2025 (2025-11-11): fine-tuned LLM reply-link prediction plus graph clustering reached Shen F 0.90 and one-to-one accuracy 0.87 on Discord developer chats outside the training domain — [Aselstyne and El Mezouar](https://conf.researchr.org/details/cascon-2025/technical-papers-track/38/Revisiting-Thread-Disentanglement-in-Technical-Chat-with-Fine-Tuned-LLMs-and-Graph-Ba).

**Branching systems**
- [PP] ContextBranch (2025-12-15): 30 synthetic software scenarios, GPT-4 generating and GPT-4 judging. Branching cut context from 31.0 to 13.0 messages (−58.1%) and raised overall score from 8.46 to 8.67 (+2.5%, p=0.010, d=0.73). Focus rose 4.6% and context awareness 6.8%; accuracy, relevance and specificity showed no significant change. **No human users and no token counts** — [arXiv 2512.13914](https://arxiv.org/html/2512.13914v1).
- [PP] Conversation Tree Architecture (2026-03-22) defines isolated node contexts and upstream, downstream and cross-node transfer, plus ephemeral branches. It has a prototype but **no empirical evaluation** — [arXiv 2603.21278](https://arxiv.org/html/2603.21278v1).
- [PP] CanvasConvo (v2 2026-09-15): five-day field deployment, 24 participants, 79 branches, SUS 78.9, **no comparison condition and no output-quality measure**. Participants were unsure what a branch inherited; one said "I didn't know what the branch knows" — [arXiv 2605.15848](https://arxiv.org/html/2605.15848).
- [PR] Sensecape (UIST 2023), within-subject, n=12: 68.3 vs 22.8 concepts explored (p=0.01) and 12.8 vs 0.7 revisits (p<0.01). It reports no answer-quality or token measure — [paper](https://ar5iv.labs.arxiv.org/html/2305.11483).
- [PR][S] Graphologue (2023) converts responses into interactive diagrams and was assessed with a small user evaluation; it is a comprehension interface rather than a context-management test — [arXiv 2305.11473](https://arxiv.org/abs/2305.11473v2).
- [A] Loom is a multiversal tree-writing interface for exploring alternative model continuations. I found no quality or token evaluation — [repository](https://github.com/socketteer/loom).
- [A][S] ChatGPT "Branch in new chat" was reported as launched on 2025-09-04; OpenAI's release-notes page blocked retrieval, and I found no published effect measurement — [Dataconomy](https://dataconomy.com/2025/09/08/chatgpt-projects-branching-feature).

### Inferences
- The clearest measured benefit of tangent isolation is smaller context. Quality gains depend on how distant and how large the tangent is, and on whether the task is restated near the query.
- ContextBranch cannot separate branching from simply removing distractors, so an automatic relevance filter on a linear log should be tested against it.
- A useful system needs both a detected segment and an explicit inheritance rule. CanvasConvo shows that users cannot reason about branches when inheritance is unclear.
- Boundary detection should be judged by downstream answer quality and token use, because boundary F1 is density-sensitive.

### Gaps
- "Clarify" could not be identified as a specific branching system; searches returned no matching paper or product.
- I found no user study comparing branching with linear chat on both answer quality and tokens.
- I found no benchmark built from naturally occurring tangents that return to the main thread. Current topic-switch datasets mostly concatenate conversations.
- I found no measured evidence about which merge strategy works best when a branch's result returns to its parent.

## Discourse structure as a retrieval signal: reply chains, question-answer pairs, coreference, contextualization and readable excerpts

### Takeaway
Adding context to retrieval units has reasonable support, while the size of the gain depends heavily on the baseline, embedding model and corpus. Reply chains and question-answer pairing are plausible and widely used, but I found little controlled evidence isolating them for retrieval.

### Cited Findings

**Unit size and turn structure**
- [PR] LongMemEval on LongMemEvalM: session units beat round units at recall@5 (0.706 vs 0.582), and adding extracted facts as keys raised both (0.732 and 0.644). Top-10 GPT-4o answer accuracy was highest for round plus facts (0.720 vs 0.700 for session plus facts). The authors report +9.4% recall and +5.4% accuracy on average from fact expansion — [LongMemEval](https://arxiv.org/html/2410.10813).
- [PR] SeCom showed granularity interacts with retriever: on LoCoMo, BM25 scored 65.58 for turns and 63.16 for sessions, while MPNet scored 57.99 and 51.18 — [SeCom](https://arxiv.org/html/2502.05589).
- [PR] SeCom's compression-based denoising added 9.46 GPT4Score points on LoCoMo (59.87 → 69.33) but only 1.30 on Long-MT-Bench+ (87.51 → 88.81) — [SeCom, Table 2](https://arxiv.org/html/2502.05589).
- [PP] Nano-Memory scores a session by its best-matching turn, then prunes to query-relevant fragments, reaching LoCoMo recall@3 69.39 versus 56.60 for HippoRAG 2 at 1,403 versus 2,991 average tokens — [arXiv 2604.11628](https://arxiv.org/html/2604.11628).

**Reply chains and threads**
- [PR][S] A 2026 email-extraction paper uses thread-preserving chunks that never split a thread, taking structure from conversation IDs — [ICNLSP 2026](https://aclanthology.org/2026.icnlsp-1.4.pdf).
- [A][S] A practitioner reports a 5–6% Slack RAG accuracy gain from thread- and time-based chunking — [DEV post](https://dev.to/criscmd/how-i-boosted-slack-rag-accuracy-by-5-6-with-smarter-chunking-1kf9).
- [PR] Reply links can be predicted well in some technical chat (Shen F 0.90) but exact conversation recovery remains hard elsewhere (45.3% F1), so inferred threads need confidence handling — [CASCON 2025](https://conf.researchr.org/details/cascon-2025/technical-papers-track/38/Revisiting-Thread-Disentanglement-in-Technical-Chat-with-Fine-Tuned-LLMs-and-Graph-Ba); [LREC 2026](https://lrec.elra.info/lrec2026-main-229).

**Coreference and standalone units**
- [PP] Coreference resolution by GPT-4o-mini before indexing (July 2025): retrieval gains were small, for example LLM2Vec nDCG 0.822 → 0.828 and bge-large unchanged at 0.789. QA varied by dataset: BoolQ was nearly flat, while SQuAD rose sharply for some models, such as Qwen2.5-7B 0.3980 → 0.7977 and Llama3.1-8B 0.5583 → 0.7827. No significance tests were reported for retrieval — [arXiv 2507.07847](https://arxiv.org/html/2507.07847v1).
- [PP] Dense X Retrieval (Dec 2023, revised Oct 2024): proposition indexing raised average recall@5 over passages from 43.0 to 52.7 with Contriever, but only from 65.2 to 68.0 with GTR. Manual checks found 3.1% of Propositionizer outputs were not standalone, multi-hop questions remained hard, and the index grew from 41M passages to 257M propositions — [arXiv 2312.06648](https://arxiv.org/html/2312.06648).
- [PP] Molecular Facts (2024-06-28) argues fully atomic facts are the wrong target for verification and defines two criteria, decontextuality and minimality; the abstract gives no numeric result — [arXiv 2406.20079](https://arxiv.org/abs/2406.20079).
- [PR] Decontextualization as a formal task dates to Choi et al. (2021), covering coreference, bridging and related edits needed for a sentence to stand alone — [TACL paper](https://preview.aclanthology.org/fix_video/2021.tacl-1.27.pdf).

**Chunk contextualization**
- [V] Anthropic Contextual Retrieval (2024-09-19): prepending 50–100 tokens of generated context cut top-20 retrieval failure from 5.7% to 3.7% (−35%), to 2.9% with contextual BM25 (−49%), and to 1.9% with reranking (−67%). Estimated cost was about $1.02 per million document tokens with prompt caching. Tests covered codebases, fiction, arXiv and science papers — [Anthropic](https://www.anthropic.com/news/contextual-retrieval).
- [PP] An independent smaller test (Apr 2025) on a 20% NFCorpus subset found modest gains: NDCG@5 0.303 baseline, 0.309 late chunking, 0.317 contextual retrieval with fusion and reranking. Generation differences were not significant, and contextualization could need about 20 GB of VRAM with the local setup — [arXiv 2504.19754](https://arxiv.org/html/2504.19754).
- [PP] Title-chain prefixes improved one production knowledge base (MRR@5 0.374 → 0.463), with the annotation caveat noted earlier — [arXiv 2608.00824](https://arxiv.org/html/2608.00824).

### Inferences
- A good excerpt format separates retrieval text from evidence: generated context or titles help find a passage, while the original span, speaker, timestamp and thread pointer keep it checkable.
- For conversations, the natural unit to test is a question-answer pair expanded to its local reply chain, then compared with fixed turns and topic segments under the same token budget.
- Coreference rewriting is most justified where passages depend heavily on earlier sentences and smaller readers are used. It should not overwrite the source text, since rewriting can introduce errors.
- Atomicity has a floor: units must be small enough to match precisely but carry enough context to be understood.

### Gaps
- I found no controlled study isolating question-answer pairing against single-message indexing. LongMemEval "round" units are the closest proxy, on my reading of a round as one user-assistant exchange.
- I found no standard benchmark for discourse-aware chunking in multi-party chat, and no controlled test of reply-chain expansion on answer quality.
- I found no accepted metric for whether an excerpt is readable out of context. Decontextualization and molecular-facts work offer criteria, but not a validated retrieval-memory measure.
- Anthropic suggests telling the model which text is added context may help generation, but I found no measurement of that effect.
