---
type: idea
status: rejected
area: selection
needs: none
---
Two opposite slogans both fail against the evidence. The claim that less context is always better is contradicted by a study that found no measurable threshold effect on two models with 1M-token windows, and the claim that the full transcript always wins is contradicted by results where focused evidence beat full history. The report also asks the project's architecture document to stop saying the whole transcript beat every selection, because only presence was measured.

## Evidence
- `reports/Cutting edge context memory research.md`: "Two opposite slogans also fail."
- `reports/Cutting edge context memory research.md`: "no measurable threshold effect on two 1M-context models"
- `reports/Cutting edge context memory research.md`: "Only presence was measured, so that sentence should say so."

## Relations
- builds-on:: [[Unrelated context can hurt answers]]
- part-of:: [[Five statements in the project docs should change]]
- tested-by:: [[Evidence sessions scored 0.870 against 0.606 for full history on LongMemEval]]
- tested-by:: [[Segment retrieval scored 88.81 at 820 tokens against 63.85 for full history]]
