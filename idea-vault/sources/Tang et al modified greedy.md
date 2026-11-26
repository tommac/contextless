---
type: source
status: external
area: selection
needs: none
year: 2020
url: https://ar5iv.labs.arxiv.org/html/2008.05391
---
An August 2020 arXiv paper that analyses modified greedy for submodular maximization under a budget: run gain-per-cost greedy, separately find the best single item, and return the better of the two. It proves a ratio above 0.405 and gives an instance achieving only about one half, so the true worst case lies between. It reports that the widely repeated proof of 0.393 had a gap and supplies a corrected one, and it gives a data-dependent upper bound on the optimum that could be logged per query.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "run gain-per-cost greedy, separately find the best single item, return the better of the two"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "They prove a ratio above **0.405**, and give an instance where it achieves only about 1/2"

## Relations
- part-of:: [[Submodular maximization under a knapsack budget]]
- supports:: [[Budgeted selection needs a monotone submodular objective]]
- tested-by:: [[Modified greedy keeps at least 0.405 of the optimum under a budget]]
