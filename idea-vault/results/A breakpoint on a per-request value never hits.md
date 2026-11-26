---
type: result
status: vendor-reported
area: caching
needs: none
measured-by: others
---
Anthropic's documentation says a cache breakpoint placed on a value that changes with every request, such as a timestamp, writes a new entry each time and never hits; one breakpoint at the end of the static content is usually enough. Claude Code's own engineering post lists past cache breaks from a timestamp in the static prompt, non-deterministic tool ordering and changed tool parameters, which is a first-party claim with no checkable method. This is the provider-side reason the brief keeps its run id and task out of the stable part.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "a breakpoint on a per-request value such as a timestamp writes a new entry each time and never hits"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Past breaks included a timestamp in the static prompt, non-deterministic tool ordering, and changed tool parameters."

## Relations
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Shared content goes in its own block ending at a breakpoint]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
