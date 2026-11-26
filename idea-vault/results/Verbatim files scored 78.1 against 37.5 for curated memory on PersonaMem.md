---
type: result
status: preprint
area: storage
needs: none
measured-by: others
---
In a third-party preprint on files as memory, verbatim session files scored 78.1 on PersonaMem against 71.9 for BM25 chunks and 37.5 for memory curated by an agent, and 77.6 against 41.2 for a condensed store on REALTALK. Organizing the files reduced search cost but did not improve answers for any agent measured. Vector and graph stores were not tested, and repeated judging moved one score by about 1.3 points.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "On PersonaMem 32k, verbatim 78.1, BM25 chunks 71.9, curated 37.5. On REALTALK, verbatim 77.6 vs condensed 41.2."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Organization reduced search cost but did not improve answers for any agent measured."
- `reports/Cutting edge context memory research.md`: "Verbatim files beat condensed memory"

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Model-written notes lose as storage]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Files are the truth for the conversation and SQLite for the machine]]
- contradicts:: [[LLM summaries as memory]]
