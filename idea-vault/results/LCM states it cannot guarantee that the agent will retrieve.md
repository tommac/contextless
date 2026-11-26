---
type: result
status: vendor-reported
area: recall
needs: none
measured-by: others
---
LCM keeps a verbatim immutable store under a summary DAG, and its authors report an average of 74.8 against 70.3 for Claude Code on OOLONG trec_coarse, both with Opus 4.6. The paper says its lossless property means retrievable, not used: it cannot guarantee the agent will retrieve, and its expand tool is limited to sub-agents. The figures are self-reported, and the same caution applies to any design, this one included, that leaves text out and relies on the model asking for it.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "the paper says it cannot guarantee the agent will retrieve, and `lcm_expand` is limited to sub-agents"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "On OOLONG trec_coarse, Volt averaged 74.8 against 70.3 for Claude Code v2.1.4, both with Opus 4.6."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "systems keep the originals but still depend on the agent choosing to retrieve them"

## Relations
- contradicts:: [[Recall by id or by search replaces memory]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Deterministic sufficiency signals word the brief]]
- source:: [[LCM Lossless Context Management]]
