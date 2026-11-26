---
type: field
status: external
area: plans
needs: none
---
Durable execution engines keep an append-only history per workflow. The orchestration code must be deterministic, each model or tool call runs as a recorded step, and on recovery the recorded result is returned instead of calling the model again. The step itself runs at least once, so a crash after acting and before the result is recorded repeats it, and side effects need idempotency keys. The granularity of recorded steps sets the cost of a crash. The evidence is documentation and design posts, with no measured recovery or duplicate-effect rates.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "the orchestration code must be deterministic, each LLM or tool call runs as a recorded step, and on recovery the recorded result is returned instead of calling the model again"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "The granularity of recorded steps sets the cost of a crash."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "No vendor or paper was found reporting measured recovery rates, duplicate-effect rates or overhead for durable-execution-backed LLM agents"

## Relations
- analog-of:: [[An append-only run log with recorded step results]]
- analog-of:: [[The engine never asks a model what to do next]]
- analog-of:: [[Determinism is the instrument]]
- part-of:: [[Append-only storage]]
