---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
PRISM's typed-path graph search and intent-sensitive edge costs made no measurable difference on LoCoMo, which fits its own estimate that about 6% of questions need an explicit two-hop bridge. SmartSearch's oracle traces put 97.0% of LoCoMo at a single hop, and its rule-based entity hop moved LoCoMo by 0.1. With no reranker and under 1,000 evidence tokens, EdgeMem's co-occurrence channel still added 2.21 points. The research note concludes that LoCoMo is a weak testbed for bridge-style mechanisms and that LongMemEval multi-session and temporal questions are the smallest credible targets.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Its typed-path graph search and intent-sensitive edge costs made no measurable difference on LoCoMo."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Oracle hop distribution on LoCoMo: 97.0% single-hop, 2.8% two-hop, 0.2% three-hop."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LoCoMo is a weak testbed for bridge-style mechanisms: two independent audits put true multi-hop at about 3% to 6% of its questions."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "73.4% of category 1 to 4 questions cite a single evidence entry, and about 6% (95% interval 2.1 to 16.5%, 50 sampled questions) need an explicit two-hop bridge"

## Relations
- supports:: [[Knowledge graphs lose to BM25 as a default store]]
- supports:: [[Feedback links and coverage did not beat search]]
- contradicts:: [[A LoCoMo leaderboard position as a headline metric]]
- part-of:: [[Personalized PageRank]]
- source:: [[PRISM]]
- source:: [[LoCoMo]]
