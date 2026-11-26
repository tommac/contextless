---
type: result
status: preprint
area: storage
needs: none
measured-by: others
---
On STALE's 1,200 queries, memory frameworks on a GPT-4o-mini backbone scored between 5.1 percent for A-MEM and 17.8 for LightMem, against 8.7 for the same model alone with truncated context. The best plain model reached 55.2 percent. A prototype that consolidates at write time reached 68.0 percent, but it relies on a predefined state schema.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "LightMem 17.8, Mem0 8.3, LiCoMemory 7.6, Zep 6.0, A-MEM 5.1, versus 8.7 for context-truncated GPT-4o-mini alone"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "A write-time consolidation prototype, CUPMem, reached 68.0%, but relies on a predefined state schema"

## Relations
- supports:: [[Stale memory is a weak spot of memory frameworks]]
- supports:: [[Write-time intelligence and read-time determinism]]
- source:: [[STALE benchmark]]
