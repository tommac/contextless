---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
CogCanvas reports that removing its LLM reranker cost 11.5 points, the largest of its ablations. It is the strongest figure for a reranker in the roadmap. LANTERN's measurement points the other way, and ContextlessAI has not taken a reranker. The figure is the paper's own ablation of its own system; the roadmap gives no benchmark name and does not say whether the paper was peer reviewed.

## Evidence
- `docs/ROADMAP.md`: "CogCanvas's largest ablation (-11.5 points without it), but LANTERN measured +1.9 points, not significant"

## Relations
- supports:: [[An LLM reranker on top of fused ranking]]
- source:: [[CogCanvas]]
