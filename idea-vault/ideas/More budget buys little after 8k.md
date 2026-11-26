---
type: idea
status: measured-kept
area: selection
needs: none
---
Between 8k and 32k of ranked text, four times the budget, the packet gains six answers of 120, and 16k adds only two. Below 8k the loss is real: 4k loses eight answers and 2k sixteen, both significant. This is why 8k is the default, with the caveat that the count is whether the source is in the packet, not whether a model answers.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "**More budget buys little after 8k:** four times the ranked text adds six answers."
- `docs/CONTEXT-COMPILER.md`: "| The same at 32k | 113 | 6, 0, 0.03 | 48,121 | 41,070 |"

## Relations
- part-of:: [[The offline floor is the baseline every selector must beat]]
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- tested-by:: [[Floor atoms at 2k held 91 of 120 answers]]
- tested-by:: [[Floor atoms at 4k held 99 of 120 answers]]
- tested-by:: [[Floor atoms at 16k held 109 of 120 answers]]
- tested-by:: [[Floor atoms at 32k held 113 of 120 answers]]
