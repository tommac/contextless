# Exact source reads from a large archive

This experiment separates the cost of storing originals from the cost of locating relevant evidence. It compares the current Go vault, a file per binary record, SQLite BLOBs, and packs of up to 1,024 records with a 12-byte offset-and-length entry for each record. A separate public-data experiment compares independently decodable lossless blocks. The programs use disposable directories and do not open an existing user vault.

## What the production code costs

At revision `04e171c`, a synthetic vault with 100 preseeded turns took 0.29 ms to append one new turn. At 1,000, 10,000 and 30,000 turns this increased to 0.69, 6.28 and 20.44 ms. Each value is the median of seven appends through `vault.AppendTurnAt`. This revision enumerates and sorts the directory to find its highest turn number. Loading and checking all 30,007 turns took 1.30 s. On that fixture, Git's first commit took 30.49 s and its subsequent single-turn commit took 168 ms. Those two Git figures each represent one observation. The synthetic messages are short, and seed generation bypasses the sequence lookup. The production encoder trims trailing whitespace before calculating the stored SHA. [Go runner](run_vault.py) · [Overlay source](physical_study.go.txt) · [Original run](production.log)

## Layouts and process recovery

Each synthetic layout stores the same original UTF-8 bytes prefixed with their SHA-256. The loose-file control already has a committed head, giving it constant-time sequence allocation. The Python files do not contain YAML, cards or a Git history. The SQLite control uses an integer primary key, WAL, synchronous=FULL, and `substr(BLOB, offset, size)`. The segment reader uses at most 16 open segment handles and reads the address table with positional I/O. All three seed in batches of 512. Each measured append opens a store, commits a group of 32 records with the configured sync operations and closes it, including SQLite's normal close behavior. The filesystem cache stays populated. Random reads compare 2,000 96-byte slices with the original bytes. Sequential replay checks every full frame digest and the order of those digests.

| 12,000 preseeded records | Open + append 32 + close, ms | 2,000 span reads, ms | Verified replay of 12,256 records, ms | Allocated MB after recovery |
|---|---:|---:|---:|---:|
| Loose files | 6.42 | 428.16 | 2,253.17 | 52.72 |
| SQLite | 5.92 | 77.18 | 151.26 | 19.61 |
| Segments | 0.92 | 4.08 | 80.68 | 17.05 |

The table uses medians over three rotating-order trials. Each trial's append statistic is itself the median of eight batches. Individual timings varied substantially on this shared machine. The raw trial results preserve that variation. [12,000-record measurements](12000.jsonl)

| One million preseeded records, two trials | Open + append 32 + close, ms | 2,000 span reads, ms | Verified replay of 1,000,256 records, seconds | Allocated GB after recovery |
|---|---:|---:|---:|---:|
| SQLite | 3.65–5.68 | 634–759 | 9.98–11.19 | 1.608–1.611 |
| Segments | 0.55–0.73 | 429–606 | 6.22–7.47 | 1.390 |

The million synthetic records contained 1,343,601,528 source bytes. At every size a child process died after staging 1,025 records, just before committing them. Reopening recovered the previous head, accepted three different records and reproduced the entire expected sequence. This exercises one process-exit boundary with one writer. Torn device writes, arbitrary kill points, simultaneous writers, external edits and power failures require separate tests. The synthetic short-span path uses a reference oracle during measurement. A single arbitrary slice cannot be authenticated by a whole-record SHA without more bytes. [Million-record measurements](million.jsonl) · [Synthetic runner](bench.py)

## Lossless compression of actual dialogue

All 500 histories from pinned LongMemEval-S contain 246,750 message occurrences and 244,779,063 UTF-8 bytes. We concatenated the bodies, cut the stream into fixed-size blocks, and stored each block raw or as an independent zlib level-1 stream. The layout includes the message offset table, encoded block offsets and lengths, block SHA-256 digests, a header, and a 32-byte index digest. A reader holding the correct index digest can validate it once, then authenticate whichever blocks it opens. A reader that gets both the archive and digest from the same untrusted source still needs a trusted snapshot identity.

| Block size | Raw archive MB | Compressed archive MB | Reduction vs matching raw control | MB fetched / decoded for 2,000 compressed span reads |
|---|---:|---:|---:|---:|
| 4 KiB | 250.37 | 113.26 | 54.8% | 3.70 / 8.40 |
| 16 KiB | 248.40 | 100.50 | 59.5% | 13.13 / 33.01 |
| 64 KiB | 247.90 | 95.51 | 61.5% | 49.51 / 131.14 |

These numbers keep every occurrence separately. Sizes are exact logical archive file lengths, summarized in decimal MB. Each row served the same 192,000 requested bytes and validated them against the public source. The 4 KiB layout decoded about 44 bytes per byte requested, the 64 KiB layout about 683. The larger blocks saved an additional 17.75 MB relative to the smaller blocks. Two random-read passes took 28–30, 88–104 and 268–1,115 ms respectively. Reads followed full verification with a warm OS cache, no decompressed block cache and an already validated, memory-mapped index. Hardware time is preliminary. The row address, block count, decoded byte and storage figures can be reproduced independently of hardware.

Across all benchmark histories, 189,515 bodies are distinct by byte equality. A separate deduplicated arm keeps an occurrence-to-body table, preserving order and multiplicity even when identical bodies repeat. At 4 KiB, that compressed archive occupies 87.66 MB. Reused benchmark material can exaggerate the savings available on a single real workspace. These figures cover message bodies and the addresses needed to find them by ordinal. Roles, session dates, session boundaries, full-text indexes and Git history would consume additional space in a workspace.

All 12 configurations restored the complete ordered stored-body stream, verified every body boundary and occurrence map, then returned the exact source bytes for every sampled span, including probes crossing blocks. A small block size is a viable choice for scattered citations. Benchmarks with real query traces should test whether grouping requested spans by block, caching decoded blocks or isolating hot small records outweighs the additional metadata. This experiment measures storage and read costs. Answer finding, ranking, and model accuracy need their own evaluation. [First public run](public.jsonl) · [Independent repeat](public-repeat.jsonl) · [Public runner](blockpack.py) · [Pinned public dataset](https://huggingface.co/datasets/xiaowu0162/longmemeval-cleaned/tree/98d7416c24c778c2fee6e6f3006e7a073259d48f)

## Reproduce

Run from the repository root, using this experiment's location as `study`. The Python synthetic runner uses POSIX pread and fork, Python 3.9+, and about 2 GB of temporary space for the million-record comparison. The public run accepts only the dataset with SHA-256 `d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442`. Recorded environment was Darwin ARM64, Go 1.26.0, Python 3.9.6 and SQLite 3.51.0.

```sh
study=docs/experiments/source_layout
python3 "$study/run_vault.py"
python3 "$study/bench.py" --records 12000 --trials 3
python3 "$study/bench.py" --records 1000000 --layouts sqlite segment --trials 2
python3 "$study/blockpack.py" /path/to/longmemeval_s_cleaned.json
```
