---
type: idea
status: built
area: index
needs: none
---
The indexer is shown the facts in force that the words of the new turns match. A fact that changes a value, such as a new port, date, price or time, supersedes the old fact. A packet then shows the current fact with the one it replaced, and every fact and decision carries the date of its source turn. The rule answers the gaps that CogCanvas states for itself and the knowledge-update questions of LongMemEval.

## Evidence
- `docs/ARCHITECTURE.md`: "The indexer is shown the facts in force that the new turns' words match; a fact that changes a value (a new port, date, price, time) supersedes the old one."
- `docs/ARCHITECTURE.md`: "Packets show the current fact with the one it replaced, and every fact and decision with the date of its source turn."
- `docs/ROADMAP.md`: "CogCanvas's stated gaps; LongMemEval's knowledge-update questions"

## Relations
- builds-on:: [[The indexer writes cards facts and decisions with checked quotes]]
- builds-on:: [[Stale memory is a weak spot of memory frameworks]]
- source:: [[CogCanvas]]
- source:: [[LongMemEval]]
