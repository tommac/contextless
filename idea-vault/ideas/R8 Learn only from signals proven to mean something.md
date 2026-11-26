---
type: idea
status: designed
area: selection
needs: model-calls
---
The hypothesis is that removing a quote-verified cited atom changes answers more often than removing a matched uncited atom, and that recall faults identify atoms the ranking wrongly left out. Stage one is that causal check, and stage two trains a handful of inspectable weights from frozen logs and replays them on locked probes. It is falsified if removing cited and uncited atoms has the same effect, in which case citations are not labels and the compiler design should drop them.

## Evidence
- `reports/Cutting edge context memory research.md`: "Removing a quote-verified cited atom changes answers more often than removing a matched uncited atom. Recall faults identify atoms that ranking wrongly left out"
- `reports/Cutting edge context memory research.md`: "Cited and uncited removal effects do not differ, in which case citations are not labels and section 7 of the compiler design should drop them"
- `reports/Cutting edge context memory research.md`: "two-stage counterfactual results needed 100K to 1M interactions"

## Relations
- builds-on:: [[Verified citations and recall faults as ranking labels]]
- builds-on:: [[Learning from recall requests and citations]]
- builds-on:: [[R7 Test whether answers depend on retrieved memory]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- analog-of:: [[Counterfactual learning to rank]]
