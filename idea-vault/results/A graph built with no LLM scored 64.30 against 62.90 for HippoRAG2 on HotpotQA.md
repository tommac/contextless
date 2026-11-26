---
type: result
status: measured
area: structure
needs: none
measured-by: others
---
LinearRAG builds a graph of passages, sentences and entities with spaCy and no relation extraction, uses zero LLM tokens for indexing and retrieval, and ranks with personalized PageRank. With a GPT-4o-mini reader its containment accuracy was 64.30 on HotpotQA against 62.90 for HippoRAG2 and 55.70 for plain top-5 retrieval, and it also led on 2Wiki and MuSiQue; indexing 2Wiki took 249.78 seconds against 1,147.01. It needs embeddings at query time, has no BM25 baseline, and shows its ablations only as charts. So a graph built without an LLM can match one built with an LLM on document multi-hop, which is not evidence about conversations.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "HotpotQA 64.30/66.50 against HippoRAG2 62.90/64.30 and plain top-5 retrieval 55.70/58.60"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "a graph of passages, sentences and entities from spaCy with no relation extraction, zero LLM tokens for indexing and retrieval"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Indexing 2Wiki took 249.78 s against 1,147.01 s for HippoRAG2. No BM25 baseline; ablations are shown only as charts."

## Relations
- supports:: [[Model-free hubs over verbatim turns]]
- supports:: [[Links help when questions need several pieces of evidence]]
- part-of:: [[Personalized PageRank]]
