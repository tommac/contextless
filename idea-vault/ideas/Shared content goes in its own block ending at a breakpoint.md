---
type: idea
status: proposed
area: caching
needs: model-calls
---
Anthropic and current OpenAI models reuse a cached prefix only up to block or message boundaries, so a shared start inside one user text block should not be assumed to hit. The portable layout is to place the shared content in its own leading block or message, close it there, and put the query or variable material in a later block. For fresh headless runs the harness's own prefix matters as much: the working directory, git snapshot, tool set, auth path and session-derived key can each prevent reuse.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Portable layout: place the shared content in its own leading block or message, close it there, and put the query or variable material in a later block."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "The working directory, git snapshot, tool set, auth path and session-derived key can each prevent reuse."

## Relations
- builds-on:: [[The brief runs from stable to volatile]]
- builds-on:: [[Prompt cache savings are a ceiling until a live test]]
- builds-on:: [[Provider CLIs run as stateless workers]]
- tested-by:: [[Anthropic caches a prefix only through a marked block]]
- tested-by:: [[OpenAI GPT-5.6 reuses a prefix only at message ends or explicit breakpoints]]
- tested-by:: [[Codex replay got 0 percent hits until an explicit breakpoint cached 98.6]]
