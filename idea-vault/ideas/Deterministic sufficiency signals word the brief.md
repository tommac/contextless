---
type: idea
status: proposed
area: selection
needs: none
---
Code can compute cheap signals of whether the packet is likely to hold the answer: the fraction of query IDF mass covered by the selected set, whether required entities and values matched in the exact or value lanes, agreement between lanes, a time cue with no dated hit. The literature says such unsupervised signals are weak predictors. So they should allocate budget and word the brief, for example by telling the model which terms were not found and to use recall, and never withhold content or suppress the recall tool.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "fraction of query IDF mass covered by the selected set; whether required entities and values were matched in the exact or value lanes; agreement between lanes in the top ranks"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The literature suggests that unsupervised signals alone will be weak, so calibrate before making behavior depend on them."

## Relations
- builds-on:: [[Recall by id or by search replaces memory]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- part-of:: [[A query planner and EXPLAIN for context]]
- tested-by:: [[Unsupervised signals correlate 0.2 to 0.3 with the benefit of retrieval]]
- alternative-to:: [[Learned sufficiency gates]]
