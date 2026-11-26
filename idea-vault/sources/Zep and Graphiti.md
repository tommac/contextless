---
type: source
status: external
area: structure
needs: none
year: 2025
url: https://arxiv.org/html/2501.13956
---
Zep, whose self-hosted form is Graphiti, stores a knowledge graph whose edges carry both valid time and transaction time. When a new edge contradicts an older one the older edge is closed at the newer edge's valid time and kept, and raw episodes stay linked to the facts derived from them. Its headline numbers are vendor-run with no ablation that isolates the graph, but its validity fields are the most developed model of supersession for records derived from immutable turns.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "on a temporally overlapping contradiction the older edge is invalidated at the newer edge's valid time, preserving history"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "a fact or edge can be closed by a later record while the source turn remains untouched. Zep's bitemporal fields map cleanly to this."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "It has **no ablation** isolating graph traversal, and single-session-assistant questions fell 17.7% with GPT-4o"

## Relations
- supports:: [[Facts supersede facts]]
- part-of:: [[Knowledge graphs lose to BM25 as a default store]]
- analog-of:: [[Bitemporal data model]]
- alternative-to:: [[Mem0]]
- tested-by:: [[Zep reports 71.2 against 60.2 for full context on LongMemEval-S]]
- tested-by:: [[Mem0 and Zep LoCoMo scores were disputed and never reconciled]]
