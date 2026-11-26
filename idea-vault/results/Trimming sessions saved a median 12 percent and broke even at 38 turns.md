---
type: result
status: preprint
area: caching
needs: none
measured-by: others
---
In Contextual Memory Virtualisation, sessions form a DAG of immutable snapshots that can be branched and trimmed. Across 76 sessions from one user, trimming saved a median of 12 percent of tokens, but each trim cost a cache miss with a median penalty of 0.32 dollars and a median break-even of 38 turns. Reasoning accuracy was not measured, so the result prices the edit and says nothing about answers.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Across 76 sessions from one user, trimming saved a mean 20% and median 12% of tokens. The median cache-miss penalty was $0.32 and median break-even 38 turns. Reasoning accuracy was not measured"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "A deterministic builder should prefer append-at-tail and infrequent, batched changes to earlier context."

## Relations
- supports:: [[The append-only history as the cache prefix]]
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Catalog spine and epoch window make the prefix repeat]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
