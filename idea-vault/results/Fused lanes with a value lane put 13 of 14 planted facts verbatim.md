---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
With exact, full-text and value lanes fused and each lane's winner kept, the source turn of 13 of the 14 planted facts was in the packet verbatim, up from 11 before the lanes were fused. A recall search for the question found 12 of the 14. The test conversation has 800 turns and about 620k tokens, and no model was run. The 14 facts were planted by the test itself, so the figures show the lanes working on that set, not on real conversations.

## Evidence
- `docs/ROADMAP.md`: "Exact, full-text and value lanes, each lane's winner kept: 13 of 14 verbatim, and 12 of 14 found by a search for the question."
- `docs/ARCHITECTURE.md`: "the source turn of 13 of 14 planted facts is in the packet verbatim"

## Relations
- supports:: [[Fused ranking over exact full-text and value lanes]]
- supports:: [[The value lane finds stated values near the question's words]]
- supports:: [[Recall by id or by search replaces memory]]
- supports:: [[No model has yet answered from a packet]]
