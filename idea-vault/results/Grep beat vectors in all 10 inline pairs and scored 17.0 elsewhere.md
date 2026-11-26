---
type: result
status: disputed
area: recall
needs: none
measured-by: others
---
Two studies disagree on grep against vectors. On a 116-question subset of LongMemEval-S, grep beat vector search in all 10 model and harness pairs when tool results were returned inline, and vectors won 5 of 10 when results went to files. A peer-reviewed paper's grep agent scored only 17.0 on LongMemEval-S against 57.0 for its hybrid. Models, harnesses and question sets differ, so the two figures should not be averaged.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "With tool results returned inline, grep beat vector in **all 10** model-harness pairs"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Models, harnesses and question sets differ from the study above, so the two figures should not be averaged."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Harness effects were as large as retriever effects"

## Relations
- supports:: [[Benchmark scores are not comparable across papers]]
- supports:: [[Raw text plus an index the agent may choose to use]]
- source:: [[Chen et al AACL 2026 cross-scenario memory evaluation]]
