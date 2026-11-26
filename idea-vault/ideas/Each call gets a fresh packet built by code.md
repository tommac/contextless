---
type: idea
status: built
area: packets
needs: none
---
Every model call runs in a packet directory that the engine built for that one call from the records in the vault. Deterministic code builds it by following edges one hop out from the new turn: the session record, the new message, recalled turns, decisions in force, recent turns, ranked paragraphs, linked facts and entities, thread turns and catalog cards. If a record is not in the packet, the model does not know it. The run directory keeps the packet exactly as the model saw it, with a manifest of every record sent, so any call can be inspected afterwards.

## Evidence
- `docs/ARCHITECTURE.md`: "Each call runs in a packet directory the engine built for it. If a record is not in the packet, the model does not know it."
- `docs/ARCHITECTURE.md`: "The packet is built by following edges one hop out from the new turn, in this order"
- `docs/ARCHITECTURE.md`: "Every model call gets a fresh, sized packet built from those records by deterministic code."
- `docs/ARCHITECTURE.md`: "ws/brief.md ws/.cx/handoff.json the packet exactly as the model saw it"

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[Provider CLIs run as stateless workers]]
- builds-on:: [[The vault layers records from session down to verbatim turn]]
- part-of:: [[Write-time intelligence and read-time determinism]]
- source:: [[BookWriter]]
- tested-by:: [[At 1,200 turns every turn is in the packet verbatim or named in it]]
