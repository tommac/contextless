---
type: idea
status: built
area: plans
needs: none
---
One isolated workspace per writer (a worktree, clone or virtual machine) with merge or review at commit has the best field support of the structures the survey compares: writers never collide physically, and conflicts surface and stop. Shared mutable boards, locks held by agents and peer chat failed or showed no benefit in the field. Isolation does not prevent semantic conflicts or duplicate work, so the verdict is to adopt it and add a check after merge. No measured conflict rates have been published.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Adopt; add a check after merge"
- `reports/Parallel AI agents data structures.md`: "Isolation does not prevent semantic conflicts or duplicate work"
- `reports/Parallel AI agents data structures.md`: "the field evidence against shared mutable coordination state is the most consistent finding in this survey"
- `research_notes/Parallel AI agents data structures/production_systems.md`: "The dominant mechanism is physical isolation before the fact (one worktree, VM or repo copy per agent) plus partitioning by file ownership; almost no product performs an automated semantic merge."

## Relations
- tested-by:: [[Twenty lock-coordinated agents ran at the throughput of one to three]]
- tested-by:: [[CodeCRDT converged every time with 5 to 10 percent semantic conflicts]]
- alternative-to:: [[An open blackboard that every agent reads and writes]]
- alternative-to:: [[CRDT merges for agent edits]]
- analog-of:: [[Snapshot isolation and write skew]]
- source:: [[Cursor planner and worker agents]]
