---
type: idea
status: measured-kept
area: recall
needs: none
---
A recall search that returned one paragraph where whole-turn mode returned one turn put back the answer far less often. Counted by whether the returned text holds the answer it found 108 of 150 against 126 for whole turns, so a search now returns four paragraphs for each turn it would have returned, which brings it level at 127. The probe searches at depth five while the engine puts back three hits, so real recall is lower than both figures.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "Counted by whether the returned text holds the answer, that was 108 of 150 against 126 for whole turns."
- `docs/CONTEXT-COMPILER.md`: "The probe searches at depth five; the engine puts back three hits per search, so real recall is lower than both figures."

## Relations
- builds-on:: [[Recall by id or by search replaces memory]]
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Recovery probes measure selection without a model]]
- tested-by:: [[Recall search found 108 of 150 at one paragraph and 127 at four]]
