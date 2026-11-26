---
type: source
status: external
area: selection
needs: none
year: 2026
url: https://arxiv.org/html/2610.10265
---
A controlled revision benchmark (7 October 2026, 1,200 queries) on retrieving the current value after a fact changes. With validity and participant keys, BM25 retrieved the current value every time with no stale exposure at a median of 0.10 ms; without validity it exposed a stale value in about 80 percent of cases. The hard part is assigning keys: model key assigners had high key recall but removed current values through false merges. The benchmark is a small synthetic template family.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The benchmark is a small synthetic template family"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The hard part is key assignment."

## Relations
- supports:: [[Validity filters need an intent router for history questions]]
- supports:: [[Facts supersede facts]]
- supports:: [[Stale memory is a weak spot of memory frameworks]]
- tested-by:: [[BM25 with validity keys cut stale exposure from 0.802 to zero]]
- tested-by:: [[LLM key assigners merged falsely and removed current values]]
