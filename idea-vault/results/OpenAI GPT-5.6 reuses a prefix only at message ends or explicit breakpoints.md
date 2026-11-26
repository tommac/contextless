---
type: result
status: vendor-reported
area: caching
needs: none
measured-by: others
---
For GPT-5.6 and later, OpenAI documents implicit breakpoints at the end of eligible messages. Extending a message moves the old endpoint inside it, and without an explicit breakpoint there the prefix is not reused. Implicit lookup considers up to 20 earlier eligible message endings, and the minimum is 1,024 visible input tokens. Earlier OpenAI models place breakpoints at model-dependent intervals, and the documentation does not say whether a hit can end inside a block.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Implicit breakpoints fall at the end of eligible messages. Extending a message moves the old endpoint inside it, and without an explicit breakpoint there the prefix is not reused"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "GPT-5.6+ implicit lookup considers up to 20 earlier eligible message endings plus the end of the initial developer block."

## Relations
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Shared content goes in its own block ending at a breakpoint]]
