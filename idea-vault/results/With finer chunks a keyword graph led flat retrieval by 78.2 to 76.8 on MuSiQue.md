---
type: result
status: measured
area: index
needs: none
measured-by: others
---
KET-RAG's keyword layer links every non-stop word to the sub-chunks that contain it, with no LLM. In the low-cost setting it beat flat retrieval on MuSiQue coverage by 50.8 to 26.4, but in the high-accuracy setting, with chunk length 150 instead of 1,200, the gap nearly closed to 78.2 against 76.8. On HotpotQA the gap stayed, at 82.2 against 74.0. The research note concludes that part of the low-cost gain is an effect of finer granularity and not of the graph.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "in the paper's high-accuracy setting (chunk length parameter 150 instead of 1,200) the gap nearly closed on MuSiQue (78.2 against 76.8) and stayed on HotpotQA (82.2 against 74.0), so part of the low-cost gain is a finer-granularity effect"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the keyword layer links every non-stop word to the sub-chunks containing it, with no LLM"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "it beat flat retrieval on MuSiQue (coverage 50.8 against 26.4, exact match 7.0 against 3.0)"

## Relations
- supports:: [[Atoms are paragraph-sized addresses into a turn]]
- supports:: [[Same-harness baselines every new structure must beat]]
- supports:: [[Model-free hubs over verbatim turns]]
