---
type: result
status: measured
area: packets
needs: none
measured-by: us
---
With the whole conversation in front of it, the answering model stated 0.946 of the needed passages on the shorter conversation (176,000 tokens, 12 questions) and 0.719 on the longer (295,000 tokens, 18 questions). Over the 30 questions an exact packet of about 1,200 tokens beat the whole conversation by 0.172, with 16 won and 1 lost. Two conversations cannot separate length from content, but the direction matches the first run: sending everything is not the ceiling, and it gets worse as the conversation grows.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "- The whole conversation scored 0.719 on the longer conversation (295,000"
- `docs/CONTEXT-COMPILER.md`: "0.172 (16 won, 1 lost): the first run's finding, repeated."

## Relations
- supports:: [[An exact packet scored 0.997 against 0.839 for the whole conversation]]
