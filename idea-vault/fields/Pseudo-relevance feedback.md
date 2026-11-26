---
type: field
status: external
area: selection
needs: none
---
Pseudo-relevance feedback is the classical information retrieval technique of expanding a query with terms from its own first results; RM3 and Rocchio are the standard models and need no language model. The research notes report a small but free gain over BM25 from RM3, and a larger gain from model-written expansion at a model-call cost. A 2025 study calls the gains of both kinds moderate and unstable, which the notes saw only as a search snippet.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Classical feedback gives a small but free gain over BM25 (+2.2 nDCG@20 points with RM3), and LLM expansion gives more (+8 to +9) at a model-call cost."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "A generalized feedback study describes RM3 and HyDE-type gains over BM25 as moderate and unstable"

## Relations
- analog-of:: [[Feedback expansion from first-pass hits needs no model]]
- source:: [[Jedidi and Lin 2026]]
- tested-by:: [[RM3 feedback raised BM25 from 36.5 to 38.7 nDCG at 20]]
