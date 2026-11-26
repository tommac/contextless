---
type: idea
status: rejected
area: plans
needs: none
---
Published frameworks converged from 2023 to 2026 on explicit graphs with typed state and checkpoints, and protocols such as A2A and MCP standardize the envelope between agents. The convergence is of design and not measured superiority: no framework has published a measurement of its concurrency model, and the protocols define no evidence or provenance fields. The survey says not to adopt a superstep framework, an actor runtime, A2A or topology search, because the engine already has explicit control flow in code. It also says to hold fan-out and recursive sub-planning where they are until the plan-against-single measurement shows a gain.

## Evidence
- `reports/Parallel AI agents data structures.md`: "do not adopt a superstep framework, an actor runtime, A2A or topology search: the engine already has explicit control flow in code"
- `reports/Parallel AI agents data structures.md`: "This is convergence of design, not measured superiority"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "neither defines confidence, evidence or provenance fields, so a hand-off contract still has to be designed on top"
- `reports/Parallel AI agents data structures.md`: "Do not raise fan-out beyond a handful of units or add recursive sub-planning until the plan-against-single measurement shows a gain at the current size."

## Relations
- alternative-to:: [[The engine never asks a model what to do next]]
- analog-of:: [[Bulk synchronous parallel supersteps]]
- analog-of:: [[Actor model and supervision]]
- source:: [[LangGraph]]
