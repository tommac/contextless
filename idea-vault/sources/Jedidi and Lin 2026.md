---
type: source
status: external
area: selection
needs: none
year: 2026
url: https://arxiv.org/html/2603.11008
---
A March 2026 preprint measuring feedback on BM25 across 13 low-resource BEIR sets by nDCG@20. Classical RM3 and Rocchio feedback, which need no model, raise BM25 from 36.5 to 38.7 and 37.3, and model-judged or model-generated feedback reaches 43 to 45. The same authors' 2025 study found that HyDE by naive concatenation adds almost nothing and that term selection and weighting produce the benefit.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "No feedback 36.5; **RM3 38.7; Rocchio 37.3** (both [NO MODEL])."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Term selection and weighting, not generation alone, produced the benefit"

## Relations
- part-of:: [[Pseudo-relevance feedback]]
- supports:: [[Feedback expansion from first-pass hits needs no model]]
- tested-by:: [[RM3 feedback raised BM25 from 36.5 to 38.7 nDCG at 20]]
