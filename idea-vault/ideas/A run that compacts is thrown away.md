---
type: idea
status: built
area: compression
needs: none
---
Every provider CLI is run with its compaction switch off. If a compaction signal appears anyway, or the run's peak use passes 90 percent, the run is tainted: its output is never used and the call is retried on a larger profile. A watchdog also kills a run at 85 percent of the window. So no stored answer comes from a run whose context was compacted.

## Evidence
- `docs/ARCHITECTURE.md`: "A tainted run (compaction seen, peak over 90 percent) is never used; it is retried on a larger profile."
- `docs/ARCHITECTURE.md`: "with the watchdogs from BookWriter (kill at 85 percent of the window, taint on any compaction event)."

## Relations
- builds-on:: [[Provider CLIs run as stateless workers]]
- builds-on:: [[A larger window is tried before any record is left out]]
- builds-on:: [[The plan cap is 60 to 70 percent of the window]]
- builds-on:: [[Lossy compression loses constraints and exact text]]
- source:: [[BookWriter]]
