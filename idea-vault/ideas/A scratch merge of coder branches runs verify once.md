---
type: idea
status: designed
area: plans
needs: none
---
For a plan with several coder units, after all units pass verify the engine can create a scratch branch from the shared base, merge the unit branches in plan order, run the verify command once, report the result and delete the branch. Nothing lands. It would catch textual conflicts and write skew between units that each passed alone, as a merge queue of length one. It rests on argument and not on a measured conflict rate, so the survey says to build it only once measurements show that multi-coder plans are used and that their branches conflict.

## Evidence
- `reports/Parallel AI agents data structures.md`: "create a scratch branch from the shared base, merge the unit branches in plan order, run the verify command once, and report the result, deleting the scratch branch afterward"
- `reports/Parallel AI agents data structures.md`: "It rests on argument, not on a measured conflict rate, so build it only after the measurements below show multi-coder plans are used and that their branches conflict."
- `docs/ARCHITECTURE.md`: "a check that merges several coder branches on a scratch branch and runs verify once"

## Relations
- builds-on:: [[Coder units work in their own git worktree]]
- builds-on:: [[Isolated workspaces beat shared mutable state]]
- analog-of:: [[Merge queues]]
- analog-of:: [[Snapshot isolation and write skew]]
- alternative-to:: [[An LLM integrator that merges coder branches on its own]]
