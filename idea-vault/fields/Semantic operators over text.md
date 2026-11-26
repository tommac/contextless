---
type: field
status: external
area: plans
needs: none
---
Semantic operators extend the relational operators (filter, join, top-k, aggregate, group-by, map) so that each is carried out by a language model over rows of text, with an optimiser choosing cheaper plans. LOTUS, DocETL and Palimpzest are the main engines, and table-augmented generation hand-writes such pipelines over tables. They are LLM-per-tuple systems, none has been evaluated on conversational memory, and their papers as far as read do not address view maintenance. For this project they mark the opposite design: the model does the set-level work at query time, where the target system makes no model calls at query time.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the semantic-operator engines are LLM-per-tuple systems whose papers, as far as read, do not address view maintenance"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the reference algorithm is a nested loop with one LLM call per pair; on BioDEX the optimised join plus rank used 5,869 LLM calls for rank-precision@5 0.265, against an estimated 6,092,500 calls for the naive plan"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No evaluation of LOTUS, DocETL or Palimpzest on conversational memory was found."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "FEVER, 1,000 claims: 91.0% accuracy in 190.0 s batched, against FacTool 80.9% in 5,396.11 s."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Split records document id and chunk number, which is lineage in all but name"

## Relations
- part-of:: [[Query planning in databases]]
- analog-of:: [[Plans run in waves with dependencies and shared decisions]]
- analog-of:: [[One schema-checked call per item over data larger than any window]]
- alternative-to:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[Plain retrieval scored 0.00 exact match on TAG table questions]]
