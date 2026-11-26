---
type: result
status: measured
area: selection
needs: none
measured-by: others
---
On a web test set of 11,857 queries, adding a pairwise proximity term over adjacent and non-adjacent query bigrams raised BM25 from 34.34 to 36.00 NDCG at 10, and from 32.05 to 36.04 on tail queries. The same paper's review of earlier work records caveats: one study had almost negligible results, span-based scoring gave mixed results on the web, and proximity features can be neutral or even negative when the base ranker is already strong. It is the best measurement found for proximity and it is from web search in 2010; no use on conversation spans was found.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "BM25 24.60/27.74/34.34; BM25 plus a pairwise proximity term (adjacent and non-adjacent query bigrams) 26.06/29.54/36.00; adjacent bigrams only 25.27/28.72/35.35"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "On tail queries BM25 scored 21.23/25.13/32.05 and the proximity variant 23.21/28.73/36.04."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Web NDCG@10 34.34 to 36.00 over BM25, 2010."

## Relations
- supports:: [[Cover density across adjacent spans]]
- supports:: [[The value lane finds stated values near the question's words]]
- part-of:: [[Proximity scoring]]
