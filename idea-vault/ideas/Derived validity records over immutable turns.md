---
type: idea
status: proposed
area: storage
needs: none
---
The most transferable temporal design is a validity record derived from immutable turns: a fact or edge can be closed by a later record while the source turn stays untouched. Supersession should be an explicit relation between records that can be inspected, because letting the latest value win automatically is a policy and not evidence of truth. In one small synthetic study it was validity metadata, not ranking, that removed stale values, and assigning keys was the hard part.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The most transferable temporal design is **derived validity records over immutable turns**"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Supersession should be an explicit, inspectable relation between records."
- `reports/Cutting edge context memory research.md`: "Validity metadata beats ranking for stale values"

## Relations
- supports:: [[Facts supersede facts]]
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- analog-of:: [[Bitemporal data model]]
- tested-by:: [[BM25 with validity keys cut stale exposure from 0.802 to zero]]
- source:: [[Zep and Graphiti]]
- source:: [[Deng and Sinchaisri 2026]]
- tested-by:: [[LLM key assigners merged falsely and removed current values]]
