---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
On GroupMemBench with GPT-5.5, T-Mem scored 63.09 against 49.26 for BM25 with expansion, at about 5.6 times the tokens. Arranging the same BM25 messages as a forest instead of a flat list added 3.09 to 7.60 points. The edge labels come from a model call per query, so the gain is not free at read time.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "T-Mem beat BM25 with expansion by 13.83 points on GroupMemBench under GPT-5.5 (63.09 vs 49.26), at about 5.6× the tokens"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Arranging the same BM25 messages as a forest rather than a flat list added 3.09 to 7.60 points."

## Relations
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Raw text beats rewritten memory]]
- source:: [[T-Mem]]
