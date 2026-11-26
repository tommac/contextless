---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
On full LoCoMo at k = 50, hybrid BM25 plus dense retrieval and vector-only retrieval tied at about 90.2 and 90.3 percent mean evidence recall, as measured by the FluctlightDB authors in 2026. A small-embedder hybrid stage in another system raised accuracy from 0.360 to 0.390 at 3.7 times the retrieval latency. So the hybrid gain is benchmark dependent and can vanish at a generous cutoff.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "On full LoCoMo at k = 50, hybrid BM25 + dense and vector-only tied at about 90.2% and 90.3% mean evidence recall"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "accuracy 0.360 → 0.390 at 3.7× retrieval latency (96.7 vs 25.7 ms)"

## Relations
- contradicts:: [[Hybrid lexical and embedding retrieval beats either alone]]
- supports:: [[Reranking pays only when the budget is tight]]
