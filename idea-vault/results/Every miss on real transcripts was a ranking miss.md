---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
In the first real run every miss was a ranking miss and not a coverage miss: the source was named in the packet by its catalog line, recallable by id, but ranked 7th to 48th for six slots, and a search for the question's words did not find it in the top five either. The causes named are paraphrase without stemming, questions that restate the fact in other words, lowercase names the value lane does not treat as values, a colon read as a sentence end, and topic words shared by 20 to 80 other turns. Names and values were never missed.

## Evidence
- `docs/ARCHITECTURE.md`: "Every miss is a ranking miss, not a coverage miss: the source was named in the packet by its catalog line (recallable by id) but ranked 7th to 48th for six slots"
- `docs/ARCHITECTURE.md`: "The five distinct misses were two decisions, two numbers and a branch name."

## Relations
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[A record goes in whole or is listed by id]]
- supports:: [[The atom index is stemmed and searches card text]]
- supports:: [[Atom titles as APIs]]
