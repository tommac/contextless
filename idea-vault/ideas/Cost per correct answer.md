---
type: idea
status: proposed
area: cost
needs: model-calls
---
Cost per correct answer divides the amortized build cost plus all serving cost by the number of correct answers, with an interval taken from the same resampled items as accuracy. Because systems that do their work at write time depend on query volume, it is reported as a curve at 1, 10, 100 and 1,000 queries per conversation. Two published break-even results differ widely, so the answer depends on the model, the cache behaviour and the workload.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "**Cost per correct answer** = (amortized build cost + all serving cost, including retries, reranking and judge-free pipeline calls) ÷ correct answers"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "cost per correct answer at 1, 10, 100 and 1,000 queries per conversation, since write-time systems depend on query volume"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Report a curve, not one number."

## Relations
- builds-on:: [[Billed-equivalent tokens]]
- part-of:: [[A billed cost-quality frontier for verbatim memory]]
- tested-by:: [[Break-even against full history ranged from 0 turns to never]]
- source:: [[Total Recall at What Cost]]
