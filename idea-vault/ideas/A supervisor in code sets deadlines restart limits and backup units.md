---
type: idea
status: proposed
area: plans
needs: none
---
The survey takes the supervisor from the actor model as ordinary code: per-unit deadlines, restart limits, and optionally a duplicate of a straggler on another lane, taking the first valid result. A wandering agent is treated as a crashed one after a deadline, which is what the termination and repetition failures in the MAST census call for. It is the survey's lowest priority, carried over from supervision and MapReduce backup tasks and unmeasured for LLM work.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Per-unit deadlines, restart limits, and an optional duplicate of a straggler on another lane"
- `reports/Parallel AI agents data structures.md`: "Carried over from supervision and MapReduce backup tasks; unmeasured for LLM work"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "The supervisor is ordinary code holding the restart policy and the retry budget; a wandering agent is treated as a crashed one after a deadline."

## Relations
- analog-of:: [[Actor model and supervision]]
- analog-of:: [[MapReduce re-execution and atomic commit]]
- tested-by:: [[MAST census puts 44 percent of failures in system design and 24 in verification]]
- builds-on:: [[An append-only run log with recorded step results]]
- builds-on:: [[Lanes cap concurrent runs per provider across the machine]]
