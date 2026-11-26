---
type: result
status: measured
area: caching
needs: none
measured-by: us
---
In the 10 source sessions, the 254 gaps between user messages have a median of 16 minutes: 37 percent are within 5 minutes, 59 percent within 30 and 68 percent within an hour. If caches can be made to work, those are the hit rates a 5-minute, 30-minute or 1-hour cache lifetime allows before keep-alives. The command cx gaps reproduces it, and sessions imported from now on keep their message times.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "254 gaps between user messages have a median of 16 minutes; 37 percent are within 5 minutes, 59 percent within 30, 68 percent within an hour and 94 percent within 12 hours."

## Relations
- supports:: [[Message gaps set the cache hit rate]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Prefix-priced context selection]]
