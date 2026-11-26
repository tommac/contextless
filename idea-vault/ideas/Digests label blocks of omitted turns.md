---
type: idea
status: built
area: index
needs: none
---
When even catalog cards have to be dropped for budget, the packet still says that those turns exist. Code writes one digest line per block of up to 50 dropped catalog turns, giving the block's thread, its id ranges and its most common keywords. No model writes it. The idea is LCM's summary of summaries, built by code as ARC insists.

## Evidence
- `docs/ARCHITECTURE.md`: "a dropped catalog card in a digest line per block of up to 50 turns that gives the block's thread, its id ranges and its most common keywords (built by code, no model)"
- `docs/ROADMAP.md`: "One line per block of up to 50 dropped catalog turns: thread title, id ranges, top keywords."
- `docs/ROADMAP.md`: "LCM's summary-of-summaries, built by code as ARC insists"

## Relations
- builds-on:: [[Catalog cards are an index not a replacement]]
- builds-on:: [[Threads group turns by topic]]
- builds-on:: [[Drop tiers set the order in which records are left out]]
- part-of:: [[Levels of detail like map tiles]]
- alternative-to:: [[Summary trees help only on very long histories]]
- source:: [[LCM Lossless Context Management]]
- source:: [[ARC Addressable Recall Compaction]]
