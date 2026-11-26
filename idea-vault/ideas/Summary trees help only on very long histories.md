---
type: idea
status: external
area: structure
needs: none
---
Trees and summary DAGs show real but modest gains that depend on the setup, and the designs that hold up keep raw text at the leaves. MemTree lost to full history at 15 rounds and won only at 200, and LCM trailed Claude Code at 8K and 16K and led from 32K up. Most tree papers leave BM25 and full context out of their main tables, so for ContextlessAI a summary may only ever be a navigation aid over text that is kept whole.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Trees and summary DAGs show real but modest and setup-dependent gains, and the designs that hold up keep raw text at the leaves."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Claude Code leads at 8K (+13.1 vs +11.2) and 16K (+26.3 vs +25.0); Volt leads from 32K up"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Most tree papers omit BM25 and full-context baselines on their main tables, so added value over simple retrieval is often unknown."

## Relations
- supports:: [[A short history is sent whole]]
- supports:: [[Catalog cards are an index not a replacement]]
- tested-by:: [[MemTree lost to full history at 15 rounds and won at 200]]
- tested-by:: [[Decomposition gained at 256K tokens and trailed at 8K and 16K]]
- source:: [[MemTree]]
- source:: [[RAPTOR]]
- source:: [[LCM Lossless Context Management]]
