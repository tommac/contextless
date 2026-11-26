---
type: result
status: vendor-reported
area: structure
needs: none
measured-by: others
---
LangChain reports that a deterministic evidence graph, built from existing citations with no extra model calls, added 27 percent accuracy over its wiki alone and lowered the cost per correct answer. It is a vendor report on a synthetic corpus, and the question subset behind the accuracy figure is not stated. On the same corpus, connected questions scored 0.10 for keyword search and 0.52 for the wiki.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "LangChain's deterministic evidence graph, built from existing citations with no extra model calls, reported +27% accuracy"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "the question subset for the accuracy figure is not stated"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Connected questions: search 0.10, wiki 0.52. The baseline was keyword search only"

## Relations
- supports:: [[Model-free hubs over verbatim turns]]
- supports:: [[Search first then bounded query-filtered expansion]]
- supports:: [[Links help when questions need several pieces of evidence]]
