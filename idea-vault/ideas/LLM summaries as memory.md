---
type: idea
status: rejected
area: compression
needs: none
---
Using summaries written by a model as the memory that later calls read is one of the ten popular ideas the report rejects for this design. Summaries tied or lost to simply masking old tool output on a coding benchmark, and compaction studies show that they drop constraints. The decision is to keep summaries out of the read path, so a summary may only ever be a navigation aid, stored as an artifact keyed to the span it covers.

## Evidence
- `reports/Cutting edge context memory research.md`: "Keep out of the read path. A summary may only ever be a navigation aid"
- `reports/Cutting edge context memory research.md`: "No summarizer is needed for coding history"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "a summary should be treated as a stored artifact keyed by source span, model, prompt and parameters"

## Relations
- contradicts:: [[Turns are stored verbatim and never rewritten]]
- part-of:: [[Lossy compression loses constraints and exact text]]
- alternative-to:: [[Digests label blocks of omitted turns]]
- alternative-to:: [[Views with lineage instead of summaries]]
- tested-by:: [[Observation masking matched LLM summaries on SWE-bench Verified]]
- tested-by:: [[Compaction damage ranged from 5 to 64 percent across 14 summarizers]]
- tested-by:: [[Compactors kept 17 percent of injected side constraints]]
