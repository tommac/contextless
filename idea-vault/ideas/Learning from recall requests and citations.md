---
type: idea
status: designed
area: selection
needs: model-calls
---
Signals the system already produces can tune selection with no model in the loop: a recall request marks atoms that should have been selected, a citation marks atoms that were useful, and probes give offline labels. They would tune lane weights, lane depth, the per-class budget and the expansion rules per workspace by replaying past messages, and the result is a small versioned config, so selection stays deterministic. The research notes found no published work that uses citations or recall requests as ranking labels, and warn that a citation means used, not needed. Every tuning has to be checked on probes it did not see.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "| `needs_recall` ids | those atoms should have been selected for that message | strong |"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "I found **no published work that uses LLM citations or recall requests as ranking labels for context selection**"
- `docs/CONTEXT-COMPILER.md`: "the value lane was tuned on 14 synthetic probes and added only one probe out of 30 on real ones"

## Relations
- builds-on:: [[Recall by id or by search replaces memory]]
- builds-on:: [[Recovery probes measure selection without a model]]
- part-of:: [[Counterfactual learning to rank]]
- builds-on:: [[Determinism is the instrument]]
- part-of:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[The value lane added one probe on two of three real conversations]]
