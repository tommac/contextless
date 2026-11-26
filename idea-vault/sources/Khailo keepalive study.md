---
type: source
status: external
area: caching
needs: none
year: 2026
url: https://arxiv.org/pdf/2607.19214
---
A single-author preprint of July 2026 that tested cache retention on Anthropic, OpenAI, Google and DeepSeek with prefixes of 40k and 100k tokens and idle gaps up to 40 minutes. No provider evicted before its documented lifetime, and keepalive pings held caches warm at a cost the paper models with a break-even idle time. The research notes flag that it is one author's study, that its price inputs may be stale, and that its Google read ratio conflicts with Vertex's documented discount.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Khailo tested Anthropic, OpenAI, Google and DeepSeek, with prefixes of 40k and 100k tokens and idle gaps up to 40 minutes."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "The keepalive result is one author's study, and its price inputs may be stale."

## Relations
- supports:: [[Message gaps set the cache hit rate]]
- tested-by:: [[No provider evicted a cache before its documented lifetime]]
- tested-by:: [[Keepalive pings break even at about 46 minutes idle on Anthropic]]
