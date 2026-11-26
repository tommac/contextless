---
type: idea
status: rejected
area: structure
needs: none
---
Using a knowledge graph of extracted triples as the default memory store is one of the ten popular ideas the report rejects for this design. On accurate retrieval BM25 scored well above GraphRAG and Zep, and stores that keep only triples lose information the source text held. The decision is to extract no triples and to allow links only as an index over atoms.

## Evidence
- `reports/Cutting edge context memory research.md`: "On accurate retrieval, BM25 scored 60.5 against 40.9 for GraphRAG and 37.5 for Zep"
- `reports/Cutting edge context memory research.md`: "No extracted triples. Links only as an index over atoms"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Graph gains are clearest on multi-hop and temporal questions, while triple-only stores lose information; the best-supported designs keep source text linked to graph elements."

## Relations
- builds-on:: [[Raw text beats rewritten memory]]
- alternative-to:: [[Links help when questions need several pieces of evidence]]
- supports:: [[The offline floor is the baseline every selector must beat]]
- tested-by:: [[BM25 tied HippoRAG-v2 overall and beat GraphRAG and Zep on retrieval]]
- tested-by:: [[BM25 at 43.22 beat Mem0 and GraphRAG on GroupMemBench]]
- tested-by:: [[HippoRAG 2 stored form could answer about 10 percent of AMABench questions]]
- source:: [[MemoryAgentBench]]
