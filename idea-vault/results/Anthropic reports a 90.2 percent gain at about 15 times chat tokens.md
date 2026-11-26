---
type: result
status: vendor-reported
area: plans
needs: none
measured-by: others
---
Anthropic reports that a Claude Opus 4 lead with Claude Sonnet 4 subagents outperformed single-agent Opus 4 by 90.2% on an internal research eval, and that multi-agent systems use about 15 times the tokens of a chat. The task set is unpublished and the comparison is not budget-matched. The two figures come from different measurements, so they should not be combined into a gain per token.

## Evidence
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "multi-agent beat single-agent Opus 4 by 90.2%; agents use about 4x the tokens of chat and multi-agent systems about 15x"
- `research_notes/Parallel AI agents data structures/production_systems.md`: "The 90.2% figure and the 15x figure are from different measurements (an internal eval and a usage comparison)"
- `reports/Parallel AI agents data structures.md`: "Vendor self-report, unpublished task set, not budget-matched"

## Relations
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- supports:: [[Plans run in waves with dependencies and shared decisions]]
- source:: [[Anthropic multi-agent research system]]
