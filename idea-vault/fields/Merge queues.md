---
type: field
status: external
area: plans
needs: none
---
A merge queue keeps a mainline that always passes its tests by testing each change against the tip it will land on, one at a time or in batches. It trades landing throughput for a green mainline, as a deliberate serial stage bounded by test time per batch. Cursor's integrator, the nearest reported analogue with agents, became a bottleneck at hundreds of workers and was removed. ContextlessAI's proposed scratch check is a merge queue of length one in which nothing lands.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "It is a deliberate serial stage; throughput is bounded by test time per batch."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Merge queues give correctness at the cost of a serial stage that became a bottleneck at hundreds of agents."
- `reports/Parallel AI agents data structures.md`: "it is a merge queue of length one without the bottleneck that sank Cursor's integrator"

## Relations
- analog-of:: [[A scratch merge of coder branches runs verify once]]
- analog-of:: [[An LLM integrator that merges coder branches on its own]]
- alternative-to:: [[Optimistic concurrency control]]
