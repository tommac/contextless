---
type: source
status: external
area: selection
needs: none
year: 2011
url: https://aclanthology.org/P11-1052/
---
A Class of Submodular Functions for Document Summarization (ACL 2011) is the design's cited source for treating budgeted selection as a submodular problem. The selection notes read the same authors' 2010 NAACL paper, which casts summarization as budgeted submodular maximization, selects by gain divided by scaled cost, and shows that greedy without a safeguard has an unbounded approximation factor. For a cost exponent of 1 that paper states a bound of about 0.39.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "| Lin and Bilmes (2011) | budgeted selection as a submodular problem | the selection step |"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Lin and Bilmes cast summarization as budgeted submodular maximization."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "show that greedy without a safeguard has an unbounded approximation factor"

## Relations
- part-of:: [[Submodular maximization under a knapsack budget]]
- supports:: [[Budgeted selection needs a monotone submodular objective]]
- tested-by:: [[Modified greedy keeps at least 0.405 of the optimum under a budget]]
