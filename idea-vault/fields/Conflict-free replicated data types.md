---
type: field
status: external
area: plans
needs: none
---
CRDTs give strong eventual consistency: replicas that have received the same updates are in the same state, with no coordination or rollback, because the merge is commutative, associative and idempotent. With LLM agents this means every edit lands and replicas agree, with no guarantee that the result compiles or makes sense. A git conflict stops and makes the agent look, and a CRDT never stops. The simplest CRDT, an add-only set merged by union, is the safe shape for shared findings, at the cost that nothing is ever retracted and supersession needs explicit entries.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Strong eventual consistency: replicas that have received the same set of updates are in the same state, with no coordination or rollback."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "A conflict is reported and stops the merge, which for code is a feature: the agent has to look. A CRDT never stops."
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Merge is union, always safe, never loses anything; the cost is that nothing is ever retracted"

## Relations
- analog-of:: [[CRDT merges for agent edits]]
- analog-of:: [[Facts supersede facts]]
- analog-of:: [[Append-only storage]]
