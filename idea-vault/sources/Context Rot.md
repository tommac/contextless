---
type: source
status: external
area: packets
needs: none
year: 2025
url: https://research.trychroma.com/context-rot
---
A vendor report by Chroma of 14 July 2025 that tested 18 models on Needle in a Haystack, LongMemEval and Repeated Words. Performance degraded with input length, faster when the needle and the question were less similar, and even one distractor hurt. Many results appear only in charts and only two haystack topics were used, so it is vendor-reported evidence.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Performance degraded with input length, faster with lower needle-question similarity, and even one distractor hurt."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Many results appear only in charts, and only two haystack topics were used"

## Relations
- supports:: [[Unrelated context can hurt answers]]
- contradicts:: [[The append-only history as the cache prefix]]
- tested-by:: [[Every model scored higher on focused input than on full LongMemEval prompts]]
