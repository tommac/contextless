---
type: idea
status: rejected
area: plans
needs: none
---
Giving agents a lock on shared coordination state buys exclusive access in theory. In the one detailed field report the locks were held too long, never released or lost on a crash, agents wrote without taking them, and twenty agents ran at the throughput of one to three. The survey's verdict is do not build: a claim has to be enforced in a tool or store, never in a prompt. In ContextlessAI the planner assigns units and the engine owns commits, so there is nothing for a lock to protect.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Measured failure: held too long, never released, lost on crash; 20 agents at the throughput of 1 to 3"
- `reports/Parallel AI agents data structures.md`: "Do not build agent-held locks, leases with fencing tokens or a claim protocol; the planner assigns units, the engine owns commits, and one process writes, so there is nothing for a claim to protect."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Enforce claims in a tool (database row, git ref, atomic file create), never in the prompt."

## Relations
- tested-by:: [[Twenty lock-coordinated agents ran at the throughput of one to three]]
- alternative-to:: [[One writer per surface]]
- alternative-to:: [[Claims and leases enforced by a tool or store]]
- analog-of:: [[Leases and fencing tokens]]
- source:: [[Cursor planner and worker agents]]
