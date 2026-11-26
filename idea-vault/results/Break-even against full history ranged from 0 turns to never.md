---
type: result
status: preprint
area: cost
needs: none
measured-by: others
---
A 2026 preprint that bills input tokens across ingest, retrieval and answer found that break-even against full history ranged from 0 to 86 turns for one memory system, 0 to 342 for another, and from 60 turns to never within 400 for a third. A second preprint found memory became cheaper after about ten turns at 100k tokens once prompt caching was modelled. The answer depends on the model, the cache behaviour and the workload, and the first paper reports no latency and no formal Pareto comparison.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Break-even against full history: Mastra OM 0 to 86 turns, Mem0 0 to 342, Hindsight 60 to never within 400"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "With prompt caching modeled, memory became cheaper after about ten turns at 100k tokens"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "no latency, no formal Pareto or budget-matched comparison"

## Relations
- supports:: [[A billed cost-quality frontier for verbatim memory]]
- supports:: [[Cost per correct answer]]
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- source:: [[Total Recall at What Cost]]
