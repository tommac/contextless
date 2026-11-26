---
type: source
status: external
area: packets
needs: none
year: 2023
url: https://arxiv.org/abs/2307.03172
---
Liu et al., 2023, published in TACL: on multi-document question answering and key-value retrieval, performance is often highest when the relevant information is at the start or end of the input and drops when it is in the middle of long contexts. The design takes from it the rule of focus last and stable material first.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "performance is often highest when relevant information is at the start or end and drops when it is in the middle of long contexts"
- `docs/CONTEXT-COMPILER.md`: "| Lost in the Middle (2023) | models use the start and end of long inputs best | focus last, stable material first |"

## Relations
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Unrelated context can hurt answers]]
- tested-by:: [[Models use the start and end of long inputs best]]
