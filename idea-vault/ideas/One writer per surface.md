---
type: idea
status: built
area: plans
needs: none
---
Each thing that can be written has exactly one writer: one process writes a workspace at a time, the engine owns commits, each coder unit has its own branch, and a single assembler writes the answer. The systems with measurements behind them pass checkable artifacts by reference, keep one writer per surface and put scheduling and reduction in code, while those that made agents share and negotiate state retreated from it. Because of this rule the vault needs no concurrency control. The survey treats it as an asset to keep.

## Evidence
- `reports/Parallel AI agents data structures.md`: "One process writes a workspace at a time, so there are no concurrent writers to the vault and no need for vault-level concurrency control."
- `reports/Parallel AI agents data structures.md`: "Systems that pass checkable artifacts by reference, keep one writer per surface, and put scheduling and reduction in code are the ones with measurements behind them"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Single-writer-per-surface (the data-discovery blackboard: main agent writes requests, helpers write to a response board only the main agent reads; LCM: sub-agents return one final answer) avoids conflict by construction."

## Relations
- tested-by:: [[Twenty lock-coordinated agents ran at the throughput of one to three]]
- alternative-to:: [[An open blackboard that every agent reads and writes]]
- alternative-to:: [[Locks held by agents]]
- source:: [[Cognition Dont Build Multi-Agents]]
- source:: [[Cursor planner and worker agents]]
