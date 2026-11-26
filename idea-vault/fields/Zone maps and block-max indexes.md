---
type: field
status: external
area: index
needs: none
---
A block-max index stores the maximum score of each block of postings so that whole blocks can be skipped with results identical to exhaustive scoring. Small materialized aggregates, the database form of the same idea, precompute per-bucket aggregates such as minimum and maximum. The research note judges that at the scale of one conversation these speed structures, with skip lists and learned indexes, solve a problem the system does not have, and the one published use on conversational memory is for speed only. The analogue worth testing is for quality: a summary row per session, holding its first and last turn, term maxima and value types present, used to guarantee that every session able to contribute is at least considered.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "a block-max index stores the maximum score of each compressed posting block so whole blocks can be skipped, with results identical to exhaustive scoring"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "At 10^5 to 10^6 tokens per conversation, block-max, skip lists and learned indexes solve a problem the system does not have."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A zone-map analogue for quality is a per-session summary row (first and last turn, term maxima, value types present) used to guarantee that every session able to contribute is at least considered."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "On LongMemEval the cascade skipped 63% of queries and was 2.67x faster (19.9 ms against 53.2 ms) at equal judged accuracy"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Absolute judged accuracy was low (0.254 hybrid), so this is speed evidence only."

## Relations
- analog-of:: [[Digests label blocks of omitted turns]]
- analog-of:: [[Levels of detail like map tiles]]
- analog-of:: [[Completeness by accounting against code-found candidates]]
