---
type: source
status: external
area: storage
needs: none
year: 2025
url: https://arxiv.org/html/2504.19413
---
Mem0 extracts memories with a model that chooses to add, update, delete or do nothing, so a contradicted memory can be removed; its graph variant Mem0g marks obsolete relations invalid instead. Its own LoCoMo table puts full context above both variants on quality while using far fewer tokens. Independent benchmarks place it below plain BM25, and it is the clearest example of the fact replacement this project avoids.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "base Mem0 uses LLM-selected ADD, UPDATE, DELETE and NOOP, so contradicted memories can be removed"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "full context 72.90, Mem0g 68.44, Mem0 66.88, Zep 65.99, best RAG 60.97"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "It can serve as an index or ordering over verbatim turns, which fits this engine far better than Mem0-style fact replacement."

## Relations
- contradicts:: [[Turns are stored verbatim and never rewritten]]
- alternative-to:: [[Zep and Graphiti]]
- tested-by:: [[Mem0 reports full context above its own memory on LoCoMo]]
- tested-by:: [[BM25 at 43.22 beat Mem0 and GraphRAG on GroupMemBench]]
- tested-by:: [[Memory frameworks scored 5 to 18 percent on STALE]]
- tested-by:: [[Mem0 and Zep LoCoMo scores were disputed and never reconciled]]
