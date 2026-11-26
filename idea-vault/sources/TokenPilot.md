---
type: source
status: external
area: caching
needs: none
year: 2026
url: https://arxiv.org/abs/2606.17016
---
Xu et al., June 2026, on cache-efficient context management for LLM agents: stable prefixes and eviction of expired segments, reported as 56 to 87 percent cheaper. The design takes the append-only prefix from it and differs by evicting by relevance. The saving is the paper's own figure and appears only in the design's prior-art table, so this vault holds no detail of how it was measured.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "| TokenPilot (2026) | stable prefixes and eviction of expired segments, 56 to 87 percent cheaper | the append-only prefix; eviction by relevance |"

## Relations
- supports:: [[The append-only history as the cache prefix]]
- supports:: [[Catalog spine and epoch window make the prefix repeat]]
- supports:: [[The brief runs from stable to volatile]]
