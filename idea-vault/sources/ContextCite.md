---
type: source
status: external
area: evaluation
needs: none
year: 2024
url: https://arxiv.org/html/2409.00729v1
---
ContextCite, from NeurIPS 2024, attributes an answer to its sources by fitting a sparse linear model over random removals of those sources. It separates sources that contribute to an answer from sources that merely agree with it. It measures influence and not mechanism, and the notes found no 2026 work applying it to a long-conversation memory system, which is the opening for atom-level causal evaluation.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "fits a sparse linear surrogate over random source ablations, separates contributive from corroborative attribution"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "ContextCite's limits: influence rather than mechanism, ablations may not be neutral, and refusals can be misattributed"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "I did not find a 2026 evaluation applying ContextCite-style ablation directly to a long-conversation memory system."

## Relations
- supports:: [[Atom-level causal evaluation]]
- part-of:: [[R7 Test whether answers depend on retrieved memory]]
- alternative-to:: [[Value-swap test]]
- supports:: [[Determinism is the instrument]]
