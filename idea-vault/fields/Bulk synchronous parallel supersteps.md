---
type: field
status: external
area: plans
needs: none
---
In the Pregel model every node triggered in a step runs concurrently, and all wait at a synchronization barrier before state is committed and the next step starts. LangGraph and Microsoft Agent Framework both use it and checkpoint at the barrier. Its cost is that a fast path cannot advance until the slowest node in the step completes, and work inside a long node is not recoverable. ContextlessAI's waves have the same shape: a simple barrier per wave, whose cost is waiting for the slowest unit and should be measured before anything more elaborate is considered.

## Evidence
- `reports/Parallel AI agents data structures.md`: "A simple barrier per wave is adequate at this scale; its cost is waiting for the slowest unit, which should be measured before anything more elaborate is considered."
- `research_notes/Parallel AI agents data structures/frameworks_state_structures.md`: "nodes triggered in the same superstep run concurrently behind a synchronization barrier, with typed state channels or buffered state updates"
- `research_notes/Parallel AI agents data structures/frameworks_state_structures.md`: "There is a direct trade-off between barrier latency and checkpoint granularity."

## Relations
- analog-of:: [[Plans run in waves with dependencies and shared decisions]]
- analog-of:: [[An append-only run log with recorded step results]]
- alternative-to:: [[Actor model and supervision]]
