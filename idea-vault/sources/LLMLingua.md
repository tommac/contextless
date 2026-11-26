---
type: source
status: external
area: compression
needs: none
year: 2024
url: https://arxiv.org/abs/2403.12968
---
LLMLingua is a family of token-level prompt compressors; the year and link here are those of LLMLingua-2 (Findings of ACL 2024), the one the research notes cite. Its authors claim that LLMLingua-2 is 3 to 6 times faster than prior compressors and cuts end-to-end latency at 2 to 5 times compression. As an agent-history baseline in a competing method's paper it scored 39.3 on AppWorld against 56.0 uncompressed, while holding up on question answering. The design names it as the kind of compression it does not propose.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "LLMLingua-2 is 3x to 6x faster than prior compressors and cuts end-to-end latency 1.6x to 2.9x at 2x to 5x compression"
- `docs/CONTEXT-COMPILER.md`: "Not proposed: prompt compression (LLMLingua and similar), summaries in place of text, embeddings before you approve a local model download."

## Relations
- alternative-to:: [[Turns are stored verbatim and never rewritten]]
- part-of:: [[Lossy compression loses constraints and exact text]]
- tested-by:: [[LLMLingua scored 39.3 on AppWorld against 56.0 uncompressed]]
- part-of:: [[Token-level prompt compression]]
