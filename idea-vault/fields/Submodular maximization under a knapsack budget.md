---
type: field
status: external
area: selection
needs: none
---
Choosing spans under a token budget is, in optimization terms, maximizing a set function under a knapsack constraint. The research notes report that the only guarantee which transfers cleanly to a token budget is for a monotone submodular objective, selected by gain per token with a fallback to the best single item. Greedy on raw gain under a knapsack budget has an unbounded approximation ratio, so the common heuristic carries no guarantee.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The only guarantee that transfers cleanly to a token budget is for a **monotone submodular** objective selected by **gain per token plus a best-single-item fallback** (at least 0.405 of optimal)."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "**raw-gain greedy under a knapsack budget has an unbounded approximation ratio**"

## Relations
- analog-of:: [[Budgeted selection needs a monotone submodular objective]]
- analog-of:: [[Prefix-priced context selection]]
- source:: [[Lin and Bilmes 2011]]
- source:: [[Tang et al modified greedy]]
