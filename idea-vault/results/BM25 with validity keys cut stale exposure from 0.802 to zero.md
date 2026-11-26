---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
In a controlled revision benchmark of 1,200 queries, BM25 with validity and participant keys retrieved the current value in every case, exposed no stale value and had clean retrieval of 0.873, within the preset margin of the reference ranker at 0.877, at a median of 0.10 ms. The same BM25 without validity exposed a stale value in 0.802 of cases, with clean retrieval of 0.195. Stores with no keys exposed a superseded value in 70.3 percent of prompts. The benchmark is a small synthetic template family, measured by its authors.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "BM25 + validity + participants: current value retrieved 1.000, stale exposure 0.000, clean retrieval 0.873, **p50 0.10 ms**, within the preset ±0.02 of the reference ranker (0.877)."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The same BM25 without validity: stale exposure 0.802, clean 0.195."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Keyless stores exposed a superseded value in 70.3% of prompts."

## Relations
- supports:: [[Validity filters need an intent router for history questions]]
- supports:: [[Facts supersede facts]]
- supports:: [[Stale memory is a weak spot of memory frameworks]]
- supports:: [[Write-time intelligence and read-time determinism]]
- source:: [[Deng and Sinchaisri 2026]]
