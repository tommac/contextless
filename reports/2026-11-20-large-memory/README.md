# How an agent could work across a growing archive with a fixed context window

The [fixed-window follow-up](fixed-window-results.md) measures which answer words survive in public conversations, tests seekable lossless storage and reports two optimizations to the Go CLI.

**The design I would pursue is a reader that can resume a query against original records.** The engine keeps the text, the search indexes, intermediate result sets, and a record of what it has already checked. The model sees its question, a small page of evidence, and handles for further work. After new data arrives it can ask “what changed since I checked?” An answer that requires every match, a latest value, or a defensible zero gets an explicit scope. A loosely worded question can return useful leads with unresolved scope.

On November 20, four offline experiments exposed different constraints. The strongest immediate implementation is a paged omission catalog. The most distinctive research direction is carrying a verified query and its evidence forward while the archive changes.

## 1. Move the catalog outside the model window

The existing `cx packet` path reads the turn collection, constructs the candidate set, and describes omitted turns in groups of 50. At 512, 2,048 and 8,192 synthetic turns we measured warm median CLI latency of **64 ms, 235 ms and 1,074 ms** across four fresh processes per size. The estimated brief size actually fell from 14,180 to 9,909 tokens, and each packet still held the planted fact. In a separate measurement, the production renderer needed **880,001 bytes and 264,001 estimated tokens** to name one million contiguous omitted turns without keywords. That quantity is the omitted catalog alone. [Benchmark](../../docs/experiments/packet_scale/main.go) · [Packet selection](../../internal/packet/select.go) · [Digest renderer](../../internal/packet/brief.go)

The prototype divides a snapshot into pages of 4,096 turn IDs. An empty or full page uses a marker. A sparse page stores positions to include or exclude, and a fragmented page can use a bitmap. The packet carries the snapshot, omitted count, and a digest identifying the external directory.

| Omitted pattern at one million turns | Handle in packet | External directory and membership payload |
|---|---:|---:|
| Every turn | 217 bytes | 8,575 bytes |
| Every 1,031st turn | 213 bytes | 10,513 bytes |
| Every other turn | 216 bytes | 133,575 bytes |

Paged enumeration reproduced the exact omitted IDs in all three cases. Once the directory was open, a point lookup needed at most one explicit payload read on these cases. Building this experimental directory still examines the membership set. A handle also lacks the descriptive words of a catalog line. A live reader needs both `search(snapshot, words)` for discovery and `open(snapshot, id)` for exact access. Measure whether the reader finds previously unnamed evidence and count the cost of those extra calls. [Prototype](../../docs/experiments/packet_scale/catalog.go)

**Production sequence.** Keep Markdown as the original source. Maintain a versioned table of turn addresses, byte lengths, segmentation versions and body hashes at ingest. Use that table to look up the latest turn and choose candidates. Load bodies after selection. Pin a snapshot for each packet and every following recall. If a database lookup fails or an index is behind that snapshot, return the error or its usable range. Today the search path can turn a SQL error into an empty hit list. This matters when a user asks whether something ever happened. [Current next-turn lookup](../../internal/engine/probe.go) · [Value and phrase selection](../../internal/packet/atoms.go) · [SQLite search](../../internal/store/store.go)

## 2. Resume questions whose answers can actually be checked

Consider “Which tickets in project 7 were still blocked at time T?” One source quote can confirm that a ticket was blocked. Answering the full question also requires cancellations, reopened tickets, late-arriving corrections and a rule for events with equal effective times.

The executable prototype stores every event's original bytes. Each append transaction seals a new revision. A query record binds its operator, project, effective-time cutoff and parser version to a revision and its source digest. On resume, an indexed range finds newly recorded events in scope. The engine updates the winning assertion for each affected item and retains earlier versions. The reader gets a count, two example records, and a cursor tied to that particular query and revision. All examples and subsequent pages resolve back to the source bytes. [Implementation](../../docs/experiments/coverage/receipts.py)

| Test on a stream starting with 1,000,000 events | New typed candidates examined for the query | Exact count returned | First result page |
|---|---:|---:|---:|
| First query | 15,625 | 1,301 | 473 bytes |
| Append 10,000 events for another project | 0 | 1,301 | 473 bytes |
| Append five events including corrections and a future-effective update | 5 | 1,302 | 473 bytes |
| Append 10,000 events concentrated on 64 in-scope items | 10,000 | 1,293 | 476 bytes |
| Append one uninterpretable record of unknown scope | 0, plus one unknown | Withheld | 476 bytes |

