---
type: field
status: external
area: storage
needs: none
---
A log-structured merge tree is a storage-engine design that writes immutable sorted segments and merges them in the background. The notes find the analogy attractive for append-only history, with immutable turn segments, sealed indexes per segment and background merges. A targeted search found no agent-memory paper that implements it, so it is an untested proposal here, and a merge must not change retrieval results unless the index version is an input to the build.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "An LSM analogy is still attractive as engineering: immutable turn segments, sealed per-segment indexes and background merges match append-only history."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "This is an untested proposal for this domain, and merge policy must not change retrieval results unless the index version is part of the build input."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "A targeted search found LSM-tree storage-engine papers and adjacent agent-memory work, but no agent-memory paper implementing LSM tiers"

## Relations
- analog-of:: [[Turns are stored verbatim and never rewritten]]
- analog-of:: [[Every derived index version is an input to the context build]]
- part-of:: [[Append-only storage]]
