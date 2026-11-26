---
type: source
status: external
area: plans
needs: none
year: 2023
url: https://arxiv.org/abs/2312.04511
---
A Berkeley system (December 2023, ICML 2024) in which a planner emits a graph of tool calls where a dependent task names an earlier output by placeholder, and a Task Fetching Unit with no model involved dispatches each task as soon as its dependencies resolve. It measured 1.8 to 3.7 times speedup against sequential ReAct, and its failure breakdown put 8% of failures on the planner and 28% on the final answer step. The measurements are on 2023 models, where the baseline is slower than today's native parallel tool calls.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "a Task Fetching Unit, with no model involved, dispatches each task as soon as its dependencies resolve and substitutes real values"
- `reports/Parallel AI agents data structures.md`: "Measured on 2023 models; baseline is slower than today's native parallel tool calls"

## Relations
- supports:: [[A task list with dependency placeholders resolved by code]]
- supports:: [[The join fails more often than the plan]]
- supports:: [[The engine never asks a model what to do next]]
