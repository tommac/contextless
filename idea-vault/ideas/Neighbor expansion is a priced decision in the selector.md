---
type: idea
status: proposed
area: selection
needs: none
---
Sending the paragraphs around a hit makes an excerpt easier to read but takes budget from other hits. In this project's first experiment, adding each atom's neighbours doubled the median cost and lowered recovery within 2k and 4k, and an outside system withdrew a 99 percent result because neighbour expansion credited adjacent spans. The notes therefore propose treating a neighbour as a candidate with its own token cost and marginal gain, and scoring recall only for what is inside the budget. As built, the paragraphs around an excerpt come back by id instead.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Make neighbor expansion a **priced decision in the selector**: a neighbor is a candidate with token cost and marginal gain"
- `docs/CONTEXT-COMPILER.md`: "It is still 2.5 to 3.5 times cheaper than whole turns, and it is the price of readability, so end-to-end tests decide it (section 9)."

## Relations
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- builds-on:: [[Budgeted selection needs a monotone submodular objective]]
- tested-by:: [[Adding each atom's neighbors doubled the median cost]]
- tested-by:: [[Neighbor expansion made a trivial BM25 baseline score about 99 percent]]
