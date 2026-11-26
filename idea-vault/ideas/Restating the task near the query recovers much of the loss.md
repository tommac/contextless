---
type: idea
status: proposed
area: packets
needs: model-calls
---
When unrelated context sits between a task statement and the question, repeating the task just before the query recovers much of the lost accuracy. The quality gain from isolating a tangent therefore depends on whether the task is restated near the query, as well as on how distant and how large the tangent is. The report adds restating as an arm of its tangent experiment.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Repeating the task just before the query recovered much of that loss: Gemini Flash rose from 0.128 to 0.438, and the GPT-4o-mini drop shrank from about 56% to 8%."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Quality gains depend on how distant and how large the tangent is, and on whether the task is restated near the query."

## Relations
- builds-on:: [[Unrelated context can hurt answers]]
- builds-on:: [[The brief runs from stable to volatile]]
- alternative-to:: [[Branching raises answer quality]]
- part-of:: [[R13 Tangent filter against branching]]
- tested-by:: [[Restating the task raised Gemini Flash from 0.128 to 0.438]]
