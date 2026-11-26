---
type: idea
status: proposed
area: cost
needs: model-calls
---
The proposal is to report verbatim memory on a frontier of answer quality against billed cost, using the provider's real cache fields, latency and paired quality, with a Pareto analysis. The closest prior art is Total Recall at What Cost and Beyond the Context Window, which already ask the cost question, and the notes found no memory paper with formal Pareto reporting. The report therefore judges it only partly new and says speed matters. Raw context tokens are not cost, so the frontier needs billed-equivalent tokens and cost per correct answer.

## Evidence
- `reports/Cutting edge context memory research.md`: "Real cache fields, latency, paired quality and a Pareto analysis"
- `reports/Cutting edge context memory research.md`: "**Partly new.** No formal Pareto report was found, but the question is already being asked, so speed matters"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Raw context tokens are not cost."

## Relations
- builds-on:: [[Billed-equivalent tokens]]
- builds-on:: [[Cost per correct answer]]
- builds-on:: [[Usage records cannot yet produce billed tokens]]
- builds-on:: [[Prompt cache savings are a ceiling until a live test]]
- source:: [[Total Recall at What Cost]]
- source:: [[Token Reduction Is Not Cost Reduction]]
