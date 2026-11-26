---
type: source
status: external
area: selection
needs: none
year: 2025
url: https://arxiv.org/pdf/2602.23372
---
SPRIG (Qizhi Wang, PingCAP; arXiv 2602.23372) builds an entity to passage bipartite graph weighted by TF-IDF with no LLM and runs personalized PageRank from BM25 seeds. It reports RM3 and two-step entity expansion as model-free baselines on HotpotQA and 2WikiMultiHopQA, with paired bootstrap intervals and wins, ties and losses per query, a practice the research note recommends copying. Quality caveats: a single author, a 2Wiki BM25 figure that differs between two of its tables, and wrong author names in its reference list, so the research note treats its 2Wiki figures as weak.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "an entity to passage bipartite graph weighted by TF-IDF, personalized PageRank with restart 0.15 and 5 iterations"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the reference list carries wrong author names, so treat the 2Wiki figures as weak"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Pruning the top 1% of entities by document frequency removed 1.76% (HotpotQA) and 0.49% (2Wiki) of gold-title mentions and cut query time 16% to 28% with negligible recall change."

## Relations
- part-of:: [[Personalized PageRank]]
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Paired tests with conversation-level resampling]]
- tested-by:: [[PageRank from BM25 seeds reached 0.775 against 0.742 for BM25 alone]]
- tested-by:: [[RM3 feedback cut HotpotQA recall at 10 from 0.742 to 0.675]]
