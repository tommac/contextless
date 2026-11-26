---
type: field
status: external
area: selection
needs: none
---
Counterfactual learning to rank trains a ranker from logged implicit feedback such as clicks while correcting for the fact that people only react to what they were shown. The foundations weight each click by its propensity to correct position bias, and later work extends this to metrics such as DCG. The research notes call the toolkit mature but note that its evidence uses interaction counts a single-user tool will not reach, which favours few parameters and strong priors.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Joachims et al. (2017) correct click position bias with propensity-weighted learning, with click models supplying propensities"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "the two-stage CLTR evidence uses interaction counts a single-user tool will not reach, which favors few parameters and strong priors"

## Relations
- analog-of:: [[Learning from recall requests and citations]]
- analog-of:: [[Verified citations and recall faults as ranking labels]]
- analog-of:: [[Spans that never enter the context can never be cited]]
