---
type: result
status: measured
area: recall
needs: none
measured-by: us
---
The first build's recall search returned one paragraph where whole-turn mode returned one turn, and its probe column counted any paragraph of the source turn as a hit. Counted by whether the returned text holds the answer, that was 108 of 150 against 126 for whole turns. Returning four paragraphs per turn brought it to 127 of 150. The counting defect was found by the independent review; it had made atoms look level when they were 18 probes behind.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "A search now returns four paragraphs per turn it would have returned: 127 of 150 (28, 27, 27, 22, 23 by set, against 27, 25, 29, 23, 22)."
- `docs/ARCHITECTURE.md`: "one paragraph per hit held the answer far less often than a whole turn (108 against 126 of 150 real probes; 127 at four)"
- `docs/ROADMAP.md`: "atoms looked level with whole turns and were 18 probes behind"

## Relations
- supports:: [[Recall search needs more paragraphs than turns]]
- supports:: [[Recall by id or by search replaces memory]]
- supports:: [[Independent review on another provider]]
