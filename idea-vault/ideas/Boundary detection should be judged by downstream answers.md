---
type: idea
status: proposed
area: evaluation
needs: none
---
Detecting where a topic begins and ends should be judged by the quality of later answers and by token use, not by boundary F1. Boundary F1 depends on how many boundaries a method emits: the choice of threshold moved scores more than the choice of method, and periodic or random baselines could compete when their density matched. Stronger models detect clean pivots well, but real conversational boundaries and thread reconstruction remain unreliable.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Boundary detection should be judged by downstream answer quality and token use, because boundary F1 is density-sensitive."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Threshold choice moved scores more than method choice, and periodic or random baselines could compete when density matched."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Stronger models detect clean topic pivots well, yet real conversational boundaries, thread reconstruction and context carryover remain unreliable."

## Relations
- builds-on:: [[Threads group turns by topic]]
- part-of:: [[R13 Tangent filter against branching]]
- tested-by:: [[o3 detected topic pivots at F1 0.973 on TopiOCQA and 0.763 on MSC]]
- tested-by:: [[Reply links reached Shen F 0.90 but exact thread recovery only 45.3 percent]]
- source:: [[SeCom]]
