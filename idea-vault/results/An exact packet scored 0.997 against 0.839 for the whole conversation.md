---
type: result
status: measured
area: evaluation
needs: none
measured-by: us
---
In the 2026-11-25 answer test, a model given only the needed paragraphs stated 0.997 of them. Given the whole conversation (about 248,000 tokens) it stated 0.839 on the 15 questions that had such an answer: 8 questions won by the exact packet, none lost. Sending everything is not the ceiling.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "oracle against full: +0.161 (interval +0.059 to +0.286), 8 won, none lost."
- `docs/CONTEXT-COMPILER.md`: "| full | the whole conversation (15 of the 30 questions) | 0.839 | 247,856 |"

## Relations
- supports:: [[A map of checked notes that the model asks paragraphs from]]
