---
type: idea
status: external
area: structure
needs: model-calls
---
The claim that branching a conversation raises answer quality is treated as unproven. Its only measured support is a small gain on synthetic scenarios with a GPT-4 judge and no token counts, and that gain is confounded with simply removing distractors. The clearest measured benefit of isolating a tangent is a smaller context, so the report tests a filter on the linear log first.

## Evidence
- `reports/Cutting edge context memory research.md`: "Treat as unproven; test a filter on the linear log first"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "The clearest measured benefit of tangent isolation is smaller context."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "I found no user study comparing branching with linear chat on both answer quality and tokens."

## Relations
- builds-on:: [[Unrelated context can hurt answers]]
- alternative-to:: [[Threads group turns by topic]]
- tested-by:: [[Branching cut context 58 percent and raised a judged score 2.5 percent]]
- source:: [[ContextBranch]]
