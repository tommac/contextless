---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/html/2609.33226
---
Threader (Cai et al., Tsinghua; arXiv 27 Sep 2026) keeps raw interactions and groups them into topic segments with a small BERT classifier, then retrieves with embeddings and a reranker. Its headline is LongMemEval 90.60 against 84.80 for the best baseline. It is the one source found that reports complete-evidence coverage beside mean recall. It is a single-group preprint, and its segmenter, embeddings and reranker put it outside the model-free setting.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "keeps raw interactions and groups them into topic segments with a small BERT classifier"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LongMemEval 90.60 against 84.80 for the best baseline"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Threader is the one source found that reports complete-evidence coverage beside mean recall"

## Relations
- supports:: [[Threads group turns by topic]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- alternative-to:: [[SeCom]]
- tested-by:: [[Topic segments scored 90.60 against 85.80 for single messages]]
- tested-by:: [[With BM25 alone complete coverage was 73.19 against 81.83 mean recall]]
