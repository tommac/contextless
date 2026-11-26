# Evaluating Long-Conversation Memory and Context Systems: Benchmarks, Critiques, and a Methods Toolkit (notes as of 2026-11-18)

Evidence labels used throughout:

- **[PEER]** measured, in a paper confirmed at a peer-reviewed venue.
- **[PREPRINT]** measured by the authors, not confirmed as peer reviewed.
- **[VENDOR]** self-reported by a party selling the system, or an evaluator with a commercial stake.
- **[AUDIT]** third-party check, usually not peer reviewed.
- **[DISPUTED]** a named party contests the number.
- **[SNIPPET]** seen only in a search-result summary. The full source was not read.

"Reader" means the model that generates the answer. "Judge" means the model that grades it.

## Benchmarks: what each tests, size, scoring, flaws, full-context baseline, best scores

### Takeaway
No single number is a "state of the art" on LoCoMo or LongMemEval, because the reader, judge, prompt and question subset move scores by more than most claimed gains. The full-context baseline on LoCoMo ranges from 56.92% to 92.62% across setups, and the benchmarks built since 2025 mostly show full context or simple retrieval matching or beating dedicated memory products.

### Cited Findings

#### Benchmark inventory

| Benchmark (date) | Tests | Size | Scoring | Full-context baseline? | Known issues |
|---|---|---|---|---|---|
| **MSC** ([ACL 2022](https://aclanthology.org/2022.acl-long.356/)) | Human-human multi-session open-domain chat | [237k train, 25k validation examples for sessions 1-4](https://parl.ai/projects/msc/) | [Automatic and human evaluation](https://parl.ai/projects/msc/) | Long-context models were studied | A dialogue dataset, not a memory QA test |
| **LoCoMo** ([ACL 2024](https://aclanthology.org/2024.acl-long.747/)) | QA, event summarization, multimodal dialogue | [10 conversations; 1,540 answerable questions in categories 1-4](https://arxiv.org/html/2608.03463); [446 adversarial category 5 questions](https://github.com/dial481/locomo-audit) | F1 originally; LLM judge in most later work | Not native; added by later papers | See audit below |
| **LongMemEval** ([ICLR 2025, submitted 2024-10-14](https://arxiv.org/abs/2410.10813)) | Extraction, multi-session reasoning, knowledge updates, temporal reasoning, abstention | [500 questions, each with its own history; S is about 115k tokens, M about 500 sessions; 30 abstention questions](https://github.com/xiaowu0162/LongMemEval) | [GPT-4o (2024-08-06) judge with per-type prompts, over 97% agreement with experts](https://arxiv.org/html/2410.10813) | **Yes**, plus oracle | [Histories were cleaned in a 2025/09 update](https://github.com/xiaowu0162/LongMemEval), so pin the file version |
| **LifelongAgentBench** ([2025-05, v3 2025-05-30](https://arxiv.org/html/2505.11942)) | Skill transfer across Database, OS, Knowledge Graph tasks | [1,396 tasks (500/500/396)](https://arxiv.org/html/2505.11942) | Task success, checked by execution | [No; baseline is no replay](https://arxiv.org/html/2505.11942) | [Only 10% of tasks manually reviewed; single values without variance](https://arxiv.org/html/2505.11942) |
| **MemoryAgentBench** ([2025-07, v4 2026-06-28](https://arxiv.org/html/2507.05257)) | Accurate retrieval, test-time learning, long-range understanding, selective forgetting | [2,071 questions, about 103K to 1.44M tokens of context](https://arxiv.org/html/2507.05257) | [Accuracy, Recall@5, F1 by task](https://arxiv.org/html/2507.05257) | **Yes**, six long-context models | Mixes reused datasets; venue not confirmed on the paper page |
| **MemoryBench** ([2025-10-20, v7 2026-06-03](https://arxiv.org/abs/2510.17281)) | Continual learning from user feedback | [About 20k cases, 11 datasets, 4:1 train/test](https://arxiv.org/html/2510.17281) | [Native metrics merged through a 1-10 LLM judge, then normalized](https://arxiv.org/html/2510.17281) | Vanilla no-memory baseline | [Feedback is simulated by an LLM](https://arxiv.org/html/2510.17281) |
| **BEAM** ([2025-10-31, v2 2026-02-21](https://arxiv.org/abs/2510.27246)) | Ten abilities at 100K to 10M tokens | [100 conversations, 2,000 questions](https://arxiv.org/html/2510.27246) | [Nuggets scored 0/0.5/1 by an LLM judge; Kendall tau-b for event order](https://arxiv.org/html/2510.27246) | **Partial**: [no tested model holds 10M, so those runs use a recent segment](https://arxiv.org/html/2510.27246) | Generated conversations; judge model not found in the paper text I searched |
| **OOLONG** ([2025-11-04](https://arxiv.org/abs/2511.02817)) | Aggregation over long context, not memory as such | [Synth: 400 test questions per length, 1K to 4M tokens. Real: D&D transcripts, about 55K to 1.3M tokens](https://arxiv.org/html/2511.02817) | [Exact match; numeric partial credit 0.75^error](https://arxiv.org/html/2511.02817) | **Yes**, direct long-context is the main condition | Tests counting more than recall |

#### Newer benchmarks (2026)

- **LongMemEval-V2** (2026-05-12) [PREPRINT]: [451 curated questions over web-agent trajectories, about 25M and 115M token tiers, deterministic scoring for structured answers, GPT-5.2 judge for gotcha and abstention items](https://arxiv.org/html/2605.12493v1). [No-retrieval baseline about 1.3%; best no-history frontier model 14.1%; best system 72.5%; Codex baseline 69.3%; best RAG 48.5%, all with a Qwen3.5-9B reader](https://arxiv.org/html/2605.12493v1).
- **MemoryArena** (2026-02, v2 2026-09-17) [PREPRINT]: [701 interdependent agent tasks; full-context success averages only 0.09 to 0.19, and external memory does not consistently beat full context](https://arxiv.org/html/2602.16313).
- **MemOps** (2026-07-14) [PREPRINT]: [2,006 QA pairs over remember, forget, update, reflect and trajectory operations, GPT-4o judge. Long-context accuracy at about 61K tokens: GPT-4.1-mini 0.798, GPT-4o 0.794, Claude Sonnet 4.5 0.790. Session-level BM25 RAG 0.845, MemOS 0.785, Mem0 0.543](https://arxiv.org/html/2607.12893v1).
- **MemTrace** (2026-06-15) [PREPRINT]: [835 knowledge points probed over age, question type and evidence condition; 200,453 scored answers; GPT-4o judge with Gemini-3-Flash second judge (kappa 0.772 and 0.703)](https://arxiv.org/html/2606.17328). [73.3% of sampled failures were reachable but unsolved, against 7.0% unreachable](https://arxiv.org/html/2606.17328).
- **STALE** (2026-05-07) [PREPRINT]: [400 conflict scenarios, 1,200 queries, contexts up to 150K tokens, top overall accuracy 55.2%](https://arxiv.org/abs/2605.06527). [SNIPPET: [LightMem 17.8%, Mem0 8.3%, Zep 6.0%](https://www.emergentmind.com/papers/2605.06527).]
- **Compact-Bench** (2026-10-28) [PREPRINT]: [paired full-history versus summary-only replay on 35 SWE-bench repairs, judged by repository tests. Stable damage ranged from 5% to 64% of 22 validated cases across 14 summarizers](https://arxiv.org/html/2610.31710v2).
- **FEB-30** (2026-10-15) [PREPRINT, AWS co-authors]: [30 fact-evolution trajectories scored directly against the memory store](https://arxiv.org/html/2610.22209).
- **MemContract** (NeurIPS 2026 poster): [1,080 tasks, six operations including revise, erase and prove-origin; GPT-4o macro average: Graph Memory 67.6% (CI 65.8 to 69.4), best hybrid 62.4%, RAG 40.6%](https://neurips.cc/virtual/2026/loc/sydney/poster/154728).
- **DolphinBench** (NeurIPS 2026 workshop): [600 action tasks verified as solvable only with memory; cost and latency are mandatory](https://neurips.cc/virtual/2026/165653). Its author list includes [Deshraj Yadav](https://neurips.cc/virtual/2026/165653), the same name as [Mem0's co-founder and CTO](https://github.com/getzep/zep-papers/issues/5), and [Mem0 publishes a 70.67% score on it](https://mem0.ai/library/agent-memory/ai-memory-benchmarks-in-2026). Treat it as vendor-linked until confirmed otherwise.

#### LoCoMo audit findings [AUDIT, February to April 2026]

- [99 of 1,540 gold answers (6.4%) are wrong, giving a ceiling of 93.57%](https://github.com/dial481/locomo-audit).
- [The gpt-4o-mini judge accepted 62.81% of vague but topical wrong answers and 10.61% of specific wrong answers](https://github.com/dial481/locomo-audit).
- [About 56% of adjacent per-category comparisons are indistinguishable under Wilson 95% intervals; open-domain (n=96) needs a gap of 15+ points](https://github.com/dial481/locomo-audit).
- Conflict: the audit says [category 5 is not evaluated in any published result](https://github.com/dial481/locomo-audit), but [Mem0 alleged in May 2025 that Zep's 84% included it](https://github.com/getzep/zep-papers/issues/5).
- Context length is reported inconsistently: [about 9K tokens in the original abstract](https://arxiv.org/abs/2402.17753), [about 16K in LeanMem](https://arxiv.org/html/2608.03463), and [26,031 tokens per full-context query in the Mem0 paper](https://arxiv.org/html/2504.19413).

#### What a plain full-context or full-text baseline scores

| Benchmark | Reader | Judge | Score | Source and status |
|---|---|---|---|---|
| LoCoMo cat. 1-4 | GPT-4.1-mini | GPT-4.1-mini | **56.92%** | [LeanMem, 2026-08](https://arxiv.org/html/2608.03463) [PREPRINT] |
| LoCoMo cat. 1-4 | gpt-4o-mini | Unnamed | **72.90% ±0.19** | [Mem0 paper, 2025-04](https://arxiv.org/html/2504.19413) [VENDOR] |
| LoCoMo | GPT-4.1-mini, chain-of-thought answer prompt | Not captured | **92.62%** | [locomo-audit](https://github.com/dial481/locomo-audit) [AUDIT] |
| LongMemEval-S | gpt-4o-mini | GPT-4o | **55.4%** | [Zep paper, 2025-01](https://arxiv.org/html/2501.13956) [VENDOR] |
| LongMemEval-S | GPT-4o | GPT-4o | **60.2%** | [Zep paper](https://arxiv.org/html/2501.13956) [VENDOR] |
| LongMemEval-S | GPT-4o | GPT-4o | **60.6%; 64.0% with Chain-of-Note** | [LongMemEval paper](https://arxiv.org/html/2410.10813) [PEER] |
| LongMemEval-S | GPT-4.1-mini | GPT-4.1-mini | **77.40%** | [LeanMem](https://arxiv.org/html/2608.03463) [PREPRINT] |
| MemoryAgentBench overall | GPT-5-mini | Not applicable across tasks | **60.6** (Mem0 21.1, Zep 24.0, best RAG 41.6, BM25 41.5) | [MemoryAgentBench v4](https://arxiv.org/html/2507.05257) [PREPRINT] |
| DMR | gpt-4-turbo / gpt-4o-mini | Unnamed | **94.4% / 98.0%** (Zep 94.8% / 98.2%) | [Zep paper](https://arxiv.org/html/2501.13956) [VENDOR] |

Oracle (evidence only) on LongMemEval-S: [GPT-4o 87.0%, or 92.4% with Chain-of-Note](https://arxiv.org/html/2410.10813) [PEER]. Mastra separately lists [an 82.4% gpt-4o oracle](https://mastra.ai/research/observational-memory) [VENDOR], a 4.6-point conflict with the benchmark paper's no-Chain-of-Note figure.

#### Best reported scores

| Benchmark | System | Score | Reader / judge | Status |
|---|---|---|---|---|
| LongMemEval-S | Mastra Observational Memory | [94.87%; 84.23% with gpt-4o](https://mastra.ai/research/observational-memory) | gpt-5-mini or gpt-4o / gpt-4o | [VENDOR], 2026-02 |
| LongMemEval | Mem0 | [94.4, 6.8K tokens, 1.09s p50](https://mem0.ai/library/agent-memory/ai-memory-benchmarks-in-2026) | Reader not stated | [VENDOR], [reported at top_200 per AgentOS](https://agentos.sh/blog/memory-benchmark-transparency-audit/) |
| LongMemEval | Maximem Synap | [92.0% (460/500)](https://arxiv.org/html/2607.21503) | gpt-5-mini / gpt-5-mini | [VENDOR], 2026-07 |
| LongMemEval-S | LeanMem | [91.80%, 5 runs](https://arxiv.org/html/2608.03463) | GPT-4.1-mini / GPT-4.1-mini | [PREPRINT], 2026-08 |
| LongMemEval-S | BFR-MV | [82.2%, against Mem0 at 72.4% in the same store and harness](https://arxiv.org/html/2610.09348) | Extractive answerer / gpt-5.5 | [PREPRINT], 2026-10, strongest protocol found |
| LoCoMo | LeanMem | [84.87%](https://arxiv.org/html/2608.03463) | GPT-4.1-mini / GPT-4.1-mini | [PREPRINT] |
| LoCoMo | Maximem Synap | [93.2%, categories 1-4](https://arxiv.org/html/2607.21503) | gpt-5-mini / gpt-5-mini | [VENDOR] |
| LoCoMo | Mem0 | [92.5, 7.0K tokens](https://mem0.ai/library/agent-memory/ai-memory-benchmarks-in-2026) | Not stated | [VENDOR] |
| LoCoMo | Zep, ByteRover, ZeroMemory | [94.7, 96.1, 96.1 as listed by Mem0](https://mem0.ai/library/agent-memory/ai-memory-benchmarks-in-2026) | Not stated | [VENDOR], second-hand via a competitor |
| LoCoMo | EverMemOS | [92.32% claimed; a third party reported 38.38%](https://github.com/dial481/locomo-audit) | Not captured | [DISPUTED] |
| BEAM 1M / 10M | Mem0 | [64.1 (700 questions) / 48.6 (200 questions)](https://mem0.ai/library/agent-memory/ai-memory-benchmarks-in-2026) | Not stated | [VENDOR] |
| BEAM average | LIGHT | [best 0.359 at 500K and 0.266 at 10M, Llama-4-Maverick reader; Vanilla row 0.283 and 0.104](https://arxiv.org/html/2510.27246) | Judge not found | [PREPRINT] |
| OOLONG | Direct long context | [Gemini-3-Pro 78.69 synth, 57.93 real; GPT-5 70.75, 47.00](https://arxiv.org/html/2511.02817) | Deterministic scoring | [PREPRINT] |

Other relevant results:

- [BM25 plus code execution under GPT-5 scored 16 to 18 points below direct GPT-5 on Oolong-real](https://arxiv.org/html/2511.02817).
- [None of A-Mem, Mem0 or MemoryOS consistently beat simple RAG on MemoryBench, and vanilla beat all of them on one task partition](https://arxiv.org/html/2510.17281).
- [FactConsolidation multi-hop stays at 2 to 5 for most RAG and memory agents, with GPT-5-mini at 28.0](https://arxiv.org/html/2507.05257).

### Inferences
- Scores above about 93.6 on LoCoMo (for example 94.7 and 96.1) exceed the audit's ceiling. If the audit is right, those systems are being credited for matching wrong gold answers or through judge leniency, so a higher LoCoMo score above that level is not evidence of better memory.
- The 36-point spread in LoCoMo full-context results (56.92 to 92.62) is larger than almost any claimed system gain. A LoCoMo result without a same-harness full-context row is uninterpretable.
- Reader choice alone moved Mastra's own system by 10.6 points (84.23 to 94.87). Reader and judge must be held fixed across arms.
- If each BEAM conversation has 20 questions, Mem0's 700 and 200 question tiers correspond to about 35 and 10 conversations, so their effective sample is small.

### Gaps
- I found no independent, same-harness replication of the top LongMemEval-S or LoCoMo scores (above about 90%).
- BEAM's judge model was not identifiable in the HTML text I searched. Its Listing 20 and appendix may name it.
- I did not confirm that DMR derives from MSC, or verify a peer-reviewed venue for MemoryAgentBench ([one secondary source says ICLR 2026](https://www.cognee.ai/agent-memory-evaluation-datasets-benchmarks)).
- The locomo-audit 92.62% judge and the third-party 38.38% setup were not checked at their detailed source files. [EverMemOS issue #73](https://github.com/dial481/locomo-audit) was not opened.

## Critiques: contamination, inflated scores, judge reliability, memory-free answerability, recall versus correctness

### Takeaway
The most serious documented problems are test-set tuning, undisclosed configuration, lenient judges and retrieval metrics presented as answer accuracy. Judges need human calibration on the specific task, and closed-book controls for LoCoMo and LongMemEval-S remain a cheap unfilled gap.

### Cited Findings

#### Disputed vendor scores
- **Mem0 versus Zep on LoCoMo [DISPUTED, May 2025]:** [Mem0's paper put Zep at 65.99%, Mem0 at 66.88%, Mem0g at 68.44%, and full context at 72.90%](https://arxiv.org/html/2504.19413). [Mem0's CTO argued Zep's 84% was really 58.44% ±0.20 after removing category 5; Zep conceded a calculation error, reported 75.14% ±0.17 over ten runs, and disputed Mem0's setup (user roles, timestamps, sequential search). The issue closed on 2025-05-19 without agreement](https://github.com/getzep/zep-papers/issues/5).
- **A misattribution to avoid:** [AgentOS describes arXiv:2512.13564 as an independent reproduction measuring Zep at 63.8%](https://agentos.sh/blog/memory-benchmark-transparency-audit/). [That paper is a survey](https://arxiv.org/abs/2512.13564), its HTML text does not contain "63.8" (my text search, 2026-11-18), and [63.8% is Zep's own gpt-4o-mini result](https://arxiv.org/html/2501.13956). The gap is a reader difference, not a failed reproduction.
- **MemPalace [AUDIT, April 2026]:** [its 100% LoCoMo used top_k=50, effectively the whole conversation; 100% LongMemEval needed repeated iterations with reranking; its 96.6% is recall_any@5, with end-to-end QA about 67.2%](https://arxiv.org/pdf/2604.21284). [Maintainers later replaced the headline with 98.4% Recall@5 on 450 held-out questions and flagged the last 0.6% as teaching to the test](https://arxiv.org/pdf/2604.21284). Table 6 in that same paper mixes metric types under one R@5 column, so do not reuse it.
- **Mem0 reproduction:** [a third-party failure to reproduce LoCoMo accuracy on Mem0's platform was closed as completed on 2026-03-27](https://agentos.sh/blog/memory-benchmark-transparency-audit/).
- **Token claims:** [EverMemOS's README claims 2,298 tokens per question while its paper's table shows 6,669 with GPT-4.1-mini, so the reduction from full context is 67%, not 89%](https://github.com/dial481/locomo-audit) [AUDIT].

#### Contamination and test-set adaptation
- [AgentOS discloses that its 85.6% LongMemEval-S router was calibrated on the same 500 questions; fixed gpt-4o gives 84.2%. The same pipeline without tuning scored 49.9% on LoCoMo](https://agentos.sh/blog/memory-benchmark-transparency-audit/) [VENDOR].
- [Mastra says configurations were run repeatedly during development and the latest run reported, without stating whether LongMemEval served as a validation set](https://mastra.ai/research/observational-memory).
- [The BFR paper states that all 10 LoCoMo conversations were visible during development, so LoCoMo only shows stability, and it kept LongMemEval-S unscored until its protocol was frozen](https://arxiv.org/html/2610.09348). This is the model disclosure.
- Selection can bias a benchmark too: [LongMemEval-V2 kept questions only once at least two of four frontier models failed them](https://arxiv.org/html/2605.12493v1), and [UTILMEM's authors caution about a selection effect from their evidence-sensitivity filter](https://arxiv.org/pdf/2608.30508) [SNIPPET].

#### Questions answerable without memory
- [LongMemEval-V2 reports a no-retrieval baseline near 1.3% and a best no-history frontier result of 14.1%](https://arxiv.org/html/2605.12493v1).
- [OOLONG's authors tested whether models could answer the real split from memory alone and concluded contamination is unlikely](https://arxiv.org/html/2511.02817).
- [DolphinBench says each task was verified solvable with memory and unsolvable without it](https://neurips.cc/virtual/2026/165653).

#### Judge reliability
- [Lee et al. (v3 2026-02-09) show raw judge accuracy overstates low true accuracy and understates high true accuracy, and provide a sensitivity and specificity correction with valid intervals. Their example needs about 200 calibration items for an interval under 0.1](https://arxiv.org/html/2511.21140v3).
- Task-specific agreement varies: [LANTERN's GPT-5 Nano judge agreed with the human majority 91% of the time (kappa 0.81; inter-annotator kappa 0.78) on 100 items](https://arxiv.org/html/2606.05182), and [MemTrace reports kappa 0.772 and 0.703 between two LLM judges](https://arxiv.org/html/2606.17328). Note that two agreeing LLMs do not establish validity.
- A stricter rubric matters: [AgentOS reports 0 false positives on 100 LoCoMo probes with gpt-4o-2024-08-06](https://agentos.sh/blog/memory-benchmark-transparency-audit/) [VENDOR].
- Same-model reader and judge is common: [Maximem uses gpt-5-mini for both](https://arxiv.org/html/2607.21503), and [LeanMem uses GPT-4.1-mini as judge alongside a GPT-4.1-mini backbone](https://arxiv.org/html/2608.03463).
- **Do not cite as settled:** ["Reliability without Validity" (21 judges, about 541,000 judgments) was withdrawn by arXiv admin on 2026-11-03 for not meeting content quality standards](https://arxiv.org/abs/2606.19544), even though it still circulates in search results.

#### Recall versus answer correctness
- [LeanMem's FullText baseline has 100.00 recall on LongMemEval-S but 77.40% accuracy](https://arxiv.org/html/2608.03463), so evidence presence does not guarantee a correct answer.
- [MemPalace pairs 96.6% recall_any@5 with about 67.2% QA accuracy](https://arxiv.org/pdf/2604.21284).
- [MemTrace found gold evidence raised reached-but-unsolved cases to 80.4 to 83.9%](https://arxiv.org/html/2606.17328), which shows failures on both sides of retrieval.
- [BFR argues top-k relevance does not ensure the retrieved set jointly suffices, and measures set-level sufficiency](https://arxiv.org/html/2610.09348).
- [FEB-30 found pass/fail agreement between direct store inspection and downstream QA near chance (kappa 0.118; 60.6% against 48.8% accuracy)](https://arxiv.org/html/2610.22209).
- [SNIPPET: [MutMem reports recall 1.000 at k=20 with 91.8% judged correct on LongMemEval](https://arxiv.org/pdf/2608.02843).]

### Inferences
- A credible paper should report three separate layers: evidence reached the context, evidence was used, and the answer was correct. Collapsing them is the root of most disputes above.
- Any large top-k result should be reported with tokens delivered to the reader, because large k approaches full context.
- The AgentOS 0-of-100 result still allows a false-positive rate up to about 3 to 4% at 95% confidence, so it does not show the judge is error-free.

### Gaps
- I found no published closed-book (no-history) score for LoCoMo or LongMemEval-S with named reader and judge. This is cheap to run and useful to report.
- I found no systematic study measuring how much memorized benchmark content inflates scores on these datasets.
- I did not read Bavaresco et al. directly. It appeared only through [secondary descriptions](https://arxiv.org/pdf/2609.13824).

## Evaluation design: recovery probes, causal tests, faithfulness, updates, abstention, distractors

### Takeaway
The strongest designs pair each item across conditions and intervene on the memory itself, using deterministic scoring where possible. Removing or altering a passage tests whether it was actually used, which correctness and citation checks cannot show.

### Cited Findings

#### Recovery probes
- **LANTERN (arXiv stamp 2026-04-18) [PREPRINT]:** [100 ShareGPT conversations reduced to 94, 1,894 LLM-extracted facts, one probe question per fact, and a fact kept only if its answer appears in the source. Compaction occurs at 50% of the conversation, with all methods given the same 6,000-character restore budget](https://arxiv.org/html/2606.05182).
- [Recovery: Lantern-Rerank 78.3%, Lantern 76.3%, MemGPT-Faithful 72.4%, summarization 63.1%, neural RAG 61.3%, GPT-5 Nano judge](https://arxiv.org/html/2606.05182).
- [Live accuracy on 137 questions per model rose by 7.3 to 10.2 points: GPT-4o Mini 33.6 to 41.6, Gemini 2.5 Flash 32.8 to 43.1, Claude Sonnet 4.5 40.1 to 48.2, GPT-5 Nano 38.7 to 46.0](https://arxiv.org/html/2606.05182).
- [Stated limits: LLM-derived ground truth, ShareGPT only, no LoCoMo or LongMemEval, no direct Letta, Mem0 or Zep comparison](https://arxiv.org/html/2606.05182).
- **Compact-Bench design:** [each case is labeled preserved, damaged, rescued or unsolved, so gains cannot hide losses; damage needs failure in at least 2 of 3 runs; 120 of 126 first-round damaged pairs were confirmed](https://arxiv.org/html/2610.31710v2).
- [Deterministic retention by exact substring found the best LLM summarizer at 16.0% against 79.1% for a mechanical extractor](https://arxiv.org/html/2610.31710v2), measured only on literal recurrence.
- [SNIPPET: ["Lost in Context Compaction" reports 31.8% constraint compliance from compacted context alone and 99.7% when the constraint is appended afterward](https://arxiv.org/pdf/2608.11242).]

#### Causal and counterfactual attribution
- **ContextCite (NeurIPS 2024):** [fits a sparse linear surrogate over random source ablations, separates contributive from corroborative attribution, and evaluates with top-k log-probability drop and linear datamodeling score. It matches or beats baselines from 32 ablations](https://arxiv.org/html/2409.00729v1).
- [ContextCite's limits: influence rather than mechanism, ablations may not be neutral, and refusals can be misattributed](https://arxiv.org/html/2409.00729v1).
- **AT2 (2025-04-18):** [learns attention-head weights from ablation signal and reports parity with many-ablation methods at lower cost](https://arxiv.org/abs/2504.13752). It requires model internals, so it does not apply to closed APIs.
- **ARC-JSD (2025) [SNIPPET]:** [Jensen-Shannon divergence attribution without a surrogate model](https://arxiv.org/html/2505.16415v4).

#### Faithfulness and citation verification
- [Wallat et al. (ICTIR 2025) distinguish citation correctness from faithfulness and report up to 57% of citations unfaithful for Command-R+ in adversarial tests](https://arxiv.org/abs/2412.18004). The cause is post-rationalization, where a model answers from prior knowledge and then attaches a fitting source.
- [MemOps scores provenance support, leakage rate and stale value rate alongside accuracy](https://arxiv.org/html/2607.12893v1).

#### Knowledge updates, abstention and distractors
- [MemTrace probes current, historical and change-trajectory questions under present, missing and contradicted evidence. Mem0 abstained on 99.3% of boundary probes but resolved only 14.6% of conflict probes](https://arxiv.org/html/2606.17328), so high abstention can hide poor correction.
- [MemContract uses distractor sessions, alias rotation, blinded, stale and delayed controls, and a 25-probe residue audit. Residue leakage bound: 3.9% for Graph Memory, 27.4% for RAG](https://neurips.cc/virtual/2026/loc/sydney/poster/154728).
- [SNIPPET: [STALE includes adversarial requests built only from the outdated state](https://www.emergentmind.com/papers/2605.06527).]
- [LoCoMo category 5 measures abstention, and including it moves scores by ten points or more](https://arxiv.org/html/2607.21503), so report it separately.
- [OP-Bench (EMNLP 2026) tests the opposite failure, overuse of memory, with 1,700 verified instances](https://aclanthology.org/2026.emnlp-main.1168/).

### Inferences
A practical protocol for a small system, combining the designs above:

1. **Conditions per item:** closed book, full context, oracle evidence, system, and system with its top evidence removed.
2. **Necessity test:** remove the cited or top-ranked memory and re-ask. A correct answer that survives removal did not depend on that memory.
3. **Sufficiency test:** provide only the selected memory and ask again.
4. **Value-swap test:** change the answer-bearing value in the source (a port number, a date) and check that the answer follows. This separates use of memory from prior knowledge and is the cleanest fit for closed APIs, where AT2-style internals are unavailable.
5. **Update and stale tests:** state a fact, supersede it, then score current answer, historical answer and stale-value rate separately.
6. **Distractors:** insert near-duplicate entities and values, and report precision and tokens as well as recall.
7. **Scoring:** prefer exact or normalized matching for identifiers, code and commands, and use a calibrated judge only for free-form items.

### Gaps
- I found no standard, validated value-swap protocol for conversational memory. The steps above are my synthesis of the attribution and benchmark designs, so they would need piloting.
- I did not find a 2026 evaluation applying ContextCite-style ablation directly to a long-conversation memory system.

## Statistics for small paired experiments

### Takeaway
With 30 to 100 items, only large effects are detectable: at 100 items and 20% discordant pairs, the smallest difference with 80% power is about 13 points. Use paired per-item tests, resample at the conversation level, freeze a held-out set, and correct for judge error.

### Cited Findings
- **Paired binary outcomes:** [Fagerland et al. (2013) recommend the McNemar mid-p test over the exact conditional test, which is conservative; mid-p did not exceed the nominal level in any of 9,595 scenarios studied](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC3716987/).
- **Current good practice:** [BFR uses exact two-sided McNemar for judged accuracy, a 10,000-sample question-level paired bootstrap for F1, clustered 95% intervals for ablations, Holm correction for type-level tests, and conversation-level five-fold analysis](https://arxiv.org/html/2610.09348). Example outcome: [BFR-MV against MRAgent-Flat* on LoCoMo (n=200) was +6.0 points, p=0.043](https://arxiv.org/html/2610.09348).
- [LANTERN uses paired two-sided Wilcoxon tests on per-conversation rates, a 10,000-resample paired bootstrap with fixed seed, and Cohen's d. Lantern against MemGPT-Faithful: p=0.005, d=0.26](https://arxiv.org/html/2606.05182).
- [Compact-Bench uses Wilson intervals for damage: about 1 to 22% for Opus 4.8 and 43 to 80% for DeepSeek-Flash on n=22](https://arxiv.org/html/2610.31710v2), which shows how wide small-sample intervals are.
- **Independence:** [Total Recall at What Cost states its Wilson intervals assume independent questions](https://arxiv.org/html/2608.11879), an assumption that questions sharing a conversation violate.
- **Power:** [Card et al. (EMNLP 2020) found underpowered experiments common in NLP and recommend power analysis before evaluation](https://web.stanford.edu/~jurafsky/pubs/2020.emnlp-main.745.pdf).
- **Test choice:** [Dror et al. (ACL 2018) found significance testing often ignored or misused in ACL and TACL papers, and give a protocol for choosing tests](https://aclanthology.org/P18-1128/).
- **Pre-registration:** [van Miltenburg et al. (NAACL 2021) propose preregistration questions by study type](https://arxiv.org/abs/2103.06944v2). [Hofman et al. (2023) propose a staged template completed before test data is accessed](https://arxiv.org/abs/2311.18807v1). [Repeated holdout checks let analysts adapt to it](https://research.google/blog/the-reusable-holdout-preserving-validity-in-adaptive-data-analysis/).
- **Judge-corrected inference:** [Rogan-Gladen-style correction with calibration-aware intervals (Lee et al.)](https://arxiv.org/html/2511.21140v3). [SNIPPET: [prediction-powered approaches combine a small human-labeled set with many judge labels](https://glide-py.readthedocs.io/en/latest/tutorials/ppi/).]

### Inferences

**Own calculation, not from a source.** Exact conditional McNemar, two-sided, 80% power, computed by summing over the binomial distribution of discordant pairs. "Discordant" is the share of items where the two systems differ in correctness. Cells show the smallest true accuracy difference detectable; a dash means 80% power is unreachable at that discordance. The mid-p test would be slightly more powerful.

| Items | Discordant 10%, alpha .05 | 20%, alpha .05 | 30%, alpha .05 | 10%, alpha .0125 | 20%, alpha .0125 | 30%, alpha .0125 |
|---|---|---|---|---|---|---|
| 30 | - | - | 27.5 pts | - | - | - |
| 50 | - | 18.0 | 22.0 | - | - | 25.5 |
| 100 | 9.0 | 13.0 | 16.0 | - | 15.0 | 18.5 |
| 200 | 6.5 | 9.5 | 11.5 | 7.5 | 11.0 | 13.5 |
| 300 | 5.5 | 7.5 | 9.5 | 6.5 | 9.0 | 11.0 |
| 500 | 4.5 | 6.0 | 7.5 | 5.0 | 7.0 | 8.5 |

Alpha .0125 is Bonferroni for four planned comparisons; Holm is at least as powerful.

**Arithmetic floor.** With 5 discordant pairs all favoring one system, the exact two-sided p-value is 2 × 0.5^5 = 0.0625, so fewer than 6 discordant pairs can never reach p < 0.05 with the exact test.

**Own calculation.** Wilson 95% half-width for a single accuracy near 70%: n=30 ±15.6 points, 50 ±12.3, 100 ±8.8, 200 ±6.3, 300 ±5.2, 500 ±4.0, 1,540 ±2.3.

**Own calculation, illustrative intraclass correlations only.** Design effect = 1 + (m - 1) × ICC, for m items per conversation:

| Design | ICC 0.05 | ICC 0.10 |
|---|---|---|
| 500 items in 50 conversations | effective n about 345 | about 263 |
| 500 items in 10 conversations | about 145 | about 85 |
| LoCoMo: 1,540 items in 10 conversations | about 178 | about 94 |

The relevant ICC is that of the paired difference, which can be lower than for raw accuracy, so estimate it from pilot data.

Recommended analysis plan:

1. **Primary endpoint:** one pre-specified comparison, such as system versus full context at matched or lower cost, with the McNemar mid-p and the discordant counts reported.
2. **Interval:** paired bootstrap of the accuracy difference, resampling conversations rather than questions, with 10,000 resamples. With fewer than about 20 to 30 conversations, cluster bootstrap intervals can be unreliable, so add an exact sign or permutation test at conversation level.
3. **Multiplicity:** Holm across secondary comparisons, Benjamini-Hochberg only for exploratory per-category screens, and no significance claims for categories too small to support them.
4. **Non-inferiority:** a claim of "same quality at lower cost" requires a pre-set margin (for example, the lower confidence bound for the difference above -3 points). A non-significant difference is not evidence of equivalence.
5. **Repeats:** run the stochastic pipeline several times, and treat repeats as nested within items rather than as extra sample size.
6. **Holdout:** split by conversation into development and locked test sets, hash the locked file and analysis plan, publish the registration before the single final run, and report every unlocked rerun.
7. **Judge:** double-label a stratified sample of at least about 200 items, blind to condition, covering both judged-correct and judged-wrong cases, then report corrected accuracy and whether the sign of the paired difference survives.

### Gaps
- I did not retrieve the findings of [Berg-Kirkpatrick et al. (EMNLP 2012)](https://aclanthology.org/D12-1091/) because the page exposed only metadata, so it is listed as a reading target rather than evidence.
- The conversation-count threshold for cluster bootstrap and the Benjamini-Hochberg and Holm recommendations above are from general statistical knowledge. I did not retrieve a dedicated source for them in this pass.
- I found no published ICC values for paired differences on memory benchmarks.

## Cost-aware metrics

### Takeaway
Raw context tokens are not cost. Credible reporting separates build and serving cost, applies cache-aware pricing, amortizes over a stated number of queries, and compares methods at matched budgets. I found no memory paper with a formal cost-quality Pareto analysis, so this is an open opportunity.

### Cited Findings
- **Total Recall at What Cost? (2026-08-12) [PREPRINT]:** [compares Mem0, Hindsight and Mastra OM against a last-10-turn floor and a full-transcript ceiling, using billable input tokens across ingest, retrieval and answer, on dialogues up to 400 turns](https://arxiv.org/html/2608.11879).
  - [Accuracy on 665 LoCoMo questions with a gpt-oss-120b judge: Mem0 0.214 to 0.516, Mastra OM 0.308 to 0.502, Hindsight 0.493 to 0.541](https://arxiv.org/html/2608.11879).
  - [Cost per correct answer at N=100, L=100: Mastra OM $0.028 on gpt-oss low, Mem0 $0.037 on Gemma medium; backbone choice matters as much as the memory system](https://arxiv.org/html/2608.11879).
  - [Break-even against full history: Mastra OM 0 to 86 turns, Mem0 0 to 342, Hindsight 60 to never within 400](https://arxiv.org/html/2608.11879).
  - [Limits: no latency, no formal Pareto or budget-matched comparison, cost-model held-out error of 18 to 69%, and cache misses on fallback providers billed at the full rate](https://arxiv.org/html/2608.11879).
- **Beyond the Context Window (2026-03-05) [PREPRINT]:** [long-context GPT-5-mini recalled more facts than a Mem0-based system on LongMemEval and LoCoMo. With prompt caching modeled, memory became cheaper after about ten turns at 100k tokens](https://arxiv.org/abs/2603.04814).
- **Build versus inference tokens:** [LeanMem reports 117.61K build tokens per conversation and 3.62K inference tokens per question on LongMemEval-S, against 112.50K per question for full text; latency 2.16s against 9.92s](https://arxiv.org/html/2608.03463). It does not report money.
- **Latency:** [Mem0's paper reports full-context p95 of 17.117s against 1.440s for Mem0 on LoCoMo](https://arxiv.org/html/2504.19413) [VENDOR]. [LongMemEval-V2 query latency runs from about 0.1s for RAG to 108 to 140s for AgentRunbook-C and 177 to 186s for Codex](https://arxiv.org/html/2605.12493v1).
- **Budget matching precedent:** [LANTERN gives every method the same restore budget, and recovery rose from 36.6% at 2,000 characters to 73.2% at 6,000 and 77.7% at 8,000 before plateauing](https://arxiv.org/html/2606.05182). [BFR constrains retrieval calls under a fixed budget](https://arxiv.org/html/2610.09348).
- **Caching:** [SNIPPET: ["Don't Break the Cache" evaluates prompt caching on long agent tasks and includes a January 2026 cache price table](https://arxiv.org/pdf/2601.06007).] [SNIPPET: [TokenPilot notes that aggressive truncation breaks prefix caching](https://arxiv.org/html/2606.17016).]
- **Omission:** [Maximem reports no measured token cost or latency on benchmarks and projects savings from an illustrative model](https://arxiv.org/html/2607.21503).

### Inferences
The two break-even results above differ widely (about ten turns against 0 to 342), which suggests the answer depends heavily on backbone, cache behavior and workload. Report a curve, not one number.

Recommended metrics:

- **Billed-equivalent tokens** = uncached input + (cache-write multiplier × cache-write tokens) + (cache-read multiplier × cache-read tokens) + (output multiplier × output tokens). Publish raw counts for every class, so results can be repriced.
- **Cost per correct answer** = (amortized build cost + all serving cost, including retries, reranking and judge-free pipeline calls) ÷ correct answers, with a bootstrap interval computed from the same resampled items as accuracy. Exclude evaluation-only judge cost and report it separately.
- **Amortization curve:** cost per correct answer at 1, 10, 100 and 1,000 queries per conversation, since write-time systems depend on query volume.
- **Budget-matched baselines:** give BM25, embedding RAG, recency truncation and a simple summary the same delivered-token budget as the system, then sweep budgets (for example 1K, 2K, 4K, 8K, 16K and full).
- **Pareto report:** plot accuracy against cost and p95 latency with intervals, mark non-dominated methods, and state dominance only where the paired interval supports it.
- **Failure accounting:** count timeouts, truncations and refusals as wrong, and report their rates.

### Gaps
- I found no memory benchmark paper with formal Pareto reporting. This reflects my searches and does not prove none exists.
- Provider prices change often. Any dollar comparison must carry a dated price table, and I did not compile current prices here.

## What counts as a publishable contribution in 2026

### Takeaway
For a small system, the most defensible route is an evaluation contribution rather than a claim of a new best score: an audit, a causal protocol, or a cost-aware replication that changes a conclusion. NeurIPS's renamed Evaluations & Datasets track explicitly welcomes this, and does not require beating a baseline.

### Cited Findings

#### Stated venue expectations
- **NeurIPS 2026 Evaluations & Datasets:** [the former Datasets & Benchmarks track now treats evaluation as an object of study, welcomes critique, audits, reproductions and negative results, and does not require outperforming prior work](https://blog.neurips.cc/2026/03/23/introducing-the-evaluations-datasets-track-at-neurips-2026/).
- [Reviewers assess evaluative claims; benchmark papers are checked for rigorous design, reproducibility, saturation or overfitting, and whether they enable reliable comparison. Methodology papers are checked for statistical soundness and whether the method shifts conclusions on existing benchmarks. Audit papers need systematic analysis, and negative results need deep, controlled experiments](https://neurips.cc/Conferences/2026/EvaluationsDatasetsReviewerGuidelines).
- [Datasets need a valid Croissant file and Responsible AI metadata; evaluation tools generally must provide runnable anonymized code](https://neurips.cc/Conferences/2026/EvaluationsDatasetsReviewerGuidelines). [Accepted papers must make code and datasets public by camera-ready](https://neurips.cc/Conferences/2026/CallForEvaluationsDatasets) [SNIPPET].
- [The guidelines name no specific significance test or number of runs](https://neurips.cc/Conferences/2026/EvaluationsDatasetsReviewerGuidelines).

#### Recent examples, with acceptance evidence

| Paper | Venue evidence | Contribution |
|---|---|---|
| OP-Bench | [EMNLP 2026 main, confirmed on Anthology page](https://aclanthology.org/2026.emnlp-main.1168/) | New failure construct: over-personalization, 1,700 instances |
| Mem-Gallery | [ACL 2026 long, confirmed on Anthology page](https://aclanthology.org/2026.acl-long.1892/) | Multimodal memory benchmark, twelve systems |
| RealMem | [Findings of ACL 2026](https://aclanthology.org/2026.findings-acl.703/) [SNIPPET] | Project-based long-term interaction |
| Memora ("From Recall to Forgetting") | [Findings of ACL 2026](https://aclanthology.org/2026.findings-acl.1337/) [SNIPPET] | Weeks-to-months memory, with consolidation and updates |
| CloneMem | [ACL 2026 long](https://aclanthology.org/2026.acl-long.1549.pdf) [SNIPPET] | Long-term memory for AI clones |
| MemoryCD | [Findings of EMNLP 2026](https://aclanthology.org/2026.findings-emnlp.1804/) [SNIPPET] | Cross-domain user memory |
| UTILMEM | [Reported in EMNLP 2026 proceedings](https://aclanthology.org/volumes/2026.emnlp-main/) [SNIPPET, exact entry not opened]; [preprint](https://arxiv.org/pdf/2608.30508) | Evidence utilization rather than recall, 1,717 instances |
| MemContract | [NeurIPS 2026 poster, track not stated](https://neurips.cc/virtual/2026/loc/sydney/poster/154728) | Contract-based evaluation of mutable memory, with human audit and upper bound |
| "When Retrieval Is Not Enough" | [NeurIPS 2026 page](https://neurips.cc/virtual/2026/169567) [SNIPPET] | Separates missed updates from unused retrieved updates |
| EvoMemBench, MedMemoryBench, WorldMemArena, OmniMemBench, SMMBench | [Listed on the NeurIPS 2026 Sydney E&D page; titles only](https://neurips.cc/virtual/2026/loc/sydney/events/evaluations-datasets-2026) [SNIPPET] | Shows how crowded new benchmark submissions have become |
| HIPPOCAMPUS, AgenticCache, Ontology-Guided Long-Term Agent Memory | [Listed in an MLSys 2026 session](https://mlsys.org/virtual/2026/session/3674) [SNIPPET] | Systems work; AgenticCache page reports 22% higher task success, 65% lower latency and 50% fewer tokens, with paper attribution not fully separable in the snippet |

### Inferences

Likely reviewer objections for a small system, drawn from the guidelines and disputes above:

- No same-harness full-context, oracle, closed-book or BM25 baseline.
- Tuning on the reported test data.
- Only LoCoMo, given its small context and audited labels.
- An uncalibrated judge, or one model acting as reader and judge.
- Item-level intervals that ignore conversation clustering.
- Vendor systems run in weak configurations, the central complaint in the Mem0 and Zep dispute. Share configurations and, where possible, invite maintainers to check them.
- Token savings without cache-aware billing or build costs.

Contribution options, ranked for feasibility with a small system:

1. **Causal evaluation of memory use.** Apply removal, sufficiency and value-swap interventions to several systems and show how often correct answers do not depend on the cited or retrieved memory. This addresses a real gap and suits the E&D criterion that a method should shift conclusions.
2. **Controlled replication and sensitivity study.** Rerun full context, oracle, closed book, BM25 and two or three open systems while varying reader, judge and prompt. Decompose variance, then show which published rankings survive. The LoCoMo full-context spread provides the motivation.
3. **Judge-corrected re-scoring.** Human-calibrate judges on LongMemEval-S and LoCoMo, correct for sensitivity and specificity, and report which pairwise conclusions change.
4. **Cost-quality frontier.** Budget-matched, cache-aware Pareto analysis, a good fit for MLSys or an E&D paper when backed by real billing traces.
5. **A narrow system result.** Viable when a mechanism gives a pre-registered gain over budget-matched baselines on held-out data, across at least two benchmarks and two readers. A generic new benchmark is the weakest option, given the crowded field.

Minimum evidence package, as my judgment rather than a venue rule: at least two independent datasets, two readers from different providers, a locked test split, full-context, oracle, closed-book, BM25 and embedding baselines, human-calibrated judging, paired clustered intervals, cost and latency, per-item outputs, and a one-command rerun.

### Gaps
- I found no public OpenReview reviewer comments on a specific 2026 memory benchmark, so reviewer objections above are inferred from guidelines and public disputes, not observed reviews.
- I did not retrieve ACL Rolling Review or MLSys 2026 reviewer criteria, or acceptance rates for memory papers.
- Several acceptance entries are search-level only, as marked. Verify each before citing.
- Reading limits: most paper pages were read through an automated extractor, and some long papers were truncated. I checked key figures against raw arXiv HTML for LeanMem, Mem0, Zep, MemoryAgentBench, BEAM, LongMemEval, BFR and LongMemEval-V2. Other figures should be checked against the source before publication.
