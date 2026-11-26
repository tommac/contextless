---
type: result
status: measured
area: plans
needs: none
measured-by: others
---
On PlanCraft, a sequential planning task, all four multi-agent topologies lost to the single agent, by 70.0%, 50.3%, 41.5% and 39.1%. On SWE-bench Verified all four also scored below the single agent's 0.522, by 2.1% to 14.9%. When steps depend on each other in sequence, splitting the work among agents hurts whatever the topology. This is the measurement behind the engine's rule that a plain chain is not run as a plan.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "On PlanCraft every multi-agent variant lost: -70.0%, -50.3%, -41.5%, -39.1%."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "hurts when steps are sequentially dependent (PlanCraft), where every topology lost 39% or more"
- `reports/Parallel AI agents data structures.md`: "on SWE-bench Verified all four topologies scored below the single agent's 0.522, by 2.1% to 14.9%"

## Relations
- supports:: [[Plans that buy nothing are answered in one call]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- source:: [[Towards a Science of Scaling Agent Systems]]
