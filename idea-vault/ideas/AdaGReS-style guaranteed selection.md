---
type: idea
status: rejected
area: selection
needs: none
---
AdaGReS presents greedy selection with a redundancy penalty as carrying a guarantee, and the report rejects that claim for this design. Its bound counts items and not their token cost, and plain MMR beat it in a later comparison. The decision is to use MMR as the baseline any new selector must beat, and the report adds that subtracting a penalty breaks the monotonicity a real bound needs.

## Evidence
- `reports/Cutting edge context memory research.md`: "plain MMR beat it, 0.9222 against 0.7771"
- `reports/Cutting edge context memory research.md`: "Use MMR as the baseline to beat"
- `reports/Cutting edge context memory research.md`: "which gives at least 0.405 of optimal; a subtractive penalty breaks monotonicity"

## Relations
- alternative-to:: [[Budgeted selection needs a monotone submodular objective]]
- alternative-to:: [[Ranked atoms are sent as excerpts under a token budget]]
- part-of:: [[R5 Knapsack-correct lexical selector]]
- source:: [[Tang et al modified greedy]]
- part-of:: [[Submodular maximization under a knapsack budget]]
- tested-by:: [[Plain MMR beat AdaGReS at a 1024 token budget]]
