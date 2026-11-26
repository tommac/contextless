---
type: idea
status: built
area: plans
needs: none
---
A plan unit, and each part of an oversized message, hands back one record of fixed shape instead of free text: its output, the claims it relies on (each with evidence ids, a copied quote and a support level), the choices another unit could have made differently, what it could not settle, and notes. The engine checks every claim against the vault before the assembler or reviewer reads it, and the check only ever lowers a claim. It is the first change taken from the survey of parallel agents, whose clearest measured result is that a typed record per unit carries most of the gain of splitting work. The survey's verdict on this structure is to adopt it, with validation at the boundary and capable models.

## Evidence
- `docs/ARCHITECTURE.md`: "A plan unit, and each part of an oversized message, hands back a record rather than free text (`packet.UnitSchema`):"
- `docs/ROADMAP.md`: "whose clearest measured result is that a typed record per unit carries most of the gain of splitting work"
- `reports/Parallel AI agents data structures.md`: "Adopt, with validation at the boundary and capable models"
- `docs/ARCHITECTURE.md`: "The check only ever lowers a claim."

## Relations
- tested-by:: [[Removing the structured record cut English QA from 41.23 to 25.93]]
- tested-by:: [[JSON between agents cut a 7B model from 75.0 to 45.0 percent]]
- source:: [[LLM x MapReduce]]
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[The indexer writes cards facts and decisions with checked quotes]]
- analog-of:: [[Proof-carrying memory]]
- alternative-to:: [[One schema-checked call per item over data larger than any window]]
