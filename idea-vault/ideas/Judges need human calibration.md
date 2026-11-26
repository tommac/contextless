---
type: idea
status: proposed
area: evaluation
needs: model-calls
---
A model used as a judge has to be calibrated against human labels on the specific task before its scores are trusted. Raw judge accuracy overstates low true accuracy and understates high true accuracy, and the correction needs a human-labelled sample of about 200 items. Two model judges agreeing with each other does not establish validity, and using one model as both reader and judge is common in the papers surveyed.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "raw judge accuracy overstates low true accuracy and understates high true accuracy, and provide a sensitivity and specificity correction with valid intervals"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "double-label a stratified sample of at least about 200 items, blind to condition"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Note that two agreeing LLMs do not establish validity."

## Relations
- supports:: [[LLM reviewers are lenient or noisy]]
- supports:: [[Independent review on another provider]]
- part-of:: [[Minimum evidence package for a credible memory study]]
- tested-by:: [[A LoCoMo judge accepted 62.81 percent of vague wrong answers]]
- tested-by:: [[One system scored 61.01 or 77.86 depending on the judge]]
