---
type: idea
status: proposed
area: selection
needs: none
---
Classical pseudo-relevance feedback takes terms from the top first-pass hits and adds them to the query, using counts and arithmetic only. The research notes propose an RM3-style lane for this engine, gated by a confidence check because feedback can reinforce a poor first pass, with term ties broken stably so it stays deterministic. The published gain is small, 2.2 nDCG@20 points over BM25 on document collections, and may not hold on short conversational spans.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Add **RM3-style feedback as a [NO MODEL] lane candidate**, gated by a confidence check, because feedback can reinforce a poor first pass."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The published gain is on document collections and may not hold on short conversational spans."

## Relations
- part-of:: [[Pseudo-relevance feedback]]
- source:: [[Jedidi and Lin 2026]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- tested-by:: [[RM3 feedback raised BM25 from 36.5 to 38.7 nDCG at 20]]
- tested-by:: [[Feedback links and coverage did not beat search]]
