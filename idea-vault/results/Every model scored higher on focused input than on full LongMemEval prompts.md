---
type: result
status: vendor-reported
area: packets
needs: none
measured-by: others
---
Chroma's Context Rot report tested 18 models. On LongMemEval's 306 prompts, about 113k tokens in full against about 300 tokens focused, every model scored higher with the focused input, and performance degraded with input length. It is a vendor report with many results only in charts. For this project it is evidence that the whole history is not automatically the best packet even when the answer is in it.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "On LongMemEval's 306 prompts (about 113k tokens full against about 300 focused), every model scored higher with focused input."

## Relations
- supports:: [[Unrelated context can hurt answers]]
- supports:: [[Ranked atoms are sent as excerpts under a token budget]]
- contradicts:: [[The append-only history as the cache prefix]]
- contradicts:: [[A short history is sent whole]]
- source:: [[Context Rot]]
