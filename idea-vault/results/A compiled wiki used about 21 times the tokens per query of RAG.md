---
type: result
status: preprint
area: cost
needs: none
measured-by: others
---
In a preregistered comparison on 24 papers and 13 questions, a wiki compiled by a model used 1,651,357 tokens per query against 78,093 for single-round RAG, and took 22.0 minutes against 3.3. The wiki did beat RAG on four synthesis questions, but a decomposition baseline recovered most of that advantage at about 491k tokens per query. No advantage in citation support was established, and an early result in the wiki's favour was withdrawn after a scoring bug.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "wiki 1,651,357 tokens per query versus 78,093 for RAG (about 21×), 22.0 versus 3.3 minutes"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "A decomposition-RAG baseline recovered most of the wiki's advantage at about 491k tokens per query."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "On synthesis questions (n=4) the wiki beat single-round RAG on inter-paper mapping by +6.25 on a 10-point scale."

## Relations
- supports:: [[Model-written notes lose as storage]]
- supports:: [[Raw text beats rewritten memory]]
- source:: [[Karpathy LLM Wiki]]
