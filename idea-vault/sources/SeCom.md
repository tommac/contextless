---
type: source
status: external
area: structure
needs: none
year: 2025
url: https://arxiv.org/html/2502.05589
---
SeCom, from ICLR 2025, builds memory from topical segments of a conversation instead of single turns or whole sessions, and found that full history scored below focused retrieval. The quality of the segmenter mattered, and the best unit size depended on the retriever. It is the report's peer-reviewed support for small verbatim units, measured on older readers.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Segmenter quality mattered: on LoCoMo with MPNet, GPT-4 segments scored 69.33 and RoBERTa 61.84."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "SeCom showed granularity interacts with retriever: on LoCoMo, BM25 scored 65.58 for turns and 63.16 for sessions, while MPNet scored 57.99 and 51.18"

## Relations
- supports:: [[Threads group turns by topic]]
- supports:: [[Unrelated context can hurt answers]]
- supports:: [[Atoms are paragraph-sized addresses into a turn]]
- tested-by:: [[Segment retrieval scored 88.81 at 820 tokens against 63.85 for full history]]
