---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
A test plants 14 facts in an 800-turn conversation of about 620k tokens and checks where each fact's source turn lands in the packet. While a full-text match took its thread's tier and was dropped oldest first, only 3 of the 14 source turns were in the packet verbatim. With ranked turns at tier 3, above thread turns at tier 2, 11 of 14 were. No model was run, and the set is 14 facts planted by the test itself, so this shows the change working on that set and says little about real conversations.

## Evidence
- `docs/ROADMAP.md`: "The probes showed it: 3 of 14 planted facts verbatim. With ranked turns at tier 3 above thread turns at tier 2: 11 of 14."
- `docs/ARCHITECTURE.md`: "facts is in the packet verbatim (3 of 14 before the ranking and tier change)"

## Relations
- supports:: [[Ranked turns outlive the bulk of their thread]]
- supports:: [[Drop tiers set the order in which records are left out]]
- supports:: [[Recovery probes measure selection without a model]]
- supports:: [[No model has yet answered from a packet]]
