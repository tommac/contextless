---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
EdgeMem opens its time channel only when the query carries an explicit month and four-digit year, and then reads the month buckets that its hubs keep. The gated channel added 0.58 strict-judge points on LoCoMo, with an interval of 0.07 to 1.15, and 0.45 when used alone on flat BM25. In SYNAPSE, switching off temporal decay dropped LoCoMo temporal F1 from 50.1 to 14.2. The gate's effect is small but measured with no model, which is the cheap end of time handling.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "a date gate that opens the time channel only when the query carries an explicit month and four-digit year"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the gated time channel as a whole adds +0.58 strict-judge points on LoCoMo"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "In SYNAPSE, switching off temporal decay dropped LoCoMo temporal F1 from 50.1 to 14.2."

## Relations
- supports:: [[Validity filters need an intent router for history questions]]
- supports:: [[Model-free hubs over verbatim turns]]
- source:: [[EdgeMem]]
- source:: [[Synapse]]
