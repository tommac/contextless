---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
Routing through sessions and growing exact spans represented all annotated messages for 380 of 470 LongMemEval-S questions, compared with 343 for the original packet. On the 157 cases where a literal answer occurs in an annotated message and the question omits it, the new packets retained the answer in 139 cases versus 137. These are retrieval measurements.

## Evidence
- `reports/2026-11-20-large-memory/fixed-window-results.md`: "At the same 8,000-byte cap, routing and growing exact excerpts recovered all annotated messages on 37 additional questions and lost none against the original packets."
- `reports/2026-11-20-large-memory/fixed-window-results.md`: "Dropping cases where the question itself names the answer leaves six gains and four losses."

## Relations
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
