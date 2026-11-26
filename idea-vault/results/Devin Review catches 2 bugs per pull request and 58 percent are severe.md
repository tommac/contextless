---
type: result
status: vendor-reported
area: review
needs: none
measured-by: others
---
Cognition reports that Devin Review, a reviewer that deliberately gets no shared context and looks only at the diff, catches an average of 2 bugs per pull request, about 58% of them severe. The coding agent, which holds the user's instructions, filters the reviewer's findings. It is a vendor number with no false-positive rate and no comparison of a same-model against a different-model reviewer. ContextlessAI's arrangement, where the writer revises with the reviewer's surviving problems, mirrors it.

## Evidence
- `research_notes/Parallel AI agents data structures/production_systems.md`: "Devin Review catches an average of 2 bugs per PR, about 58% of them severe."
- `research_notes/Parallel AI agents data structures/production_systems.md`: "The reviewer deliberately gets no shared context and looks only at the diff; the coding agent, which holds the user's instructions, filters the reviewer's findings to avoid loops and out-of-scope work."
- `reports/Parallel AI agents data structures.md`: "The writer filtering the reviewer's surviving problems mirrors the Devin Review arrangement."

## Relations
- supports:: [[Independent review on another provider]]
- supports:: [[One writer per surface]]
- part-of:: [[Maker-checker separation]]
- source:: [[Cognition Dont Build Multi-Agents]]
