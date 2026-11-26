---
type: idea
status: built
area: index
needs: none
---
The atom index uses the porter stemmer, so a question word meets other forms of the same word, and each atom is searched together with the card text of its turn. The second part came from a measured failure: on the synthetic conversation atoms lost a planted decision whose question shared its words with the card and not with the paragraph. A card match names the turn and not the paragraph, so a turn found only through its card goes in whole when it fits.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "a stemmed atom index that also searches each atom with its turn's card text"
- `docs/CONTEXT-COMPILER.md`: "A card match says which turn, not which paragraph, so a turn found only through its card goes in whole when it fits the budget"

## Relations
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Catalog cards are an index not a replacement]]
- part-of:: [[Fused ranking over exact full-text and value lanes]]
- tested-by:: [[Atoms lost a planted decision until searched with card text]]
- tested-by:: [[Stemming helped atoms by 6 to 17 percent and hurt whole turns]]
