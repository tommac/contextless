---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
A study titled Does Memory Need Graphs compared index shapes on LongMemEval-S by recall at 5: a flat index over summaries, facts and keywords scored 0.9045, a similarity-edge graph 0.8424, a triple graph 0.9116 and a description graph 0.9331. One-hop expansion under the best re-ranking moved recall at 10 from 0.9904 to 0.9928, which the authors call marginal, and under a weaker re-ranking it cut recall at 5 from 0.9546 to 0.8615. Raw sessions as keys gave 0.9021, and temporal adjacency was listed but not tested. The study uses an LLM at write time and embeddings, so this is not a model-free result.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "One-hop expansion under the best re-ranking moved R@10 from 0.9904 to 0.9928, which the authors call marginal; under a weaker re-ranking it cut R@5 from 0.9546 to 0.8615."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "flat index over summaries, facts and keywords 0.9045; similarity-edge graph 0.8424; triple graph 0.9116; description graph 0.9331"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Raw sessions as keys gave R@5 0.9021. Temporal adjacency was listed but not tested."

## Relations
- contradicts:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Feedback links and coverage did not beat search]]
- part-of:: [[Personalized PageRank]]
- source:: [[LongMemEval]]
