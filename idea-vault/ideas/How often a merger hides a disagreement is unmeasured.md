---
type: idea
status: proposed
area: evaluation
needs: model-calls
---
No paper reports the rate at which a combine step asserts agreement that the unit outputs do not support. ContextlessAI can estimate it locally by recording, per plan, the number of cross-unit conflicts found by code against the number mentioned in the final answer. The same instrumentation records the share of claims that failed the quote check, the `no_information` rate, and the claims the assembler cited or omitted with and without a reason. The engine already records what the assembler left unaccounted for and the counts of possible disagreements, so the numbers can be collected from real plans.

## Evidence
- `reports/Parallel AI agents data structures.md`: "The last ratio is a local estimate of the quantity no paper reports: how often a merger hides a disagreement."
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "No measured rate found for a merge step asserting agreement that the unit outputs do not support."
- `docs/ARCHITECTURE.md`: "so how often a merge drops a result can be measured"

## Relations
- builds-on:: [[The assembler accounts for every claim]]
- builds-on:: [[Code compares unit records before the assembler writes]]
- builds-on:: [[The join fails more often than the plan]]
- part-of:: [[cx compare measures plans against single calls blind]]
