---
type: source
status: external
area: cost
needs: none
year: 2026
url: https://arxiv.org/abs/2607.12161
---
Weinberger and Hozez, July 2026: 2,848 analysed Claude Code runs across 103 tasks and 7 repositories on three models, using provider-billed cost. One tool-output compressor cut estimated raw tool-output tokens by 38.4 percent and the bill rose 6.8 percent; the authors attribute that to changed agent trajectories and do not show cache invalidation. In the baseline bill, cache creation was 44.3 percent and cache reads 35.4 percent, with uncached input only 1.3 percent.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "2,848 analyzed Claude Code runs of 2,908, across 103 tasks and 7 repositories, on Haiku 4.5, Sonnet 5 and Opus 4.8, using provider-billed cost"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Baseline bill: cache creation 44.3%, cache reads 35.4%, output 10.4%, uncached input 1.3%, unattributed 8.7%."
- `docs/CONTEXT-COMPILER.md`: "| Token Reduction Is Not Cost Reduction (2026) | removing tokens can raise the bill | measure billed cost |"

## Relations
- supports:: [[Measure the bill not the tokens]]
- supports:: [[Lossy compression loses constraints and exact text]]
- tested-by:: [[Removing 38.4 percent of tool output tokens raised the bill 6.8 percent]]
- tested-by:: [[Compression cut cost 59 percent and patch applicability from 27 to 15 of 40]]
