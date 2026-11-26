---
type: result
status: preprint
area: storage
needs: none
measured-by: others
---
Stashbird keeps, for every derived entity and relation, the set of episodes that support it: repeated evidence extends the set, a correction supersedes the active relation and keeps the old version, and deleting an episode drops only the relations whose support becomes empty. With this reconciliation the final state was correct in 20 of 20 update sequences against 8 of 20 without, and all checks passed across 263 deletion interventions. It is the only measured consequence found for the event-sourcing pattern, and it measures consistency, not retrieval. Extraction in Stashbird is by LLM, so the result supports the bookkeeping and not a model-free pipeline.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "With reconciliation the final state was correct in 20 of 20 update sequences, against 8 of 20 without; all checks passed across 263 deletion interventions, where deleting every affected record would have lost 441 still-supported relation occurrences on LoCoMo."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The only measured consequence found is consistency, not retrieval: Stashbird's 20 of 20 correct final states under updates against 8 of 20."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "a correction supersedes the active relation and keeps the old version"

## Relations
- supports:: [[Facts supersede facts]]
- supports:: [[Derived validity records over immutable turns]]
- supports:: [[Stale memory is a weak spot of memory frameworks]]
- supports:: [[Views with lineage instead of summaries]]
- part-of:: [[Provenance and lineage]]
- source:: [[Stashbird]]
