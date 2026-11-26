---
type: result
status: preprint
area: storage
needs: none
measured-by: others
---
On MemOps, a July 2026 preprint with 2,006 question-answer pairs over remember, forget, update and reflect operations, session-level BM25 retrieval scored 0.845. Long-context models at about 61K tokens scored about 0.79 to 0.80, MemOS 0.785 and Mem0 0.543. Simple retrieval over stored sessions beat both full context and the dedicated memory products on this benchmark.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Session-level BM25 RAG 0.845, MemOS 0.785, Mem0 0.543"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Long-context accuracy at about 61K tokens: GPT-4.1-mini 0.798, GPT-4o 0.794, Claude Sonnet 4.5 0.790."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "the benchmarks built since 2025 mostly show full context or simple retrieval matching or beating dedicated memory products"

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[The offline floor is the baseline every selector must beat]]
- supports:: [[Unrelated context can hurt answers]]
- supports:: [[Full text alone matched all lanes on real probes]]
