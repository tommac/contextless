---
type: idea
status: measured-kept
area: cost
needs: none
---
Sending the entire history puts every answer's source in the packet, which no selection arm did, but at about ten times the input of the 8k atom packet. It is the upper bound that selection is measured against, and the reason selection is worth doing only when the history does not fit or costs too much. None of this shows that a model answers better from the whole history, since no model has been run.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "Whole history finds all 120 at about ten times the input."
- `docs/CONTEXT-COMPILER.md`: "| Whole history | 120 | 13, 0, under 0.001 | 297,525 | 219,541 |"

## Relations
- builds-on:: [[A short history is sent whole]]
- part-of:: [[The offline floor is the baseline every selector must beat]]
- alternative-to:: [[Ranked atoms are sent as excerpts under a token budget]]
- tested-by:: [[Floor whole history held all 120 answers at 297525 tokens]]
- tested-by:: [[The 140 turn conversation sent whole held all 30 sources at 83100 tokens]]
