---
type: result
status: measured
area: packets
needs: none
measured-by: others
---
Lost in the Middle (Liu et al., 2023, TACL): on multi-document question answering and key-value retrieval, performance is often highest when the relevant information is at the start or the end of the input and drops when it sits in the middle of a long context. The research notes add that placing the question last is sensible and cache-friendly but does not cancel the length penalty.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "performance is often highest when relevant information is at the start or end and drops when it is in the middle of long contexts"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "placing the question last is sensible and cache-friendly, but it does not cancel the length penalty"

## Relations
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Unrelated context can hurt answers]]
- source:: [[Lost in the Middle]]
