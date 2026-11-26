---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/html/2607.29377
---
Zero-Mem (Xiao et al., PolyU; arXiv 31 Jul 2026) keeps raw traces and builds an entity to context graph with co-occurrence edges and adjacency edges between neighbouring units, plus a hierarchy of turn, window, episode and local span. It retrieves with BM25 plus BGE-M3 embeddings and personalized PageRank, and a closure step adds neighbouring turns. LoCoMo F1 averaged 59.15 against 53.75 for the best baseline, though it was not best on multi-hop. It needs embeddings at query time and has no ablation on LoCoMo.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "an entity to context graph with co-occurrence edges and adjacency edges between neighbouring units; a hierarchy of turn, window, episode and local span"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LoCoMo F1 with GPT-4o-mini: 59.15 average against GAM 53.75, HippoRAG 47.92 and plain retrieval 44.73; on multi-hop it was not best (41.61 against 42.29)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Zero-Mem has no LoCoMo ablation"

## Relations
- part-of:: [[Personalized PageRank]]
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- alternative-to:: [[EdgeMem]]
- tested-by:: [[On HotpotQA a graph alone scored 62.50 F1 against 54.88 for a hierarchy alone]]
