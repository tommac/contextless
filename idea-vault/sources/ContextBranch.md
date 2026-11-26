---
type: source
status: external
area: structure
needs: none
year: 2025
url: https://arxiv.org/html/2512.13914v1
---
ContextBranch tested conversation branching on 30 synthetic software scenarios, with GPT-4 both generating and judging. Branching cut context by more than half and raised the overall judged score slightly, with no significant change in accuracy and with no token counts or human users. It cannot separate branching from simply removing distractors, so the report treats the quality claim as unproven.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "30 synthetic software scenarios, GPT-4 generating and GPT-4 judging"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "ContextBranch cannot separate branching from simply removing distractors, so an automatic relevance filter on a linear log should be tested against it."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "accuracy, relevance and specificity showed no significant change. **No human users and no token counts**"

## Relations
- part-of:: [[Branching raises answer quality]]
- supports:: [[Unrelated context can hurt answers]]
- part-of:: [[R13 Tangent filter against branching]]
- tested-by:: [[Branching cut context 58 percent and raised a judged score 2.5 percent]]
