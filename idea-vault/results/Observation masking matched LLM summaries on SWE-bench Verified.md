---
type: result
status: preprint
area: compression
needs: none
measured-by: others
---
The Complexity Trap (JetBrains Research and TUM): with SWE-agent on SWE-bench Verified, 500 instances per model, replacing old tool output with a placeholder gave solve rates about equal to LLM summarization at similar or lower cost. Only two solve-rate changes are significant, and both are losses on Gemini 2.5 Flash with thinking: 4.0 points for masking and 9.0 for summaries. Summary calls took up to 7.2 percent of instance cost and lengthened runs, and neither method reliably beats a full-context agent.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "For coding agents, dropping or masking old tool output is about as good as LLM summarization and cheaper, but neither reliably beats a full-context agent on solve rate."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "both are losses on Gemini 2.5 Flash with thinking: masking −4.0 pp (95% CI −7.8 to −0.2) and summary −9.0 pp (CI −13.0 to −5.2)"

## Relations
- supports:: [[A record goes in whole or is listed by id]]
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Measure the bill not the tokens]]
