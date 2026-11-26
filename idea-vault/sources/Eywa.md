---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/html/2605.30771
---
Eywa (Joshi, arXiv 29 May 2026) extracts facts with an LLM at write time and answers through an LLM-free read path: a hand-coded planner classifies the shape of the query and sets channel weights over vector search, SQLite FTS5 BM25, a validity-window date filter and entity graph traversal. It keeps SQLite as the authority for evidence, facts and provenance and treats vector indexes and graph checkpoints as projections that can be rebuilt from it. It reports LongMemEval-S 88.2% overall with GPT-4o and publishes an error analysis of its 59 remaining failures. No ablation of the planner is reported, and it is a single-group preprint.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "has an LLM-free read path: a hand-coded planner classifies query shape"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LongMemEval-S 88.2% overall, multi-session 84.2%, temporal 87.2% with GPT-4o"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Eywa keeps SQLite as the authority for evidence, facts and provenance and treats vector indexes and graph checkpoints as"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No ablation of the planner is reported."

## Relations
- supports:: [[Write-time intelligence and read-time determinism]]
- supports:: [[A query planner and EXPLAIN for context]]
- part-of:: [[Event sourcing]]
- tested-by:: [[Typed value anchors appeared on 59 of 2541 chat facts]]
- tested-by:: [[A third of Eywa's residual errors were packing or multi-evidence errors]]
- tested-by:: [[Set-level BEAM categories scored 64.07 to 79.49 percent against 93.21]]
