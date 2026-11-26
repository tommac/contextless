---
type: idea
status: built
area: recall
needs: none
---
A model that needs a turn it was not given answers `needs_recall`, with ids taken from the catalog or with search words when no catalog line points to what it needs. The engine searches every earlier turn, rebuilds the packet with what was asked for pinned verbatim, tells the run what each search found and which ids do not exist, and runs again, for at most two rounds. Recall by search came from LCM's lcm_grep, and the note for a bad id that names the valid range came from ARC. With this the model never has to remember: it can always ask for the exact text.

## Evidence
- `docs/ARCHITECTURE.md`: "A model that needs a turn it was not given answers `needs_recall` with the ids from the catalog or, when no catalog line points to what it needs, with search words."
- `docs/ARCHITECTURE.md`: "add a note of what each search found and which ids do not exist, run again (at most 2 rounds)"
- `docs/ROADMAP.md`: "`recall_queries` in the handoff; the engine searches every earlier turn and reports what each search found. Bad ids get a note naming the valid range (ARC)."
- `docs/ARCHITECTURE.md`: "rebuilds the packet with the turns asked for and the paragraphs found pinned verbatim"

## Relations
- builds-on:: [[Catalog cards are an index not a replacement]]
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- analog-of:: [[Recall as paging]]
- source:: [[LCM Lossless Context Management]]
- source:: [[ARC Addressable Recall Compaction]]
- tested-by:: [[Fused lanes with a value lane put 13 of 14 planted facts verbatim]]
