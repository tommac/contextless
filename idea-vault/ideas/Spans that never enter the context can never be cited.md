---
type: idea
status: proposed
area: selection
needs: model-calls
---
Learning from citations has an exposure problem: a span the policy never showed cannot be cited, so citation labels cover only what was already selected. Recall requests partly solve it, since they surface items the policy did not show, but they are still chosen by the agent's behaviour and are biased positives. Propensity methods need variation in exposure, which conflicts with determinism. The notes suggest seeding any exploration from a hash of the query and policy version, in a small reserved slice of the budget, with the propensity logged, and say no estimator was verified for that design.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The main obstacle is exposure: spans that never enter the context can never be cited."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "seed any exploration from a hash of the query and policy version, restrict it to a small reserved slice of budget, and log the propensity with every decision"

## Relations
- part-of:: [[Learning from recall requests and citations]]
- part-of:: [[Counterfactual learning to rank]]
- builds-on:: [[Determinism is the instrument]]
- builds-on:: [[Recall by id or by search replaces memory]]
