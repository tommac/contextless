---
type: field
status: external
area: structure
needs: none
---
A materialized view is a query result that a database stores and keeps up to date, so that a question is answered by a lookup instead of a scan. The research note found no work that keeps deterministic, incrementally updated views (per topic, per entity, timeline, decision log) over a verbatim conversation log and measures their retrieval value alone. The nearest systems are EdgeMem's append-only hubs, DMF's deterministic cards with no ablation, Bitacorista's regex decision log with no measurement, Stashbird's reconciliation over LLM extraction and AutoViewMem's LLM-induced views. For this project the candidate views are postings per key, a timeline per key, a value sequence per key and a decision log, each a pure function of the stored turns that can be rebuilt or audited by replay. A view keyed by a subject would turn a question about everything tried for that subject, or about how a plan changed, into a point lookup.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No work was found that maintains deterministic, incrementally updated views (per topic, per entity, timeline, decision log) over a verbatim conversation log and measures their retrieval value in isolation."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A view keyed by X turns both into point lookups."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A view that is a pure function of the log is reproducible and can be rebuilt or audited by replay, which fits a deterministic prompt builder"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "RisingWave and Materialize propose materialized views over conversation history for preference summaries and session statistics. No independent benchmark was found for any of the three."

## Relations
- analog-of:: [[Views with lineage instead of summaries]]
- analog-of:: [[Catalog cards are an index not a replacement]]
- analog-of:: [[Idea notes as verified keys over verbatim paragraphs]]
- analog-of:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[DMF's deterministic cards scored 0.7753 against 0.6883 for Mem0 on LoCoMo]]
- tested-by:: [[LLM-induced views scored 0.837 against 0.821 for full history on LoCoMo]]
- tested-by:: [[Plain retrieval scored 0.00 exact match on TAG table questions]]
