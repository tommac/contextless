---
type: field
status: external
area: selection
needs: none
---
Information foraging theory borrows the marginal value theorem from ecology: a forager should leave a patch of information when the rate of gain from staying drops to the average rate available elsewhere. Information scent is the perceived value of the cues that guide where to go next. The only LLM-era measurement found is InForage, a trained search policy in which nothing is model-free, and the patch model itself was read only through a 2026 review. For this project the theory gives a stopping rule for neighbour expansion that fixed windows lack: treat a session as a patch and keep reading adjacent spans while the gain per token stays above the average gain of jumping to the next best seed.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "a forager should leave an information patch when the rate of gain from staying drops to the average rate available elsewhere, with the richness of the patch and the cost of switching setting the stopping point"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Information scent is the perceived value of cues that guide where to go next."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "eight-set average 41.0, against 35.6 for the strongest trained baseline and 26.5 for plain retrieval"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No model-free use of foraging theory for retrieval or packing was found."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Pirolli and Card's paper was not read directly; the patch-model description is from a 2026 review."

## Relations
- analog-of:: [[Neighbor expansion is a priced decision in the selector]]
- analog-of:: [[Recall as paging]]
- analog-of:: [[Atom titles as APIs]]
- analog-of:: [[Catalog cards are an index not a replacement]]
