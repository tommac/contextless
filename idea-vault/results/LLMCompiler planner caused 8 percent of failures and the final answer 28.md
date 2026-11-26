---
type: result
status: measured
area: plans
needs: none
measured-by: others
---
In LLMCompiler's breakdown of 36 failures on ParallelQA, the executor caused 64% (mostly math-tool errors), the final answer step 28% and the planner 8%. On that benchmark the task graph was rarely wrong, and the join was a more frequent source of error than the decomposition. It is one benchmark on 2023 models.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "LLMCompiler failure breakdown on ParallelQA (36 failures, about 10.6%): executor 64% (mostly math-tool errors such as wrong attributes or unit conversions), final answer step 28%, planner 8%."
- `reports/Parallel AI agents data structures.md`: "In LLMCompiler's failure breakdown the planner caused 8% of failures and the final answer step 28%"

## Relations
- supports:: [[The join fails more often than the plan]]
- supports:: [[The assembler accounts for every claim]]
- supports:: [[A task list with dependency placeholders resolved by code]]
- source:: [[LLMCompiler]]
