---
type: idea
status: proposed
area: selection
needs: none
---
Link each span to the rare keys it contains, seed the graph with the first-pass BM25 hits, and spread score over it, keeping only keys whose document frequency lies in a band so that common keys do not connect everything. No embeddings are used at any step. The research note finds this unmeasured on conversational data: EdgeMem does a single hub lookup, and Zero-Mem and LinearRAG use dense similarity inside the propagation. The expected gain is modest and link noise is a documented risk. A regex for capitalised spans may be able to replace the tagger that finds the keys, which is still to be confirmed on conversational text.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Propagation from BM25 seeds over a span to rare-key graph with a document-frequency band and no embeddings. Prior art uses embeddings in the propagation (Zero-Mem, LinearRAG) or a single lookup (EdgeMem)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Personalized PageRank seeded by BM25 over a span to rare-key graph has not been measured on conversational data without embeddings."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Every model-free graph method that works applies a band-pass on document frequency: EdgeMem's 2 to 10 band, SPRIG's inverse-square-root hub penalty and top 1% pruning, SYNAPSE's fan effect."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The expected gain is modest (SPRIG +3.3 R@10 on HotpotQA) and link noise is a documented risk."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A regex for capitalised spans gave R@10 0.812 against 0.810 for spaCy in SPRIG's hybrid setting"

## Relations
- builds-on:: [[The link graph and its expansion lane]]
- builds-on:: [[Model-free hubs over verbatim turns]]
- builds-on:: [[Search first then bounded query-filtered expansion]]
- builds-on:: [[Whole-conversation questions are the weak spot]]
- analog-of:: [[Personalized PageRank]]
- tested-by:: [[PageRank from BM25 seeds reached 0.775 against 0.742 for BM25 alone]]
- source:: [[SPRIG]]
