---
type: result
status: preprint
area: evaluation
needs: none
measured-by: others
---
In Beyond Memory Leaderboards, a preprint whose authors include their own system, a graph memory's lead on one benchmark came with 2.6 million retrieved characters per query and disappeared after controlling for budget. On a second benchmark, adding BM25 to dense retrieval was the largest single intervention, and a simple hybrid, Mem0 and the authors' system tied. Controlling the budget changed the conclusions, which is why accuracy has to be reported with evidence recall and tokens used.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "on PAIM, Graphiti's lead came with 2.6M retrieved characters per query and disappeared after controlling for budget"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "On PTr, adding BM25 to dense retrieval was the largest intervention, and hybrid Simple RAG, Mem0 and Theoria tied within 0.03 points"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Evaluation should report retrieved evidence recall and tokens used alongside answer accuracy."

## Relations
- supports:: [[Budget-matched baselines]]
- supports:: [[Same-harness baselines every new structure must beat]]
- supports:: [[Benchmark scores are not comparable across papers]]
- supports:: [[Knowledge graphs lose to BM25 as a default store]]
- supports:: [[Hybrid lexical and embedding retrieval beats either alone]]
