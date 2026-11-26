---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
PRISM routes queries by intent with keyword regexes first: 32.2% of LoCoMo queries were settled by keywords, and 82.6% of temporal queries matched keyword triggers. Eywa's hand-coded planner classifies the shape of a query and sets channel weights, and EdgeMem raises the number of distinct sessions sampled from 3 to 5 for enumerative questions. None of these routing rules has an isolated ablation. The research note suggests cue words for enumerative or change-tracking questions as the model-free way to set the balance between relevance and diversity per query.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "32.2% of LoCoMo queries were settled by keywords and 82.6% of temporal queries matched keyword triggers"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "a hand-coded planner classifies query shape (exact, recommendation, inference, count, interval, boolean, list, factoid) and sets channel weights"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A deterministic proxy for the query type (enumerative or change-tracking cue words, as in EdgeMem's trigger) is the model-free counterpart."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No isolated ablation; keyword gates settle 32.2% of LoCoMo queries, May 2026."

## Relations
- supports:: [[A query planner and EXPLAIN for context]]
- supports:: [[Validity filters need an intent router for history questions]]
- part-of:: [[Query planning in databases]]
- source:: [[PRISM]]
- source:: [[Eywa]]
