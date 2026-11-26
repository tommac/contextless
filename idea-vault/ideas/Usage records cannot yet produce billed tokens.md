---
type: idea
status: designed
area: cost
needs: none
---
When the report was written, its reading of the code found that the usage records could not produce billed tokens, which blocks honest billing comparisons. The usage type had no cache-write field, the Claude adapter added cache-creation tokens into input so a write at a premium looked like ordinary input, and the Codex adapter ignored the cache-write field that Codex emits. The compare path totalled only input and output and had no full-history, BM25 or recency arm. Repairing this is the first step of the agenda, and another note in this vault says runs now record cache reads and writes, so the code should be checked before this finding is relied on.

## Evidence
- `reports/Cutting edge context memory research.md`: "`Usage` in `internal/providers/types.go` has input, cached and output fields but no cache-write field."
- `reports/Cutting edge context memory research.md`: "so a write at a premium looks like ordinary input"
- `reports/Cutting edge context memory research.md`: "repair the instrument (the usage records cannot currently produce billed tokens)"

## Relations
- part-of:: [[R1 Repair the instrument and record the offline floor]]
- contradicts:: [[Measure the bill not the tokens]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- builds-on:: [[Provider CLIs run as stateless workers]]
- builds-on:: [[cx compare measures plans against single calls blind]]
