---
type: result
status: preprint
area: index
needs: none
measured-by: others
---
In the same revision benchmark the hard part was assigning each statement to the right key. A rule extractor reached 0.807 key recall and 0.781 clean retrieval, while LLM key assigners reached 0.980 to 0.996 key recall but only 0.560 to 0.644 clean retrieval, because false merges removed current values. On a labelled LongMemEval subset merge recall never exceeded 0.062. So a validity filter is only as good as its keys, and a conservative rule beat a model at write time here.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "A rule extractor reached 0.807 key recall and 0.781 clean retrieval. LLM key assigners reached 0.980 to 0.996 key recall but only 0.560 to 0.644 clean retrieval because false merges removed current values."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "merge recall never exceeded 0.062 and stale exposure stayed between 0.594 and 0.719"

## Relations
- supports:: [[Validity filters need an intent router for history questions]]
- supports:: [[Stale memory is a weak spot of memory frameworks]]
- contradicts:: [[Write-time intelligence and read-time determinism]]
- source:: [[Deng and Sinchaisri 2026]]
