---
type: field
status: external
area: plans
needs: none
---
If a fraction p of the work can run in parallel and the rest is serial, speedup on N workers is 1 / ((1 - p) + p/N), which can never pass 1/(1 - p). Nobody has fitted it to an agent system, but the shapes it predicts appear in field reports. For agents the serial fraction lives in three places: planning before fan-out, the join waiting for the slowest unit, and integration and review afterward. The formula is the standard one, stated from the literature and not re-read.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "If a fraction p of the work can be parallelised and (1 - p) is serial, speedup on N workers is S(N) = 1 / ((1 - p) + p/N), with the limit 1/(1 - p) as N grows."
- `reports/Parallel AI agents data structures.md`: "The serial fraction lives in three places: planning before fan-out, the join waiting for the slowest unit, and integration and review afterward."
- `reports/Parallel AI agents data structures.md`: "Nobody has fitted Amdahl's law or the Universal Scalability Law to an agent system, but the shapes those laws predict all appear in field reports"

## Relations
- analog-of:: [[Plans run in waves with dependencies and shared decisions]]
- analog-of:: [[Plans that buy nothing are answered in one call]]
- part-of:: [[Universal Scalability Law]]
