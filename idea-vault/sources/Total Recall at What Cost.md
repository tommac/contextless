---
type: source
status: external
area: cost
needs: none
year: 2026
url: https://arxiv.org/html/2608.11879
---
This August 2026 preprint compares three memory systems against a floor of the last ten turns and a ceiling of the full transcript, counting billable input tokens across ingest, retrieval and answer. It finds that the choice of model matters as much as the memory system for cost per correct answer. It reports no latency and no formal Pareto comparison, which is the gap the project's cost-quality frontier would fill.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "compares Mem0, Hindsight and Mastra OM against a last-10-turn floor and a full-transcript ceiling, using billable input tokens across ingest, retrieval and answer"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "backbone choice matters as much as the memory system"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "no latency, no formal Pareto or budget-matched comparison"

## Relations
- supports:: [[A billed cost-quality frontier for verbatim memory]]
- supports:: [[Cost per correct answer]]
- supports:: [[Measure the bill not the tokens]]
- tested-by:: [[Break-even against full history ranged from 0 turns to never]]
