---
type: idea
status: proposed
area: index
needs: download
---
Keep, for each key the value lane sees, an ordered list of (turn, value, span id) with a validity interval derived for each value, maintained incrementally as turns arrive. A question about how something changed is then answered by reading the chain of change points with their neighbours, which is closer to a group-by on a key and an order-by on time than to hopping between passages. The turn index gives an exact transaction time, and the interval of each value runs from the turn that asserts it to the next different value for the same key. The research note proposes testing it on LongMemEval knowledge-update and temporal questions and on BEAM event ordering and knowledge update. Prior art is LLM-derived (Zep, Stashbird), unablated (DMF) or unmeasured (Bitacorista).

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A per-key change ledger maintained incrementally from the value lane: ordered (turn, value, span id) with derived validity intervals, queried as the chain of change points plus neighbours."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Test on LongMemEval knowledge-update and temporal questions and on BEAM event ordering and knowledge update. Prior art is LLM-derived (Zep, Stashbird), unablated (DMF) or unmeasured (Bitacorista)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "for a key with a sequence of asserted values, the interval of each value runs from its assertion turn to the next different value for the same key. Transaction time is the turn index and is always exact."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "closer to a group-by on a key followed by an order-by on time than to bridge hopping"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No blame or bisect analogue for conversational facts was found, in papers or tools."

## Relations
- part-of:: [[R11 Value chains and time intent]]
- builds-on:: [[The value lane finds stated values near the question's words]]
- builds-on:: [[Derived validity records over immutable turns]]
- builds-on:: [[Whole-conversation questions are the weak spot]]
- analog-of:: [[Version control blame and bisect]]
- analog-of:: [[Bitemporal data model]]
- part-of:: [[Materialized views]]