Every positive item ID and its winning source record matched a full raw-event scan at every stage. After closing and reopening the database, the first snapshot still returned its original witnesses. Hand-labeled fixtures checked a late correction, an equal-time tie, an unknown integer encoding and a cursor used against the wrong snapshot. The cursor binding is included in the byte figures. They exclude model instructions, the original question and framing by a provider.

The final archive held 49.9 MB of original event bytes. In this laptop run, writing and indexing the first million took 4.95 s. Maintaining the query after the unrelated append took 0.80 ms after ingestion. A conventional incremental SQL control took 0.15 ms on the same delta. On the concentrated append those figures were 14.9 ms and 2.7 ms. SQL is already an excellent execution engine. The added capability here is a reader retaining its historical evidence, resuming at the correct revision and being told when its exact answer has become unknowable. Times are single-run observations, and candidate counts measure logical records rather than physical disk reads. [Million-event measurements](million-event-run.json) · [Smaller independent scale run](65536-event-run.json) · [Runner and SQL controls](../../docs/experiments/coverage/bench.py)

This prototype recognizes a narrow, deterministic event grammar and trusts its SQLite store. A source digest identifies the input snapshot; it cannot attest by itself that a parser understood a sentence or an indexer indexed it correctly. An unknown-scope row conservatively withholds an exact count, even if a person would later classify it as irrelevant. The same contract can serve typed tool events, build outcomes, explicitly recorded decisions and domain records. Free-form conversation needs a separate evidence-discovery step. For an arbitrary natural-language predicate, a single decisive exception can occur anywhere; a system may have to inspect the entire corpus. Fresh bounded model calls can keep each context small during such a scan while total computation still grows.

## 3. Index the work the ranker repeats

The value lane currently searches the neighborhood of every date, name, number or identifier at query time. The prototype builds postings from a normalized term to the *occurrence* of a value within four tokens in the same sentence. This preserves the difference between three terms surrounding the same value and three terms scattered across unrelated values. At query time the engine visits the relevant occurrence lists and computes the original score.

Across **568,887 query–atom pairs on all 500 public LongMemEval-S questions**, the indexed and production scores agreed on every pair. Scanning took 6.283 s. Building indexes for all 500 histories took 8.759 s and the subsequent lookups took 15.2 ms in total. A one-query history therefore lost about 40% overall on this component. Reuse changes the economics. On the 50,000-atom synthetic workload the compressed index added about 70% of the raw text size and the common-word query still visited 200,000 postings. Benchmark repeated queries, append cost, disk footprint, resident memory and **complete packet latency** before enabling this in a live workspace. [Algorithm, controls and reproduction](../../docs/experiments/valuepostings/README.md)

