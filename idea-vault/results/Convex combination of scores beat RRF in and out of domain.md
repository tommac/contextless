---
type: result
status: measured
area: selection
needs: none
measured-by: others
---
A peer-reviewed study in ACM TOIS found that a convex combination of lexical and semantic scores beat reciprocal rank fusion in-domain and out-of-domain, that RRF was sensitive to its parameters, and that a small labelled set was enough to tune the single weight. RRF discards score magnitude, which the exact and value lanes may carry usefully, so the notes suggest testing weighted RRF against a convex combination of normalized scores on this project's probes. The result comes from document retrieval, not memory.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Convex combination of scores beat RRF in-domain and out-of-domain, RRF was sensitive to its parameters, CC was largely agnostic to normalization choice once learned, and a small labeled set sufficed to tune its single weight"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Bruch's result comes from document retrieval, not memory, so confirm it locally."

## Relations
- contradicts:: [[Fused ranking over exact full-text and value lanes]]
- supports:: [[Learning from recall requests and citations]]
