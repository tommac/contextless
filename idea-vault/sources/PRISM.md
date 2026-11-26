---
type: source
status: external
area: selection
needs: none
year: 2026
url: https://arxiv.org/html/2605.12260v1
---
PRISM (Peng et al., arXiv 12 May 2026) is a conversational memory system that routes queries by intent, with keyword regexes first, and searches a typed-path graph with edge costs that depend on the intent. It uses an LLM at write time and in part at query time. On LoCoMo the keyword rules settled about a third of queries, while the typed-path search and intent-sensitive edge costs made no measurable difference. Its annotation of LoCoMo is one of the two audits the research note uses to put true multi-hop at about 3% to 6% of that benchmark's questions.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "routes by intent with keyword regexes first: 32.2% of LoCoMo queries were settled by keywords and 82.6% of temporal queries matched keyword triggers"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "graph path search and intent-weighted edges were null on LoCoMo, consistent with its estimate that about 6% of questions need a bridge"

## Relations
- supports:: [[Knowledge graphs lose to BM25 as a default store]]
- supports:: [[A query planner and EXPLAIN for context]]
- tested-by:: [[Graph path search made no measurable difference on LoCoMo]]
- tested-by:: [[Keyword rules settled 32.2 percent of LoCoMo queries with no model]]
