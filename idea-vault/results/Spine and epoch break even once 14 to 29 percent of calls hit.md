---
type: result
status: measured
area: caching
needs: none
measured-by: us
---
Using the figures of the project's cache replay, the report calculates that turning the catalog spine and epoch on saves 28.6, 37.9 and 28.1 percent if every call hits the cache and costs 6.4, 6.2 and 11.6 percent more if none do. The change therefore breaks even once 18, 14 and 29 percent of calls hit. This is the report author's arithmetic on a replay, not a measured hit rate, and it gives the live test its target.

## Evidence
- `reports/Cutting edge context memory research.md`: "**The change breaks even once 18, 14 and 29 percent of calls hit** (my arithmetic, at read price 0.1 and write price 1.0, excluding the CLI's own prefix)."
- `reports/Cutting edge context memory research.md`: "turning the spine and epoch on saves 28.6, 37.9 and 28.1 percent if every call hits and costs 6.4, 6.2 and 11.6 percent more if none do"
- `reports/Cutting edge context memory research.md`: "it must distinguish a hit rate above roughly 14 to 29 percent from zero"

## Relations
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Message gaps set the cache hit rate]]
- supports:: [[R2 Cache truth table]]
