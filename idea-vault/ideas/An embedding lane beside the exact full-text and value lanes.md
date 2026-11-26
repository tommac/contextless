---
type: idea
status: proposed
area: selection
needs: download
---
An embedding lane would add a ranking by an embedding model to the exact, full-text and value lanes. The idea is in LANTERN and CogCanvas. It was considered in Stage 2b and not taken yet, because it needs a local embedding model, which is a download to ask about first. Two things in the papers suggest it can wait: LCM leaves embeddings out on purpose, and LANTERN's own misses were mostly coverage, 361 of 420, not ranking, 59 of 420.

## Evidence
- `docs/ROADMAP.md`: "It needs a local embedding model, which is a download to ask about first. LCM leaves embeddings out on purpose, and LANTERN's own misses were mostly coverage (361 of 420), not ranking (59 of 420)."
- `docs/ROADMAP.md`: "**An embedding lane** (LANTERN, CogCanvas)."

## Relations
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[Hybrid lexical and embedding retrieval beats either alone]]
- source:: [[LANTERN]]
- source:: [[CogCanvas]]
