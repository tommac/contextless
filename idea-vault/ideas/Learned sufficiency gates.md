---
type: idea
status: rejected
area: recall
needs: none
---
A learned gate that judges the context sufficient and then suppresses further retrieval is rejected for this design. Unsupervised signals correlate only weakly with whether retrieval helps, and a replication of one such gate labelled most questions ambiguous. The decision is to use these signals to allocate budget and to word the brief, and never to suppress recall.

## Evidence
- `reports/Cutting edge context memory research.md`: "Unsupervised signals correlate about 0.2 to 0.3 with retrieval benefit"
- `reports/Cutting edge context memory research.md`: "a CRAG replication labeled 88.3% of ARC questions Ambiguous"
- `reports/Cutting edge context memory research.md`: "Use signals to allocate budget and word the brief, never to suppress recall"

## Relations
- contradicts:: [[Recall by id or by search replaces memory]]
- alternative-to:: [[Recall as paging]]
- alternative-to:: [[Packets are sized before the model starts]]
- alternative-to:: [[Deterministic sufficiency signals word the brief]]
- tested-by:: [[Unsupervised signals correlate 0.2 to 0.3 with the benefit of retrieval]]
