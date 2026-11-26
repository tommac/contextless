---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
In a 2026 study of budgeted selection on clinical and biomedical text, which selector won depended on the budget: document order was strong at 256 to 512 tokens and MMR dominated near 1,024, while on long PubMedQA contexts document order was the worst selector at every budget from 64 to 256 tokens. The authors found that selector choice mattered more than unit choice and that cluster-based units hurt. It is a single-group preprint with no independent replication the notes could find.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Selector choice mattered more than unit choice; cluster-based units hurt"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "On MIMIC, Lead (document order) is strong at 256 to 512 tokens and MMR dominates near 1,024."

## Relations
- supports:: [[Budgeted selection needs a monotone submodular objective]]
- supports:: [[Atoms match whole turns for fewer tokens within noise]]
- supports:: [[Reranking pays only when the budget is tight]]
