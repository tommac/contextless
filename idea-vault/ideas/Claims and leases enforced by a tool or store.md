---
type: idea
status: rejected
area: plans
needs: none
---
A claim on a task, enforced atomically by a tool or store and given an expiry, prevents duplicate work when workers pick their own tasks. It worked at small scale, with claim files and git push as the arbiter across 16 agents, and a synthetic study with agents that were not LLMs cut duplicate work from 78% to 0% with locks plus shared completion state, seen only through a secondary summary. The survey's verdict is that it is needed only when workers pick their own tasks. A planner that assigns work removes the need, so ContextlessAI does not build a claim protocol.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Needed only when workers pick their own tasks. A planner that assigns work removes the need"
- `reports/Parallel AI agents data structures.md`: "Small-scale success with claim files and git push as arbiter across 16 agents"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "A task claim that only prevents duplicated work can be a cheap advisory lease"
- `research_notes/Parallel AI agents data structures/production_systems.md`: "ran 32 concurrent synthetic deterministic agents (not LLMs): no coordination gave a 78% duplicate-work rate, 410 conflicting edits and 2.33 tasks/round"

## Relations
- analog-of:: [[Leases and fencing tokens]]
- alternative-to:: [[Locks held by agents]]
- alternative-to:: [[Plans run in waves with dependencies and shared decisions]]
- source:: [[Anthropic C compiler built by parallel Claudes]]
