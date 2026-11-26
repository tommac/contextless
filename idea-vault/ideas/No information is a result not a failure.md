---
type: idea
status: built
area: plans
needs: none
---
A unit that finds nothing for its task returns the status `no_information` and says what it looked for, instead of inventing an answer. Code leaves such units out of the comparison, and the assembler sees one line for each. The survey calls this explicit status the measured remedy for units that fabricate from irrelevant input.

## Evidence
- `docs/ARCHITECTURE.md`: "A unit that finds nothing for its task says so with `no_information` instead of inventing an answer."
- `reports/Parallel AI agents data structures.md`: "The explicit `no_information` status is the measured remedy for units that fabricate from irrelevant input."
- `docs/ARCHITECTURE.md`: "a `no_information` unit is one line with what it looked for"

## Relations
- part-of:: [[Units return typed records with checked claims]]
- tested-by:: [[Refusal training raised correct refusals from 51.0 to 78.6 percent]]
- source:: [[LLM x MapReduce]]
- source:: [[LongAgent]]
