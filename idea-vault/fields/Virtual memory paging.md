---
type: field
status: external
area: recall
needs: none
---
The design treats a recall request the way an operating system treats a page fault, and takes three mechanisms from paging. The working set keeps recently recalled or cited atoms in the next packets, which is temporal locality. Prefetch brings in the graph neighbours of a faulted atom, which is spatial locality, and the fault rate per call becomes a health metric.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "**Recall as paging**: working set, prefetch, fault rate (section 6)."
- `docs/CONTEXT-COMPILER.md`: "After a fault on an atom, its graph neighbors join the next packet (spatial locality)."

## Relations
- analog-of:: [[Recall as paging]]
- analog-of:: [[Recall by id or by search replaces memory]]
