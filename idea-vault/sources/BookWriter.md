---
type: source
status: external
area: packets
needs: none
---
BookWriter is the system whose version 2 engine ContextlessAI generalizes from books to any conversation. Its core idea is that the model's context window is never the memory. ContextlessAI copied its provider layer whole and re-implemented its packet sizing, drop tiers and brief for chat records. Its vault design, handoff contract, routing table, cooldown ladder, judge independence and crash recovery were carried over as well, and its measurements on a book seed the order of providers for each role.

## Evidence
- `docs/ARCHITECTURE.md`: "ContextlessAI is a general chat assistant built on BookWriter v2's core idea:"
- `docs/ARCHITECTURE.md`: "the model's context window is never the memory."
- `docs/ARCHITECTURE.md`: "BookWriter `internal/bw/packet` (re-implemented for chat records)"
- `docs/ARCHITECTURE.md`: "Provider adapters, sandbox, watchdogs, compaction switches, usage parsing"

## Relations
- supports:: [[Roles map to ordered provider profiles kept in config]]
- alternative-to:: [[MultiAIWorkspace]]
- tested-by:: [[Codex medium integrated 61 of 61 summaries and 130 of 130 annotations]]
