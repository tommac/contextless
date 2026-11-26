---
type: source
status: external
area: caching
needs: none
year: 2026
url: https://arxiv.org/abs/2601.06007
---
Lumer et al., January 2026, an evaluation of prompt caching for long-horizon agentic tasks. On DeepResearch Bench with a 10,000-token system prompt and 40 sessions per model and condition, caching cut cost by 27.8 to 81.4 percent depending on model and strategy; time to first token mostly improved and regressed in some settings. The research notes caution that the abstract's ranges do not match all table values and that the advice to put dynamic values last is guidance, not an isolated experiment. This design takes the layout of its brief from it.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "| Don't Break the Cache (2026) | caching cut cost 27.8 to 81.4 percent, by model and strategy | the layout in section 5 |"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "DeepResearch Bench, 10,000-token system prompt, 40 sessions per model and condition, more than 500 total."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "The recommendation to put dynamic values at the end of the system prompt is guidance, not an isolated experiment"

## Relations
- supports:: [[The brief runs from stable to volatile]]
- supports:: [[Measure the bill not the tokens]]
- tested-by:: [[Prompt caching cut cost 27.8 to 81.4 percent on agentic tasks]]
