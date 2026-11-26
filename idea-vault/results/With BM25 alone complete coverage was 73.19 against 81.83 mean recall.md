---
type: result
status: preprint
area: evaluation
needs: none
measured-by: others
---
Threader is the one source found that reports complete-evidence coverage beside mean recall. With BM25 as the only signal, 73.19% of queries had all their evidence in the top 5 units while mean evidence recall was 81.83%, a gap of 8.6 points; at top 12 the figures were 84.68 and 91.06. With the full pipeline the top-5 figures were 82.55 and 90.26. The table does not name its benchmark, though its accuracies match the paper's LongMemEval results, and the row is at segment granularity from a single paper. The research note calls it the clearest published statement of headroom for a lexical system.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "With BM25 as the only signal (dense and reranker removed): 73.19 against 81.83 at top 5 and 84.68 against 91.06 at top 12."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Threader's BM25-only gap of 8.6 points between mean recall and complete coverage at top 5 is the clearest published statement of headroom for a lexical system."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Full-R denotes the fraction of queries with complete evidence coverage, and Mean-R denotes average evidence recall"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Threader's BM25-only row is at segment granularity and from a single paper"

## Relations
- supports:: [[Whole-conversation questions are the weak spot]]
- supports:: [[Complete-evidence coverage at a fixed budget is the outcome measure]]
- supports:: [[Completeness by accounting against code-found candidates]]
- supports:: [[Slot allocation over deterministic aspects]]
- source:: [[Threader]]
