---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
In one comparable table of Recall@10, BM25 alone scored 48.3 on LoCoMo and 81.3 on LongMemEval, a 0.6B embedder alone 46.9 and 84.8, and the two combined with normalized equal weight 54.3 and 90.5. Token-level matching over the same small model's vectors did better still, at about 20 to 30 ms and roughly 100 times BM25's storage or more. The table is by the authors of the best method in it.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "| BM25 | 48.3 / 36.7 | 81.3 / 73.0 | 40.0 | [NO MODEL] |"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "| BM25 + 0.6B (normalized, equal weight) | 54.3 / 42.1 | 90.5 / 81.5 | 46.5 | [LOCAL MODEL] |"

## Relations
- supports:: [[Hybrid lexical and embedding retrieval beats either alone]]
- supports:: [[An embedding lane beside the exact full-text and value lanes]]
- source:: [[Luo et al 2026]]
