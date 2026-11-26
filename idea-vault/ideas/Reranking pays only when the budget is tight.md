---
type: idea
status: external
area: selection
needs: model-calls
---
Selection effort pays most when the budget is tight relative to the candidate pool. In a pre-registered 2026 study, reranking's gain over cosine order was 17.4 points on LoCoMo and 9.1 on LongMemEval when 3 of 30 candidates were kept, and 1.5 and 1.1 when 20 of 30 were kept. The selector there is a model called per query over an embedding shortlist, so this is not a no-model result; the lesson taken here is to measure selection at the engine's actual pool ratio.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Selection effort pays most when the budget is tight relative to the candidate pool, and that effect is large (+17.4 points at 3 of 30 candidates, +1.5 at 20 of 30)."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Sharma and Lall suggest effort is worth spending when the budget admits about 10% of candidates and is close to wasted at about 67%."

## Relations
- source:: [[Sharma and Lall 2026]]
- builds-on:: [[Hybrid lexical and embedding retrieval beats either alone]]
- tested-by:: [[Reranking gains fall from 17.4 points to 1.5 as the budget loosens]]
