---
type: idea
status: built
area: structure
needs: none
---
A thread is a record for one topic or workstream of the conversation, marked open or resolved. The indexer assigns each new turn to its threads, and the edge is stored on the turn. Threads shape the packet: the other turns of the new message's threads go in verbatim, facts and entities are picked through its threads, and a digest line names the thread of each block of omitted turns. Open or resolved status also covers part of what reminder and todo records would do.

## Evidence
- `docs/ARCHITECTURE.md`: "L1 topics / workstreams, open or resolved"
- `docs/ARCHITECTURE.md`: "the other turns of the focus's threads, verbatim"
- `docs/ROADMAP.md`: "Threads with open or resolved status cover part of it."

## Relations
- part-of:: [[The vault layers records from session down to verbatim turn]]
- builds-on:: [[The indexer writes cards facts and decisions with checked quotes]]
- alternative-to:: [[Reminder and todo records in the vault]]
