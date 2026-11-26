---
type: field
status: external
area: index
needs: none
---
Incremental view maintenance updates a stored query result from the changes alone instead of recomputing it. DBSP defines the incremental form of a query so that for linear operators the work is proportional to the size of the change, and the counting algorithm keeps for each derived row the number of ways it can be derived, so a deletion removes the row only when the count reaches zero. The research note argues that this project has the easy case of the theory: spans are immutable and append-only, so only insertions occur and each new turn touches only the keys it mentions. Zep's community summaries, updated by an approximate rule, drift from full recomputation, which argues for exact incremental rules wherever a derived structure feeds the prompt.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "For linear operators the work is proportional to the size of the change."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the counting algorithm keeps, for each derived tuple, the number of alternative derivations, so a deletion removes a tuple only when its count reaches zero"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "only insertions occur, the candidate views (postings per key, per-key timeline, per-key value sequence, decision log) are monotone in the log, and each new turn touches only the keys it mentions"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Zep's drift note argues for exact incremental rules over heuristic ones wherever a derived structure feeds the prompt."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "on TPC-DI it incrementalised 100% of 8 datasets, 6 of 8 with substantial speedups, and the cost model chose correctly in 7 of 8 cases"

## Relations
- part-of:: [[Materialized views]]
- analog-of:: [[The indexer writes cards facts and decisions with checked quotes]]
- analog-of:: [[Views with lineage instead of summaries]]
- analog-of:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[Support sets gave 20 of 20 correct final states against 8 of 20]]
