---
type: result
status: measured
area: selection
needs: none
measured-by: others
---
On the TREC 2010 Web track diversity task with official sub-topics as aspects, PM-2 reached alpha-NDCG 0.4546, xQuAD 0.4074, MMR 0.3349 and the query-likelihood baseline 0.3236. PM-2 treats result slots as seats in an election and gives aspects seats in proportion to their weight. In the earlier xQuAD study MMR fell below undiversified BM25, 0.120 against 0.159 at alpha-NDCG at 5, and the PM-2 authors call MMR the least effective. These are peer-reviewed web search results, and no application of either method to context selection or conversational memory was found.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Web track 2010: 0.3236, 0.3349, 0.4074, 0.4546; with aspects from query suggestions xQuAD 0.3582 and PM-2 0.4374."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "BM25 0.159/0.186/0.288; BM25 plus MMR 0.120/0.150/0.224; plus IA-Select 0.110/0.119/0.180; plus xQuAD 0.208/0.227/0.324"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No application of xQuAD or PM-2 to retrieval-augmented generation or to conversational memory was found."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Wins and losses against the baseline for PM-2 were 34/10 in 2010 and 19/19 in 2009."

## Relations
- supports:: [[Slot allocation over deterministic aspects]]
- part-of:: [[Result diversification]]
- source:: [[PM-2]]
