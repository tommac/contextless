---
type: idea
status: proposed
area: selection
needs: none
---
Use an ordinary compressor to estimate how much a candidate span adds to the spans already selected: the extra compressed bytes it contributes. That number could serve as the redundancy term in greedy packing where MMR would use embedding similarity, with no model. The research note found no retrieval measurement of a compression distance, only text classification, and notes that short spans are a known weak spot of compressors.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A compression distance is a model-free novelty estimate: the extra compressed bytes a span adds to the already selected set."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "It could serve as the redundancy term in greedy packing where MMR would use embedding similarity."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No measured use of description length or compression distance for context selection was found."

## Relations
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- alternative-to:: [[Budgeted selection needs a monotone submodular objective]]
- alternative-to:: [[Slot allocation over deterministic aspects]]
- analog-of:: [[Minimum description length]]
