---
type: field
status: external
area: selection
needs: none
---
Region algebra treats the answer to a query as a set of text intervals and combines intervals with seven operators, among them containing, contained in and followed by. Cover density ranking, from the same group, scores a document by the shortest intervals that cover the query terms. The research note found no use of either on conversational memory and no measured application of interval joins to conversation spans. Lifted from words to spans, cover density becomes the shortest run of consecutive spans that covers the query terms, which would return contiguous evidence windows.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "queries return sets of text intervals; seven operators combine intervals (containing, not containing, contained in, not contained in, one of, both of, followed by)"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The same group's shortest-substring and cover density ranking scores a document by the shortest intervals that cover the query terms. No use on conversational memory was found."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No measured application of interval joins, region algebra or cover density to conversation spans was found."

## Relations
- part-of:: [[Proximity scoring]]
- analog-of:: [[Atoms are paragraph-sized addresses into a turn]]
- analog-of:: [[Quotes are pinned to the paragraphs that hold them]]
