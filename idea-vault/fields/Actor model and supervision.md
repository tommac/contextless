---
type: field
status: external
area: plans
needs: none
---
An actor has private state and a mailbox, and on a message it can send messages, create actors and choose its behaviour for the next message, with no shared state. Erlang adds supervision: a supervisor restarts failed children under a strategy with a restart limit, after which the failure escalates, and corrupt state is not repaired in place. This suits LLM agents because a long, confused context is corrupt state, and restarting with a fresh context from the last recorded handoff is the cheap repair. No measurement of supervision trees for LLM agents was found.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "a long, confused context is corrupt state; restarting with a fresh context from the last recorded handoff is the cheap repair"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "a supervisor restarts failed children according to a strategy (one-for-one, one-for-all, rest-for-one) with a restart-intensity limit, after which the failure escalates upward"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "No published measurement of Erlang-style supervision trees applied to LLM agents (restart rates, recovery success) was found."

## Relations
- analog-of:: [[Actor mailboxes between agents]]
- analog-of:: [[A supervisor in code sets deadlines restart limits and backup units]]
- analog-of:: [[Each call gets a fresh packet built by code]]
- analog-of:: [[A run that compacts is thrown away]]
