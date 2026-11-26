---
type: result
status: claimed
area: caching
needs: none
measured-by: others
---
In a Codex issue opened 2026-07-25, reporters replayed requests with a stable prefix of about 9k tokens and a changed final user message: unmodified requests got 0 percent cache hits, while an explicit breakpoint on the last stable block cached 8,937 tokens, 98.6 percent. Within one session, append-only turns reached about 98 percent. The measurement is by issue reporters and is not confirmed by OpenAI, and other users report that some backends reject the breakpoint field.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "unmodified requests got 0% hits, while an explicit breakpoint on the last stable block cached 8,937 tokens (98.6%)"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Within one session, append-only turns reached about 98%."

## Relations
- supports:: [[Shared content goes in its own block ending at a breakpoint]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[The append-only history as the cache prefix]]
