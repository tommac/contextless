---
type: idea
status: rejected
area: plans
needs: none
---
Conflict-free replicated data types let several agents edit one text without coordination and always converge. Measured with LLM agents, they converged every time with zero merge failures and still produced 5% to 10% semantically wrong results, and speed moved from 21.1% faster to 39.4% slower depending on the task. The survey's verdict is do not build unless humans and agents co-edit one text in real time. For code, git branches with a conflict that stops are the better fit.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Do not build unless humans and agents co-edit one text in real time"
- `reports/Parallel AI agents data structures.md`: "Do not adopt CRDTs; git branches with a conflict that stops are the better fit for code, and convergence without validation is measured to be wrong 5% to 10% of the time."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Convergence of bytes is solved; agreement of meaning is not, so a CRDT or git merge needs a validation step after it."

## Relations
- tested-by:: [[CodeCRDT converged every time with 5 to 10 percent semantic conflicts]]
- analog-of:: [[Conflict-free replicated data types]]
- alternative-to:: [[Isolated workspaces beat shared mutable state]]
- alternative-to:: [[Coder units work in their own git worktree]]
