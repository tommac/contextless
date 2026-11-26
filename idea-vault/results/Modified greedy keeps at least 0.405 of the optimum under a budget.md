---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
For a monotone submodular objective under a budget, running gain-per-cost greedy, separately finding the best single item and returning the better of the two has a proven ratio above 0.405, and an instance exists where it achieves only about one half. Density greedy alone has no bounded ratio. This is a proof in a 2020 arXiv paper, not an experiment, and it is the bound the design quotes for its selection step.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "They prove a ratio above **0.405**, and give an instance where it achieves only about 1/2, so the true worst case lies between about 0.405 and 0.5."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Density greedy alone has no bounded ratio."
- `docs/CONTEXT-COMPILER.md`: "a monotone submodular objective, selection by gain per token, and a fallback to the best single item give at least 0.405 of the optimum under a token budget"

## Relations
- supports:: [[Budgeted selection needs a monotone submodular objective]]
- source:: [[Tang et al modified greedy]]
- supports:: [[R5 Knapsack-correct lexical selector]]
- supports:: [[The priced knapsack with a prefix constraint]]
