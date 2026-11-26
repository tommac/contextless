---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
SmartSearch keeps raw conversation text and adds a light feedback pass: nouns and proper nouns that occur in at least 2 of the top 10 passages, at weight 0.5, for one extra pass. On LongMemEval-S (baseline 78.4) feedback added 6.0 accuracy points and a rule-based entity hop 1.4, together 9.2; on LoCoMo (baseline 91.0) the changes were -0.1, +0.1 and -0.3. The research note reads this as feedback acting as vocabulary repair in a single user's conversation. The gain was obtained with a strong neural reranker downstream, so whether it survives in a purely lexical pipeline is untested.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LoCoMo (baseline 91.0): feedback -0.1, entity discovery +0.1, both -0.3. LongMemEval-S (baseline 78.4): feedback +6.0, entity discovery +1.4, both +9.2, with temporal questions rising from 67.7 to 80.5."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The published gain was obtained with a strong neural reranker downstream; whether it survives in a purely lexical pipeline is untested."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Its feedback pass adds nouns and proper nouns that occur in at least 2 of the top 10 passages, weight 0.5, one extra pass."

## Relations
- supports:: [[Feedback expansion from first-pass hits needs no model]]
- supports:: [[Feedback gated by the top-score gap]]
- contradicts:: [[Feedback links and coverage did not beat search]]
- part-of:: [[Pseudo-relevance feedback]]
- source:: [[SmartSearch]]
