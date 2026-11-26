---
type: idea
status: measured-kept
area: caching
needs: none
---
A cached prefix is read only if the next message arrives before the cache entry expires, so the gaps between user messages bound the hit rate of any cache lifetime. In the 10 source sessions the median gap is 16 minutes: 37 percent of gaps are within 5 minutes, 59 percent within 30 and 68 percent within an hour. Those are the hit rates a 5-minute, 30-minute or 1-hour lifetime allows before keep-alives, if caches can be made to work at all.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "254 gaps between user messages have a median of 16 minutes; 37 percent are within 5 minutes, 59 percent within 30, 68 percent within an hour and 94 percent within 12 hours."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Estimate hit probability from the gap distribution against measured retention, not only the nominal TTL."

## Relations
- builds-on:: [[Prompt cache savings are a ceiling until a live test]]
- source:: [[Khailo keepalive study]]
- tested-by:: [[254 message gaps had a median of 16 minutes]]
- tested-by:: [[No provider evicted a cache before its documented lifetime]]
- tested-by:: [[Keepalive pings break even at about 46 minutes idle on Anthropic]]
