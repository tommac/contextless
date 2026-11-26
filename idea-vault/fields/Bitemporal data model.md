---
type: field
status: external
area: storage
needs: none
---
A bitemporal store, a database idea, records for every fact both when it was true in the world and when the system learned it, and it closes old versions instead of deleting them. Zep's edges carry four such timestamps, and another store in the notes keeps an identity node unchanged while version nodes carry the two time intervals. For this project it is the model for letting one fact supersede another while the source turn stays untouched, though the notes found no benchmark that isolates it from the quality of extraction and retrieval.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Temporal graphs offer the most developed model of supersession (bitemporal validity plus invalidation instead of deletion)"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "an identity node is never changed, content sits in version nodes with valid-time and transaction-time intervals, and updates close the current version"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "I did not find a benchmark that isolates bitemporal modelling from extraction quality and retrieval."

## Relations
- analog-of:: [[Facts supersede facts]]
- analog-of:: [[Derived validity records over immutable turns]]
- analog-of:: [[Validity filters need an intent router for history questions]]
