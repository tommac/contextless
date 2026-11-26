---
type: result
status: preprint
area: cost
needs: none
measured-by: others
---
With a read ratio, a re-prefill ratio and a ping interval, the study models the idle time up to which a keepalive ping is cheaper than letting the cache entry die: about 46 minutes for Anthropic's 5-minute tier at a 240-second interval, about 36 minutes for OpenAI and DeepSeek, and about 12 minutes for Google at the paper's read ratio. In its measurements a 30-second keepalive held a median of at least 98 percent cached, and a 240-second interval cut ping spend 7.8 times. It is one author's preprint and its price inputs may be stale.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "about 46 minutes for Anthropic's 5-minute tier at `τ≈240 s`, about 36 minutes for OpenAI and DeepSeek, and about 12 minutes for Google at the paper's `r=0.25`"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "A 30-second keepalive held a median of at least 98% cached in every tested group."

## Relations
- supports:: [[Message gaps set the cache hit rate]]
- supports:: [[Prefix-priced context selection]]
- source:: [[Khailo keepalive study]]
