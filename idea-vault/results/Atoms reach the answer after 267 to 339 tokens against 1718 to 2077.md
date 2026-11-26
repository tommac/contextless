---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
On three real conversations and 90 probes, units were ranked by full-text bm25 and added in rank order until one held the answer. With atoms the answer entered after a median of 267 to 339 tokens, against 1,718 to 2,077 for whole turns, about six times fewer; within 2,000 tokens atoms held the answer for 26 of 30 probes on each conversation and whole turns for 14 to 17. Any atom size from about 20 to 2,400 tokens beat whole turns. It is one lane with no fusion and no model, on questions written by a model that had read the source.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "the answer enters the packet after a median of 267 to 339 tokens, against 1,718 to 2,077 for whole turns, about six times fewer"
- `docs/CONTEXT-COMPILER.md`: "| | atoms | 26 of 30 | 28 | 29 | 267 |"
- `docs/CONTEXT-COMPILER.md`: "**Any atom size beats whole turns.** Sizes from about 20 to 2,400 tokens per atom all did"

## Relations
- supports:: [[Atoms are paragraph-sized addresses into a turn]]
- supports:: [[Ranked atoms are sent as excerpts under a token budget]]
- supports:: [[The sixfold saving did not carry over to packets]]
