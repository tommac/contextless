---
type: field
status: external
area: packets
needs: none
---
The design borrows from databases the planner that turns a query into a plan, and the EXPLAIN command that prints the plan. Applied to context, the plan is the list of passes that build a packet: candidate lanes, expansion, budgeted selection, a cache-friendly layout, then a cost and time estimate per provider. The design counts this as the first of its unusual ideas.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "**Build each packet like a database query planner**: candidate lanes, expansion, budgeted selection, a layout that keeps the prompt cache warm, then a cost and time estimate per provider."
- `docs/CONTEXT-COMPILER.md`: "**A query planner and `EXPLAIN` for context** (section 4)."

## Relations
- analog-of:: [[A query planner and EXPLAIN for context]]
- analog-of:: [[Packets are sized before the model starts]]