A production index could append postings in immutable segments, compress sorted integer deltas, and rebuild only segments whose tokenizer changes. Block-level upper bounds could stop a search once the remaining blocks cannot beat the current score, provided the bound also preserves ties. Fusing truncated lists adds a second correctness problem. Recent work on exact adaptive RRF shows that anti-correlated lists can require full traversal. [Exact Adaptive Hybrid Retrieval, August 2026](https://arxiv.org/abs/2608.07152)

## 4. Separate finding candidate addresses from buying words in the packet

On the pinned public benchmark, 343/470 answerable questions had text from every annotated message in the original 8,000-byte packet. Of the 127 misses, **108 lacked a required message anywhere in the top 24**. Both rankers missed the same 108, even after taking their union. The other 19 already contained a feasible solution. Giving each source message one chance before taking a second atom from it reached **355/470** at the same budget, with 15 wins and three losses. This is an exploratory result on an already examined dataset. The label identifies the source message; a selected paragraph can still omit the words needed to answer. [Diagnostic and per-category results](candidate-ceiling.json) · [Runnable comparison](../../docs/experiments/longmem/candidate_ceiling.py)

A second search route found evidence the first could not. It searched whole sessions, searched messages in the three highest-ranking sessions and interleaved their addresses with the original atom candidates. That made 20 previously unreachable questions reachable and lost two formerly reachable ones. Actual complete packets **fell from 343 to 338**. Only five of the previously unreachable questions became complete within the final byte budget. Using a whole message in a small packet can displace the passage that made the original ranking work. [Session-routing results](session-routing.json) · [Routing code](../../docs/experiments/longmem/session_routing.py)

**Next experiment.** Keep the alternate route's addresses on disk. Give the first reader a source-diverse packet. When the reader identifies a missing prerequisite, such as the effective date or the later reversal of a decision, allow one targeted second read from that address set. Fetch an exact sentence and its neighboring sentences, with an option to open the full original paragraph. Compare against more plain BM25 results, unconditional adjacent-turn expansion and a budget-matched second BM25 call. Fix total input bytes, call allowance and judge first. Measure regressions on previously correct answers as well as recovered misses. A cheap second route is valuable only when its evidence survives packing and changes the answer correctly.

## Where recent research draws the line

| Prior work | What it already shows | Consequence for this project |
|---|---|---|
| [GrepMem, Nov 18 preprint](https://arxiv.org/abs/2611.25123) | A trained 9B reader with an append-only log and regex recall scored 95–97% on ContextBench. The tasks and reader training matter. | Include an ordinary log with grep as a competitive baseline. |
| [Folio, Nov 12 preprint](https://arxiv.org/abs/2611.16627) | Bounded reading with source handles and hierarchical recovery solved 66/75 LOCA-bench tasks versus 54/75 for compaction in the same runtime. | Addressable external memory has close prior art. Measure the additional value of a resumable, scoped query. |
| [LEDGER, Nov 16 preprint](https://arxiv.org/abs/2611.20927) and [Read Boundaries in Agent Memory, Nov 14 preprint](https://arxiv.org/abs/2611.19264) | These evaluate changing facts and the risk of combining evidence read at different revisions. | Include historical snapshots, correction order and concurrent recall in any live test. |
| [TallyMem, October revision](https://arxiv.org/abs/2610.19474v2) | Generated event specifications and deterministic aggregation already operate over a source-preserving trajectory. | Test specification errors independently of execution and compare against fresh recomputation. |
| [Evergreen, July revision](https://arxiv.org/abs/2604.26180v2), [SemViews, October preprint](https://arxiv.org/abs/2610.12768), and [DBToaster](https://arxiv.org/abs/1207.0137) | Prior systems verify semantic aggregates, reuse labels with statistical guarantees, and incrementally maintain database queries. | The research claim must survive a conventional incremental SQL baseline and distinguish parser correctness from index completeness. |
| [QO-Bench, Findings of EMNLP 2026](https://aclanthology.org/2026.findings-emnlp.1653/) and [AGGBench, EMNLP 2026](https://aclanthology.org/2026.emnlp-main.1026/) | These cover operator-sensitive questions and exhaustive evidence collection over public text. | Test intersection, counting and distributed evidence on unfamiliar corpora alongside conversation QA. |

## The public result worth pursuing

Publish a replay in which the reader asks a question, opens exact evidence, resumes after an unrelated append, revises after a backdated correction and declines an unsupported completeness claim. A second reader should be able to reopen the first revision and see exactly why the earlier answer was right at that time.

Three integration gates would justify investing further. These are proposed targets for new experiments.

1. **Scale.** Run the real packet-and-recall path on 1M and then 10M addressable records, with a fixed reader window. Publish p50/p95 end-to-end latency, logical records touched, physical bytes read, index size and ingest throughput, for cold and warm stores. Include a common-term workload and a corpus with many small threads. Require exact recall of every requested source span and reject mismatched snapshots.
2. **Correctness under change.** Compare full recomputation, a maintained SQL view, grep over raw history, and the resumable reader. Inject late corrections, changed predicates, malformed records and index lag. A false claim of completeness on any supported deterministic fixture fails the test. For natural-language labels, independently adjudicate a sample and publish extraction errors and unresolved questions.
3. **Reader utility.** Freeze the policy on development material, then test on unseen conversations and one external operator benchmark. Match the reader, judge and total context allowance across arms; report answer correctness, contradiction, abstention, failures, cost per correct answer and extra reads. The November 25 private protocol remains a separate frozen experiment.

All experiments in this report used public data or synthetic fixtures and ran without model calls. The catalog, postings and resumable-query prototypes live outside the production path. Results establish specific scaling and correctness properties under the stated workloads; a public end-to-end agent result is still ahead.

## Reproduction

Run from the repository root. The first three commands generate their own synthetic material and clean up their temporary stores.

```sh
go run ./docs/experiments/packet_scale -catalog
go test ./docs/experiments/packet_scale
python3 docs/experiments/coverage/bench.py --records 1000000
```

For the public-data experiments, follow the [pinned dataset and rank-generation instructions](../../docs/experiments/longmem/README.md) first. Using the same shell variables from those instructions:

```sh
python3 docs/experiments/longmem/candidate_ceiling.py "$data" "$work/ranks.jsonl" --by-type --check-cases reports/2026-11-20-experiments/public-cases.json
python3 docs/experiments/longmem/session_routing.py "$data" "$work/ranks.jsonl"
python3 docs/experiments/valuepostings/run.py --public-data "$data"
```
