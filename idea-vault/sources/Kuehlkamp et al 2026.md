---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/html/2610.10170
---
A placebo-controlled test of chunk structure (7 October 2026) on 951 Wikipedia and 4,303 QASPER queries with dense retrieval, run by a group independent of the methods tested. Structure-aligned chunks with real heading paths beat LLM-contextualized windows, real paths beat shuffled placebo paths, and routing to a section before ranking lost to flat search. Effects are small and parent-child reranking was not tested.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "951 Wikipedia and 4,303 QASPER queries, dense retrieval, coverage-nDCG@10"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Effects are small (dz about 0.06 to 0.11), and parent-child reranking was not tested"

## Relations
- supports:: [[Deterministic keys on spans with verbatim output]]
- supports:: [[Route-first hierarchy loses to flat search]]
- tested-by:: [[Real heading paths beat LLM-written context and shuffled placebo paths]]
- tested-by:: [[Route to section then rank lost to flat search by 0.033 and 0.015]]
