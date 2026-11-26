---
type: idea
status: proposed
area: index
needs: none
---
Useful structure can be built with no model: session and time links, entity and keyword hubs, and entity graphs without relations all have measured results. A hub has a simple relational form, a hub id plus the ids of its member turns, and it serves as an index over text that stays verbatim. Identifiers taken from code may suit this engine better than general entity recognition, which was the weakest retriever in one benchmark. The report counts the idea as engineering with close prior art, not as a novel contribution.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "useful structure can be built with no LLM: session and time links, entity and keyword hubs, BM25, and relation-free entity graphs all have measured results"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Hyperedges have a simple relational form here: a hub ID plus member turn IDs."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Code-oriented identifiers may suit this engine better than general NER, but that needs testing."

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Links help when questions need several pieces of evidence]]
- alternative-to:: [[Knowledge graphs lose to BM25 as a default store]]
- tested-by:: [[EdgeMem lifted flat BM25 from 49.74 to 61.01 with no LLM at build]]
- tested-by:: [[A deterministic evidence graph added 27 percent accuracy over a wiki]]
- source:: [[EdgeMem]]
- supports:: [[The link graph and its expansion lane]]
