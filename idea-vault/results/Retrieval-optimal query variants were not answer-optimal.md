---
type: result
status: measured
area: evaluation
needs: none
measured-by: others
---
In a SIGIR 2026 study on 56 TREC-RAG 2024 queries with 30 model-written reformulations each, choosing the variant with the best oracle nDCG@5 gave a BM25 Nugget-Strict of 0.344, below the 0.377 from a pre-retrieval selector. A post-retrieval score correlated about 0.33 with nDCG@5 and about zero with answer nuggets. So a good retrieval score need not mean a good answer, which is the gap this project's probes leave open.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Choosing variants by oracle nDCG@5 gave BM25 Nugget-Strict of 0.344, below the 0.377 from a pre-retrieval selector"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "so a good retrieval score need not imply a good answer"

## Relations
- supports:: [[No model has yet answered from a packet]]
- supports:: [[Excerpt readability is the main open question]]
- supports:: [[Atom-level causal evaluation]]
