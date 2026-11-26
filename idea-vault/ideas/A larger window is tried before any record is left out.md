---
type: idea
status: built
area: packets
needs: none
---
When a packet does not fit the plan cap of the profile it was routed to, the engine first tries a larger profile that the role allows. The README gives the ladder as claude 200k, codex 500k and kimi 1M. Moving to a larger window keeps the packet whole; leaving records out is the second step and splitting the work the third.

## Evidence
- `docs/ARCHITECTURE.md`: "If it does not fit it tries a larger window, then leaves records out by tier (listing each one it left out, by id), then splits the work into units."
- `docs/ARCHITECTURE.md`: "try a larger profile allowed for the role; then drop tier 1, 2, 3, 4 records"

## Relations
- builds-on:: [[Packets are sized before the model starts]]
- builds-on:: [[Roles map to ordered provider profiles kept in config]]
- source:: [[BookWriter]]
