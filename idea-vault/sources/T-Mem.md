---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/html/2610.12874
---
T-Mem builds a forest over the original messages at query time, with Update, Expand and Root edges, and never changes the stored conversations. On GroupMemBench it beat BM25 with expansion by about 14 points at roughly 5.6 times the tokens, and its edge labels come from a model call per query. It shows that a graph can be an ordering over verbatim turns and need not replace them.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "builds a query-time forest over original messages with Update, Expand and Root edges, without changing stored conversations"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Arranging the same BM25 messages as a forest rather than a flat list added 3.09 to 7.60 points. Edge labels come from per-query LLM calls"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "EdgeMem and T-Mem together suggest a graph need not replace text."

## Relations
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- alternative-to:: [[EdgeMem]]
- tested-by:: [[T-Mem beat BM25 with expansion by 13.83 points at 5.6 times the tokens]]
