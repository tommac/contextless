---
type: result
status: vendor-reported
area: structure
needs: none
measured-by: others
---
Zep and Graphiti put four timestamps on each fact edge, created and expired on the transaction timeline and valid and invalid on the event timeline, and an LLM detects contradictions and sets the older edge's invalid time. In the vendor's LongMemEval-S results with gpt-4o, temporal reasoning rose from 45.1 to 62.4 against full context, multi-session from 44.3 to 57.9 and knowledge update from 78.2 to 83.3, while single-session-assistant fell from 94.6 to 80.4. Raw episodes are kept and linked both ways to derived facts, but those links were not evaluated. No ablation separates the bitemporal model from LLM fact extraction, so the gain cannot be credited to the time model alone.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "with gpt-4o temporal reasoning 45.1 to 62.4, multi-session 44.3 to 57.9, knowledge update 78.2 to 83.3, single-session-assistant 94.6 to 80.4"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "put four timestamps on each fact edge (created and expired on the transaction timeline, valid and invalid on the event timeline); an LLM detects contradictions and sets the older edge's invalid time"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No ablation was found that separates bitemporal modelling from LLM fact extraction."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Raw episodes are kept and linked both ways to derived facts, but those links were not evaluated."

## Relations
- supports:: [[Facts supersede facts]]
- supports:: [[Derived validity records over immutable turns]]
- part-of:: [[Bitemporal data model]]
- source:: [[Zep and Graphiti]]
