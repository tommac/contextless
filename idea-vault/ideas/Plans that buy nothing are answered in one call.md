---
type: idea
status: built
area: plans
needs: none
---
The engine refuses by rule a plan of one unit, or of units that would only run one after another. The message is then answered in one call, told why, and offered no plan. The reason is that sequential work is where every measured multi-agent topology lost to a single agent. A coder unit changing a repository still runs as a plan whatever its shape.

## Evidence
- `docs/ARCHITECTURE.md`: "a plan of one unit, or of units that would only run one after another, is not run (sequential work is where every measured multi-agent topology lost to a single agent)"
- `reports/Parallel AI agents data structures.md`: "A plan whose units form a pure chain is a sequential task and should run as one responder call with continuations, since that is where every measured topology lost."
- `docs/ROADMAP.md`: "refused by rule when one unit or a plain chain (answered in one call)"

## Relations
- tested-by:: [[Every multi-agent topology lost 39 to 70 percent on sequential planning]]
- tested-by:: [[Skeleton-of-Thought doubled speed and got worse on math and coding]]
- builds-on:: [[Multi-agent gains are narrow and cost more tokens]]
- builds-on:: [[A cascade tries one agent first and escalates on hard cases]]
- part-of:: [[Plans run in waves with dependencies and shared decisions]]
- analog-of:: [[A short history is sent whole]]
