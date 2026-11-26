---
type: result
status: measured
area: plans
needs: none
measured-by: others
---
On Finance-Agent, a decomposable analysis task where the single agent scored 0.349, a centralized multi-agent team scored 0.631, a relative gain of 80.8%. It is the largest controlled gain in the survey, and it marks the shape where splitting work pays: independent sub-analyses and a weak single-agent baseline. The result is not confirmed budget-matched, and coordination gave diminishing or negative returns once the single-agent baseline passed about 45%.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Change against the single-agent baseline ranges from +80.8% (Finance Agent, centralized: 0.631 vs 0.349) to -70.0% (PlanCraft, independent)."
- `reports/Parallel AI agents data structures.md`: "Measured, controlled; not confirmed budget-matched; predictive model weakened from R² 0.524 to 0.373 between versions"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Capability saturation: coordination gives diminishing or negative returns once the single-agent baseline exceeds about 45%"

## Relations
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- supports:: [[Plans run in waves with dependencies and shared decisions]]
- source:: [[Towards a Science of Scaling Agent Systems]]
