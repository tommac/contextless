---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/html/2609.05553v1
---
EdgeMem keeps verbatim turns as vertices and adds month, entity or keyword, and session hubs as hyperedges, all built with no model. On LoCoMo under a strict judge it lifted flat BM25 by about 11 points with zero build tokens. The report calls it the most relevant structural result for this project, and also close prior art for model-free hubs and verbatim addressable spans.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "verbatim turns are vertices; month, entity or keyword, and session hubs are hyperedges"
- `reports/Cutting edge context memory research.md`: "The most relevant structural result, and close prior art"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "LoCoMo: 0 build tokens, 1.8 min build; retrieval P95 10.73 ms."

## Relations
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Model-free hubs over verbatim turns]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- alternative-to:: [[HippoRAG 2]]
- tested-by:: [[EdgeMem lifted flat BM25 from 49.74 to 61.01 with no LLM at build]]
- tested-by:: [[One system scored 61.01 or 77.86 depending on the judge]]
