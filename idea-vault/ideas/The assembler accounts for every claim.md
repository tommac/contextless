---
type: idea
status: built
area: plans
needs: none
---
The assembler reads the comparison, then the records, and must list each claim under `claims_used` or under `claims_omitted` with a reason. A reply that leaves claims out of both lists, or names ids that do not exist, is run once more with the engine's check, and what is still unaccounted for is recorded in `assembly.json`. This makes the two join failures seen in the literature visible to code: dropping a correct unit result, and asserting agreement the units do not support. It also lets the project measure how often a merge drops a result.

## Evidence
- `docs/ARCHITECTURE.md`: "must account for every claim: `claims_used` for what its reply relies on, `claims_omitted` with a reason for the rest"
- `docs/ARCHITECTURE.md`: "What is still unaccounted for is recorded, not hidden, in `artifacts/<plan-run>/assembly.json`"
- `reports/Parallel AI agents data structures.md`: "This makes the two join failures seen in the literature, discarding a correct unit result and asserting agreement that the units do not support, visible to code."

## Relations
- builds-on:: [[Code compares unit records before the assembler writes]]
- builds-on:: [[Units return typed records with checked claims]]
- builds-on:: [[The join fails more often than the plan]]
- tested-by:: [[LLMCompiler planner caused 8 percent of failures and the final answer 28]]
- analog-of:: [[Completeness by accounting against code-found candidates]]
