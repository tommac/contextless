---
type: idea
status: external
area: selection
needs: none
---
Irrelevant earlier context can measurably lower answer accuracy, and removing it saves a large amount of context. Focused evidence beat full history on LongMemEval and on SeCom's benchmark, and unrelated chat cost several points on the GPQA question set. The effect is not universal, since a 2026 study found no threshold effect on two models with 1M-token windows, so whole history is neither automatically the quality ceiling nor automatically harmful.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Irrelevant prior context can measurably hurt answers, and removing it saves large amounts of context, but the quality gain from branching itself is so far small and tested only in synthetic LLM-judged settings."
- `reports/Cutting edge context memory research.md`: "Whole history is not automatically the quality ceiling"
- `reports/Cutting edge context memory research.md`: "no measurable threshold effect on two 1M-context models"

## Relations
- supports:: [[Each call gets a fresh packet built by code]]
- contradicts:: [[A larger window is tried before any record is left out]]
- tested-by:: [[Evidence sessions scored 0.870 against 0.606 for full history on LongMemEval]]
- tested-by:: [[Segment retrieval scored 88.81 at 820 tokens against 63.85 for full history]]
- tested-by:: [[Unrelated chat cost 3.5 to 13.1 percent of accuracy on GPQA]]
- source:: [[SeCom]]
- source:: [[LongMemEval]]
