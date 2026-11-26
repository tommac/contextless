---
type: result
status: preprint
area: cost
needs: none
measured-by: others
---
Ingest cost differs by orders of magnitude between memory systems. EdgeMem built its LoCoMo index with no model, 0 build tokens and 1.8 minutes, while the baselines it compares with used 0.46M to 11.77M build tokens and 17 minutes to 10.1 hours. The cost of memory derived by a model is often deferred or hidden, because many papers report throughput or relative tokens instead of calls per stored item.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "LoCoMo: 0 build tokens, 1.8 min build; retrieval P95 10.73 ms. Baselines 0.46M to 11.77M build tokens and 17 min to 10.1 h"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The cost of LLM-derived memory is often deferred or hidden, because many papers report throughput or relative tokens instead of calls per stored item."

## Relations
- supports:: [[Model-free hubs over verbatim turns]]
- supports:: [[Measure the bill not the tokens]]
- supports:: [[A billed cost-quality frontier for verbatim memory]]
- source:: [[EdgeMem]]
