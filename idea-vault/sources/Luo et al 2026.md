---
type: source
status: external
area: selection
needs: none
year: 2026
url: https://arxiv.org/pdf/2610.13277
---
A preprint (version 3, 7 November 2026) with one comparable table of retrieval methods on LoCoMo, LongMemEval and REALTALK, from BM25 through small and large embedders, hybrids and late interaction to the authors' own training-free sparse token matching. BM25 plus a 0.6B embedder beats either alone, and a 4B embedder alone loses to BM25 by 22 points on LongMemEval. The authors' own method is the last row of the table, so its margin deserves caution.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "the authors' own method is the last row, so its margin deserves caution"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The bigger 4B embedder alone lost to BM25 by 22 points on LongMemEval, so model size is not a proxy for memory retrieval quality"

## Relations
- supports:: [[Hybrid lexical and embedding retrieval beats either alone]]
- supports:: [[An embedding lane beside the exact full-text and value lanes]]
- tested-by:: [[BM25 plus a small embedder gained 9 recall points on LongMemEval]]
