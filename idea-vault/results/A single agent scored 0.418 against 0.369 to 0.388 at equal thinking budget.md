---
type: result
status: measured
area: plans
needs: none
measured-by: others
---
Tran and Kiela (Stanford, April 2026) matched thinking tokens between a single agent and multi-agent variants on multi-hop question answering. At a budget of 1,000 thinking tokens the single agent averaged 0.418 against 0.369 to 0.388 for the multi-agent variants. The match is on thinking tokens only, the tasks are question answering with no tool use, and at a 100-token budget most multi-agent variants led. A second budget-matched study found self-consistency beating debate at equal budget.

## Evidence
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Where budgets are matched, a single agent usually ties or wins (average 0.418 against 0.369 to 0.388 for multi-agent variants at a 1,000 thinking-token budget on multi-hop question answering)."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Tran and Kiela match thinking tokens only, not total tokens, and cover multi-hop QA only, with no tool use."
- `reports/Parallel AI agents data structures.md`: "Measured, budget-matched; QA only, no tool use"

## Relations
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- supports:: [[cx compare measures plans against single calls blind]]
- supports:: [[A cascade tries one agent first and escalates on hard cases]]
