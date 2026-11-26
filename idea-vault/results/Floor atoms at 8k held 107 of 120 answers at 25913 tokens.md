---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
The reference arm, the current default of 8k ranked tokens, held 107 of 120 with a median brief of 25,913 tokens and 22,260 input tokens per replayed message. Every other arm of the floor is compared with it. The median ranked text before the answer is 400 tokens with full text alone and 432 with all lanes.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "| The same at 8k | 107 | reference | 25,913 | 22,260 |"
- `docs/CONTEXT-COMPILER.md`: "the median ranked text before the answer is 400 tokens with full text alone and 432 with all lanes"

## Relations
- supports:: [[The offline floor is the baseline every selector must beat]]
- supports:: [[Ranked atoms are sent as excerpts under a token budget]]
- supports:: [[Atoms match whole turns for fewer tokens within noise]]
- supports:: [[Full text alone matched all lanes on real probes]]
