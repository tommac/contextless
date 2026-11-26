---
type: source
status: external
area: structure
needs: none
year: 2024
url: https://arxiv.org/abs/2401.18059
---
RAPTOR builds a tree over a corpus by bottom-up recursive clustering and summarization. Its abstract claims a 20-point accuracy gain on QuALITY with GPT-4, a claim the notes did not check against the paper's table. In later runs by other authors it scored below full text and below BM25, so for this project it is evidence for caution about summary trees and not a design to copy.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "bottom-up recursive clustering and summarization. Its abstract claims a 20-point absolute accuracy improvement on QuALITY when paired with GPT-4"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "QuALITY full text 65.1 vs GraphRAG 62.8, MemTree 59.8, RAPTOR 59.0"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "overall BM25 41.5 vs HippoRAG-v2 41.6, MemoRAG 30.9, RAPTOR 27.0, GraphRAG 23.4"

## Relations
- part-of:: [[Summary trees help only on very long histories]]
- alternative-to:: [[MemTree]]
- tested-by:: [[MemTree lost to full history at 15 rounds and won at 200]]
- tested-by:: [[BM25 tied HippoRAG-v2 overall and beat GraphRAG and Zep on retrieval]]
