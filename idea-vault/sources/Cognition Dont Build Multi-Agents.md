---
type: source
status: external
area: plans
needs: none
year: 2025
url: https://cognition.com/blog/dont-build-multi-agents
---
Walden Yan's June 2025 essay argued against multi-agent systems because actions carry implicit decisions and conflicting decisions carry bad results, and copying the original task to each subagent does not fix that. The April 2026 revision keeps the objection for parallel-writer swarms and endorses setups where several agents contribute intelligence while writes stay single-threaded. The essay is an argument with examples and not a measurement. The survey takes two things from it: one writer, and shared decisions written down before units run.

## Evidence
- `research_notes/Parallel AI agents data structures/production_systems.md`: "Cognition's position changed in scope, not in principle."
- `research_notes/Parallel AI agents data structures/production_systems.md`: "A report should not cite the 2025 essay as Cognition's current view without the April 2026 qualifier."
- `reports/Parallel AI agents data structures.md`: "copying the original task to each unit does not prevent divergent implicit decisions"

## Relations
- supports:: [[One writer per surface]]
- supports:: [[Plans run in waves with dependencies and shared decisions]]
- supports:: [[The handoff document is a free-text field in the typed record]]
