---
type: idea
status: built
area: index
needs: none
---
Each turn gets a one-line card that says what it is about and names its threads and entities, tied to the sha of the turn's body. The catalog of cards tells a model what exists so that it can ask for it by id. A card never stands in for text that was selected: the text goes in verbatim and the card stays in the catalog. The model-written line is used only as a pointer to stored text, never as the text.

## Evidence
- `docs/ARCHITECTURE.md`: "Each turn gets a one-line card (what it is about, its threads, its entities) tied to the sha of the turn body. Cards tell a model what exists so it can ask for it."
- `docs/ARCHITECTURE.md`: "A card never stands in for text that was selected: a turn ranked for the message goes in as exact excerpts (whole when all of it is picked), and its card stays in the catalog."
- `docs/ARCHITECTURE.md`: "cards for every turn not sent; a long turn's card gives its size and parts"

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[The indexer writes cards facts and decisions with checked quotes]]
- builds-on:: [[Model-written notes lose as storage]]
- part-of:: [[Views with lineage instead of summaries]]
- source:: [[BookWriter]]
- tested-by:: [[At 1,200 turns every turn is in the packet verbatim or named in it]]
