---
type: idea
status: built
area: storage
needs: none
---
Across independent and multi-system studies, raw text searched with BM25, dense retrieval or file tools repeatedly matches or beats memory that a model has rewritten. The pattern the notes draw is that keeping the information comes first and structure second, because systems that discard wording, speaker or thread context lose to raw retrieval. This is the strongest outside support for the verbatim store ContextlessAI is built on, and it is not evidence about this system itself.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "raw text with BM25, dense retrieval, or file tools repeatedly matches or beats LLM-transformed memories"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The most consistent pattern is **information preservation first, structure second**."
- `reports/Cutting edge context memory research.md`: "LLM-rewritten memory, token-level compression and default knowledge graphs repeatedly tie or lose to BM25 over raw text"

## Relations
- supports:: [[Turns are stored verbatim and never rewritten]]
- contradicts:: [[LLM summaries as memory]]
- tested-by:: [[Verbatim files scored 78.1 against 37.5 for curated memory on PersonaMem]]
- tested-by:: [[BM25 at 43.22 beat Mem0 and GraphRAG on GroupMemBench]]
- tested-by:: [[Compaction damage ranged from 5 to 64 percent across 14 summarizers]]
- tested-by:: [[Mem0 reports full context above its own memory on LoCoMo]]
- source:: [[GroupMemBench]]
