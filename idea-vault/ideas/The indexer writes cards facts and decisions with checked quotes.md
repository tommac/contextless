---
type: idea
status: built
area: index
needs: none
---
After an answer, in the background and off the critical path, an indexer run reads the new turns and returns ledger operations: thread assignment, new entities, facts and decisions, and a card per turn. The model does not write the vault itself. The engine checks each quote against the verbatim turn before it applies an operation and commits. So everything a model adds to the vault is an index entry tied to exact stored text.

## Evidence
- `docs/ARCHITECTURE.md`: "An `indexer` run reads the new turns and returns ledger operations: thread assignment, new entities, facts and decisions, and a card per turn. The engine checks each quote against the verbatim turn before applying it and commits."
- `docs/ARCHITECTURE.md`: "L2 stated facts with their source turn"

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[The engine never asks a model what to do next]]
- part-of:: [[Write-time intelligence and read-time determinism]]
- part-of:: [[Proof-carrying memory]]
- alternative-to:: [[CogCanvas]]
- source:: [[BookWriter]]
