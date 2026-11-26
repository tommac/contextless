---
type: source
status: external
area: plans
needs: none
year: 2025
url: https://arxiv.org/abs/2512.08296
---
Kim and colleagues (Google and collaborators, December 2025, third version April 2026) compared a single agent with independent, centralized, decentralized and hybrid teams over 260 configurations. It is the survey's main controlled source for when multi-agent coordination helps and what it costs: the mean change, the gain on decomposable analysis, the loss on sequential planning, error amplification, token overhead and a near-quadratic law for turns. Its comparisons are not confirmed budget-matched, and its predictive model weakened between versions.

## Evidence
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "v3 abstract: 260 configurations, 6 benchmarks, 5 architectures (single agent plus Independent, Centralized, Decentralized, Hybrid), 3 LLM families."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "The predictive rule is weaker than first advertised: R² dropped from 0.524 to 0.373 when two benchmarks were added."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Treat the performance deltas as NOT confirmed budget-matched."

## Relations
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- supports:: [[Plans that buy nothing are answered in one call]]
- supports:: [[The join fails more often than the plan]]
