---
type: idea
status: proposed
area: structure
needs: none
---
Some of what is credited to a graph is really chronology. In Synapse, removing the graph structure halved temporal F1 but left multi-hop flat, and its edges include links between consecutive episodes. A baseline that simply expands to adjacent turns should therefore be run before any gain is credited to semantic links, and the report makes it the required control for the project's own link graph.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "A baseline with adjacent-turn expansion should be included before crediting semantic links."
- `reports/Cutting edge context memory research.md`: "Adjacent-turn expansion is the required control"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Time-adjacency and agent-controlled expansion have better support than generic semantic links"

## Relations
- alternative-to:: [[Links help when questions need several pieces of evidence]]
- part-of:: [[R12 Link graph against the adjacency control]]
- tested-by:: [[Removing graph structure halved temporal F1 and left multi-hop flat]]
- source:: [[Synapse]]
- tested-by:: [[Adding each atom's neighbors doubled the median cost]]
- builds-on:: [[Neighbor expansion is a priced decision in the selector]]
