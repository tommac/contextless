---
type: result
status: claimed
area: cost
needs: none
measured-by: others
---
In a Codex issue, one user's AWS costs for 1 to 23 July 2026 show 258.8M cache-write tokens at $1,780.33, which is 90.0 percent of $1,977.58 of model spend, against 28.6M read tokens at $15.71. The reporter says this does not prove a single cause, and it is not a controlled study. It shows the case a cache-friendly layout must avoid: paying the write premium again and again for prefixes that are never read.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "258.8M cache-write tokens at $1,780.33, or 90.0% of $1,977.58 model spend, with 28.6M read tokens at $15.71"

## Relations
- supports:: [[Measure the bill not the tokens]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Prefix-priced context selection]]
