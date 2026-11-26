---
type: result
status: measured
area: structure
needs: none
measured-by: others
---
In the ablation of Synapse on LoCoMo, a peer-reviewed result, removing the graph structure cut temporal F1 from 50.1 to 25.4 while multi-hop barely moved, from 35.7 to 35.2. Its edges include temporal links between consecutive episodes, so part of what looks like a gain from a graph is chronology. The report concludes that expanding to adjacent turns is the control any link graph must beat.

## Evidence
- `reports/Cutting edge context memory research.md`: "removing graph structure halves temporal F1 (50.1 to 25.4), leaves multi-hop flat (35.7 to 35.2)"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Its edges include consecutive-episode temporal links, abstraction links, and semantic links above 0.92 similarity"

## Relations
- supports:: [[Adjacent-turn expansion is the control for any link graph]]
- contradicts:: [[Links help when questions need several pieces of evidence]]
- source:: [[Synapse]]
