---
type: idea
status: proposed
area: selection
needs: model-calls
---
The proposal is to use two signals the system already produces as labels for ranking: citations whose quoted text a program has confirmed, and recall requests that reveal an atom the ranking wrongly left out. The closest prior art in the notes is counterfactual learning to rank and generative relevance feedback, and the report judges the idea likely new and the strongest of its candidates, since no use of model citations or recall requests as labels was found. It must first be shown that a quote-checked citation predicts causal use, because citations can be attached after the fact to an answer the model already had.

## Evidence
- `reports/Cutting edge context memory research.md`: "Measuring whether a quote-checked citation predicts causal use, then training on it"
- `reports/Cutting edge context memory research.md`: "**Likely new.** No use of LLM citations or recall requests as labels was found. The strongest candidate"
- `reports/Cutting edge context memory research.md`: "up to 57% of citations can be unfaithful"

## Relations
- builds-on:: [[Learning from recall requests and citations]]
- builds-on:: [[Quotes are pinned to the paragraphs that hold them]]
- builds-on:: [[Units return typed records with checked claims]]
- builds-on:: [[Atom-level causal evaluation]]
- analog-of:: [[Counterfactual learning to rank]]
- tested-by:: [[Up to 57 percent of citations were unfaithful in adversarial tests]]
