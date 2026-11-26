---
type: result
status: preprint
area: compression
needs: none
measured-by: others
---
In a single-shot SEARCH/REPLACE study with Sonnet 4.6 on 40 SWE-bench-derived Go tasks, compression cut billable tokens 73 percent and cost 59 percent, but patch applicability fell from 27 of 40 to 15 of 40 and resolved tasks from 2 to 1. An edit needs the exact bytes of the text it replaces, and compression does not keep them. The sample is small.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "compression cut billable tokens 73% and cost 59%, but patch applicability fell from 27/40 to 15/40, and resolved tasks went from 2/40 to 1/40"

## Relations
- supports:: [[Lossy compression loses constraints and exact text]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Raw text beats rewritten memory]]
- source:: [[Token Reduction Is Not Cost Reduction]]
