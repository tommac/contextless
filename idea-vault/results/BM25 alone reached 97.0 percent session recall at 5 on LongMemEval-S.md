---
type: result
status: vendor-reported
area: selection
needs: none
measured-by: others
---
A practitioner ablation published with the open-memex tool ran SQLite FTS5 BM25 on cleaned LongMemEval-S (470 non-abstention questions, session level, retrieval only) and found recall at 5 of 97.0% and MRR 0.909 with no model. Switching term matching from OR to AND collapsed recall at 5 to 26.0%, removing prefix matching raised MRR to 0.919, and a second-chance synonym round fired on 0 of 470 queries. Pooling all 22,419 sessions into one corpus dropped recall at 5 to 40.6%. Other BM25-only reports on the same benchmark count a hit when any gold session is returned, so they say nothing about complete evidence coverage.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Baseline R@1 86.0%, R@5 97.0%, R@10 98.3%, MRR 0.909. Switching term matching from OR to AND collapsed R@5 to 26.0% and MRR to 0.257."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Removing prefix matching raised MRR to 0.919, so prefix matching was mildly harmful."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Pooling all 22,419 sessions into one corpus dropped R@5 to 40.6%."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "These count a hit when any gold session is returned, so they say nothing about complete evidence coverage."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A second-chance synonym round fired on 0 of 470 queries at limits up to 20 and rescued nothing."

## Relations
- supports:: [[Full text alone matched all lanes on real probes]]
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[The offline floor is the baseline every selector must beat]]
- source:: [[open-memex ablation]]
- source:: [[LongMemEval]]
