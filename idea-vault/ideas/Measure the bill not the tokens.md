---
type: idea
status: built
area: cost
needs: none
---
Billed input is not the same as tokens sent, because providers charge cached prefix tokens at a fraction of the normal price, so the order of the packet can matter as much as its size. Two 2026 results make the point: prompt caching cut cost by 27.8 to 81.4 percent, and removing 38.4 percent of tool-output tokens raised the billed cost by 6.8 percent. The project now records cache reads, cache writes and reported cost per run, and its evaluation rule is to report billed cost and time, not token counts.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "So measure the bill, not the tokens."
- `docs/CONTEXT-COMPILER.md`: "Providers charge cached prefix tokens at a fraction of the normal price, so the order of the packet can matter as much as its size."
- `docs/ROADMAP.md`: "runs record cache reads, writes and reported cost"

## Relations
- source:: [[Token Reduction Is Not Cost Reduction]]
- source:: [[Dont Break the Cache]]
- part-of:: [[A billed cost-quality frontier for verbatim memory]]
- builds-on:: [[Provider CLIs run as stateless workers]]
- tested-by:: [[Removing 38.4 percent of tool output tokens raised the bill 6.8 percent]]
- tested-by:: [[Prompt caching cut cost 27.8 to 81.4 percent on agentic tasks]]
