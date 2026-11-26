---
type: idea
status: proposed
area: selection
needs: none
---
Record, for each span that reaches the packet, which lanes and which query terms derived it, and score the span by the number of independent derivations. This borrows the counting semiring from database provenance and would give lane fusion a principled form. The research note marks it as a proposal: no measurement of provenance-based ranking or packing was found.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Provenance polynomials give a principled form for lane fusion: record for each packed span which lanes and query terms derived it, then score under a counting semiring (number of independent derivations)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "This is a proposal; nothing measured was found."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No measurement of provenance-based ranking or packing was found."

## Relations
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[A query planner and EXPLAIN for context]]
- analog-of:: [[Provenance and lineage]]
