---
type: result
status: claimed
area: cost
needs: none
measured-by: us
---
From documented prices, a cache write pays back after more than (w - 1) / (1 - r) reads, where w is the write price and r the read price as multiples of normal input. With reads at 0.1, a write at 1.25 needs one read and a write at 2 or 2.5 needs two; for a single possible reuse, writing at 1.25 pays when the chance of reuse exceeds about 28 percent. These are derivations by the research notes' author from documented prices, not published or measured results.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "For one possible future reuse, writing at `w=1.25, r=0.1` pays when reuse probability exceeds about 28%."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "With `r=0.1`: `w=1.25` needs 1 read, `w=2` needs 2, `w=2.5` needs 2."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "These are my derivations from the documented prices, not published results."

## Relations
- supports:: [[Prefix-priced context selection]]
- supports:: [[Message gaps set the cache hit rate]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- part-of:: [[Rent-or-buy decisions]]
