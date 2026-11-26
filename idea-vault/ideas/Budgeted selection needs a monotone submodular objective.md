---
type: idea
status: designed
area: selection
needs: none
---
Filling a token budget greedily has a quality guarantee only under conditions: the objective must be monotone submodular, items must be picked by gain per token, and the result compared with the best single item that fits, which gives at least 0.405 of the optimum. A subtractive redundancy penalty breaks monotonicity, so redundancy has to be written as saturating coverage of query terms, values and entities. The selection notes recommend this with no model: a concave lexical coverage function, lazy greedy on gain per token, ties broken by span id. It sits in phase 3 of the build plan.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "A subtractive redundancy penalty breaks monotonicity, so redundancy has to be expressed as saturating coverage for the bound to apply."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "A positive-weight concave-over-modular function is monotone submodular, so the 0.405 bound applies, and the run is fully deterministic."

## Relations
- source:: [[Lin and Bilmes 2011]]
- source:: [[Tang et al modified greedy]]
- part-of:: [[Submodular maximization under a knapsack budget]]
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- tested-by:: [[Modified greedy keeps at least 0.405 of the optimum under a budget]]
- tested-by:: [[Plain MMR beat AdaGReS at a 1024 token budget]]
- tested-by:: [[Feedback links and coverage did not beat search]]
