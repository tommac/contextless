---
type: idea
status: proposed
area: selection
needs: model-calls
---
An LLM reranker would spend one more model call per message to reorder what the code ranked. It was considered in Stage 2b and not taken yet. The papers pull two ways: it is CogCanvas's largest ablation, 11.5 points lost without it, but LANTERN measured a gain of 1.9 points that was not significant. The cost is plain for this project, which runs on subscriptions that already run out.

## Evidence
- `docs/ROADMAP.md`: "CogCanvas's largest ablation (-11.5 points without it), but LANTERN measured +1.9 points, not significant, for one more model call per message on subscriptions that already run out."

## Relations
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[Reranking pays only when the budget is tight]]
- source:: [[CogCanvas]]
- source:: [[LANTERN]]
- tested-by:: [[CogCanvas lost 11.5 points without its LLM reranker]]
- tested-by:: [[LANTERN measured 1.9 points from an LLM reranker, not significant]]
