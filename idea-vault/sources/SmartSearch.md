---
type: source
status: external
area: selection
needs: none
year: 2026
url: https://arxiv.org/html/2603.15599v1
---
SmartSearch (Derehag et al., Midbrain; arXiv 16 Mar 2026) keeps raw conversation text with no ingest step, recalls by exact substring match with part-of-speech term weights, and reranks with a cross-encoder plus ColBERT. Its title claim is that ranking, not structure, carries conversational retrieval, and its reranker ladder moved LoCoMo from 76.8 with no reranker to 91.9. The research note draws on it for feedback and entity-hop ablations, the hop distribution of LoCoMo, a null result for MMR and the compilation bottleneck. It is a single-group preprint, and it depends on a named-entity tagger and neural rerankers at query time.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "keeps raw conversation text with no ingest step, recalls by exact substring match with part-of-speech term weights, and reranks with a cross-encoder plus ColBERT"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The title claim is that ranking, not structure, carries conversational retrieval."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Its reranker ladder moved LoCoMo from 76.8 with no reranker to 91.9"

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- alternative-to:: [[EdgeMem]]
- tested-by:: [[Index-free SmartSearch scored 88.4 on LongMemEval-S against 63.8 for Zep]]
- tested-by:: [[One feedback pass added 6.0 points on LongMemEval-S and none on LoCoMo]]
- tested-by:: [[Only 22.5 percent of gold evidence survived truncation to the budget]]
- tested-by:: [[MMR reranking did not help SmartSearch on conversational memory]]
