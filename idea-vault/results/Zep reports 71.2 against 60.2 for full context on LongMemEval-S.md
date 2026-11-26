---
type: result
status: vendor-reported
area: structure
needs: none
measured-by: others
---
Zep's own paper reports 71.2 percent on LongMemEval-S with gpt-4o against 60.2 for full context, using about 1.6k context tokens against 115k. The figures are vendor-run, there is no ablation that isolates the graph, and Zep trails full context on single-session-assistant questions. On the DMR benchmark its margin over the full conversation is negligible.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "LongMemEval-S gpt-4o 71.2 vs full context 60.2; gpt-4o-mini 63.8 vs 55.4; about 1.6k context tokens vs 115k"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "DMR margins are negligible (gpt-4o-mini 98.2 vs full conversation 98.0)"
- `reports/Cutting edge context memory research.md`: "No ablation isolating the graph"

## Relations
- supports:: [[Unrelated context can hurt answers]]
- contradicts:: [[Knowledge graphs lose to BM25 as a default store]]
- source:: [[Zep and Graphiti]]
