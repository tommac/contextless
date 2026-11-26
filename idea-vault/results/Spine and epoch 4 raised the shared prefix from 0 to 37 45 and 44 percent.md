---
type: result
status: measured
area: caching
needs: none
measured-by: us
---
cx cache replays each user message and measures the prefix shared with the previous brief, with no model call. On the three real conversations in selection mode the old layout shared 0 percent. With the catalog spine and a recent-window epoch of 4, the median shared prefix was 37, 45 and 44 percent, and probe recovery did not change on any of the five sets. A shared prefix is not a cache hit: whether a provider serves it from cache is untested.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "On the three real conversations, in selection mode, the old layout shared 0 percent. With both steps at an epoch of 4 the median shared prefix was 37, 45 and 44 percent."
- `docs/CONTEXT-COMPILER.md`: "`cx cache` replays each user message and measures the prefix shared with the previous brief, with no model call."

## Relations
- supports:: [[Catalog spine and epoch window make the prefix repeat]]
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
