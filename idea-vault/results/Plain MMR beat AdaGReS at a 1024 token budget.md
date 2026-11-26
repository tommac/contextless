---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
In the one 2026 head-to-head the notes found, at a 1,024-token budget on a viewpoint-coverage task, recovered coverage was 0.9588 for the learned LMVU selector, 0.9222 for plain MMR, 0.8566 for dense top-k and 0.7771 for AdaGReS, against 0.9997 for an exact oracle. AdaGReS subtracts a growing redundancy term and selects on raw gain, so its headline guarantee does not apply to a token budget. The evaluation is by the LMVU authors on small query counts and is about viewpoint coverage, not fact recall.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "LMVU 0.9588 / 0.8795, **MMR 0.9222 / 0.8314**, dense top-k 0.8566 / 0.7695, **AdaGReS 0.7771 / 0.6362**, exact oracle 0.9997 / 0.9558"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "AdaGReS's objective subtracts a growing redundancy term, so it is not monotone, and it selects on raw gain."

## Relations
- supports:: [[Budgeted selection needs a monotone submodular objective]]
- supports:: [[Learning from recall requests and citations]]
- contradicts:: [[AdaGReS-style guaranteed selection]]
