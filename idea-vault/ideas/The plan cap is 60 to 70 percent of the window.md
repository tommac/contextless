---
type: idea
status: built
area: packets
needs: none
---
The engine does not fill a provider profile's window. It plans each packet to 60 to 70 percent of it, the plan cap, and a packet must fit the cap of the profile it is routed to. Two more fractions guard the run itself: a watchdog kills a run at 85 percent of the window, and a run whose peak passes 90 percent is tainted and not used. The plan, kill and taint fractions came from BookWriter's packet model.

## Evidence
- `docs/ARCHITECTURE.md`: "plans to 60 to 70 percent of that profile's window"
- `docs/ARCHITECTURE.md`: "The packet must fit the routed profile's plan cap."
- `docs/ARCHITECTURE.md`: "kill at 85 percent of the window, taint on any compaction event"
- `docs/ARCHITECTURE.md`: "Packet cost model, plan/kill/taint fractions, drop tiers,"

## Relations
- part-of:: [[Packets are sized before the model starts]]
- source:: [[BookWriter]]
- tested-by:: [[At 1,200 turns every turn is in the packet verbatim or named in it]]
