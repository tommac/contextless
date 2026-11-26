---
type: idea
status: proposed
area: selection
needs: none
---
Score a run of consecutive paragraph spans by how short a run covers the query terms, and use the shortest covering runs as candidates. It is cover density lifted from words to spans, and it would return contiguous evidence windows instead of isolated paragraphs. The research note expects proximity inside one paragraph-sized span to add little, so the form across spans is the variant left to test, compared with fixed neighbour reading at equal tokens. No prior art was found.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The untested variant is proximity across adjacent spans: the shortest run of consecutive spans that covers the query terms, which is cover density lifted from words to spans and would return contiguous evidence windows."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Cover density across adjacent spans as a candidate generator, compared with fixed neighbour reading at equal tokens. No prior art found."

## Relations
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[Whole-conversation questions are the weak spot]]
- analog-of:: [[Region algebra and cover density]]
- alternative-to:: [[Neighbor expansion is a priced decision in the selector]]
- alternative-to:: [[A marginal-gain stopping rule for neighbour expansion]]
