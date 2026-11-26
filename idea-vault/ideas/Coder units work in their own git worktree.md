---
type: idea
status: built
area: plans
needs: none
---
A workspace can attach git repositories, each with a verify command and a path allowlist. A coder unit that names one runs in a fresh git worktree on its own branch inside the packet directory, told to edit only that copy and never run git. The engine does the rest: it diffs against the base, rejects the unit if any path falls outside the allowlist, commits once and runs the verify command. Nothing is merged and the user's checkout and branch are never touched. The unit runs like a MultiAIWorkspace worker, reduced to one unit.

## Evidence
- `docs/ARCHITECTURE.md`: "`git worktree add -b cx/<run> <packet>/repo <base>` from the repo's HEAD"
- `docs/ARCHITECTURE.md`: "The engine stages everything and diffs against the base (so a commit the agent made anyway is still checked), rejects the unit if any path falls outside the allowlist (or is under `.git`), and commits once."
- `docs/ARCHITECTURE.md`: "Nothing is merged: that is the user's call."
- `docs/ROADMAP.md`: "Agents never run git: the engine diffs, rejects any change outside the allowlist, commits, runs the verify command, reports branch, diffstat and verify result, and removes the worktree."

## Relations
- part-of:: [[Plans run in waves with dependencies and shared decisions]]
- builds-on:: [[Isolated workspaces beat shared mutable state]]
- builds-on:: [[One writer per surface]]
- builds-on:: [[The engine never asks a model what to do next]]
- source:: [[MultiAIWorkspace]]
