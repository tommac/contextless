# ContextlessAI experiment results · November 20, 2026

## Public retrieval

The project ranker and plain BM25 each represented every annotated source message on **343 of 470** answerable LongMemEval-S questions. Two questions favored each ranker and 466 tied. At least one annotated message was represented on 423 and 424 questions, respectively. Measured at session granularity, the apparent coverage was much higher: 417 and 416 questions had all annotated sessions represented. Adjacent material from the right session can still miss the answer.

We ran the actual Go atom splitter and ranker over the pinned cleaned release. The comparison used the same FTS index and an 8,000-byte final evidence budget. No reader or judge saw the 500 questions in this run. The 30 abstention cases are in the case manifest and excluded from the source-coverage denominator. All 500 are eligible for the later answer test. [Per-question outcomes](public-cases.json) · [Aggregate measurements](public-retrieval.json) · [Method and reproduction](../../docs/experiments/longmem/README.md)

The matched-passage answer runner now supports verbatim packing, one-time lossy compression on overflow and one overview followed by a single request for originals. Each of these three treatments starts from the same 24 candidates. Plain BM25, full history and the two pinned native competitors have separate arms. Offline tests exercised the budget, routing and resume logic. Live public answer accuracy, native baseline quality and provider costs remain unmeasured. The native SDK integrations still require a provisioned end-to-end check. [Comparison driver](../../docs/experiments/longmem/compare.py) · [Public replay](../../examples/replay/run.py)

## SQLite through the real CLI

Both experimental builds ran the ordinary `cx new`, `cx import` and `cx reindex` against disposable on-disk workspaces. Both used the project's Go SQLite driver with WAL and explicit FULL synchronous commits. The candidate changed only the record search key and its FTS addressing. Each size used four paired trials, alternating which build ran first. The checker reopened each store after each operation, compared record IDs, types and full indexed text, checked the top 30 results for one fixed BM25 query and ran SQLite's integrity check.

| Corpus | Operation | Existing median | Row-ID median | Change in elapsed time |
|---|---|---:|---:|---:|
| Original 511 notes | Import | 1.103 s | 1.001 s | −9.3% |
| Original 511 notes | Reindex | 0.462 s | 0.390 s | −15.6% |
| Four copies, 2,044 turns | Import | 6.376 s | 5.043 s | −20.9% |
| Four copies, 2,044 turns | Reindex | 3.261 s | 1.615 s | −50.5% |
| Repaired 520 notes | Import | 0.964 s | 1.048 s | +8.8% |
| Repaired 520 notes | Reindex | 0.383 s | 0.352 s | −8.0% |
| Four copies, 2,080 turns | Import | 6.059 s | 4.397 s | −27.4% |
| Four copies, 2,080 turns | Reindex | 2.955 s | 1.506 s | −49.1% |

The large-corpus reindex improvement reproduced across both snapshots. Small-corpus import timings changed direction. The four-copy workloads repeat note bodies and one laptop supplied all timings. A production migration needs explicit handling of existing index rows and uninterrupted access to run history. The experimental patch is intended for fresh benchmark databases only. [All 32 trials](sqlite-trials.jsonl) · [Benchmark runner](../../docs/experiments/sqlite-rowid/bench.py) · [Candidate patch](../../docs/experiments/sqlite-rowid/candidate.patch)

To reproduce on the current idea vault, run each command from the repository root:

```sh
python3 docs/experiments/sqlite-rowid/bench.py
python3 docs/experiments/sqlite-rowid/bench.py --copies 4
```

## Obsidian retrieval and evidence integrity

The original vault had 81 references to nine missing note titles and two evidence quotes that no longer matched their source. Added the nine source-backed notes and updated the cache observation to the source's current 254 message gaps, median 16 minutes. The updated vault contains 520 notes. Its checker now finds all 1,410 quotations and resolves all 2,199 relations. This verifies the text and destinations. It does not judge whether a quotation entails a note's interpretation. [Vault checker](../../docs/experiments/vault_graph.py)

Twelve analyst-written known-item questions each named one target result before running the comparison. The lexical index used titles and prose with relations removed. The two graph arms used the same three seeds, edge weight rule, and five-result cutoff. One followed only `tested-by` relations in either direction and one used every relation type. Repeating after the vault repairs gave the same outcomes.

| Method | Named target in top five | Mean reciprocal rank at five |
|---|---:|---:|
| Flat BM25 | 5/12 | 0.176 |
| BM25 plus tested-by links | 4/12 | 0.194 |
| BM25 plus all links | 4/12 | 0.174 |

The typed links moved some targets nearer the top but displaced another. Keep the graph available for inspecting sources. This small known-item test has no independent human relevance judgments and gives no basis for enabling link expansion by default. Run `python3 docs/experiments/link_retrieval.py` to inspect each question and rank. [Queries and scoring](../../docs/experiments/link_retrieval.py) · [Obsidian review views](../../idea-vault/Research%20review.base)

## Frozen private experiment

The existing one-time Claude Desktop task was enabled for **November 25 at 5:15 a.m. Eastern** when inspected. The local preflight checked the presence and nonzero size of all 195 assigned answer packets and 30 gold files, using only their generated identifiers. The protocol calls for 195 answer invocations and 60 judge invocations. It requires the Claude workflow and checks both the weekly all-models usage and the answering model's usage against a 40%-used ceiling. It forbids sending the private packets to an outside service. No private answer or judge run was started in this review. The scheduled run's success and quota headroom are unverified.

## Public release readiness

The README now starts with a model-free replay of a decision that changes twice. It exercises the normal import, packet generation and one-shot recall commands, verifies the excerpts byte-for-byte on the fixture text and shows the consequences of opening an outdated record. The paragraph choices are scripted. A reader model's ability to make those choices remains a benchmark question.

The local history screen covered all nine reachable commits, their branch/tag/remote history and the additional local Codex snapshot refs. The branch/remote/tag history contained 225 distinct blobs. Including local snapshot refs raised this to 751. The scan checked common provider-key/private-key formats, email-like strings, home-directory paths and common local-network ranges. It also checked 526 pre-repair untracked research files. One research note contained a local home path and was represented in a local snapshot ref. The scan found no matches for the credential patterns; semantic privacy review is still needed for project names, proprietary material, illustrative examples and commit metadata. Treat the local snapshots and untracked research as separate from the material selected for release.

All 35 provider Go files have counterparts in the BookWriter checkout. Seventeen currently match byte-for-byte. No tracked license or notice file was found in the BookWriter or MultiAIWorkspace source checkouts inspected. MIT is a reasonable choice for the project-owned implementation once redistribution rights and required third-party notices are established. No repository visibility or licensing change was made.
