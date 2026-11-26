---
type: idea
status: external
area: storage
needs: none
---
Memory that a model maintains is weak at noticing that an old note has gone stale. On the STALE benchmark, memory frameworks scored far below the best plain model, and they usually retrieved the old entry without judging that it needed an update. Staleness is therefore as much a problem of deciding at write time as of retrieving, which bears directly on how this project lets one fact supersede another.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "model-maintained memory is currently weak at noticing that old notes have gone stale"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Staleness looks like a write-time adjudication problem as much as a retrieval problem, since systems often retrieve the old note and still fail to revise it."

## Relations
- supports:: [[Facts supersede facts]]
- tested-by:: [[Memory frameworks scored 5 to 18 percent on STALE]]
- tested-by:: [[Only 3.3 percent of recalled stale entries were judged to need an update]]
- tested-by:: [[BM25 with validity keys cut stale exposure from 0.802 to zero]]
- source:: [[STALE benchmark]]
