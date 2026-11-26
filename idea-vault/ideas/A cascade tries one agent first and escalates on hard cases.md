---
type: idea
status: built
area: plans
needs: none
---
The best-supported practical policy in the survey is a cascade: try the single agent first and escalate to multi-agent only on hard or failed cases. In one study it beat both pure strategies on most tasks, for example 71.2 against 62.9 and 62.3 on DS1000, while cutting tokens against always using several agents. That paper has internal inconsistencies, so the direction is supported and the exact numbers are soft. ContextlessAI's flow, in which the responder proposes a plan only when it judges one necessary, is already such a cascade, and the engine's gate makes it enforceable.

## Evidence
- `reports/Parallel AI agents data structures.md`: "The best-supported practical policy is a cascade."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "a cascade that tries the single agent first and escalates to multi-agent beat both on most tasks (for example DS1000 71.2 against 62.9 and 62.3; BigCodeBench-Hard 32.4 against 22.3 and 23.0)"
- `reports/Parallel AI agents data structures.md`: "That paper has internal inconsistencies in its summary figures, so treat the direction as supported and the exact numbers as soft."
- `reports/Parallel AI agents data structures.md`: "The current flow, in which the responder proposes a plan only when it judges one necessary, is already a cascade of the kind that beat both pure strategies; the gate makes it enforceable."

## Relations
- builds-on:: [[Multi-agent gains are narrow and cost more tokens]]
- tested-by:: [[Multi-agent token cost ran 1.6 to 6.2 times a single agent]]
- tested-by:: [[A single agent scored 0.418 against 0.369 to 0.388 at equal thinking budget]]
- analog-of:: [[A larger window is tried before any record is left out]]
