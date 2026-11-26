---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
With two embedders and chunk sizes from 64 to 1,024 tokens, Recall@1 on SQuAD fell from 0.642 and 0.609 at 64 tokens to 0.386 and 0.429 at 1,024, while on TechQA it rose from 0.049 and 0.045 to 0.619 and 0.715. Short fact answers favour small chunks and dispersed answers favour large ones. Several of the sets are synthetic and matching is by string, so the study shows a pattern, not a rule for conversations.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "SQuAD* falls from 0.642 / 0.609 at 64 tokens to 0.386 / 0.429 at 1,024. TechQA* rises from 0.049 / 0.045 to 0.619 / 0.715."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Short fact answers favor small chunks and dispersed answers favor large ones."

## Relations
- supports:: [[Atoms are paragraph-sized addresses into a turn]]
- supports:: [[Levels of detail like map tiles]]
- supports:: [[Whole-conversation questions are the weak spot]]
