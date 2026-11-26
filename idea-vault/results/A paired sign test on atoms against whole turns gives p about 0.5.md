---
type: result
status: measured
area: evaluation
needs: none
measured-by: us
---
On the distinct questions, atoms at 8k won 11 and lost 7 against six whole turns on the build before the review, and a paired sign test gives p about 0.5. The offline floor later reports the same comparison from the other side: six whole turns won 7 and lost 11 against the 8k default, exact p 0.48. So the four extra answers are not a demonstrated improvement.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "a paired sign test on the distinct questions (atoms won 11, lost 7 on the build before the review) gives p about 0.5"
- `docs/CONTEXT-COMPILER.md`: "| Six whole turns | 103 | 7, 11, 0.48 | 31,002 | 29,382 |"

## Relations
- supports:: [[Atoms match whole turns for fewer tokens within noise]]
- supports:: [[The offline floor is the baseline every selector must beat]]
- part-of:: [[McNemar test for paired outcomes]]
