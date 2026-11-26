---
type: idea
status: external
area: plans
needs: model-calls
---
In parallel agent systems the step that combines unit results is a larger source of error than the step that splits the work. Outputs aggregated without a check amplify errors, a final-answer step can drop or contradict a correct unit result, and agreement cannot be assumed on hard inputs. No source measures how often a merger papers over a real conflict, which the survey says is worth measuring locally.

## Evidence
- `reports/Parallel AI agents data structures.md`: "The combine step is a larger error source than the decomposition."
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "On that benchmark the graph was rarely wrong; the join was a more frequent source of error than the decomposition."
- `reports/Parallel AI agents data structures.md`: "No source measures the rate at which a merger papers over a real conflict; that gap matters and is worth measuring locally."

## Relations
- tested-by:: [[LLMCompiler planner caused 8 percent of failures and the final answer 28]]
- tested-by:: [[Independent agents amplified errors 17.2 times against 4.4 with an orchestrator]]
- tested-by:: [[Full consensus fell from 89.8 percent on MMLU to 29.4 on MATH]]
- source:: [[LLMCompiler]]
- source:: [[Towards a Science of Scaling Agent Systems]]
