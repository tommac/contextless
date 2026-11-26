---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
The arm with recent turns only, and no ranked section, held the source of none of the 120 questions: against the 8k default it won 0 and lost 107, exact p under 0.001. Its median brief is 18,616 tokens against 25,913 for the default, so the ranked section is what finds every one of the 107. It also shows that these probes ask about older history, not the last few exchanges.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "| Recent turns only | 0 | 0, 107, under 0.001 | 18,616 | 15,345 |"

## Relations
- supports:: [[The offline floor is the baseline every selector must beat]]
- supports:: [[Ranked atoms are sent as excerpts under a token budget]]
- supports:: [[Fused ranking over exact full-text and value lanes]]
