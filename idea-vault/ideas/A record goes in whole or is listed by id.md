---
type: idea
status: built
area: packets
needs: none
---
The engine never shortens a record to make it fit. What is sent is sent word for word, and a record that is left out for budget is named by its id, so the model knows it exists and can ask for it. A ranked turn may go in as exact excerpts, whole when all of it is picked, while its card stays in the catalog. Nothing in a packet is a summary standing in for text.

## Evidence
- `docs/ARCHITECTURE.md`: "Nothing is ever compacted, and no record is ever summarized to make it fit."
- `docs/ARCHITECTURE.md`: "leaves records out by tier (listing each one it left out, by id), then splits the work into units. It never rewrites a record shorter."
- `docs/ARCHITECTURE.md`: "It never rewrites a record shorter."

## Relations
- builds-on:: [[Packets are sized before the model starts]]
- builds-on:: [[Catalog cards are an index not a replacement]]
- builds-on:: [[Lossy compression loses constraints and exact text]]
- alternative-to:: [[Ranked atoms are sent as excerpts under a token budget]]
- source:: [[BookWriter]]
- tested-by:: [[At 1,200 turns every turn is in the packet verbatim or named in it]]
