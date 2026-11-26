---
type: idea
status: built
area: packets
needs: none
---
Every selector that feeds a packet has a tier, and the tier is the drop order. Tier 0 is never dropped: the session record, the new message, recalled turns and active decisions. When a packet is still too big after a larger window has been tried, the engine drops catalog cards first (tier 1), then thread turns oldest first (tier 2), then ranked text least relevant first followed by facts and entities (tier 3), and recent turns last, again oldest first (tier 4). Every record dropped is named in the packet, so the order decides what leaves view first and nothing leaves silently.

## Evidence
- `docs/ARCHITECTURE.md`: "tier = drop order, 0 is never dropped"
- `docs/ARCHITECTURE.md`: "try a larger profile allowed for the role; then drop tier 1, 2, 3, 4 records and name every one"
- `docs/ARCHITECTURE.md`: "BookWriter `internal/bw/packet` (re-implemented for chat records)"

## Relations
- part-of:: [[Packets are sized before the model starts]]
- builds-on:: [[A larger window is tried before any record is left out]]
- builds-on:: [[A record goes in whole or is listed by id]]
- source:: [[BookWriter]]
- tested-by:: [[Ranked turns at tier 3 raised planted facts verbatim from 3 to 11 of 14]]
