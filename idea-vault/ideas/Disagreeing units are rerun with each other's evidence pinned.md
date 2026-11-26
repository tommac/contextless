---
type: idea
status: designed
area: plans
needs: model-calls
---
When units disagree on a fact, the assembler should not pick silently. The measured remedy is to rerun the disagreeing units with each other's evidence pinned, which is how LongAgent's members corrected themselves, capped at one round. If the disagreement survives, the answer states both positions with their evidence. It is not built yet, and the roadmap holds it until `cx compare` shows the plan path is worth more work.

## Evidence
- `reports/Parallel AI agents data structures.md`: "The measured remedy is to rerun the disagreeing units with each other's evidence pinned, which is how LongAgent's members corrected themselves, capped at one round"
- `docs/ARCHITECTURE.md`: "Not built yet: rerunning units that disagree with each other's evidence pinned"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Members that hallucinated mostly corrected themselves after seeing the relevant chunk."

## Relations
- builds-on:: [[Code compares unit records before the assembler writes]]
- builds-on:: [[Quotes are pinned to the paragraphs that hold them]]
- alternative-to:: [[Debate rounds voting and confidence scores as a merge rule]]
- source:: [[LongAgent]]
