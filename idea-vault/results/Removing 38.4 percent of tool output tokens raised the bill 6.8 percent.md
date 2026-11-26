---
type: result
status: preprint
area: cost
needs: none
measured-by: others
---
In 2,848 Claude Code runs with provider-billed cost, the compressor arm RTK-ML cut estimated raw tool-output tokens by 38.4 percent and the billed cost rose 6.8 percent, with a 95 percent interval of 2.8 to 11.3. Task-level token reduction barely predicted cost change. The paper attributes the increase to changed trajectories such as extra searches, re-reads and turns, and does not demonstrate cache invalidation.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "RTK-ML cut estimated raw tool-output tokens by 38.4% and still cost more. Task-level reduction barely predicted cost change (Pearson r=0.15, CI crossing zero)"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "| RTK-ML | +6.8% (+2.8, +11.3) | 97.5% | 4.69 | 1.051 (1.006, 1.100) |"
- `docs/CONTEXT-COMPILER.md`: "removing 38.4 percent of tool-output tokens raised the billed cost by 6.8 percent"

## Relations
- supports:: [[Measure the bill not the tokens]]
- supports:: [[A billed cost-quality frontier for verbatim memory]]
- source:: [[Token Reduction Is Not Cost Reduction]]
