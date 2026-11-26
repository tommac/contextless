---
type: field
status: external
area: plans
needs: none
---
The classic blackboard (Hearsay-II) has three parts: a shared structured board of hypotheses, independent knowledge sources that each watch for a condition and contribute a change, and a control component that chooses which triggered source runs next. Knowledge sources never call each other, which suits ill-structured problems nobody can plan ahead. Control is the hard part. With LLM agents the hard problem moves into control and into claim discipline, which agents follow unreliably, and the 2025 revivals have thin conflict handling.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "they move the hard problem into control (who acts next) and into claim discipline, which LLM agents follow unreliably"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Limits: control/scheduling is the hard part and carries overhead; the blackboard is a single shared structure."
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "The 2025 blackboard revivals share a board of requests or messages and let agents self-select, but their conflict handling is thin"

## Relations
- analog-of:: [[An open blackboard that every agent reads and writes]]
- alternative-to:: [[Actor model and supervision]]
- alternative-to:: [[A task list with dependency placeholders resolved by code]]
