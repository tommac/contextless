---
type: idea
status: rejected
area: plans
needs: none
---
Relaying a unit's result through another model's restatement is the default in hub and spoke systems, where subagents return condensed findings through the coordinator's context. Anthropic calls the effect a game of telephone and moved to subagents writing to external storage and returning lightweight references. The survey says not to let any model summarize a unit's output for another model. The record is passed verbatim or by id, which is the engine's founding principle applied to unit outputs.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Do not let any model summarize a unit's output for another model; pass the record verbatim or by id, which is the engine's founding principle applied to unit outputs."
- `reports/Parallel AI agents data structures.md`: "Anthropic's fix is for subagents to write to external storage and return lightweight references"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "information is lost in summarization, for example a correct intermediate answer in a math debate was dropped and later agents reinforced a wrong one"

## Relations
- alternative-to:: [[Units return typed records with checked claims]]
- alternative-to:: [[Turns are stored verbatim and never rewritten]]
- contradicts:: [[Raw text beats rewritten memory]]
- source:: [[Anthropic multi-agent research system]]
