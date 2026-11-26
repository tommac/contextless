---
type: idea
status: proposed
area: index
needs: none
---
Keep one summary row per session holding its first and last turn, the maximum weight of each term and the value types present, and use the rows to guarantee that every session able to contribute to a question is at least considered. It is the zone map of a database turned from a speed device into a quality device. The published use of the same partitioning on conversational memory is for speed only, and no work was found on zone maps used for retrieval quality.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A zone-map analogue for quality is a per-session summary row (first and last turn, term maxima, value types present) used to guarantee that every session able to contribute is at least considered."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "AgentIR uses the same partitioning only for speed."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No work was found on block-max or zone maps used for retrieval quality, on learned indexes, on suffix arrays or FM-indexes, or on compressed bitmap indexes for conversational memory."

## Relations
- builds-on:: [[Digests label blocks of omitted turns]]
- builds-on:: [[Completeness by accounting against code-found candidates]]
- builds-on:: [[Whole-conversation questions are the weak spot]]
- analog-of:: [[Zone maps and block-max indexes]]
- alternative-to:: [[Slot allocation over deterministic aspects]]
