## Reproduce the ContextBench log-triage replay

Fetch the public released tasks at commit [fb533c1](https://github.com/RulinShao/ContextBench/tree/fb533c18c78a82d3f1b2047a9c1700735cf829ae). From the directory holding `replay.py`, run with Python 3.8 or newer.

```sh
set -e
work=$(mktemp -d)
base="https://raw.githubusercontent.com/RulinShao/ContextBench/fb533c18c78a82d3f1b2047a9c1700735cf829ae/tasks"
for seed in 42 43 44 45 46 47 48 49; do
    task="$base/log-triage-L12800-seed$seed"
    curl -fsSL --retry 2 "$task/environment/stream/chunks.jsonl" -o "$work/stream-$seed.jsonl"
    curl -fsSL --retry 2 "$task/tests/gold.json" -o "$work/gold-$seed.json"
done
python3 replay.py "$work"
```

Both controls operate on exactly the supplied dated log lines and batch delimiters from the stream, concatenated in order. All log batches precede the 24 questions on each seed. The indexer knows the two task schemas and builds counters and offsets without looking at the questions. Gold labels are opened only after both solvers have emitted their answers. All sixteen local input files are checked against the Git blob IDs from the pinned release. Original-byte counts include delimiters. “Source reaccess” counts bytes in answer slices after the full scan to construct the index; the scan arm's logical span includes full passes for counting and start-to-match spans for lookups. Neither tracks physical reads, RAM, model tokens, or the official Harbor streaming agent protocol. Median elapsed seconds cover five end-to-end in-memory trials per arm across all eight fixtures, with index construction inside the timed loop and JSON decoding outside it.
