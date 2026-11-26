---
type: result
status: claimed
area: plans
needs: none
measured-by: others
---
Anthropic's research system, the Claude Code agent teams documentation and practitioner reports all settle on 3 to 5 coordinated agents. It is a convergence of rules of thumb and not a measured optimum, since none of these sources reports a curve of quality or throughput against agent count. ContextlessAI's cap of five units per wave rests on this rule, and the survey proposes measuring throughput at 1, 2, 3 and 5 units to replace it with a local curve.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Convergent rule of thumb; no curve of quality against agent count exists"
- `research_notes/Parallel AI agents data structures/production_systems.md`: "is a convergence of rules of thumb, not a measured optimum; none of these sources reports a curve of quality or throughput against agent count"
- `research_notes/Parallel AI agents data structures/production_systems.md`: "No controlled measurement of marginal benefit per added agent for real LLM coding agents was found."

## Relations
- supports:: [[Plans run in waves with dependencies and shared decisions]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- part-of:: [[Universal Scalability Law]]
- source:: [[Anthropic multi-agent research system]]
