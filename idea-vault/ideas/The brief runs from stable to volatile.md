---
type: idea
status: built
area: caching
needs: none
---
The brief is laid out from most stable to least: session and standing instructions, decisions and facts in force, the catalog and digests in id order, the request-specific spans, the latest exchanges, and the message to answer last. No run id, timestamp or other per-call text comes before the volatile part, so each call can open the way the previous one did. The order costs nothing, puts the question at the end where models read best, and is the precondition for any prompt cache hit.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "The brief now follows this order, with the run id and task at the end and the message to answer last."
- `docs/CONTEXT-COMPILER.md`: "No run id, timestamp or other per-call text before the volatile part."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Stable reference first, question last satisfies both prefix caching and the best-supported position advice."

## Relations
- builds-on:: [[Each call gets a fresh packet built by code]]
- source:: [[Dont Break the Cache]]
- source:: [[Lost in the Middle]]
- tested-by:: [[Spine and epoch 4 raised the shared prefix from 0 to 37 45 and 44 percent]]
- tested-by:: [[Models use the start and end of long inputs best]]
