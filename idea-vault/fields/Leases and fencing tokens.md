---
type: field
status: external
area: plans
needs: none
---
A lease is a time-limited grant that expires on its own if the holder fails. A lease alone is unsafe, because a holder can pause past its expiry and then write, so the lock service issues a number that rises with every grant and the store rejects any write carrying a lower one. An LLM agent in the middle of a step is such a pause, lasting minutes, and it cannot be trusted to check itself, so the check must sit in the tool or store. No measurement of leases with fencing for LLM agents was found. The survey tells ContextlessAI not to build them, since nothing there needs a claim.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Stale writers are rejected mechanically. The check must sit in the tool or store, since the agent cannot be trusted to self-check."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "An LLM agent mid-step is Kleppmann's GC pause, lasting minutes."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "No published measurement of lease-plus-fencing for LLM agents was found; the recommendation is carried over from distributed-systems practice."

## Relations
- analog-of:: [[Claims and leases enforced by a tool or store]]
- analog-of:: [[Locks held by agents]]
- alternative-to:: [[Optimistic concurrency control]]
