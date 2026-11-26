---
type: result
status: vendor-reported
area: cost
needs: none
measured-by: others
---
Anthropic's own data says that on BrowseComp token usage alone explained 80% of performance variance, and that token usage, tool-call count and model choice together explained 95%. If tokens explain that much, a large part of the reported multi-agent gain is extra tokens spread over parallel context windows and not coordination as such. This is why a comparison of plans against single calls needs a budget-matched arm.

## Evidence
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "on BrowseComp, token usage alone explained 80% of performance variance and tokens plus tool-call count plus model choice explained 95%"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "if tokens explain 80% of variance, a large part of the 90.2% gain is extra tokens and parallel context windows, not coordination as such"

## Relations
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- supports:: [[cx compare measures plans against single calls blind]]
- source:: [[Anthropic multi-agent research system]]
