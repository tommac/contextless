---
type: result
status: preprint
area: caching
needs: none
measured-by: others
---
On DeepResearch Bench with a 10,000-token system prompt and 40 sessions per model and condition, prompt caching cut cost by 27.8 to 81.4 percent depending on model and strategy. Time to first token mostly improved but regressed in some settings. Differences between caching strategies were only about 2 to 4 points for most models, because the system prompt is cached in all of them.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "On long agentic tasks, prompt caching cut cost by 27.8 to 81.4 percent depending on model and strategy, and first-token time mostly improved but regressed in some settings"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Strategy differences on cost are only about 2 to 4 points for most models because the system prompt is cached in all of them."

## Relations
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Measure the bill not the tokens]]
- source:: [[Dont Break the Cache]]
