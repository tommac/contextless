---
type: result
status: measured
area: storage
needs: none
measured-by: us
---
Across 246,750 public message occurrences, 4 KiB lossless blocks used 113.26 MB including the address table. Moving to 64 KiB saved a further 17.75 MB. Fetching the same 2,000 short citations then required decoding 131.14 MB instead of 8.40 MB. Every source span matched. The measurement excludes session metadata, Git and search indexes.

## Evidence
- `reports/2026-11-20-large-memory/fixed-window-results.md`: "A 64 KiB block saved a further 17.75 MB, but those same scattered reads expanded 131.14 MB of data versus 8.40 MB with 4 KiB blocks."

## Relations
- supports:: [[Views with lineage instead of summaries]]
