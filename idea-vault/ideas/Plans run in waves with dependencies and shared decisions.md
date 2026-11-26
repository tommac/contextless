---
type: idea
status: built
area: plans
needs: none
---
A plan is a task list in which each unit has an id, a role, instructions, `needs` and `depends_on`. Every unit whose dependencies are done runs in the next wave, at most five at once, and a dependent unit gets its upstream records verbatim. The plan's `shared_decisions` (names, interfaces, formats, scope) go into every unit's packet, because the request alone does not stop parallel units from deciding the same thing differently. The engine checks the plan by rule: a dependency on a later or unknown unit is dropped, so a plan is acyclic, and coder units changing the same repository are chained. The cap of five is the convergent rule of thumb, unproven as an optimum.

## Evidence
- `docs/ARCHITECTURE.md`: "Units run in waves: every unit whose dependencies are done runs in the next wave, at most five at once, and a dependent unit gets its upstream records verbatim."
- `docs/ARCHITECTURE.md`: "since the request alone does not stop parallel units from deciding the same thing differently"
- `reports/Parallel AI agents data structures.md`: "Fan-out should be capped at about five units per wave, the convergent rule of thumb, which is unproven as an optimum"

## Relations
- builds-on:: [[A task list with dependency placeholders resolved by code]]
- builds-on:: [[Each call gets a fresh packet built by code]]
- builds-on:: [[Units return typed records with checked claims]]
- analog-of:: [[Bulk synchronous parallel supersteps]]
- tested-by:: [[Coordination turns grow with team size to the power 1.724]]
- source:: [[LLMCompiler]]
- source:: [[Cognition Dont Build Multi-Agents]]
