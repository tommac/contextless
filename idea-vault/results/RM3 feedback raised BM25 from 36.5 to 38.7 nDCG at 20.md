---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
On 13 low-resource BEIR sets, BM25 with no feedback scored 36.5 nDCG@20, with RM3 feedback 38.7 and with Rocchio 37.3, all with no model. Feedback judged or written by a language model reached 43 to 45. A second study by the same authors found that HyDE by naive concatenation raised Recall@20 from 28.8 to only 29.3 to 30.2, and to 34.0 to 34.7 with Rocchio term selection. The classical gain is small and was measured on document collections, not conversations.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "No feedback 36.5; **RM3 38.7; Rocchio 37.3** (both [NO MODEL])."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "BM25 28.8. HyDE by naive concatenation 29.3 to 30.2, so almost no gain. HyDE with Rocchio term selection 34.0 to 34.7."

## Relations
- supports:: [[Feedback expansion from first-pass hits needs no model]]
- source:: [[Jedidi and Lin 2026]]
