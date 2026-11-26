---
type: idea
status: proposed
area: caching
needs: none
---
The report calls this the sharpest open problem. An atom gets the cached price only if everything before it in the prompt is unchanged, so the cheap set is prefix-closed and ordinary knapsack selection does not apply. For a fixed reusable prefix the marginal objective is still monotone submodular, so greedy selection by gain per token with a best-single-item fallback keeps the 0.405 factor, and taking the best run over all prefixes keeps it overall. The derivation is the report author's own and unreviewed, and it stops holding with a second hard constraint or a per-token distractor penalty.

## Evidence
- `reports/Cutting edge context memory research.md`: "An atom gets the cached price only if everything before it in the prompt is unchanged, so the cheap set is prefix-closed and ordinary knapsack selection does not apply."
- `reports/Cutting edge context memory research.md`: "taking the best of the K+1 runs keeps the 0.405 factor overall"
- `reports/Cutting edge context memory research.md`: "**This is my derivation and has not been reviewed.**"

## Relations
- builds-on:: [[Prefix-priced context selection]]
- builds-on:: [[Budgeted selection needs a monotone submodular objective]]
- builds-on:: [[Catalog spine and epoch window make the prefix repeat]]
- builds-on:: [[Message gaps set the cache hit rate]]
- analog-of:: [[Rent-or-buy decisions]]
- analog-of:: [[Submodular maximization under a knapsack budget]]
- source:: [[Tang et al modified greedy]]
