---
type: field
status: external
area: selection
needs: none
---
Personalized PageRank ranks the nodes of a graph by spreading score outward from a chosen set of seed nodes, and spreading activation is a related method. Graphs built from named entities or keywords with no LLM are now established, but all except EdgeMem and one SPRIG variant need embeddings at query time. On conversational data the published ablations assign most of the gain to session and neighbour context and to time, while on Wikipedia-style multi-hop the entity graph matters more than the hierarchy. The recipe that transfers is first-pass lexical hits as seeds, then spread, with a band-pass on how common a linking key may be.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Graphs built from named-entity or keyword co-occurrence with no LLM are now established (LinearRAG, SPRIG, KET-RAG's keyword layer, EdgeMem, Zero-Mem), but all except EdgeMem and one SPRIG variant need embeddings at query time."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Query-entity seeding fails (SPRIG 0.464 against 0.742) and lexical seeding works."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Keys that are too common connect everything, and keys seen once connect nothing."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "BABILong average 80.4 for GPT-4o with MixPR against 48.3 alone"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No model-free personalized PageRank or spreading activation on conversational memory was found; SYNAPSE depends on LLM-extracted concepts and embeddings."

## Relations
- analog-of:: [[The link graph and its expansion lane]]
- analog-of:: [[Links help when questions need several pieces of evidence]]
- tested-by:: [[PageRank from BM25 seeds reached 0.775 against 0.742 for BM25 alone]]
- tested-by:: [[Graph path search made no measurable difference on LoCoMo]]
- tested-by:: [[Removing graph structure halved temporal F1 and left multi-hop flat]]
- source:: [[SPRIG]]
- source:: [[HippoRAG 2]]
