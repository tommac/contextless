---
type: field
status: external
area: selection
needs: none
---
Proximity scoring rewards a passage when the query terms occur close together, on top of term-frequency ranking such as BM25. The research note calls it the best-evidenced classic signal the target system lacks, at about 1.7 NDCG at 10 over BM25 on a web test set, with mixed results noted in the same literature. Inside paragraph-sized spans it is likely to add little, and the untested variant is proximity across adjacent spans. A two-field form of lexical scoring already exists for conversations in EdgeMem, which mixes turn-level and session-level BM25 and reads each seed turn with its neighbours.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Proximity scoring is the best-evidenced classic signal the target system lacks (about +1.7 NDCG@10 over BM25 on a web test set, with mixed results noted in the same literature)"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Inside paragraph-sized spans proximity is likely to add little"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Tao and Zhai (SIGIR 2007) compared five proximity measures and found minimum pair distance the most useful"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "EdgeMem mixes turn-level and session-level BM25 with weight 0.5 each, and reads each seed turn with its predecessor and successor"

## Relations
- analog-of:: [[The value lane finds stated values near the question's words]]
- analog-of:: [[Fused ranking over exact full-text and value lanes]]
- tested-by:: [[Proximity terms raised web NDCG at 10 from 34.34 to 36.00]]
