---
type: idea
status: built
area: evaluation
needs: model-calls
---
`cx compare` answers each message once per arm: a single call, a plan, and optionally a budget-matched single call on a named profile, since unmatched comparisons flatter multi-agent systems. Every trial runs on its own scratch copy of the workspace, and the answers are written blind with a key of tokens, time, runs and units. It spends model calls, so it starts only with `-yes`, and nothing has been measured with it yet. Its first run on real messages decides whether the plan path is worth more work.

## Evidence
- `docs/ARCHITECTURE.md`: "It spends model calls, so it starts only with `-yes`; nothing has been measured with it yet."
- `docs/ARCHITECTURE.md`: "`-matched`, the budget-matched arm, since unmatched comparisons flatter multi-agent systems"
- `docs/ROADMAP.md`: "Next: run `cx compare` on real messages (model calls, ask first). It decides whether the plan path is worth more work."
- `reports/Parallel AI agents data structures.md`: "Include a budget-matched arm where the single call gets a larger window or higher effort, because the published record says unmatched comparisons flatter multi-agent systems."

## Relations
- builds-on:: [[Multi-agent gains are narrow and cost more tokens]]
- builds-on:: [[Plans run in waves with dependencies and shared decisions]]
- analog-of:: [[Recovery probes measure selection without a model]]
- builds-on:: [[Determinism is the instrument]]
