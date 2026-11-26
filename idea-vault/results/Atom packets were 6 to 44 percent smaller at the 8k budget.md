---
type: result
status: measured
area: cost
needs: none
measured-by: us
---
Median brief at the 8k atom budget against six whole turns: 18,800 tokens against 24,100, 26,200 against 28,000, and 22,700 against 40,300 on the research session. The saving holds only for conversations with no threads: on the indexed synthetic conversation most ranked text is thread turns that are kept whole, and the brief is the same size as with whole turns. The design's own summary gives the range as the same size to 41 percent smaller, which does not match the 6 to 44 percent stated in the section itself.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "Median brief at the 8k budget against whole turns: 18,800 against 24,100, 26,200 against 28,000, and 22,700 against 40,300 on the research session"
- `docs/CONTEXT-COMPILER.md`: "the brief is the same size as with whole turns. The size result above is for conversations with no threads."
- `docs/CONTEXT-COMPILER.md`: "in packets the same size to 41 percent smaller"

## Relations
- supports:: [[Atoms match whole turns for fewer tokens within noise]]
- supports:: [[The sixfold saving did not carry over to packets]]
- supports:: [[Ranked atoms are sent as excerpts under a token budget]]
