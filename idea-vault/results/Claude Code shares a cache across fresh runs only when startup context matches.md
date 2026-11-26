---
type: result
status: vendor-reported
area: caching
needs: none
measured-by: others
---
Claude Code's documentation says fresh-run cache reuse is effectively limited to one machine and directory. Sequential sessions share a cache only when the startup git-status snapshot matches, because the system context carries auto-memory paths, working directory, platform, shell, OS, branch and recent commits. Each run of this project has its own directory, so its headless runs should not be assumed to share a cache. Headless runs default to a 1-hour lifetime on a subscription and 5 minutes with an API key.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Fresh-run reuse is effectively limited to one machine and directory. Parallel sessions in the same directory can share. Sequential sessions share only when the startup git-status snapshot matches"
- `docs/CONTEXT-COMPILER.md`: "Claude Code shares a cache across fresh runs only when its startup context matches, and each run here has its own directory."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Default TTL is 1 hour on a subscription within plan usage and 5 minutes with usage credits, API key or cloud provider."

## Relations
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Shared content goes in its own block ending at a breakpoint]]
- contradicts:: [[The append-only history as the cache prefix]]
