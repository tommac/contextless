---
type: idea
status: external
area: selection
needs: none
---
Adaptive-k, peer reviewed at EMNLP 2025, sorts similarity scores, cuts at the largest gap within the top 90 percent and adds a fixed buffer of five, with no model call or tuning. It kept about 70 percent context recall with 2 to 10 times fewer tokens than full context. It has only been shown on embedding similarities, and this engine's fused scores are rank-derived and carry almost no gap information, so the transfer is untested. The notes prefer a stopping rule inside the selector: stop when the marginal gain per token falls below a threshold.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Sort similarity scores, cut at the **largest gap** (searched within the top 90%), add a fixed buffer of B = 5. No LLM calls or tuning."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "RRF values are rank-derived and carry almost no gap information, so compute the gap on normalized lane scores."

## Relations
- alternative-to:: [[Ranked atoms are sent as excerpts under a token budget]]
- alternative-to:: [[Budgeted selection needs a monotone submodular objective]]
- part-of:: [[A query planner and EXPLAIN for context]]
