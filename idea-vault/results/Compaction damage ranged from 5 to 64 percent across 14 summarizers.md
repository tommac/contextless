---
type: result
status: preprint
area: compression
needs: none
measured-by: others
---
Compact-Bench replayed 35 SWE-bench repairs with the full history and with only a summary, and found stable damage in 5 to 64 percent of 22 validated cases across 14 summarizers. Measured by exact substring, the best model summarizer retained 16.0 percent of recurring literals against 79.1 for a mechanical extractor. The intervals are wide at this sample size, about 1 to 22 percent for the best model and 43 to 80 for the worst.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Stable damage ranged from 5% to 64% of 22 validated cases across 14 summarizers"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Deterministic retention by exact substring found the best LLM summarizer at 16.0% against 79.1% for a mechanical extractor"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "about 1 to 22% for Opus 4.8 and 43 to 80% for DeepSeek-Flash on n=22"

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Lossy compression loses constraints and exact text]]
- supports:: [[A run that compacts is thrown away]]
- contradicts:: [[LLM summaries as memory]]
- source:: [[Compact-Bench]]
