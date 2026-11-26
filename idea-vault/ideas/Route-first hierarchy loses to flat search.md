---
type: idea
status: rejected
area: selection
needs: none
---
Routing a query down a hierarchy first and ranking only inside the chosen branch is rejected for this design. In the nearest test, routing by section lost to flat search, with first-stage recall as low as 0.17, because a wrong choice at the top hides the evidence from everything below. The decision is to use structure as a score feature and never as a gate.

## Evidence
- `reports/Cutting edge context memory research.md`: "Section routing lost to flat search, -0.033 and -0.015, with first-stage recall as low as 0.17"
- `reports/Cutting edge context memory research.md`: "Use structure as a score feature, never as a gate"

## Relations
- supports:: [[A code-made zoom ladder loses to flat search]]
- alternative-to:: [[Fused ranking over exact full-text and value lanes]]
- tested-by:: [[Route to section then rank lost to flat search by 0.033 and 0.015]]
- builds-on:: [[Summary trees help only on very long histories]]
- source:: [[Kuehlkamp et al 2026]]
