---
type: result
status: measured
area: recall
needs: none
measured-by: others
---
ClawVM assembles a prompt from typed pages by installing pinned minimums and then upgrading by utility per token with deterministic tie-breaks. In its replays the mean number of explicit faults was 0, against 1.5 for a compaction hybrid, 10.8 for retrieval plus cache and 67.8 for retrieval alone. The paper is peer reviewed per its listing, but the main experiments are synthetic replays with very small budgets and baselines built by the authors, and there is no end-to-end test of task quality with a live model.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "mean explicit faults 0 vs 1.5 (compaction hybrid), 10.8 (retrieval plus cache) and 67.8 (retrieval)"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The main experiments are synthetic replays with very small budgets (120 to 500 tokens), the baselines were built by the authors, and there is no end-to-end task-quality evaluation with a live model"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "ClawVM's assembly is the closest published template for deterministic context building"

## Relations
- supports:: [[Recall as paging]]
- supports:: [[Levels of detail like map tiles]]
- supports:: [[Packets are sized before the model starts]]
- supports:: [[Each call gets a fresh packet built by code]]
- source:: [[ClawVM]]
