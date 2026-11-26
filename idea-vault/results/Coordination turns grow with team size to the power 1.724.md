---
type: result
status: measured
area: cost
needs: none
measured-by: others
---
The scaling study fitted the number of turns against the number of agents n as T = 2.72 × (n + 0.5)^1.724, close to quadratic. Mean turns rose from 7.2 for a single agent to 44.3 for the hybrid topology. Cost per added agent rises with the number already present, which is what a coherency term looks like, and it supports small teams under any fixed budget. No one has fitted a scalability law to these data.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Turn count grows as T = 2.72 × (n + 0.5)^1.724 in the number of agents, close to quadratic"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "The exponent of about 1.72 means coordination turns grow almost quadratically with team size, which supports small teams under any fixed budget."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Mean turns: single agent 7.2, Independent 11.4, Decentralized 26.1, Centralized 27.7, Hybrid 44.3."

## Relations
- supports:: [[Plans run in waves with dependencies and shared decisions]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- part-of:: [[Universal Scalability Law]]
- source:: [[Towards a Science of Scaling Agent Systems]]
