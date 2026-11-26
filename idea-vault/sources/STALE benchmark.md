---
type: source
status: external
area: evaluation
needs: none
year: 2026
url: https://arxiv.org/html/2605.06527
---
STALE is a May 2026 preprint benchmark of 400 scenarios and 1,200 queries in which a later statement makes an earlier memory outdated without saying so. The best plain model reached 55.2 percent and memory frameworks scored far lower. Its write-time diagnostic shows that systems retrieve the old entry and still fail to revise it, and it tests only one conflict pair per case.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "400 implicit-conflict scenarios and 1,200 queries. Best plain model was Gemini-3.1-pro at 55.2%."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "STALE tests one conflict pair per instance, so repeated updates and gradual drift remain unmeasured."

## Relations
- supports:: [[Stale memory is a weak spot of memory frameworks]]
- part-of:: [[R11 Value chains and time intent]]
- alternative-to:: [[LongMemEval]]
- tested-by:: [[Memory frameworks scored 5 to 18 percent on STALE]]
- tested-by:: [[Only 3.3 percent of recalled stale entries were judged to need an update]]
