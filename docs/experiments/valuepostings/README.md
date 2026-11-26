# Exact value-neighborhood postings

The value lane in ContextlessAI scores an atom by the greatest number of distinct query terms surrounding any qualifying value occurrence. The production scan repeats tokenization and the value test for every query. This experiment indexes the same relation once. The tests call the actual tokenizer, stemmer, value detector and exhaustive scorer from the packet package using a local Go overlay. The index can be discarded and rebuilt from the verbatim text.

For each value occurrence the index stores its atom ordinal and stem. For every distinct surrounding stem within four positions in the same sentence, a posting records that occurrence. At query time, visit the posting list for each query stem, skip occurrences whose own stem appears in the query, and take the largest count for each atom. Each distinct query stem contributes exactly once to any eligible occurrence. This produces the same integer score with repeated terms, repeated values or overlapping windows. Ordering after scoring still depends on the FTS ranking, turn sequence and atom number.

## Run

From the repository root:

```sh
python3 docs/experiments/valuepostings/run.py
```

Use `--sizes 1000,10000,50000` to choose the synthetic corpus sizes. To check every question against all atoms in its own public history, add `--public-data /path/to/longmemeval_s_cleaned.json`. The runner verifies the public file's pinned digest before reading it. [Download instructions for the public dataset](../longmem/README.md)

## Results for November 20, 2026

The existing SQLite-assisted production ranking and the experimental score map agreed on the fixture's score ties, historical turn-sequence cutoffs, and ranked-list depths. A seeded test exercised 80 queries on 200 punctuation- and Unicode-heavy bodies. Every score also matched on the 500 public LongMemEval-S questions, including all 26,968 positive scores among 568,887 query–atom pairs. These measurements establish parity with the existing feature on these inputs. They do not measure whether a value extracted from a sentence answers its user's question.

On the public run, scoring 500 existing histories took **6.283 s** by exhaustive scan. Building a fresh index for every history took **8.759 s**. Once built, the 500 lookups took **15.2 ms** in total and visited 69,340 postings. With one query per fresh history the build plus lookup took **8.774 s**, about 40% longer than scanning. The aggregate costs cross at about 1.4 comparable queries per built history, assuming a similar mix and full reuse. This expectation needs checking on repeated real queries. Times exclude loading the JSON dataset and splitting it into atoms.

| Synthetic lookup, 50,000 atoms | Matching atoms | Posting entries visited | Exhaustive scan | Indexed lookup |
|---|---:|---:|---:|---:|
| One serial identifier | 1 | 1 | 121.50 ms | 0.00379 ms |
| Station identifier | 100 | 100 | 96.67 ms | 0.03154 ms |
| Common words “worker timeout” | 50,000 | 200,000 | 114.88 ms | 13.61 ms |
| Absent term | 0 | 0 | 97.28 ms | 0.00175 ms |

Query timings are medians of five paired trials with alternating order, on one darwin/arm64 machine. The scorer receives already-loaded individual atom texts or an already-built in-memory index. Synthetic texts share a template and the chosen lookups deliberately span selectivities.

On that corpus the build took 156.4 ms for 200,000 value occurrences and 1,050,000 postings. The raw atom texts contained 4,991,960 bytes. The additional Go gob snapshot occupied 7,482,156 bytes, or approximately 150% of source text size. A gzip-compressed gob snapshot occupied 3,473,780 bytes, or approximately 70% extra. The runner decoded the compressed data and compared the reconstructed structures. Lookup timings exclude encoding, loading and decoding. Gob map order may change the compressed size slightly. Resident heap, append/merge cost, storage of the address table and end-to-end packet latency require separate measurements.

For integration, append posting segments by immutable turn range, retain the active tokenizer/segmentation version, and resolve only selected addresses. Compare byte-identical ranked IDs and complete packet output after appends, backdated records, restart, and a version change. Broad queries still visit many postings; an adaptive scan is worth testing if the relevant bodies are already hot. The production ranker's other lanes and packet-building work can still dominate total latency.
