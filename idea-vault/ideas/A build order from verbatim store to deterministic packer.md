---
type: idea
status: proposed
area: storage
needs: none
---
The notes suggest an order in which to build an immutable-turn, deterministic engine, drawn from the evidence they survey. It starts with a verbatim append-only turn store with content hashes, then deterministic metadata indexes, then BM25 with a pinned tokenizer, then model-free hubs. Dense retrieval and stored artifacts made by a model are optional later layers, and the last step is a deterministic packer that knows about the cache prefix.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "**Verbatim append-only turn store with content hashes** and stable turn IDs."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "This is the required baseline and the cheapest strong retriever."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "fixed quotas, pinned minimums, utility-per-token ordering, stable tie-breaks, and cache-prefix awareness, following ClawVM's pattern"

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[The offline floor is the baseline every selector must beat]]
- builds-on:: [[Model-free hubs over verbatim turns]]
- builds-on:: [[Every derived index version is an input to the context build]]
- builds-on:: [[Packets are sized before the model starts]]
- analog-of:: [[Content-addressed storage]]
- source:: [[ClawVM]]
