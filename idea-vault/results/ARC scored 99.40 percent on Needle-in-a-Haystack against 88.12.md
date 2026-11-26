---
type: result
status: preprint
area: recall
needs: none
measured-by: others
---
ARC, Addressable Recall Compaction (Dang et al., July 2026), keeps an append-only tool log addressed by id, with citations in context. Its abstract reports 99.40 percent on Needle-in-a-Haystack against 88.12 for the best baseline, and 29.97 against 28.25 on LongBench-v2 Hard. The tests use small open models and are not coding-agent benchmarks, and the research notes read only the abstract figures.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Needle-in-a-Haystack 99.40% against 88.12% for the best baseline; LongBench-v2 Hard 29.97% against 28.25%"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "The tests use Qwen3-8B and Qwen3-32B and are not coding-agent benchmarks."

## Relations
- supports:: [[Recall by id or by search replaces memory]]
- supports:: [[A record goes in whole or is listed by id]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- source:: [[ARC Addressable Recall Compaction]]
