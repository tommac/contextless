---
type: field
status: external
area: plans
needs: none
---
Work on private copies, check at commit that nothing read has changed, then write: no locks are held, and a failed check restarts the work. It is good when conflicts are rare and wastes the loser's work when they are not, which is expensive when a step costs minutes and dollars. Cursor found a version check at commit better than locks, but it did not remove confusion, and agents without ownership drifted to small safe changes. The survey's verdict on the optimistic version check is that it is good as a commit-time validation and not as a coordination scheme.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Good as a commit-time validation, not as a coordination scheme"
- `reports/Parallel AI agents data structures.md`: "Field report: better than locks, did not remove confusion, and agents without ownership drifted to small safe changes"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Three phases: read (work on private copies), validate (check that nothing read has changed), write."

## Relations
- alternative-to:: [[Leases and fencing tokens]]
- analog-of:: [[Coder units work in their own git worktree]]
- analog-of:: [[Isolated workspaces beat shared mutable state]]
