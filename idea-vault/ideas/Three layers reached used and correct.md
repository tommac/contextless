---
type: idea
status: proposed
area: evaluation
needs: none
---
A credible paper should report three separate layers: the evidence reached the context, the evidence was used, and the answer was correct. Collapsing them is the root of most benchmark disputes, because retrieval metrics get presented as answer accuracy. Presence metrics also miss the failure in which a model answers wrongly instead of abstaining when its context is insufficient.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "A credible paper should report three separate layers: evidence reached the context, evidence was used, and the answer was correct. Collapsing them is the root of most disputes above."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "so evidence presence does not guarantee a correct answer"
- `reports/Cutting edge context memory research.md`: "Presence metrics miss this failure"

## Relations
- builds-on:: [[Recovery probes measure selection without a model]]
- supports:: [[No information is a result not a failure]]
- part-of:: [[Minimum evidence package for a credible memory study]]
- tested-by:: [[Full text had 100 recall and 77.40 percent accuracy on LongMemEval-S]]
- tested-by:: [[73.3 percent of sampled memory failures were reachable but unsolved]]
- tested-by:: [[MemPalace 96.6 percent was recall while end-to-end QA was about 67.2]]
