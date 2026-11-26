# Context selection as an optimization and retrieval problem

Notes compiled 2026-11-18. Target system: a deterministic engine fusing exact-phrase, stemmed BM25 and value lanes with RRF over paragraph-sized spans, no embeddings yet.

Tags used below.
Compute: **[NO MODEL]** runs with counts and arithmetic; **[INDEX MODEL]** model download, used at write time only; **[LOCAL MODEL]** model download, used at query time, CPU feasible; **[GPU/LLM]** needs a GPU or an LLM call in normal use.
Evidence: **[M-author]** measured by the method's own authors; **[M-indep]** measured by a party with no stake in the method; **[C]** claimed without numbers I could check; **[peer]** peer-reviewed venue, otherwise preprint.
Reading depth: unless marked *(snippet)* or *(abstract)*, I read the paper's body or tables. *(snippet)* means search-result text only, so verify before relying on it.

## Budgeted selection: submodular and knapsack formulations, MMR, redundancy-aware selection, marginal utility per token, guarantees that actually apply

### Takeaway
The only guarantee that transfers cleanly to a token budget is for a **monotone submodular** objective selected by **gain per token plus a best-single-item fallback** (at least 0.405 of optimal). The most-cited recent method, AdaGReS, meets neither condition, and in the one 2026 head-to-head I found, plain MMR beat it. Selection effort pays most when the budget is tight relative to the candidate pool, and that effect is large (+17.4 points at 3 of 30 candidates, +1.5 at 20 of 30).

### Cited Findings

