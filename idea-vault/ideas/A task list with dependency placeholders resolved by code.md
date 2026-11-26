---
type: idea
status: built
area: plans
needs: none
---
The common way to express a parallel plan since 2023 is a list of tasks with ids, where an argument that depends on an earlier task is a placeholder naming that task's id, substituted by plain code once the earlier task finishes. The model plans once, code schedules and dispatches, and independent nodes run together. The survey's verdict is to adopt it; the risk is a stale up-front plan, and no source measures how often a plan must be revised. In ContextlessAI a unit names earlier units in `depends_on`, and the engine pins the upstream record verbatim into the dependent unit's packet. An explicit plan and progress ledger also beat a group chat in Magentic-One's ablation.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Adopt. Risk is a stale up-front plan; no source measures replan frequency"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "It makes the edge explicit and lets plain code, not a model, do scheduling and value substitution"
- `reports/Parallel AI agents data structures.md`: "The engine, not a model, resolves a placeholder by pinning the upstream unit's result record verbatim into the dependent unit's packet at tier 0"
- `reports/Parallel AI agents data structures.md`: "Removing Magentic-One's ledgers cut GAIA performance by 31%"

## Relations
- tested-by:: [[Independent task graphs ran 1.8 to 3.7 times faster than sequential ReAct]]
- tested-by:: [[LLMCompiler planner caused 8 percent of failures and the final answer 28]]
- builds-on:: [[The engine never asks a model what to do next]]
- alternative-to:: [[An open blackboard that every agent reads and writes]]
- source:: [[LLMCompiler]]
