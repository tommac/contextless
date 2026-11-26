---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
On the 120 distinct questions the floor ran each atom budget with full text only and with all lanes, and the count of answers in the packet was the same at every budget. The median ranked text before the answer was 400 tokens with full text alone and 432 with all lanes. The design's reading is that the exact and value lanes may matter for other questions but these 120 give them no credit.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "**The exact and value lanes change no packet result here.** Full text alone found the same number at every budget."
- `docs/CONTEXT-COMPILER.md`: "They may matter for other questions, but these 120 give them no credit"

## Relations
- supports:: [[Full text alone matched all lanes on real probes]]
- contradicts:: [[Fused ranking over exact full-text and value lanes]]
- contradicts:: [[The value lane finds stated values near the question's words]]
- supports:: [[Held-out probes written from memory without the transcript]]
