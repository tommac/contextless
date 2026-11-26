---
type: idea
status: proposed
area: selection
needs: none
---
Treat a session as a patch and keep reading spans adjacent to a hit while the lexical gain per token (new query terms, new aspects, new values) stays above the average gain of jumping to the next best seed. This replaces a fixed window: EdgeMem reads exactly one neighbour each side, and SmartSearch sizes its whole pack with one global threshold. The rule is deterministic and applies per patch. The research note found no prior art for it and no model-free use of foraging theory for retrieval or packing.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Treat a session as a patch and keep reading adjacent spans while the lexical gain per token (new query terms, new aspects, new values) stays above the average gain of jumping to the next best seed."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "an adaptive, deterministic, per-patch rule is untested"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A marginal-gain stopping rule for neighbour expansion. No prior art found."

## Relations
- builds-on:: [[Neighbor expansion is a priced decision in the selector]]
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Whole-conversation questions are the weak spot]]
- analog-of:: [[Information foraging]]
- alternative-to:: [[Cover density across adjacent spans]]
- alternative-to:: [[An adaptive cutoff at the largest score gap needs no model]]
