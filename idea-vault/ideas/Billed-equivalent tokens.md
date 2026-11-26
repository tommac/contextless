---
type: idea
status: proposed
area: cost
needs: none
---
Billed-equivalent tokens weight each class of token by its price: uncached input, cache writes, cache reads and output each get their own multiplier. Raw counts for every class are published so the results can be repriced when prices change. Fewer visible tokens need not cut the bill, so this metric and not packet size is what the project's cost claims should rest on.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "**Billed-equivalent tokens** = uncached input + (cache-write multiplier × cache-write tokens) + (cache-read multiplier × cache-read tokens) + (output multiplier × output tokens)."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Publish raw counts for every class, so results can be repriced."
- `reports/Cutting edge context memory research.md`: "Fewer visible tokens need not cut the bill"

## Relations
- builds-on:: [[Usage records cannot yet produce billed tokens]]
- builds-on:: [[Prompt cache savings are a ceiling until a live test]]
- part-of:: [[A billed cost-quality frontier for verbatim memory]]
- supports:: [[Prefix-priced context selection]]
- source:: [[Token Reduction Is Not Cost Reduction]]
- builds-on:: [[Measure the bill not the tokens]]
