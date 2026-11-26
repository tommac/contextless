---
type: result
status: preprint
area: plans
needs: none
measured-by: others
---
CodeCRDT ran 600 trials (6 tasks, 50 runs per mode) of lock-free concurrent code generation on CRDTs. Convergence was 100% with zero merge failures, the semantic conflict rate was 5% to 10%, and speed ranged from 21.1% faster to 39.4% slower depending on the task. It is a single-author preprint with six tasks, read at abstract level. It is the nearest measurement of the gap between bytes converging and meaning agreeing, and of write skew between agents.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Up to 21.1% speedup on some tasks; up to 39.4% slowdown on others. 100% convergence, zero merge failures. Semantic conflict rate 5% to 10%."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "It is a single-author preprint with six tasks."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "CodeCRDT's 5% to 10% semantic conflict rate with 0% merge failures is the measurement of that gap."

## Relations
- contradicts:: [[CRDT merges for agent edits]]
- supports:: [[Isolated workspaces beat shared mutable state]]
- supports:: [[A scratch merge of coder branches runs verify once]]
- part-of:: [[Conflict-free replicated data types]]
