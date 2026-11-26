---
type: idea
status: proposed
area: selection
needs: none
---
With validity and participant metadata written at storage time, plain BM25 retrieved the current value every time with no stale exposure, at sub-millisecond latency. A validity filter alone would break the question of what a value was before, because it hides exactly the value asked for. The notes propose a deterministic router on cue words: current-state queries prefer the latest member of a value chain, history queries retrieve the chain with the predecessor labelled by date, and change queries include the transition pair. The value lane is the natural place to build chains, favouring precision because false merges destroy current values.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "the strongest finding is that **state modeling at write time beats ranking**: with validity and participant metadata, plain BM25 matched the reference ranker at sub-millisecond latency"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "by hiding exactly the required value. Use a deterministic intent router instead:"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Favor precision in merging: the evidence shows false merges destroy current values."

## Relations
- builds-on:: [[Facts supersede facts]]
- builds-on:: [[The value lane finds stated values near the question's words]]
- source:: [[Deng and Sinchaisri 2026]]
- part-of:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[BM25 with validity keys cut stale exposure from 0.802 to zero]]
- tested-by:: [[LLM key assigners merged falsely and removed current values]]
- analog-of:: [[Bitemporal data model]]
