---
type: idea
status: measured-kept
area: compression
needs: none
---
Across the 2026 studies in the research notes, the most repeatable harm from lossy compaction is the loss of constraints and byte-exact text, not of general gist. Constraint violations rose from 0 to 30 percent after compaction, compactors kept 17 percent of injected side constraints, and a compressed edit task lost patch applicability while saving cost. All of that was measured by others, mostly in preprints. This project therefore proposes no prompt compression and no summaries in place of text.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "The most repeatable harm from lossy compaction is the loss of constraints and byte-exact text, not loss of general gist."
- `docs/CONTEXT-COMPILER.md`: "Not proposed: prompt compression (LLMLingua and similar), summaries in place of text, embeddings before you approve a local model download."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Never compress text that will be used as an exact edit anchor, an identifier, a path, or a quoted requirement."

## Relations
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[A run that compacts is thrown away]]
- source:: [[LLMLingua]]
- tested-by:: [[Constraint violations rose from 0 to 30 percent after compaction]]
- tested-by:: [[Compactors kept 17 percent of injected side constraints]]
- tested-by:: [[Compression cut cost 59 percent and patch applicability from 27 to 15 of 40]]
