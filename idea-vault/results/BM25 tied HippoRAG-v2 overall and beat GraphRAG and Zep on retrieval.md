---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
On MemoryAgentBench, a third-party benchmark whose venue is not confirmed, BM25 scored 41.5 overall against 41.6 for HippoRAG-v2, with RAPTOR at 27.0 and GraphRAG at 23.4. On the accurate-retrieval tasks BM25 scored 60.5, above GraphRAG, Zep, RAPTOR and Mem0, though HippoRAG-v2 reached 65.1. A long-context GPT-5-mini scored 60.6 overall, above every memory system listed.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "On accurate retrieval, BM25 60.5 beats GraphRAG 40.9, Zep 37.5, RAPTOR 36.8 and Mem0 32.6, while HippoRAG-v2 reaches 65.1"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "overall BM25 41.5 vs HippoRAG-v2 41.6, MemoRAG 30.9, RAPTOR 27.0, GraphRAG 23.4"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "**60.6** (Mem0 21.1, Zep 24.0, best RAG 41.6, BM25 41.5)"

## Relations
- supports:: [[Knowledge graphs lose to BM25 as a default store]]
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[The offline floor is the baseline every selector must beat]]
- source:: [[MemoryAgentBench]]
