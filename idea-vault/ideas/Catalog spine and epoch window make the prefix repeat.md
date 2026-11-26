---
type: idea
status: built
area: caching
needs: model-calls
---
Two layout steps make consecutive briefs share a long opening. CatalogSpine lists every older turn in the catalog, so the catalog only grows at its end, and recent_epoch starts the recent window at a multiple of N turns, so it does not slide on every message. Both are built and switched off: in replay at an epoch of 4 they raised the median shared prefix from nothing to 37 to 45 percent without changing probe recovery, but with no cache hits they make each message cost more.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "Two further steps are built but off: `CatalogSpine` (the catalog lists every older turn) and `recent_epoch` (the recent window starts at a multiple of N turns)."
- `docs/CONTEXT-COMPILER.md`: "Probe recovery did not change on any of the five sets."

## Relations
- builds-on:: [[The brief runs from stable to volatile]]
- builds-on:: [[Catalog cards are an index not a replacement]]
- source:: [[TokenPilot]]
- tested-by:: [[Spine and epoch 4 raised the shared prefix from 0 to 37 45 and 44 percent]]
- tested-by:: [[With no cache hits the stable layout raises input per message]]
- tested-by:: [[Spine and epoch break even once 14 to 29 percent of calls hit]]
