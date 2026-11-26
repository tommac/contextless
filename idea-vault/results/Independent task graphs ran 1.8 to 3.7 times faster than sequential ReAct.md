---
type: result
status: measured
area: plans
needs: none
measured-by: others
---
LLMCompiler measured wall-clock speedups against sequential ReAct of 1.80 times on HotpotQA, 3.74 on Movie Recommendation and 2.15 on ParallelQA, with equal or better accuracy and lower cost. OpenAI's built-in parallel function calling gave 1.61, 2.76 and 1.86 on the same three. All were measured on 2023 models, and no re-measurement was found on 2025 to 2026 models, where native parallel tool calling makes the baseline faster.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "HotpotQA 62.00% accuracy, 1.80×, 3.37× cheaper; Movie Recommendation 77.13% (ReAct 68.60%), 3.74×, 6.73× cheaper; ParallelQA 89.38%, 2.15×, 4.65× cheaper"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "OpenAI's built-in parallel function calling gave 1.61×, 2.76× and 1.86× on the first three."
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "All speedups above were measured on 2023-generation models (GPT-3.5/4, LLaMA-2)."

## Relations
- supports:: [[A task list with dependency placeholders resolved by code]]
- supports:: [[Plans run in waves with dependencies and shared decisions]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- source:: [[LLMCompiler]]
