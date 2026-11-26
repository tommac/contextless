---
type: result
status: vendor-reported
area: caching
needs: none
measured-by: others
---
As fetched on 2026-11-18, Anthropic needs a minimum of 512 to 4,096 tokens by model, keeps entries 5 minutes by default with a 1-hour option, and charges reads at 0.1 times input and writes at 1.25 or 2 times. OpenAI GPT-5.6 and later need 1,024 visible input tokens, keep entries 30 minutes by default, and charge reads at 0.1 and writes at 1.25 or 2.5 times. xAI has no fixed lifetime and can evict at any time. Prompts under the minimum are processed uncached with no error, so a short stable prefix saves nothing.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "512 to 4,096 tokens by model | 5 min default; 1 h option; refreshed on use"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "1,024 visible input tokens | `30m` default; `12h` in Responses | Read 0.1x, or 0.05x on GPT-6.1 Sol and Terra and GPT-6.2 Astra and Sol. Write 1.25x (30m), 2.5x (12h)"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Shorter prompts are processed uncached with no error."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "No fixed TTL; can be evicted at any time"

## Relations
- supports:: [[Message gaps set the cache hit rate]]
- supports:: [[Prefix-priced context selection]]
- supports:: [[Measure the bill not the tokens]]
