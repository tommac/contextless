---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
On 3,600 test questions per dataset, the correlation between cheap predictors and the actual benefit of retrieval was 0.141 on NQ and about zero on HotpotQA and TriviaQA for pre-retrieval IDF, and 0.255, 0.305 and 0.214 for the best unsupervised post-retrieval predictor shown. A supervised retrieval-stage model reached 0.30 to 0.45, and a probe of the model's internals about 0.5. So cheap signals can drive an adaptive cutoff but are weak at predicting whether the answer is present.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "IDF pre-retrieval: .141 NQ, −.013 HotpotQA, −.045 TriviaQA. Best unsupervised post-retrieval shown (REF): .255, .305, .214."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Cheap signals can drive an **adaptive cutoff** but are weak at predicting whether the answer is actually present"

## Relations
- supports:: [[Deterministic sufficiency signals word the brief]]
- supports:: [[Recall by id or by search replaces memory]]
