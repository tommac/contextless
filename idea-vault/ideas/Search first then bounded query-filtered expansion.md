---
type: idea
status: proposed
area: selection
needs: none
---
A practical default drawn from the evidence on links is to search first and then expand along links in a bounded way that the query filters. Links followed regardless of the query lowered faithfulness in one small test reported on a blog, by pulling in other domains. Links that come from existing structure, such as citations, reply edges and timestamps, are cheaper and easier to audit than links a model infers, though the support for that is mainly vendor-reported.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "A practical default is search first, then bounded, query-filtered expansion."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "which the author attributes to query-agnostic links pulling in other domains"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Deterministic links from existing structure (citations, reply edges, timestamps) are cheaper and easier to audit than LLM-inferred links, though the supporting evidence here is mainly vendor-reported."

## Relations
- builds-on:: [[Links help when questions need several pieces of evidence]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- supports:: [[Model-free hubs over verbatim turns]]
- alternative-to:: [[Route-first hierarchy loses to flat search]]
- tested-by:: [[A deterministic evidence graph added 27 percent accuracy over a wiki]]
- supports:: [[Neighbor expansion is a priced decision in the selector]]
