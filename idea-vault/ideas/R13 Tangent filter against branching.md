---
type: idea
status: designed
area: structure
needs: model-calls
---
The hypothesis is that filtering a linear log by thread and relevance matches explicit branching on answers while using fewer tokens than unfiltered history. Real tangents that return to the main thread are labelled in the owner's transcripts and questions are asked after the return, with restating the task near the query as one arm. It is falsified if no arm differs from full context on quality, in which case tangents matter only for cost.

## Evidence
- `reports/Cutting edge context memory research.md`: "Filtering a linear log by thread and relevance matches explicit branching on answers, using fewer tokens than unfiltered history"
- `reports/Cutting edge context memory research.md`: "Label real tangents that return to the main thread in the owner's transcripts. Ask questions after the return."
- `reports/Cutting edge context memory research.md`: "No arm differs from full context on quality, so tangents matter only for cost"

## Relations
- builds-on:: [[Threads group turns by topic]]
- builds-on:: [[Unrelated context can hurt answers]]
- builds-on:: [[Restating the task near the query recovers much of the loss]]
- builds-on:: [[Natural tangent-and-return evaluation]]
- builds-on:: [[Boundary detection should be judged by downstream answers]]
- alternative-to:: [[Branching raises answer quality]]
- source:: [[ContextBranch]]
