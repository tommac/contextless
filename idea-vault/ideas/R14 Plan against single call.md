---
type: idea
status: designed
area: plans
needs: model-calls
---
The hypothesis is that plans help only on decomposable requests and do not beat a budget-matched single call overall. The existing blind comparison is run after the usage repair, on 30 messages stratified by request class, so only large effects are visible. It is falsified if the plan arm wins at a matched budget across classes, which would overturn the expected result.

## Evidence
- `reports/Cutting edge context memory research.md`: "Plans help only on decomposable requests and do not beat a budget-matched single call overall, as"
- `reports/Cutting edge context memory research.md`: "An exact sign test needs the preferred arm to win about 75% of untied messages (approximate), so only large effects are visible"
- `reports/Cutting edge context memory research.md`: "The plan arm wins at matched budget across classes, which would overturn the expected result"

## Relations
- builds-on:: [[cx compare measures plans against single calls blind]]
- builds-on:: [[Plans that buy nothing are answered in one call]]
- builds-on:: [[Multi-agent gains are narrow and cost more tokens]]
- builds-on:: [[Plans run in waves with dependencies and shared decisions]]
- builds-on:: [[The assembler accounts for every claim]]
- builds-on:: [[R1 Repair the instrument and record the offline floor]]
