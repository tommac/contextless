---
type: result
status: vendor-reported
area: index
needs: none
measured-by: others
---
Anthropic reports that the top-20 miss rate fell from 5.7 percent with plain embeddings to 3.7 with contextual embeddings, 2.9 with contextual BM25 added, and 1.9 with reranking on top, at a stated $1.02 per million document tokens. A model writes the context at index time. It is the vendor's own measurement: per-dataset numbers were not in the text the notes could read, and the contribution of contextual BM25 alone is not given.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Top-20 miss rate: embeddings 5.7%, contextual embeddings 3.7% (−35%), plus contextual BM25 2.9% (−49%), plus reranking 1.9% (−67%)."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The specific contribution of contextual BM25 alone is not in the text I read."

## Relations
- supports:: [[Atom titles as APIs]]
- supports:: [[Deterministic keys on spans with verbatim output]]
- supports:: [[Write-time intelligence and read-time determinism]]
- source:: [[Anthropic contextual retrieval]]
