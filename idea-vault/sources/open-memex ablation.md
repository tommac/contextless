---
type: source
status: external
area: evaluation
needs: none
year: 2026
url: https://unpkg.com/open-memex@0.8.1/docs/retrieval-ablation-study.md
---
The open-memex retrieval ablation is a practitioner study dated 7 Oct 2026 that stores sessions in SQLite FTS5 and retrieves with BM25 and no model. On cleaned LongMemEval-S (470 non-abstention questions, session level, retrieval only) it tests term-matching options one at a time against a baseline of 97.0% recall at 5. It also reports MRR by question type, from 0.981 on knowledge-update down to 0.652 on single-session-preference. It is published with the tool and is not peer reviewed.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "open-memex retrieval ablation (practitioner study dated 7 Oct 2026; LongMemEval-S cleaned, 470 non-abstention questions, session-level, SQLite FTS5 BM25, retrieval only)"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "MRR by type: knowledge-update 0.981, multi-session 0.929, temporal-reasoning 0.851, single-session-preference 0.652."

## Relations
- supports:: [[Full text alone matched all lanes on real probes]]
- supports:: [[Recovery probes measure selection without a model]]
- supports:: [[Raw text beats rewritten memory]]
- tested-by:: [[BM25 alone reached 97.0 percent session recall at 5 on LongMemEval-S]]
