---
type: idea
status: designed
area: evaluation
needs: model-calls
---
No model has yet answered a question from a ContextlessAI packet, so every local number measures whether text reached the packet and not whether an answer was right or what it cost. Outside results show the gap matters, since evidence that is present is often not used correctly. The report says the next gain comes from measurement and from pricing context correctly, and it places a paired answer pilot third on its agenda.

## Evidence
- `reports/Cutting edge context memory research.md`: "**no model has yet answered a question from a ContextlessAI packet**, so every local number measures whether text reached the packet, not whether an answer was right or what it cost"
- `reports/Cutting edge context memory research.md`: "The next gain comes from measurement and from pricing context correctly, not from a richer memory structure."
- `reports/Cutting edge context memory research.md`: "Directly limits what `cx probe` can show"

## Relations
- builds-on:: [[Recovery probes measure selection without a model]]
- builds-on:: [[Three layers reached used and correct]]
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- tested-by:: [[Full text had 100 recall and 77.40 percent accuracy on LongMemEval-S]]
- tested-by:: [[73.3 percent of sampled memory failures were reachable but unsolved]]
