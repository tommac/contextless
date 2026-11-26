---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
EdgeMem's LoCoMo ladder under a strict judge starts at 49.74 for flat turn BM25, rises to 58.18 with session structure and neighbours, to 60.43 with co-occurrence links and to 61.01 with time. The session and neighbour step is nearly four times the co-occurrence step, and the authors write that most of what the memory contributes is the surrounding episode. The step bundles session-level BM25, distinct-session mixing and reading each seed with its two neighbours, which EdgeMem does not separate. Evidence grew only from 727 to 944 tokens across the ladder, and the result is a single-group preprint on one benchmark.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "EdgeMem's largest step is the session and neighbour step (+8.44), nearly four times the incremental co-occurrence step (+2.21)"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "most of what the memory contributes is the surrounding episode rather than the matched turn alone"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The step bundles session-level BM25, distinct-session mixing and reading each seed with its two neighbours."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Evidence grew only from 727 to 944 tokens across the ladder."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "EdgeMem does not separate neighbour reading from session-level BM25 mixing and distinct-session sampling."

## Relations
- supports:: [[Adjacent-turn expansion is the control for any link graph]]
- supports:: [[Neighbor expansion is a priced decision in the selector]]
- supports:: [[Model-free hubs over verbatim turns]]
- supports:: [[Overlapping components are judged by insertion-order ladders]]
- source:: [[EdgeMem]]
