---
type: idea
status: built
area: packets
needs: none
---
The engine estimates the token cost of every packet before it spawns a model, from the packet's bytes and a tokens-per-byte figure measured for each provider profile. It plans to 60 to 70 percent of that profile's window. A packet that does not fit is handled in a fixed order: a larger window, then records left out by tier and named, then the work split into units. Sizing takes the place of compaction, because the engine never rewrites a record shorter.

## Evidence
- `docs/ARCHITECTURE.md`: "The engine estimates every packet's token cost before it spawns a model and plans to 60 to 70 percent of that profile's window."
- `docs/ARCHITECTURE.md`: "Tokens are estimated per profile (bytes times measured tokens-per-byte). The packet must fit the routed profile's plan cap."

## Relations
- builds-on:: [[Each call gets a fresh packet built by code]]
- builds-on:: [[Lossy compression loses constraints and exact text]]
- source:: [[BookWriter]]
- tested-by:: [[At 1,200 turns every turn is in the packet verbatim or named in it]]
