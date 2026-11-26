---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
In AutoViewMem a view is an LLM-proposed schema with a name, slots and an extraction template; a determinantal point process keeps 10 low-overlap views and an LLM extracts items per view with provenance span ids. On LoCoMo with Qwen3-8B its overall judge score was 0.837 against 0.821 for full history, 0.783 for MemGAS and 0.683 for Mem0. In the ablation, random view choice scored 0.777 and raised the share of queries with no gold evidence in the retrieved pool from 5.14% to 7.03%. There is no comparison with BM25 over raw turns, and both the views and their contents are written by a model.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "overall judge 0.837 against MemGAS 0.783, A-mem 0.720, Mem0 0.683 and full history 0.821"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Ablation: single-view structured extraction 0.774, random view choice 0.777, no consolidation 0.801; share of queries with no gold evidence in the retrieved pool 5.14% (full), 5.53%, 7.03%, 9.83%."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No comparison with BM25 over raw turns."

## Relations
- supports:: [[Views with lineage instead of summaries]]
- part-of:: [[Materialized views]]
- source:: [[AutoViewMem]]
