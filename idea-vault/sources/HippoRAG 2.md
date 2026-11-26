---
type: source
status: external
area: structure
needs: none
year: 2025
url: https://arxiv.org/html/2502.14802
---
HippoRAG 2 retrieves over a graph of phrase nodes, passage nodes and synonym edges with Personalized PageRank, and a model filters the retrieved triples. It averaged a little above a dense baseline, with most of the gain on one dataset, and it tied BM25 overall on MemoryAgentBench. A peer-reviewed oracle probe found its stored form could answer only about 10 percent of sampled questions on one benchmark, which shows that structure built at ingest can destroy evidence.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "phrase nodes, passage nodes, synonym edges and Personalized PageRank, with an LLM filtering retrieved triples"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "The biggest gain was 2Wiki (71.0 vs 61.5); others were small, such as HotpotQA 75.5 vs 75.3 and NarrativeQA 25.9 vs 25.7"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "HippoRAG 2's stored form could answer about 10% of sampled AMABench questions in an oracle probe"

## Relations
- part-of:: [[Links help when questions need several pieces of evidence]]
- alternative-to:: [[EdgeMem]]
- tested-by:: [[HippoRAG 2 averaged F1 59.8 against 57.0 for dense retrieval]]
- tested-by:: [[HippoRAG 2 stored form could answer about 10 percent of AMABench questions]]
- tested-by:: [[BM25 tied HippoRAG-v2 overall and beat GraphRAG and Zep on retrieval]]
