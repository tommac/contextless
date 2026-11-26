---
type: idea
status: built
area: plans
needs: none
---
Before the assembler writes, code compares the unit records. It leaves out units that found nothing or failed, lists claims of different units whose quotes lie in overlapping paragraphs of one turn, pairs assumptions that may be the same choice, and collects open questions and conflicts. It decides nothing and only shows where to look. The survey's reason is that the systems with the strongest long-input results all moved the combine step away from an unconstrained model call.

## Evidence
- `docs/ARCHITECTURE.md`: "It decides nothing; it shows where to look."
- `docs/ARCHITECTURE.md`: "lists claims of different units whose quotes lie in overlapping paragraphs of one turn (marking pairs that share topic words but state different values)"
- `reports/Parallel AI agents data structures.md`: "The systems with the strongest long-input results all moved the combine step away from an unconstrained model call."

## Relations
- builds-on:: [[Quotes are pinned to the paragraphs that hold them]]
- builds-on:: [[Units return typed records with checked claims]]
- builds-on:: [[The join fails more often than the plan]]
- tested-by:: [[Independent agents amplified errors 17.2 times against 4.4 with an orchestrator]]
- analog-of:: [[MapReduce re-execution and atomic commit]]
- source:: [[LCM Lossless Context Management]]
