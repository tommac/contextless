---
type: result
status: preprint
area: recall
needs: none
measured-by: others
---
Pichay, a proxy that evicts old tool results and leaves retrieval handles in their place, had a fault rate of 0.0254 percent in offline replay. One live session of 681 turns had 97 percent faults and thrashed, and one collapse dropped the cache hit rate from 100 to 25 percent for a turn. The evidence is mixed and comes from a single user: paging is cheap on average and can fail badly in one session, and structural edits to earlier context can cost more than they save.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "offline replay gave a 0.0254% fault rate (354 faults in 1.393M evictions), but one 681-turn live session had 97% faults and thrashed"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "A collapse dropped the cache hit rate from 100% to 25% for one turn, costing about one recompute of 105K tokens."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Pichay and CMV both show that structural edits can cost more than they save."

## Relations
- supports:: [[Recall as paging]]
- supports:: [[The append-only history as the cache prefix]]
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
