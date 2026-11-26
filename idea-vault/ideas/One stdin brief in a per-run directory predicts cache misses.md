---
type: idea
status: built
area: caching
needs: model-calls
---
When the report was written, each run got its own directory, Claude runs kept no session, and the brief arrived as one stdin message, so the caller placed no cache breakpoints at all. The report says this combination predicts misses on Anthropic and on GPT-5.6 and later, where reuse stops at block or message boundaries. This is a prediction from reading the code and the provider documents, not a measurement, and the cache truth table is its test.

## Evidence
- `reports/Cutting edge context memory research.md`: "each run gets its own directory (`runDir/ws` in `call.go`), Claude runs use `--no-session-persistence`, and the brief arrives as one stdin message, so the caller does not place cache breakpoints at all"
- `reports/Cutting edge context memory research.md`: "where reuse stops at block or message boundaries"

## Relations
- part-of:: [[R2 Cache truth table]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- contradicts:: [[Catalog spine and epoch window make the prefix repeat]]
- alternative-to:: [[Shared content goes in its own block ending at a breakpoint]]
- builds-on:: [[Provider CLIs run as stateless workers]]
