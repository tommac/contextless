---
type: source
status: external
area: index
needs: none
year: 2019
url: https://arxiv.org/abs/1904.08375
---
Nogueira et al., 2019, Document Expansion by Query Prediction: a model writes the questions a document answers and they are indexed with it, so the model runs at index time only. The design cites it for the one to three questions an indexer could write per atom. A 2023 follow-up, Doc2Query--, filters hallucinated generated queries before indexing and leaves the inverted index unchanged at query time.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "The indexer can write, per atom, a one-line claim and one to three questions the atom answers (doc2query)."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Doc2Query-- filters hallucinated generated queries before indexing: up to +16% effectiveness, −23% mean query time, −33% index size against Doc2Query."

## Relations
- supports:: [[Atom titles as APIs]]
- part-of:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[Doc2Query-- filtering gave up to 16 percent more effectiveness]]
- tested-by:: [[Indexing atoms with the question they answer did not help overall]]
