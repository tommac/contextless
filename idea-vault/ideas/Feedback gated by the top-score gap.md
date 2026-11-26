---
type: idea
status: proposed
area: selection
needs: none
---
Apply query expansion from the first-pass results only when a first-pass confidence signal, such as the gap between the top scores, is weak, and leave confident queries alone. The research note ranks it as a sixth, lower-novelty option after its five open mechanisms; the gate itself is the untested part on conversational data, and no study of selective feedback on conversational memory was found. A cascade that skips its dense channel on the same kind of score-gap signal is the nearest working example, read only through a review page. Any such lane should be reported with a robustness index, since mean gains hide the queries it harms.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A gated feedback lane is a sixth, lower-novelty option: apply expansion only when a first-pass confidence signal such as the top-score gap is weak, and report the robustness index. The gate itself is the untested part on conversational data."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No study of selective (gated) feedback on conversational memory was found."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "skips the dense channel (52 ms) when the confidence (s0 - s1)/s0 passes a threshold"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A risk-constrained selection of expansion terms cut that loss by 21.0% to 76.9%"

## Relations
- part-of:: [[Feedback expansion from first-pass hits needs no model]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[Deterministic sufficiency signals word the brief]]
- builds-on:: [[An adaptive cutoff at the largest score gap needs no model]]
- analog-of:: [[Pseudo-relevance feedback]]
- alternative-to:: [[Slot allocation over deterministic aspects]]
- alternative-to:: [[Propagation from lexical seeds over a rare-key graph]]
