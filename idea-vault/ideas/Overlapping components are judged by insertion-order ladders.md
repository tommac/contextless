---
type: idea
status: proposed
area: evaluation
needs: none
---
When two structure components partly substitute for each other, the gain of each depends on which is added first, so each must be measured in both orders and not by a single delta. In EdgeMem, co-occurrence alone adds 5.65 points to flat BM25 but only 2.21 once session structure and neighbours are in, and the two are about 3.4 points short of additive. The comparison reviewers will expect is EdgeMem's ladder: flat span BM25, then neighbours and session mixing, then the new mechanism, with insertion-order checks. The bar a new link mechanism must clear is therefore flat BM25 plus neighbour reading plus distinct-session seeding.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Co-occurrence and adjacency partly substitute for each other (3.4 points of overlap), so their value must be judged by insertion-order ladders as EdgeMem does, not by single-component deltas."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The comparison baseline that reviewers will expect is EdgeMem's ladder: flat span BM25, then neighbours and session mixing, then the new mechanism, with insertion-order checks because structure components overlap."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The bar a new link mechanism must clear on conversations is not flat BM25. It is flat BM25 plus neighbour reading plus distinct-session seeding, which already captures most of the published gain."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Used alone on flat BM25: co-occurrence +5.65, time +0.45. Reversing the insertion order gives +5.65 then +5.04"

## Relations
- builds-on:: [[Adjacent-turn expansion is the control for any link graph]]
- builds-on:: [[Same-harness baselines every new structure must beat]]
- part-of:: [[R12 Link graph against the adjacency control]]
- tested-by:: [[Neighbour and session context added 8.44 points against 2.21 for co-occurrence]]
- source:: [[EdgeMem]]
