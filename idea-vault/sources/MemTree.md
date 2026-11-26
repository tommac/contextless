---
type: source
status: external
area: structure
needs: none
year: 2024
url: https://arxiv.org/html/2410.14052
---
MemTree inserts new content down a tree by embedding similarity and has a model rewrite every ancestor on the path. Its own tables are the main evidence that a summary tree loses to raw text on short histories and wins only at 200 rounds. The ancestor rewrites cost 3.27 model calls per insertion and do not fit an immutable store unless they are expressed as new versions.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "new content descends by embedding similarity against depth-adaptive thresholds; each ancestor on the path is rewritten by an LLM"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "MemTree's ancestor rewrites and H-MEM's mutable weights do not fit immutability unless expressed as new versions or a separate overlay."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "measures 3.27 ± 2.38 LLM calls per insertion on MultiHop RAG vs 1 for MemoryStream"

## Relations
- part-of:: [[Summary trees help only on very long histories]]
- alternative-to:: [[RAPTOR]]
- tested-by:: [[MemTree lost to full history at 15 rounds and won at 200]]
- contradicts:: [[Turns are stored verbatim and never rewritten]]
