---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/pdf/2609.21940
---
AutoViewMem (Cao et al., arXiv 18 Sep 2026) organises memory into views, each an LLM-proposed schema with a name, slots and an extraction template. Every 50 chunks an LLM proposes 10 candidate views, a determinantal point process keeps 10 with low overlap, and an LLM extracts items per view with provenance span ids. It maps retrieved items back to source span ids before scoring, so several items from one span do not inflate coverage. It has no comparison with BM25 over raw turns, and both the views and their contents are written by a model.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "every 50 chunks an LLM proposes 10 candidates, a determinantal point process keeps 10 low-overlap views, and an LLM extracts items per view with provenance span ids"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "AutoViewMem uses a determinantal point process to pick views, not passages"

## Relations
- part-of:: [[Materialized views]]
- supports:: [[Views with lineage instead of summaries]]
- tested-by:: [[LLM-induced views scored 0.837 against 0.821 for full history on LoCoMo]]
