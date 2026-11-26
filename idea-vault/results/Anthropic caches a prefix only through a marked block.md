---
type: result
status: vendor-reported
area: caching
needs: none
measured-by: others
---
Anthropic's documentation says a marked block writes one cache entry, a hash of the prefix ending at that block, and no earlier positions; a hit needs identical content through the marked block. So a shared start inside one user text block does not hit by documented design, though the docs do not address the mid-block case in those words. The prefix order is tools, then system, then messages, and a change at one level invalidates that level and the later ones. No live cache test was run for the research notes.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "A marked block writes one entry, a hash of the prefix ending at that block, and no earlier positions. Lookback checks block positions. A hit needs identical content through the marked block"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Anthropic's prefix order is `tools` → `system` → `messages`, and a change invalidates that level and later ones."

## Relations
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Shared content goes in its own block ending at a breakpoint]]
- supports:: [[The brief runs from stable to volatile]]
