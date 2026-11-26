---
type: result
status: preprint
area: caching
needs: none
measured-by: others
---
A single-author 2026 study tested Anthropic, OpenAI, Google and DeepSeek with prefixes of 40k and 100k tokens and idle gaps up to 40 minutes, and no provider evicted before its documented lifetime. Anthropic was warm through 300 seconds and gone by 360; OpenAI was about 50 percent warm at 15 to 20 minutes and fully cold by 30. Google stayed between 33 and 83 percent warm at every gap, which the author treats as routing probability and not retention. So the nominal lifetime is a fair first estimate, and hit probability should still be estimated from measured retention.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "No provider evicted before its documented TTL"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Anthropic was warm through 300 s and gone by 360 s. DeepSeek collapsed by 540 to 600 s. OpenAI was about 50% warm at 15 to 20 minutes and fully cold by 30."

## Relations
- supports:: [[Message gaps set the cache hit rate]]
- supports:: [[Prefix-priced context selection]]
- source:: [[Khailo keepalive study]]
