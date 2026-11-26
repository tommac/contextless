---
type: source
status: external
area: structure
needs: none
year: 2023
url: https://arxiv.org/abs/2312.06648
---
Chen et al., 2023, ask what retrieval granularity to use and index model-written propositions as the unit. Gains over passages concentrate in weaker or unsupervised retrievers and long-tail entities, and the index grows to about 6.3 propositions per passage. This design takes the small unit and rejects the rewriting: atoms are verbatim spans, not propositions written by a model.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "| Dense X Retrieval (2023) | propositions as retrieval units | small units; but verbatim spans, not model-rewritten propositions |"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Gains concentrate in weaker or unsupervised retrievers and long-tail entities."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Cost: 256.9M propositions against 41.4M passages (about 6.3 per passage), index about 768 GB."

## Relations
- supports:: [[Atoms are paragraph-sized addresses into a turn]]
- alternative-to:: [[Turns are stored verbatim and never rewritten]]
- tested-by:: [[Propositions helped weak retrievers and not strong ones]]
