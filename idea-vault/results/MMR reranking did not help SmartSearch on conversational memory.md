---
type: result
status: claimed
area: selection
needs: none
measured-by: others
---
SmartSearch tried MMR reranking on conversational memory and reports that it did not help, with no figure, and that pruning by cumulative score cost 2.3 points. The research note notes that this null result was obtained with a cross-encoder and at budgets where recall is already .93 to .96. Diversification should matter most where the budget is tight and there is no neural ranker, which is the target system's regime; the research note marks that as a hypothesis.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "and that pruning by cumulative score cost 2.3 points"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "SmartSearch's null result for MMR was obtained with a cross-encoder and budgets where recall is already .93 to .96."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Diversification should matter most where the budget is tight and there is no neural ranker, which is the target system's regime; that is a hypothesis."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Tried and rejected in SmartSearch"

## Relations
- supports:: [[Feedback links and coverage did not beat search]]
- part-of:: [[Result diversification]]
- source:: [[SmartSearch]]
