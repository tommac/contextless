---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
Zero-Mem's ablation on HotpotQA in a 56K-token haystack gave 72.07 F1 for the full system, 62.50 with the entity graph only, 54.88 with the hierarchy only and 67.90 without the closure step that adds neighbouring turns. The research note reads it as entity graphs paying off where evidence lives in unrelated documents, while a conversation has a strong native order so that time and session carry more of the linkage. Zero-Mem uses embeddings at query time and has no ablation on LoCoMo, where it was not best on multi-hop.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Ablation on HotpotQA in a 56K-token haystack: full 72.07 F1, graph only 62.50, hierarchy only 54.88, no closure 67.90."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Entity graphs pay off where evidence lives in unrelated documents (Zero-Mem on HotpotQA: graph only 62.50 against hierarchy only 54.88). A conversation has a strong native order, so time and session carry more of the linkage."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "on multi-hop it was not best (41.61 against 42.29)"

## Relations
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Adjacent-turn expansion is the control for any link graph]]
- part-of:: [[Personalized PageRank]]
- source:: [[Zero-Mem]]
