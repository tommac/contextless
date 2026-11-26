---
type: result
status: measured
area: selection
needs: none
measured-by: others
---
In a placebo-controlled test of chunk structure, routing a query to a section first and ranking inside it lost to flat search by 0.033 and 0.015 coverage-nDCG@10, with first-stage section recall as low as 0.17 to 0.20 on QASPER. A good span is lost whenever the first stage picks the wrong section. The notes' advice is to treat session or section evidence as a score feature in flat ranking instead.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Route-to-section-then-rank **lost** to flat search, −0.033 and −0.015, with first-stage section recall as low as 0.17 to 0.20 on QASPER."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Treat session or section evidence as a score feature in flat ranking, so low first-stage recall cannot remove a good span."

## Relations
- supports:: [[Route-first hierarchy loses to flat search]]
- supports:: [[A code-made zoom ladder loses to flat search]]
- source:: [[Kuehlkamp et al 2026]]
