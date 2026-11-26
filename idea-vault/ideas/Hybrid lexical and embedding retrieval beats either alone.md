---
type: idea
status: external
area: selection
needs: download
---
On memory benchmarks BM25 is a strong floor and adding one small embedding model is the best-supported next step: about 9 Recall@10 points on LongMemEval and 6 on LoCoMo in the one comparable table. The gain is benchmark dependent: on full LoCoMo at k = 50 hybrid and vector-only tied, and a larger embedder alone lost to BM25 by 22 points. This project has no embedding lane; the notes advise staying embedding-free until lexical selection is measured, and the build plan leaves an optional lane to phase 5 behind a model download.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "On memory benchmarks, BM25 is a strong floor and adding one small embedding model is the best-supported next step: about +9 Recall@10 points on LongMemEval and +6 on LoCoMo."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Stay embedding-free until lexical selection and query handling are measured."

## Relations
- alternative-to:: [[Fused ranking over exact full-text and value lanes]]
- source:: [[Luo et al 2026]]
- tested-by:: [[BM25 plus a small embedder gained 9 recall points on LongMemEval]]
- tested-by:: [[Hybrid and vector-only tied at about 90 percent on full LoCoMo]]