**Foundations and what the theorems say**
- 2010, [peer] NAACL. Lin and Bilmes cast summarization as budgeted submodular maximization. They select by gain divided by scaled cost, `(f(G ∪ {l}) − f(G)) / c_l^r`, and show that greedy without a safeguard has an unbounded approximation factor. For r = 1 they state a bound of (1 − 1/√e) ≈ 0.39. For r ≠ 1 they rely on an instance-dependent bound. They report beating the best DUC'04 system on ROUGE-1 — [Lin and Bilmes 2010](https://aclanthology.org/N10-1134.pdf).
- Aug 2020 (arXiv). Tang et al. analyze "modified greedy": run gain-per-cost greedy, separately find the best single item, return the better of the two. Density greedy alone has no bounded ratio. They prove a ratio above **0.405**, and give an instance where it achieves only about 1/2, so the true worst case lies between about 0.405 and 0.5. They report that the widely repeated Khuller et al. proof of 0.393 had a gap (found by Zhang et al. 2018), and they supply a corrected proof. They also give a data-dependent upper bound on OPT, usually showing above 0.9 of optimal in their viral-marketing experiments, at an extra log n factor — [Tang et al.](https://ar5iv.labs.arxiv.org/html/2008.05391).
- Partial enumeration combined with cost-weighted greedy reaches 1 − 1/e under a knapsack, at a much higher polynomial cost *(snippet)* — [Submodularity survey](https://arxiv.org/pdf/2006.09905).
- Oct 2026. The LMVU authors restate the operational point: **raw-gain greedy under a knapsack budget has an unbounded approximation ratio** — [LMVU](https://arxiv.org/html/2610.29234).

**Recent budgeted context selectors**

| Method (date) | Objective and selection rule | Guarantee status | Evaluation | Compute |
|---|---|---|---|---|
| [AdaGReS](https://arxiv.org/html/2512.25052) (31 Dec 2025) | α·Σsim(q,x) − β·Σsim(x,y). Greedy on **raw** marginal gain, skip if it does not fit. Closed-form β from pool statistics | ε-approximate submodular bound, f(S) ≥ (1−1/e)·OPT − kε/e, cited from prior work and **cardinality based**. The paper itself says classical results do not directly transfer to the token budget | NQ plus a proprietary drug corpus, against similarity top-k only. No MMR or DPP baseline. Gains of 8 to 15 points of IOU on "some hard queries", mostly in figures. Downstream QA is one example | Dense embeddings required |
| [RCD](https://arxiv.org/pdf/2605.00336) (1 May 2026) | αR(S) + βC(S) + γD(S): modular relevance, facility location, log-det diversity. Monotone submodular if k ≥ 0 and K is PSD | Claims knapsack guarantees via lazy greedy, but states no ratio I could find | MIMIC, Cochrane, L-Eval, PubMedQA. Budgets 256 to 16,384 tokens. Authors: "simpler objectives often suffice", MMR comparable to RCD | Similarity kernel |
| [LMVU](https://arxiv.org/html/2610.29234) (21 Oct 2026) | Learned linear scorer, 18 features, sequential add-if-fits | Explicitly **no** approximation bound | See below | all-MiniLM-L6-v2 features; 1.3 s CPU refit |
| [AdaGATE](https://arxiv.org/pdf/2605.05245) (4 May 2026) | Gap tracking, micro-queries, utility over gap coverage, corroboration, novelty, redundancy, relevance | None stated | HotpotQA, N = 200, budget 3,000 tokens | Embeddings; micro-query generation |

- **Head-to-head at 1,024 tokens** (Oct 2026, [M-author], viewpoint-coverage task, evaluation query counts of only 6, 97, 6 and 34). Recovered viewpoint coverage RVC1 / RVC2: LMVU 0.9588 / 0.8795, **MMR 0.9222 / 0.8314**, dense top-k 0.8566 / 0.7695, **AdaGReS 0.7771 / 0.6362**, exact oracle 0.9997 / 0.9558. Only LMVU's Perspectrum repeated-support gain over MMR (+0.0213) had a paired interval excluding zero. The authors label the full-pool evaluation exploratory. This is one task type, viewpoint coverage rather than fact recall — [LMVU](https://arxiv.org/html/2610.29234).
- **Budget determines which selector wins** (May 2026, [M-author]). On MIMIC, Lead (document order) is strong at 256 to 512 tokens and MMR dominates near 1,024. Sentence/MMR at 512 tokens matched full context on ROUGE-1 (.277 vs .277) and exceeded it on BERTScore (.117 vs .106), N = 100. On PubMedQA long contexts, Lead was the **worst** selector at every budget from 64 to 256 tokens, with shuffled order up to 20 points better. Selector choice mattered more than unit choice; cluster-based units hurt — [Qureshi et al.](https://arxiv.org/pdf/2605.00336).
- **Selection value depends on tightness** (28 Sep 2026, pre-registered, [M-author]). Reranking's gain over cosine order is +17.4 points on LoCoMo and +9.1 on LongMemEval when 3 of 30 candidates are kept, falling to +1.5 and +1.1 when 20 of 30 are kept. At the tight budget raw turns picked by one selector call were non-inferior to an LLM-extraction memory (one-sided 95% bound −3.0 against a −5 margin) and cost 3,061× less to write. At generous budgets extraction systems were more accurate. The selector, Jev, is a typed decision model called per query over a cosine shortlist, so this result is **[GPU/LLM]** and also needs embeddings — [Sharma and Lall](https://arxiv.org/pdf/2609.34227).
- **Token efficiency against evidence quality** (May 2026, [M-author], N = 200). AdaGATE reached 62.3% evidence F1 on clean data and 71.2% under redundancy injection using 360 tokens clean, 2.6× fewer than Adaptive-k (1,116), with input falling to 220 to 232 tokens under redundancy. Adaptive-k still had the highest clean answer accuracy (69.0%) with precision 0.278 and recall 0.820, so the F1 winner was not the accuracy winner — [AdaGATE](https://arxiv.org/pdf/2605.05245).

### Inferences
- AdaGReS's objective subtracts a growing redundancy term, so it is not monotone, and it selects on raw gain. Its headline guarantee therefore should not be cited for a token budget. Treat it as a heuristic, and prefer MMR as the baseline to beat.
- **Recommended [NO MODEL] design.** Make the objective a monotone submodular lexical coverage function: for each query term, stem, value token or entity `t`, score `w_t · g(count of t in S)` with concave `g` (for example sqrt or a saturating cap), where `w_t` comes from lane evidence and IDF. Add a modular term from the fused RRF score. Then run lazy greedy on gain per token, compare with the best single span that fits, and break ties by stable span ID. A positive-weight concave-over-modular function is monotone submodular, so the 0.405 bound applies, and the run is fully deterministic. Tang's upper bound can be logged per query as a free "how far from optimal could this be" diagnostic.
- Redundancy needs no embeddings at this stage: shared-stem or shingle overlap can stand in for cosine similarity. No paper I found validates lexical redundancy against embedding redundancy for context selection, so this needs a local test against MMR with a Jaccard kernel.
- Density greedy favors short spans. With paragraph units that variance is moderate, but test r < 1 as well. Only r = 1 keeps the constant-factor argument.
- Measure selection at the engine's actual pool ratio. Sharma and Lall suggest effort is worth spending when the budget admits about 10% of candidates and is close to wasted at about 67%.

### Gaps
- I found no study that compares knapsack-correct density greedy with raw-gain greedy, on the same RAG or memory task, with downstream accuracy.
- No token-budgeted selection study on conversational or agent transcripts with a purely lexical objective.
- The LMVU, RCD and AdaGATE results are all single-group preprints. None has independent replication that I found.
- I did not locate a primary numeric evaluation for "context as knapsack" outside clinical text.

## Retrieval models that run locally and cheaply: sparse, late interaction, cross-encoders, fusion, BM25 variants, local embeddings on conversational memory

### Takeaway
On memory benchmarks, BM25 is a strong floor and adding one small embedding model is the best-supported next step: about +9 Recall@10 points on LongMemEval and +6 on LoCoMo. Token-level matching over a small model's existing token vectors currently shows the largest gains (+13 to +20 over BM25), at about 20 to 30 ms and roughly 100× BM25's storage or more. Fusion should be treated as tunable: the main study found convex combination beats RRF, and hybrid gains are benchmark dependent.

### Cited Findings

**Conversational memory, one comparable table** (7 Nov 2026 v3, [M-author]; the authors' own method is the last row, so its margin deserves caution). Recall@10 / NDCG@10 in percent — [Luo et al.](https://arxiv.org/pdf/2610.13277).

| Method | LoCoMo | LongMemEval | REALTALK R@10 | Compute |
|---|---|---|---|---|
| BM25 | 48.3 / 36.7 | 81.3 / 73.0 | 40.0 | [NO MODEL] |
| Qwen3-Embedding-0.6B alone | 46.9 / 34.2 | 84.8 / 72.8 | 46.6 | [LOCAL MODEL] |
| Qwen3-Embedding-4B alone | 41.6 / 29.9 | **59.2 / 47.1** | 42.8 | heavier model |
| BM25 + 0.6B (normalized, equal weight) | 54.3 / 42.1 | 90.5 / 81.5 | 46.5 | [LOCAL MODEL] |
| ColBERTv2 | 61.8 / 49.4 | 86.3 / 79.5 | 52.6 | [LOCAL MODEL] |
| ConstBERT | 63.1 / 50.3 | 85.7 / 76.6 | 55.9 | [LOCAL MODEL] |
| MaxSim on frozen 0.6B token vectors | 64.2 / 51.3 | 92.2 / 82.5 | 53.4 | [LOCAL MODEL] |
| Sparse token matching on 0.6B, training-free (authors) | 68.0 / 53.4 | 94.5 / 87.1 | 55.1 | [LOCAL MODEL] |

- Costs on LoCoMo from the same paper (Table 5, size in MiB / query ms): BM25 0.23 / 0.76; Qwen 0.6B 2.23 / 16.66; ConstBERT 4.45 / 10.18; ColBERTv2 1.54 / 27.27; authors' method full 153.93 / 29.21; with 4-bit scalar quantization 26.17 / 21.68 at Recall@10 68.37. The bigger 4B embedder alone lost to BM25 by 22 points on LongMemEval, so model size is not a proxy for memory retrieval quality — [Luo et al.](https://arxiv.org/pdf/2610.13277).
- Shrinking dense windows did not close the gap: on LoCoMo with Qwen 0.6B, independent windows down to 64 tokens stayed at Recall@10 44.99 to 46.87, against 67.98 for token matching — [Luo et al.](https://arxiv.org/pdf/2610.13277).
- **Hybrid end-to-end** (Mar 2026, [M-author], 100 stratified LongMemEval questions, 8B reader). Cosine-only F1 36.5, BM25 + cosine 44.2 (+7.7). By type: knowledge-update +26.7 (36.3 → 63.0, n = 16), single-session-user +19.0 (n = 14), temporal +5.7 (n = 26), multi-session −2.2 (n = 27), single-session-assistant −3.0 (n = 11). Per-type samples are small — [Liu et al.](https://arxiv.org/pdf/2603.23013).
- **Hybrid does not always help** (2026, [M-author]). On full LoCoMo at k = 50, hybrid BM25 + dense and vector-only tied at about 90.2% and 90.3% mean evidence recall — [FluctlightDB](https://arxiv.org/pdf/2608.12365).
- **Small embedder in a BM25 system** (2026, [M-author], N = 100, LongMemEval-S). bge-small-en-v1.5 hybrid RRF stage: accuracy 0.360 → 0.390 at 3.7× retrieval latency (96.7 vs 25.7 ms). Multi-session +0.077, temporal +0.037 — [MEMTIER](https://arxiv.org/pdf/2605.03675).
- **Benchmark validity.** MEMTIER found LoCoMo scores insensitive to memory architecture when conversations are placed in context, and recommends LongMemEval-S as more discriminating. Its LoCoMo run is a context-stuffed setup, so this does not impugn LoCoMo retrieval-only metrics — [MEMTIER](https://arxiv.org/pdf/2605.03675).

**Fusion**
- 2022 to 2023, [peer] ACM TOIS. Convex combination of scores beat RRF in-domain and out-of-domain, RRF was sensitive to its parameters, CC was largely agnostic to normalization choice once learned, and a small labeled set sufficed to tune its single weight — [Bruch et al.](https://arxiv.org/abs/2210.11934).
- Unnormalized BM25 can swamp bounded signals: MEMTIER traced its failed weight learning to raw BM25 dominance, naming score normalization as a precondition — [MEMTIER](https://arxiv.org/pdf/2605.03675).

**Learned sparse and document expansion**
- 2023, [peer] ECIR. Doc2Query-- filters hallucinated generated queries before indexing: up to +16% effectiveness, −23% mean query time, −33% index size against Doc2Query. **[INDEX MODEL]**, inverted index unchanged at query time — [Gospodinov et al.](https://arxiv.org/abs/2301.03266).
- Apr 2025, [peer] SIGIR. Li-LSR makes query encoding a **table lookup**, so a model is needed only for documents and training. MS MARCO MRR@10 38.8 (Big) against 37.8 for SPLADE-v3-Doc; BEIR nDCG@10 48.8 against 47.0. Single-thread CPU on a Xeon Silver 4314: an LSR encoder reached MRR@10 39.80 in about 800 µs with Seismic, while query encoding averaged 24 ms, so encoding is the bottleneck. **[INDEX MODEL]** plus a learned token table — [Nardini et al.](https://arxiv.org/pdf/2505.01452).
- 19 Oct 2026, vendor blog, [M-author]. BEIR-15 nDCG@10: SlideOn 56.68 (174M params), SPLADE-v3 50.46 (133M), DenseOn dense 56.20, LateOn late interaction 57.22. Retrieval FLOPS are 30.87 for SlideOn against 5.55 for SPLADE-v3, so it is denser and costlier to traverse. No CPU latency or storage figures. **[LOCAL MODEL]**, sizeable — [LightOn](https://huggingface.co/blog/lightonai/slideon-mslideon).

**Late interaction**
- 2022. PLAID cut ColBERTv2 search latency by up to 45× on CPU and 7× on GPU, tens to a few hundred ms on CPU at up to 140M passages *(abstract)* — [PLAID](https://arxiv.org/abs/2205.09707).
- answerai-colbert-small-v1: 33.4M params, 96-dim token vectors, Apache-2.0, BEIR average 53.79 against 50.02 for ColBERTv2.0 and 53.25 for the 109M bge-base-en, per its own card ([M-author], labeled a proof of concept, no latency claims). **[LOCAL MODEL]** — [model card](https://huggingface.co/answerdotai/answerai-colbert-small-v1).

**Small cross-encoders**
- cross-encoder/ms-marco-MiniLM-L6-v2: 22.7M params, Apache-2.0, TREC DL19 NDCG@10 74.30, MS MARCO MRR@10 39.01, 1,800 docs/s. **That throughput was measured on a V100 GPU**, so it is not a CPU figure. The L2 variant gives 71.01 / 34.85 at 4,100 docs/s; L12 adds nothing (74.31) at about half the speed — [model card](https://huggingface.co/cross-encoder/ms-marco-MiniLM-L6-v2).
- CPU latency reports for MiniLM rerankers and FlashRank are blog-grade and conflict, from 8 to 25 ms to about 550 ms depending on candidate count, quantization and hardware *(snippet)* — [llms.blog](https://www.llms.blog/posts/context-reranking-engines-in-production-rag-comparing-cohere-rerank-bge-reranker-v2-flashrank-and-colbert-late-interaction), [Neural Base](https://theneuralbase.com/reranking/learn/beginner/flashrank-lightweight-option/).

**Local embedding options**
- EmbeddingGemma: about 300M parameters, 768 dimensions truncatable to 128, under 200 MB RAM when quantized, per coverage of its card *(snippet)* — [model card mirror](https://huggingface.co/Ccre/gemma-embedding-model), [OpenSourceForU](https://www.opensourceforu.com/?p=89735).

### Inferences
- **Priority order for this engine.** (1) Stay embedding-free until lexical selection and query handling are measured. (2) If a model is added, the best-supported first step is one small embedder as an extra lane. Its token vectors could later enable MaxSim-style matching on the top candidates only, which limits storage. (3) A cross-encoder is a separate decision for the final tight cut, where Sharma and Lall suggest reranking pays most.
- MaxSim or token matching over a **shortlist from the lexical lanes** would avoid most of the 154 MiB-class index cost, since token vectors can be computed for candidates on demand or cached. This is untested in the sources.
- RRF discards score magnitude, which the existing exact-phrase and value lanes may carry usefully. Worth testing: per-lane weighted RRF against a convex combination of per-query normalized scores, with weights tuned on the repo's probes. Bruch's result comes from document retrieval, not memory, so confirm it locally.
- LongMemEval at 81.3 for BM25 suggests that Recall@10 is nearly saturated for the easy classes. Evaluate by question type, and at the actual token budget rather than at k = 10.

### Gaps
- **BM25 variants:** I did not find 2025 to 2026 evidence isolating BM25 variants (BM25+, BM25L, ATIRE) on conversational or agent text. FluctlightDB claims a surprisal channel "matches or beats" Okapi BM25 in ablation, but I did not verify it.
- No independent CPU benchmark on a laptop-class machine covers these small models on memory data. Luo et al.'s latency table does not, in the part I read, establish a laptop-equivalent setup.
- Learned sparse models have not, in what I found, been evaluated on LongMemEval or LoCoMo.
- Vendor LongMemEval numbers (for example 95 to 97% at Recall@5) use differing protocols and are not comparable, so I excluded them.

## Granularity: propositions, late chunking, contextual retrieval, parent-child, chunk size. Retrieve small, read large?

### Takeaway
No single chunk size wins, and the evidence supports **indexing extra derived keys while returning verbatim text** more firmly than it supports hierarchy. Paragraph spans are a defensible unit. Cheap structural context (real headings, speaker, date) has better controlled evidence than LLM-written context. Neighbor expansion needs token-honest evaluation, because it can make trivial baselines look perfect.

### Cited Findings
- **Chunk size is dataset dependent** (May 2025, [M-author], two embedders, sizes 64 to 1,024, Recall@1, Stella / Snowflake). SQuAD* falls from 0.642 / 0.609 at 64 tokens to 0.386 / 0.429 at 1,024. TechQA* rises from 0.049 / 0.045 to 0.619 / 0.715. NQ goes from 0.172 / 0.150 to 0.349 / 0.477. Short fact answers favor small chunks and dispersed answers favor large ones. Several sets are synthetic and matching is by string — [Rethinking Chunk Size](https://arxiv.org/html/2505.21700).
- **Propositions** (Dec 2023, revised Oct 2024). Recall@5 against passages on NQ: SimCSE +12.3, Contriever +7.6, but **DPR −0.6, GTR +0.2**. On EntityQuestions: +21.1, +15.4, +11.9, +3.2. Gains concentrate in weaker or unsupervised retrievers and long-tail entities. With GTR, EM@500 rose +2.8 under a capped prompt. Cost: 256.9M propositions against 41.4M passages (about 6.3 per passage), index about 768 GB. **[INDEX MODEL]** (Flan-T5-Large propositionizer) — [Dense X Retrieval](https://arxiv.org/html/2312.06648).
- **Placebo-controlled structure test** (7 Oct 2026, [M-indep] with respect to the methods tested; 951 Wikipedia and 4,303 QASPER queries, dense retrieval, coverage-nDCG@10):
  - Structure-aligned chunks with real heading paths beat LLM-contextualized fixed windows, +0.022 and +0.012.
  - Real heading paths beat shuffled placebo paths, +0.010 and +0.016, so content rather than added tokens drives the gain.
  - Route-to-section-then-rank **lost** to flat search, −0.033 and −0.015, with first-stage section recall as low as 0.17 to 0.20 on QASPER.
  - At 512 tokens, LLM prefixes fell below no prefix.
  - Effects are small (dz about 0.06 to 0.11), and parent-child reranking was not tested — [Kuehlkamp et al.](https://arxiv.org/html/2610.10170).
- **Contextual Retrieval** (19 Sep 2024, vendor, [M-author]). Top-20 miss rate: embeddings 5.7%, contextual embeddings 3.7% (−35%), plus contextual BM25 2.9% (−49%), plus reranking 1.9% (−67%). Cost $1.02 per million document tokens under its stated assumptions. **[GPU/LLM]** at index time. Per-dataset numbers were not in the text I could read — [Anthropic](https://www.anthropic.com/news/contextual-retrieval).
- **Late chunking** (Sep 2024, revised Jul 2025) embeds the full text first and pools per chunk, which **requires a long-context embedder** *(abstract)* — [Günther et al.](https://arxiv.org/abs/2409.04701). Its reported benefit is largest for small chunks (NFCorpus nDCG@10 28.3 → 30.0 *(snippet)*), and a later benchmark's citation of a 21-point nDCG@10 drop on CovidQA is secondhand — [Kuehlkamp et al.](https://arxiv.org/html/2610.10170).
- **Memory-specific granularity** (LongMemEval, v2 Mar 2025, [M-author], LongMemEval-M). Recall@5 / NDCG@5: rounds 0.582 / 0.481, rounds with fact-augmented keys 0.644 / 0.498, sessions 0.706 / 0.617, sessions with fact keys 0.732 / 0.620. Fact keys gave an average 9.4% recall and 5.4% accuracy gain. Round and session recall have different targets, so do not read this as sessions beating rounds — [Wu et al.](https://arxiv.org/html/2410.10813).
- **Verbatim over extraction.** One 2026 study, as reported by Sharma and Lall, had verbatim chunks ahead of LLM-extracted artifacts by 15.9 points on LoCoMo and 22.0 on LongMemEval-S (secondary citation, I did not read the original) — [Sharma and Lall](https://arxiv.org/pdf/2609.34227).
- **Expansion can corrupt evaluation.** FluctlightDB withdrew an earlier 99.0% result because ±3 session-neighbor expansion credited adjacent spans. Under that policy **a trivial BM25 baseline also scored about 99%** — [FluctlightDB](https://arxiv.org/pdf/2608.12365).

### Inferences
- Retrieve-small-read-large is best implemented here as **multi-key indexing on spans with verbatim output**: keep paragraph spans as the charged unit, and add deterministic keys to BM25 only (speaker, date, file path, heading path, tool name, extracted identifiers). This is the no-LLM analogue of fact-augmented keys and Contextual BM25. Whether it delivers comparable gains is unmeasured.
- Avoid route-first hierarchy. Treat session or section evidence as a score feature in flat ranking, so low first-stage recall cannot remove a good span.
- Make neighbor expansion a **priced decision in the selector**: a neighbor is a candidate with token cost and marginal gain, for example when a hit is a question turn and the next turn contains the answer. Score recall only for what is actually inside the budget.
- Consider sentence-level trimming as a second pass on selected spans, keeping the span boundary as the citation unit. Qureshi et al. found selector choice mattered more than unitization, so prioritize this below the selector.

### Gaps
- I found no clean controlled evaluation of parent-child or small-to-big retrieval on conversational memory with token-charged reading. Kuehlkamp et al. name parent-child reranking as untested.
- No independent replication of Contextual Retrieval on agent transcripts. The specific contribution of contextual BM25 alone is not in the text I read.
- Optimal unit size for tool output and code in agent transcripts remains unmeasured in these sources.

## Query side: HyDE, multi-query, decomposition, deterministic rewriting, time-aware retrieval for "what was it before"

### Takeaway
Classical feedback gives a small but free gain over BM25 (+2.2 nDCG@20 points with RM3), and LLM expansion gives more (+8 to +9) at a model-call cost. For "what was it before", the strongest finding is that **state modeling at write time beats ranking**: with validity and participant metadata, plain BM25 matched the reference ranker at sub-millisecond latency. Without it, about 70 to 80% of prompts exposed stale values.

### Cited Findings

**Expansion and feedback**
- **Mar 2026, [M-author], 13 low-resource BEIR sets, nDCG@20 on BM25.** No feedback 36.5; **RM3 38.7; Rocchio 37.3** (both [NO MODEL]). LLM-judged corpus feedback 43.3 (RM3), 43.0 (Rocchio). HyDE-style generated feedback 43.6 (RM3), 44.5 (Rocchio). Both sources together 45.2 (Rocchio), best on BM25. The feedback model choice matters most for LLM-generated text — [Jedidi and Lin 2026](https://arxiv.org/html/2603.11008).
- **Nov 2025, [M-author], 14 datasets, BM25 Recall@20, all-dataset average.** BM25 28.8. HyDE by naive concatenation 29.3 to 30.2, so almost no gain. HyDE with Rocchio term selection 34.0 to 34.7. Term selection and weighting, not generation alone, produced the benefit — [Jedidi and Lin 2025](https://arxiv.org/html/2511.19349).
- Oct 2025. A generalized feedback study describes RM3 and HyDE-type gains over BM25 as moderate and unstable *(snippet)* — [Generalized PRF](https://arxiv.org/html/2510.25488v1).
- **Choosing among variants** (Apr 2026, [peer] SIGIR '26, only 56 TREC-RAG 2024 queries, 30 GPT-4o reformulations per query). On BM25, Nugget-All: original query 0.273, best single reformulator 0.371, **selection by pre-retrieval IDF_max 0.398**, post-retrieval NQC 0.381, oracle 0.511. Selecting the variant with the best retrieval metric was not the same as the best for the answer. The selector is [NO MODEL], the variants are [GPU/LLM] — [Arabzadeh et al.](https://arxiv.org/html/2604.22661).

**Time and knowledge updates**
- **Controlled revision benchmark** (7 Oct 2026, [M-author], 1,200 queries, k = 6, 800-character budget). BM25 + validity + participants: current value retrieved 1.000, stale exposure 0.000, clean retrieval 0.873, **p50 0.10 ms**, within the preset ±0.02 of the reference ranker (0.877). The same BM25 without validity: stale exposure 0.802, clean 0.195. Dense and hybrid with validity: clean 0.836 and 0.843 at about 18 ms. Keyless stores exposed a superseded value in 70.3% of prompts. The benchmark is a small synthetic template family — [Deng and Sinchaisri](https://arxiv.org/html/2610.10265).
- The hard part is key assignment. A rule extractor reached 0.807 key recall and 0.781 clean retrieval. LLM key assigners reached 0.980 to 0.996 key recall but only 0.560 to 0.644 clean retrieval because false merges removed current values. On a labeled LongMemEval subset (32 of 78 knowledge-update questions), merge recall never exceeded 0.062 and stale exposure stayed between 0.594 and 0.719 — [Deng and Sinchaisri](https://arxiv.org/html/2610.10265).
- **Time-aware expansion** on LongMemEval-M temporal questions, Recall@5 with a GPT-4o time-range extractor: session 0.639 → 0.654, round 0.421 → 0.451, session with fact keys 0.684 → 0.722, round with fact keys 0.489 → 0.526. Llama 3.1 8B extraction gave smaller gains — [Wu et al.](https://arxiv.org/html/2410.10813).
- **Bounded memory loses updates** (Jun 2026, [M-author]). gpt-5.4 on LongMemEval knowledge-update (n = 78): 92% with full context, 77% with bounded self-maintained memory (McNemar p = 0.0033). The cause is mostly facts compressed away or not overwritten — [Supersede](https://arxiv.org/html/2606.27472v1).
- **Showing history can hurt.** One 2026 paper reports explicit bitemporal edges scoring below prose-inline history at answer time (72/81 vs 79/81), attributed to rendering superseded edges *(snippet)* — [Ground Truth First](https://arxiv.org/pdf/2607.21962).

### Inferences
- Add **RM3-style feedback as a [NO MODEL] lane candidate**, gated by a confidence check, because feedback can reinforce a poor first pass. Its deterministic behavior fits the engine if term ties are stably broken. The published gain is on document collections and may not hold on short conversational spans.
- A validity filter alone would **break "what was it before"** by hiding exactly the required value. Use a deterministic intent router instead:
  - *Current-state queries* ("what is", no time cue): prefer the latest member of a value chain, suppress superseded members.
  - *History queries* ("before", "previously", "used to", "originally", "changed from"): retrieve the chain and include the immediate predecessor, ordered by time and labeled with dates.
  - *Change queries* ("when did", "why did it change"): include the transition pair.
- The existing value lane is the natural place to build chains, since exact values give a conservative key. Favor precision in merging: the evidence shows false merges destroy current values.
- Render old values with explicit "superseded on DATE" labels to reduce the risk that a reader takes them as current.
- Deterministic decomposition is plausible for conjunctions and "X vs Y" patterns via separate lane queries with per-subquery coverage terms in the submodular objective. I have no direct evidence for it.

### Gaps
- I did not find a benchmark built specifically around "what was the previous value" questions. LongMemEval's knowledge-update class mostly tests the latest value.
- I found no head-to-head of rule-based query rewriting against LLM multi-query on memory data.
- Query decomposition evidence on token-budgeted memory retrieval was not found in this pass. AdaGATE's micro-queries are the closest item, on HotpotQA.
- The original HyDE paper was not re-read; the figures above come from later re-evaluations.

## Predicting when retrieval is insufficient: sufficiency, QPP, adaptive budgets, Self-RAG/CRAG, recall-as-a-tool against pre-fetch, agentic against one-shot

### Takeaway
Cheap signals can drive an **adaptive cutoff** but are weak at predicting whether the answer is actually present: unsupervised post-retrieval predictors correlate about 0.2 to 0.3 with the benefit of retrieval, and pre-retrieval IDF is near zero on two of three datasets (0.14 on the third). The sound design is pre-fetch plus an exposed recall tool, with conservative triggering. One-shot and iterative retrieval have reached similar accuracy when one-shot is given enough budget, and iterative wins when context is scarce.

### Cited Findings

**Sufficiency and prediction**
- Nov 2024, revised Apr 2025, [peer] ICLR 2025. Joren et al. define sufficient context independent of a gold answer. Large models answer well when context suffices but often answer wrongly rather than abstain when it does not. Using the sufficiency signal for selective generation raised correctness among answered questions by 2 to 10% for Gemini, GPT and Gemma *(abstract)*. **[GPU/LLM]** — [Sufficient Context](https://arxiv.org/abs/2411.06037).
- **Predicting retrieval benefit** (2026, v3 Aug 2026, 3,600 test questions per dataset, Pearson r, BM25 with Falcon, Q_E5 metric). IDF pre-retrieval: .141 NQ, −.013 HotpotQA, −.045 TriviaQA. Best unsupervised post-retrieval shown (REF): .255, .305, .214. Supervised retrieval-stage BERT: .450, .375, .304. Internal LLM prober: .530, .484, .547. Post-generation Bert-Gen-2A up to .873 — [Dado et al.](https://arxiv.org/html/2604.07985).
- The apparent conflict with Arabzadeh et al., where pre-retrieval IDF_max selected good variants, reflects different tasks: **ranking variants of one query is easier than predicting absolute benefit**. NQC also correlated about 0.33 with nDCG@5 and about −0.004 with answer nuggets there, so a good retrieval score need not imply a good answer — [Arabzadeh et al.](https://arxiv.org/html/2604.22661).

**Adaptive budgets**
- **Adaptive-k** (Nov 2025, [peer] EMNLP). Sort similarity scores, cut at the **largest gap** (searched within the top 90%), add a fixed buffer of B = 5. No LLM calls or tuning. About 70% context recall, 2× to 10× fewer tokens than full context, up to 99% reduction on factoid QA, and up to +9 accuracy points over Self-Route on high-information aggregation tasks. Tested on embeddings over long contexts — [Taguchi et al.](https://aclanthology.org/2025.emnlp-main.1017.pdf).
- 2025 to 2026 alternatives: Stop-RAG treats stopping as a learned MDP; MuSiQue F1 47.0 against 44.8 for LLM-based stopping, slightly worse on HotpotQA (NeurIPS 2025 workshop, secondary page) — [Stop-RAG summary](https://lacuna.tiptreesystems.com/work/stop-rag-value-based-retrieval-control-for-iterative-rag/wrk_e98d85255f3d1f73dcc1900de53204f1). QCAR assigns retrieval depth by offline query clusters *(snippet)* — [QCAR](https://arxiv.org/abs/2609.13489v2).

**Corrective methods under replication**
- Mar 2026, [M-indep], single author, single run. CRAG on PopQA: 54.4 reproduced against 54.9 original, vanilla RAG 51.4. On ARC-Challenge 85.2 against vanilla RAG 84.8, with **88.3% of questions classified Ambiguous**. SHAP on 9 case studies suggests the T5 evaluator acts largely as an entity-alignment detector. Generator and search backend differ from the original — [Yalavarthi](https://arxiv.org/html/2603.16169v1).
- Abstention is fragile. Reranking lowered correct abstention in Sharma and Lall — [Sharma and Lall](https://arxiv.org/pdf/2609.34227). A schema-free margin gate abstained on 15.8% of answerable requests and cut clean retrieval by 0.134, while a slot-based gate abstained correctly on all 60 unanswerable requests but needs a known schema — [Deng and Sinchaisri](https://arxiv.org/html/2610.10265).

**One-shot against iterative, tool against pre-fetch**
- **Government documents** (Sep 2025, [M-author], 1,000 questions, 40,000+ documents, LLM judge). Basic top-5 81.0. One-shot filling a token budget 87.5, with a rule-based chunk filter **91.0**. Basic agentic 83.5, with fallback search 87.5, with chunk deletion **90.0**. Mean 1.59 to 1.97 retrieval calls. Combining a large initial context with iteration was worse than iteration alone: bigger starts degraded chunk deletion and encouraged "retrieval laziness". The authors prefer one-shot when tokens are sufficient and iteration when tokens are limited — [Lin et al.](https://arxiv.org/html/2509.04820).
- **Direct corpus interaction** (May 2026, [M-author]). BrowseComp-Plus with Sonnet 4.6: grep/read-style agent 80.0% against 69.0% with an embedding retriever, total cost $1,016 against $1,440. **It degrades with scale**: going from 100K to 200K documents raised tool calls per question from 38.5 to 86.9 and cost 13.6 accuracy points; at 400K accuracy was 37.5%. The gain appears to come from using partial evidence better: gold-document coverage was lower than the retriever's (28.0 against 56.7). Model and harness differences confound some comparisons — [DCI](https://arxiv.org/html/2605.05242).
- **Proactive memory** (Jul 2026, [M-author], τ²-Bench, macro accuracy, Sonnet 4.5 actor, Opus 4.6 memory agent). Baseline 57.5, passive full-bank exposure 61.5, general retrieval via Mem0 62.1, always-on injection 63.5, selective intervention 64.3. Selective and always-on are essentially tied (0.3 micro points the other way, within run variance). The memory agent's token cost is not reported — [Remember When It Matters](https://arxiv.org/html/2607.08716v1).
- Pre-fetch and on-demand recall serve different needs: pre-retrieval makes constraints visible from the first step, on-demand handles relevance that cannot be known in advance *(snippet)*, a design argument rather than a measurement — [SemaClaw](https://arxiv.org/pdf/2604.11548).

### Inferences
- **Deterministic sufficiency signals to try**, all [NO MODEL]: fraction of query IDF mass covered by the selected set; whether required entities and values were matched in the exact or value lanes; agreement between lanes in the top ranks; score gap on normalized BM25; presence of a time cue with no dated hit; whether a value chain was resolved. Combine with a small logistic model fit on local probes. The literature suggests that unsupervised signals alone will be weak, so calibrate before making behavior depend on them.
- Use the signal to **allocate budget and word the brief**, for example "coverage low: these terms were not found, use recall", rather than to withhold content. A false "sufficient" is the costly error, so never let the signal suppress the recall tool.
- A gap cutoff needs real scores. RRF values are rank-derived and carry almost no gap information, so compute the gap on normalized lane scores. Adaptive-k has only been shown on embedding similarities, so this is an untested transfer. Prefer the budget-aware stopping rule from the selector: stop when marginal gain per token falls below a threshold, which leaves unused budget for the conversation.
- Lin et al.'s "retrieval laziness" warns against filling the whole budget by default if the agent also has a recall tool. Test pre-fetch at 50%, 75% and 100% of budget with recall available, and record both answer quality and recall-call rate.
- Grep-style recall suits a single user's bounded history, since DCI's degradation starts at corpus sizes far above a typical transcript store. Exact tools complement ranked recall rather than replacing it.

### Gaps
- I found no controlled study that isolates **pre-fetch against recall-tool only against both** for coding-agent memory with token cost reported, and none measuring how often agents fail to call an available memory tool.
- I found no independent reproduction of Self-RAG's headline gains. AdaGATE's re-implementation is an implementation-dependent data point, not a replication.
- No validated QPP for short conversational spans or BM25-over-transcript settings.
- Claims of grep beating vectors for coding agents in my searches were mostly blogs, so I excluded their numbers.

## Learning from implicit feedback such as citations or recall requests; offline counterfactual learning to rank

### Takeaway
The counterfactual toolkit is mature, but I found **no published work that uses LLM citations or recall requests as ranking labels for context selection**. For this engine the realistic first step is a low-dimensional, inspectable weight model trained on versioned, frozen logs, with deterministic replay. The main obstacle is exposure: spans that never enter the context can never be cited.

### Cited Findings
- **Foundations** ([peer]). Joachims et al. (2017) correct click position bias with propensity-weighted learning, with click models supplying propensities *(snippet)* — [Unbiased LTR](https://www.cs.cornell.edu/people/tj/publications/joachims_etal_17a.pdf). Agarwal et al. (SIGIR 2019) generalize counterfactual LTR to additive metrics such as DCG, noting that naive use of implicit feedback is distorted by presentation bias *(snippet)* — [General framework](https://www.cs.cornell.edu/people/tj/publications/agarwal_etal_19b.pdf).
- **Two-stage systems** (ICTIR 2025, v4 Jan 2026, [peer]). Joint counterfactual training of a candidate generator and reranker beat independent training, and independent training beat the baseline. At K2 = 1,000 and 1M interactions, NDCG@10 was 0.504, 0.496 and 0.469. A reranker trained on the logging policy's candidates degraded when candidate generation changed. Evidence is semi-synthetic (MovieLens-1M, simulated position-based clicks), and it needed 100K to 1M interactions — [Gupta et al.](https://arxiv.org/html/2506.20854).
- **Small learned selector** (Oct 2026, [M-author]). LMVU learned 18 coefficients with a pairwise logistic loss over **marginal gains within a selection state**, refitting in about 1.3 s on CPU, with inference as one dot product per candidate. It needed viewpoint labels from development splits, and a static three-feature variant slightly beat it on one benchmark (by 0.004), so most value may lie in a few features — [LMVU](https://arxiv.org/html/2610.29234).
- **A failed learning loop** (2026, [M-author]). PPO weight adaptation gave no gain (0.382, ±0.000) because raw BM25 dominated bounded signals. Removing individual signals changed accuracy by only +0.012 to +0.014 — [MEMTIER](https://arxiv.org/pdf/2605.03675).
- **Fusion weights are sample-efficient**: a small labeled set was enough to tune convex combination's single weight — [Bruch et al.](https://arxiv.org/abs/2210.11934).
- **Retrieval-optimal is not answer-optimal.** Choosing variants by oracle nDCG@5 gave BM25 Nugget-Strict of 0.344, below the 0.377 from a pre-retrieval selector — [Arabzadeh et al.](https://arxiv.org/html/2604.22661).
- **Closest work on feedback for RAG** concerns pseudo and generative relevance feedback, not citation labels *(snippet)* — [Test-time Corpus Feedback, EACL Findings 2026](https://aclanthology.org/2026.findings-eacl.298.pdf).

### Inferences
- **Signal semantics.** These are my proposals and have no published validation.

| Signal | Likely meaning | Main bias | Suggested use |
|---|---|---|---|
| Span cited or quoted in the answer | Used, not necessarily needed | Position, salience, verbosity, model's habit of citing | Weak positive, position-corrected |
| Selected but uncited | Weak negative only | Useful background is rarely cited | Low weight or ignore at first |
| Recall returns span that was left out, then it is used | **Strong miss label**, the best available signal for unexposed items | Only observed when the agent chooses to recall | Pairwise: recalled span above the lowest-value selected spans of similar cost |
| Recall repeats content already in context | Presentation or findability problem | Agent behavior | Diagnose, do not train on it |
| User correction after an answer | Possible staleness or missed evidence | Sparse, hard to attribute | Evaluation set first |

- Recall requests partially solve the exposure problem, since they surface items the policy did not show. They are still selected by agent behavior, so treat them as biased positives.
- **Determinism and exploration conflict.** Propensity methods need variation in exposure. Options that keep replay exact: seed any exploration from a hash of the query and policy version, restrict it to a small reserved slice of budget, and log the propensity with every decision. Natural variation across policy versions can also supply overlap. I did not verify a specific estimator for this design.
- **Learning order.** First, per-query score normalization, or the MEMTIER failure mode can recur. Second, a handful of global weights: lane weights, recency, redundancy penalty, cost exponent. Third, a pairwise model on marginal-gain features, in the LMVU style. Keep the weights versioned, human-readable, and promoted only when they beat the current policy on a frozen replay set, stratified by question type.
- Evaluate on answer-level or nugget outcomes where possible, because retrieval metrics can pick the wrong policy.
- Volume matters: the two-stage CLTR evidence uses interaction counts a single-user tool will not reach, which favors few parameters and strong priors.

### Gaps
- I found no paper treating LLM citation behavior as relevance feedback for retrievers, and no study measuring how citation correlates with counterfactual usefulness. A citation-removal test on the local corpus would answer this directly: remove a cited span, regenerate, and check whether the answer changes.
- No published propensity model for position in an LLM context window as applied to citations.
- No offline evaluation protocol designed specifically for token-budgeted selection from agent logs.
