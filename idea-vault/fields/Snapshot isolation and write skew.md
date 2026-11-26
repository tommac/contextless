---
type: field
status: external
area: plans
needs: none
---
Under snapshot isolation each transaction reads from a consistent snapshot taken at its start, the first committer wins a write-write conflict, and readers never block writers. Its known blind spot is write skew: two transactions read overlapping data, write disjoint items, and together break a constraint neither breaks alone. An isolated workspace per agent is a snapshot per writer, so two units can make individually valid changes that jointly break an invariant. A text merge does not see it, and tests or a reviewer after the merge must. No agent study measures it by that name.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Anomaly: write skew, where two transactions read overlapping data, write disjoint items, and together break a constraint neither breaks alone."
- `reports/Parallel AI agents data structures.md`: "two units make individually valid changes that jointly break an invariant. A text merge does not see it. Tests or a reviewer after the merge must."
- `reports/Parallel AI agents data structures.md`: "This is carried over from database theory; no agent study measures it by that name, and CodeCRDT's 5% to 10% semantic conflict rate is the nearest evidence."

## Relations
- analog-of:: [[Isolated workspaces beat shared mutable state]]
- analog-of:: [[Coder units work in their own git worktree]]
- analog-of:: [[A scratch merge of coder branches runs verify once]]
