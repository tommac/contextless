---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
Dense X Retrieval indexed model-written propositions in place of passages. Recall@5 on NQ rose 12.3 points with SimCSE and 7.6 with Contriever, but fell 0.6 with DPR and rose only 0.2 with GTR; on EntityQuestions the gains were 21.1, 15.4, 11.9 and 3.2. The gains concentrate in weaker or unsupervised retrievers and long-tail entities, and they cost about 6.3 propositions per passage and a model at index time.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Recall@5 against passages on NQ: SimCSE +12.3, Contriever +7.6, but **DPR −0.6, GTR +0.2**. On EntityQuestions: +21.1, +15.4, +11.9, +3.2."

## Relations
- supports:: [[Atoms are paragraph-sized addresses into a turn]]
- source:: [[Dense X Retrieval]]
