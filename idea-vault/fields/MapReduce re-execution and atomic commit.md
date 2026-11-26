---
type: field
status: external
area: plans
needs: none
---
MapReduce recovers from failure by re-executing a failed task, and each task writes to private temporary files and commits by atomic rename, so only one execution's output is visible. With deterministic functions the output equals that of a sequential run, and with non-deterministic operators only a weaker guarantee holds: different tasks may correspond to different runs. LLM workers are not deterministic, so what survives is immutable inputs, an isolated workspace per worker, one atomic output per task, an idempotent commit, and the recorded output as the source of truth. The reduce step therefore has to check consistency across tasks, and a merge is safe to regroup only if it is associative and commutative. Backup duplicates of stragglers carry over too, unmeasured for LLM work.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "LLM workers are not deterministic, so the properties that survive are immutable inputs, isolated per-worker workspaces, one atomic output per task, and idempotent commit, with the recorded output (not re-execution) as the source of truth."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "The reduce step therefore has to check cross-task consistency, which is real work and part of the serial fraction."
- `reports/Parallel AI agents data structures.md`: "Free-text summarize-the-summaries is neither associative nor commutative, so order and grouping change the result."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "No source measured the benefit of speculative duplicate agents (backup tasks) for LLM work."

## Relations
- analog-of:: [[Code compares unit records before the assembler writes]]
- analog-of:: [[An append-only run log with recorded step results]]
- analog-of:: [[Units return typed records with checked claims]]
- analog-of:: [[Turns are stored verbatim and never rewritten]]
- analog-of:: [[Coder units work in their own git worktree]]
