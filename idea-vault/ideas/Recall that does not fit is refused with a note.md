---
type: idea
status: built
area: recall
needs: none
---
When the turns a model asks to recall do not fit together in the packet, the engine refuses the ones that do not fit and says so in a note. The overflow is never treated as a message that is too big. Before this rule it was, and the engine split the user's short message into parts. The rule is ARC's recall budget.

## Evidence
- `docs/ARCHITECTURE.md`: "Turns that do not fit together are refused with a note; the overflow is never taken for an oversized message"
- `docs/ROADMAP.md`: "and the engine split the user's short message into parts."
- `docs/ARCHITECTURE.md`: "ARC (Fujitsu, 2026): chunked recall, nearest-match answers, recall budgets"

## Relations
- builds-on:: [[Recall by id or by search replaces memory]]
- builds-on:: [[Packets are sized before the model starts]]
- builds-on:: [[A message too large for any window is split at paragraph boundaries]]
- source:: [[ARC Addressable Recall Compaction]]
