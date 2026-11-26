---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
TAG tested 80 queries adapted from BIRD that need world knowledge or reasoning over tables. Exact match was 0.17 for Text2SQL, 0.00 for retrieval-augmented generation, 0.02 for retrieval plus LLM rank, 0.13 for Text2SQL plus LLM and 0.55 for hand-written pipelines on LOTUS; aggregation queries were not scored. The research note reads this as the database-side statement of the problem: set-level questions are not answerable from a top-k sample. The winning pipelines were written by hand and call an LLM at query time.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Exact match: Text2SQL 0.17, retrieval-augmented generation 0.00, retrieval plus LLM rank 0.02, Text2SQL plus LLM 0.13, hand-written TAG pipelines on LOTUS 0.55."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "set-level questions are not answerable from a top-k sample"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Aggregation queries were not scored. The winning pipelines were written by hand."

## Relations
- supports:: [[Whole-conversation questions are the weak spot]]
- supports:: [[Views with lineage instead of summaries]]
- supports:: [[A per-key change ledger from the value lane]]
- part-of:: [[Semantic operators over text]]
