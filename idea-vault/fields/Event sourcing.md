---
type: field
status: external
area: storage
needs: none
---
Event sourcing records every state change as an ordered event, so the log allows a complete rebuild, temporal queries that reconstruct past states, and replay to correct mistakes; snapshots let a system start from a recent point instead of replaying everything. Agent papers that follow the pattern report reproducibility and consistency, not retrieval quality. The research note states that the target system is already event sourced, and that the transferable part is the snapshot: a periodic materialised state of the value ledger, so that a question about the state as of a given turn does not require a scan.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "record every state change as an ordered event; the log allows a complete rebuild, temporal queries that reconstruct past states, and replay to correct mistakes; snapshots let a system start from a recent point instead of replaying everything"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The target system is already event sourced. The transferable part is the snapshot: a periodic materialised state of the value ledger"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Event-sourcing papers for agents report reproducibility and consistency, not retrieval quality."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the append-only log is the source of truth and the working graph a deterministic projection"

## Relations
- part-of:: [[Append-only storage]]
- analog-of:: [[Durable execution by journal and replay]]
- analog-of:: [[Turns are stored verbatim and never rewritten]]
- analog-of:: [[Each call gets a fresh packet built by code]]
- analog-of:: [[Views with lineage instead of summaries]]
- tested-by:: [[Support sets gave 20 of 20 correct final states against 8 of 20]]
