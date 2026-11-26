---
type: idea
status: measured-kept
area: selection
needs: none
---
A turn that ranking picks for the new message is dropped later than the ordinary turns of its thread. Before this change a full-text match inside a large thread took the thread's tier and was dropped oldest first with the rest of it. The recovery probes showed the fault: only 3 of 14 planted facts stayed in the packet verbatim. With ranked turns at tier 3 above thread turns at tier 2, 11 of 14 stayed, and the change was kept.

## Evidence
- `docs/ROADMAP.md`: "A full-text match in a large thread used to take the thread's tier and be dropped oldest first."
- `docs/ROADMAP.md`: "The probes showed it: 3 of 14 planted facts verbatim. With ranked turns at tier 3 above thread turns at tier 2: 11 of 14."

## Relations
- builds-on:: [[Drop tiers set the order in which records are left out]]
- builds-on:: [[Threads group turns by topic]]
- builds-on:: [[Recovery probes measure selection without a model]]
- analog-of:: [[Route-first hierarchy loses to flat search]]
- tested-by:: [[Ranked turns at tier 3 raised planted facts verbatim from 3 to 11 of 14]]
