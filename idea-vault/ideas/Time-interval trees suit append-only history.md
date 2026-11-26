---
type: idea
status: proposed
area: structure
needs: none
---
A tree whose parent nodes cover time intervals that contain their children's intervals suits a history that is only ever appended to. New turns mostly affect the rightmost path, and a window that has closed need never change. This is the notes' own reasoning from two temporal-tree systems, and neither paper tests it against an immutable log.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Time-interval containment (TiMem, MemForest) suits append-only history: new turns mostly affect the rightmost path, and sealed windows need never change. This is my reasoning; neither paper tests it against an immutable log."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Insertion touches the new leaf plus its ancestor path, O(log_k N) per affected tree; refresh cost scales with distinct dirty nodes"

## Relations
- builds-on:: [[Immutable leaves plus disposable derived nodes]]
- builds-on:: [[Summary trees help only on very long histories]]
- supports:: [[Catalog spine and epoch window make the prefix repeat]]
- supports:: [[Digests label blocks of omitted turns]]
- supports:: [[The vault layers records from session down to verbatim turn]]
- analog-of:: [[Append-only storage]]
