---
type: idea
status: proposed
area: evaluation
needs: model-calls
---
The proposal is to evaluate tangent handling on tangents that really occurred and returned to the main thread, judged by the answers that follow. The notes found no benchmark built from naturally occurring tangents, since current topic-switch datasets mostly join separate conversations and the one branching study is synthetic. The report judges the gap real but the publishability weak, because the transcripts are private and new benchmarks are crowded.

## Evidence
- `reports/Cutting edge context memory research.md`: "Tangents that really occurred and return to the main thread, judged by downstream answers"
- `reports/Cutting edge context memory research.md`: "**Gap is real, publishability is weak**: the transcripts are private and"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "I found no benchmark built from naturally occurring tangents that return to the main thread. Current topic-switch datasets mostly concatenate conversations."

## Relations
- builds-on:: [[Threads group turns by topic]]
- builds-on:: [[Unrelated context can hurt answers]]
- builds-on:: [[Boundary detection should be judged by downstream answers]]
- source:: [[ContextBranch]]
