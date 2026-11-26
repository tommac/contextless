---
type: idea
status: designed
area: plans
needs: none
---
An append-only event log with checkpoints and recorded step results gives resume, replay and audit. Because rerunning an LLM step produces a different output, the recorded output is the truth and a step is never regenerated on recovery. The survey's verdict is to adopt this at unit granularity: the unit record is the unit of durability, an interrupted plan reuses every stored record and reruns only the rest, and coder commits are keyed by unit id. Records are already stored with the plan, but resuming a plan from them is not built yet.

## Evidence
- `reports/Parallel AI agents data structures.md`: "For recovery, the unit record is the unit of durability."
- `reports/Parallel AI agents data structures.md`: "Re-running an LLM unit produces a different output, so recovery must reuse the recorded result, and commits should be idempotent and keyed by unit id so a retry or duplicate cannot land twice."
- `reports/Parallel AI agents data structures.md`: "Checkpoint unit is the whole step, so a crash repeats it"
- `docs/ARCHITECTURE.md`: "Not built yet: rerunning units that disagree with each other's evidence pinned; resuming a plan from stored records"

## Relations
- analog-of:: [[Durable execution by journal and replay]]
- analog-of:: [[MapReduce re-execution and atomic commit]]
- part-of:: [[Append-only storage]]
- builds-on:: [[Units return typed records with checked claims]]
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- source:: [[LangGraph]]
