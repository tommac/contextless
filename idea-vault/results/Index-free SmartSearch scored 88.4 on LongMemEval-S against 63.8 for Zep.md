---
type: result
status: preprint
area: storage
needs: none
measured-by: others
---
SmartSearch keeps raw conversation text with no ingest step and no index, and scored 88.4 on LongMemEval-S against 63.8 for Zep and 65.6 for full context. The benchmark separates systems most on the categories that need several pieces of evidence: multi-session 84.2 against 47.4 for Zep and 51.1 for full context, and temporal reasoning 82.7 against 54.1 and 60.2. Its substring scan resolved 98.9% of LoCoMo oracle traces, and its 650 ms per query is dominated by rerankers. The ranking is done by a cross-encoder and ColBERT, so this is raw text with neural ranking and not a model-free result.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LongMemEval-S 88.4 against Zep 63.8 and full context 65.6"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "multi-session 84.2 (SmartSearch index-free) against 47.4 (Zep) and 51.1 (full context); temporal reasoning 82.7 against 54.1 and 60.2"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "SmartSearch scans the raw conversation by substring with no index, and that scan resolved 98.9% of LoCoMo oracle traces; its 650 ms per query is dominated by rerankers"

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Knowledge graphs lose to BM25 as a default store]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Unrelated context can hurt answers]]
- source:: [[SmartSearch]]
- source:: [[LongMemEval]]
