---
type: idea
status: designed
area: structure
needs: none
---
A packet would show each region of history at one of five levels, from atom through turn, exchange and thread block to the whole conversation: the focus and latest exchanges in full, the hits as atoms, the exchanges around the hits as cards, other threads as digests. Every level has ids, so the model zooms in by recall. Unlike a summary tree, the upper levels only navigate and anything the model relies on arrives verbatim. It is designed and not built in this form.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "far history at low resolution, the relevant spot at full resolution, and zoom by recall"
- `docs/CONTEXT-COMPILER.md`: "here the upper levels are only navigation, and anything the model relies on arrives verbatim"

## Relations
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Digests label blocks of omitted turns]]
- builds-on:: [[Catalog cards are an index not a replacement]]
- source:: [[RAPTOR]]
- analog-of:: [[Map tiles and level of detail]]
- tested-by:: [[A code-made zoom ladder loses to flat search]]
