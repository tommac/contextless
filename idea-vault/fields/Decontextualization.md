---
type: field
status: external
area: index
needs: none
---
Decontextualization is the task, from computational linguistics, of editing a sentence so that it stands alone by resolving coreference, bridging and similar dependencies. It dates as a formal task to 2021, and later work on molecular facts adds the criteria of decontextuality and minimality. Resolving coreference before indexing gave only small retrieval gains in one test, and for this project the lesson is that any rewriting belongs in the retrieval key and never in the stored evidence.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Decontextualization as a formal task dates to Choi et al. (2021), covering coreference, bridging and related edits needed for a sentence to stand alone"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "retrieval gains were small, for example LLM2Vec nDCG 0.822 → 0.828 and bge-large unchanged at 0.789"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Coreference rewriting is most justified where passages depend heavily on earlier sentences and smaller readers are used."

## Relations
- analog-of:: [[Excerpts separate retrieval text from evidence]]
- analog-of:: [[Atomicity has a floor]]
- analog-of:: [[Excerpt readability is the main open question]]
