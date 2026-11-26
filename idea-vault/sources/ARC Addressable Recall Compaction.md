---
type: source
status: external
area: recall
needs: none
year: 2026
---
ARC is a 2026 paper from Fujitsu. It has an append-only store, a view rebuilt before every call, and recall by address. ContextlessAI took three things from it: parts of a long turn that can be recalled one at a time, a note for a bad id that names the valid range, and a recall budget, so that recall that does not fit is refused. ContextlessAI's digest lines are also built by code, as ARC insists.

## Evidence
- `docs/ROADMAP.md`: "ARC (Fujitsu: append-only store, a view rebuilt before every call, recall by address)"
- `docs/ARCHITECTURE.md`: "ARC (Fujitsu, 2026): chunked recall, nearest-match answers, recall budgets"
- `docs/ARCHITECTURE.md`: "Parts of a long turn, notes for bad ids, refusing recall that does not fit"
- `docs/ROADMAP.md`: "Bad ids get a note naming the valid range (ARC)."

## Relations
- alternative-to:: [[LCM Lossless Context Management]]
- alternative-to:: [[LANTERN]]
- alternative-to:: [[CogCanvas]]
