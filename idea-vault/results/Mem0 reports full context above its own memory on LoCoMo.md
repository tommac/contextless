---
type: result
status: vendor-reported
area: storage
needs: none
measured-by: others
---
Mem0's own paper puts full context at 72.90 on LoCoMo, above its graph variant at 68.44 and base Mem0 at 66.88. Mem0 used 1,764 tokens against 26,031 for full context and was much faster, so its case rests on cost and speed and not on quality. The graph variant barely moved the score and nearly doubled latency.

## Evidence
- `reports/Cutting edge context memory research.md`: "Its own table favors full context on quality"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Mem0 used 1,764 tokens vs 26,031 for full context"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Mem0's graph variant barely moved the LoCoMo judge score (68.44 vs 66.88)"

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[A short history is sent whole]]
- supports:: [[Knowledge graphs lose to BM25 as a default store]]
- contradicts:: [[Unrelated context can hurt answers]]
- source:: [[Mem0]]
