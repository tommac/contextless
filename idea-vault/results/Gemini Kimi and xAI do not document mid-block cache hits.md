---
type: result
status: vendor-reported
area: caching
needs: none
measured-by: others
---
On the pages read on 2026-11-18, Google, Moonshot and xAI do not say whether a shared start inside one block or part is served from cache. Gemini gives no rule on content, part or token granularity; Kimi says content after a changed position cannot be reused; xAI says to append new messages and that editing, deleting or reordering earlier ones breaks the prefix. For these silent cases the notes advise a small live probe: send A+B then A+C in one text block, repeat with A as a separate block, and compare cache fields over several trials.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Google, Moonshot and xAI do not document mid-block or mid-part hit behaviour on the pages read."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "send `A+B` then `A+C` in one text block, with `A` above the minimum, and repeat with `A` as a separate block"

## Relations
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Shared content goes in its own block ending at a breakpoint]]
