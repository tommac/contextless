---
type: source
status: external
area: index
needs: none
year: 2024
url: https://www.anthropic.com/news/contextual-retrieval
---
A vendor article of 19 September 2024: a model writes a short context for each chunk at index time, and the chunk is embedded and indexed for BM25 with it. The vendor reports the top-20 miss rate falling from 5.7 percent to 1.9 with contextual embeddings, contextual BM25 and reranking together, at a stated cost of $1.02 per million document tokens. The research notes found no independent replication on agent transcripts, and a later placebo-controlled test found that real heading paths beat model-written context.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Top-20 miss rate: embeddings 5.7%, contextual embeddings 3.7% (−35%), plus contextual BM25 2.9% (−49%), plus reranking 1.9% (−67%)."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "No independent replication of Contextual Retrieval on agent transcripts."

## Relations
- supports:: [[Deterministic keys on spans with verbatim output]]
- part-of:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[Contextual retrieval cut the top-20 miss rate from 5.7 to 1.9 percent]]
- tested-by:: [[Real heading paths beat LLM-written context and shuffled placebo paths]]
