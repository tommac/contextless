---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f
---
Karpathy's gist of April 2026 describes immutable raw sources, a linked Markdown wiki kept up by a model, and a schema file, with ingest, query and lint operations. It contains no measurements, and its commenters report dropped facts, near-duplicate pages and stale values. The notes read the pattern as a precomputed synthesis cache that is useful for repeated cross-source questions and risky for exact values.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "It defines immutable raw sources, an LLM-maintained linked Markdown wiki, and a schema file, with ingest, query and lint operations"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "The LLM wiki behaves like a precomputed synthesis cache: useful for repeated cross-source questions, risky for exact values, and not a replacement for raw sources."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "It claims the index approach works at about 100 sources and hundreds of pages."

## Relations
- part-of:: [[Model-written notes lose as storage]]
- supports:: [[Catalog cards are an index not a replacement]]
- alternative-to:: [[A-MEM]]
- analog-of:: [[Zettelkasten]]
- tested-by:: [[A compiled wiki used about 21 times the tokens per query of RAG]]
