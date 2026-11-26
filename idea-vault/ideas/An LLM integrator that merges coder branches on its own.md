---
type: idea
status: rejected
area: plans
needs: none
---
A single agent that resolves conflicts and merges every worker's branch looks like the natural join for coding agents. At Cursor the integrator became a bottleneck with hundreds of workers and was removed, and requiring a fully correct tree before each commit serialized the system. The survey says not to build an LLM integrator, a merge queue beyond the scratch check, or automatic merging of any kind. In ContextlessAI nothing lands by itself and the user still decides.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Do not build an LLM integrator that merges coder branches on its own, a merge queue beyond the scratch check, or automatic merging of any kind."
- `research_notes/Parallel AI agents data structures/production_systems.md`: "Merge queues as such are not described in any of the vendor sources read. The closest reported analogue, Cursor's integrator, was removed."
- `reports/Parallel AI agents data structures.md`: "Nothing lands; the user still decides."

## Relations
- alternative-to:: [[A scratch merge of coder branches runs verify once]]
- analog-of:: [[Merge queues]]
- source:: [[Cursor planner and worker agents]]
