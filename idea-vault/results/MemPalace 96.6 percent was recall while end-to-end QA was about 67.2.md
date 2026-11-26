---
type: result
status: disputed
area: evaluation
needs: none
measured-by: others
---
An audit of MemPalace found that its 100 percent on LoCoMo used a top-k of 50, which is effectively the whole conversation, and that its 96.6 percent was recall at 5 while end-to-end question answering was about 67.2 percent. A retrieval metric had been presented as answer accuracy. Any result with a large top-k should be reported with the tokens delivered to the reader.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "its 100% LoCoMo used top_k=50, effectively the whole conversation"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "MemPalace pairs 96.6% recall_any@5 with about 67.2% QA accuracy"

## Relations
- supports:: [[Three layers reached used and correct]]
- supports:: [[Benchmark scores are not comparable across papers]]
- supports:: [[Budget-matched baselines]]
- contradicts:: [[A LoCoMo leaderboard position as a headline metric]]
