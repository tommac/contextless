---
type: idea
status: rejected
area: plans
needs: none
---
Letting units debate, vote or attach numeric confidence scores is a common way to settle disagreement between agents. Debate rarely beats chain-of-thought once the baseline is fair, its gains flatten after the second round, and aggressive judges flip correct answers to wrong. In the LLM x MapReduce ablation calibration added only 2 to 5 points, and the paper gives no rule for how scores settle a conflict. The survey says not to add any of them: support is a three-level class tied to a checkable quote, because a number the engine cannot check is not worth asking for.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Do not add debate rounds, voting among units or numeric confidence scores as a merge rule."
- `reports/Parallel AI agents data structures.md`: "so a number the engine cannot check is not worth asking for"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Against chain-of-thought or self-consistency with similar spend, the average debate effect is near zero or negative on 2024-era small and mid models."

## Relations
- tested-by:: [[Debate beat chain-of-thought in at most 13.9 percent of 36 settings]]
- tested-by:: [[Removing the structured record cut English QA from 41.23 to 25.93]]
- alternative-to:: [[Disagreeing units are rerun with each other's evidence pinned]]
- alternative-to:: [[Code compares unit records before the assembler writes]]
