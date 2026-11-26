---
type: field
status: external
area: plans
needs: none
---
Gunther's law gives relative capacity on N workers as C(N) = N / (1 + α(N - 1) + βN(N - 1)), where α is contention and β is coherency, the cost of keeping copies consistent. With β above zero throughput peaks and then falls. No one has fitted it to a multi-agent LLM system, yet the measured near-quadratic growth of coordination turns is what a coherency term looks like in practice. The survey proposes fitting it locally from throughput at 1, 2, 3 and 5 units on a fixed class of task, to see where adding units stops paying.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "α is contention (serialisation, queueing; linear in N), β is coherency (cost of keeping copies consistent; quadratic in N)."
- `reports/Parallel AI agents data structures.md`: "throughput at 1, 2, 3 and 5 units on a fixed class of task, which is enough to fit the contention and coherency terms of the Universal Scalability Law and see where adding units stops paying"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "No published fit of Amdahl or USL parameters to any LLM multi-agent system was found."

## Relations
- analog-of:: [[Plans run in waves with dependencies and shared decisions]]
- analog-of:: [[Multi-agent gains are narrow and cost more tokens]]
- analog-of:: [[cx compare measures plans against single calls blind]]
