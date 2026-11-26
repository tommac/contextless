---
type: idea
status: designed
area: caching
needs: model-calls
---
The hypothesis is that selecting over reuse boundaries lowers billed-equivalent input by at least 20 percent against cache-blind selection at equal packet recovery. Every user message is replayed with the boundary solver, using the hit behaviour measured in R2 and the message gaps from R1, and it runs live only if R2 finds a working route. It is falsified if savings are under 10 percent after real hit rates, if quality falls by more than the pre-set margin, or if simple fixed epochs match the solver.

## Evidence
- `reports/Cutting edge context memory research.md`: "Selecting over reuse boundaries lowers billed-equivalent input by at least 20% against cache-blind selection, at equal packet recovery"
- `reports/Cutting edge context memory research.md`: "Savings are under 10% after real hit rates, or quality falls by more than the pre-set margin, or simple fixed epochs match the solver"
- `reports/Cutting edge context memory research.md`: "First read Pricing the Context Window and TokenPilot in full"

## Relations
- builds-on:: [[Prefix-priced context selection]]
- builds-on:: [[The priced knapsack with a prefix constraint]]
- builds-on:: [[R2 Cache truth table]]
- builds-on:: [[R1 Repair the instrument and record the offline floor]]
- builds-on:: [[Catalog spine and epoch window make the prefix repeat]]
- builds-on:: [[Plans run in waves with dependencies and shared decisions]]
- builds-on:: [[Billed-equivalent tokens]]
