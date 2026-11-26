---
type: idea
status: proposed
area: index
needs: none
---
Build determinism needs more than deterministic code: the version of every derived index or artifact must be an input to the context build. Approximate nearest-neighbour search, floating-point ties and changing language-processing models are named as specific risks. A summary regenerated at build time would make the same request non-reproducible, so anything a model makes has to be generated off the read path and frozen before use.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Build determinism requires more than deterministic code: every derived index or artifact version must be an input to the context build."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Approximate nearest-neighbor search, floating-point ties and changing NLP models are specific risks."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "These should be generated off the read path and frozen before use."

## Relations
- supports:: [[Determinism is the instrument]]
- supports:: [[Write-time intelligence and read-time determinism]]
- builds-on:: [[Each call gets a fresh packet built by code]]
- builds-on:: [[The atom index is stemmed and searches card text]]
