---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
In a benchmark that builds topic pivots by joining separate conversations, o3 detected pivots with F1 0.973 on TopiOCQA and 0.763 on the harder MSC, and GPT-4o reached only 0.695 on the easier set. Weaker models carried stale context over at nearly every true pivot. Answer quality and tokens were not measured, so the result says detection is feasible on clean pivots and nothing yet about answers.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Pivot F1 on TopiOCQA without cues: o3 0.973, Claude 3.7 Sonnet 0.941, Gemini 2.5 Pro 0.871, GPT-4o 0.695, Mistral-7B 0.214"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "MSC is harder: o3 0.763, Claude 3.7 Sonnet 0.703."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Over-carry of stale context at true pivots reached 1.000 for GPT-4o-mini and 0.991 for Mistral-7B."

## Relations
- supports:: [[Threads group turns by topic]]
- supports:: [[Boundary detection should be judged by downstream answers]]
- supports:: [[Natural tangent-and-return evaluation]]
