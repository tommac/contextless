---
type: source
status: external
area: plans
needs: none
year: 2025
url: https://www.anthropic.com/engineering/multi-agent-research-system
---
Anthropic's June 2025 engineering post on its Research feature: a lead agent plans, spawns 3 to 5 subagents in parallel with explicit briefs, receives condensed findings and hands the result to a citation pass. It reports a 90.2% gain over single-agent Opus 4 at about 15 times the tokens of a chat, a vendor self-report on an unpublished task set that is not budget-matched. Its lasting design points are the self-contained brief, storing subagent work outside the lead's context and returning lightweight references, and that only the reading is parallel while a single agent writes the report.

## Evidence
- `research_notes/Parallel AI agents data structures/production_systems.md`: "a lead agent that plans, spawns 3 to 5 subagents in parallel with explicit briefs, receives condensed findings, and hands the result to a separate citation pass"
- `reports/Parallel AI agents data structures.md`: "Vendor self-report, unpublished task set, not budget-matched"
- `research_notes/Parallel AI agents data structures/production_systems.md`: "The internal eval's task set, size and scoring rubric are not published, so the 90.2% number cannot be reproduced."

## Relations
- supports:: [[Each call gets a fresh packet built by code]]
- supports:: [[One writer per surface]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- contradicts:: [[A model summarizing one unit's output for another]]
