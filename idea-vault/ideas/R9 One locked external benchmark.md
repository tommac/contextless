---
type: idea
status: designed
area: evaluation
needs: download
---
The hypothesis is that the best compiler configuration beats budget-matched BM25 on billed cost per correct answer and is non-inferior to full context on accuracy. A LongMemEval importer with no model at ingest is added, 200 questions are used for tuning and 300 are locked, and the plan is frozen with a hash before a single run with two readers from different providers. It is falsified if the compiler is dominated by BM25 at a matched budget or misses the non-inferiority margin against full context. It needs a dataset download and thousands of model calls, with disk space checked first.

## Evidence
- `reports/Cutting edge context memory research.md`: "The best compiler configuration beats budget-matched BM25 on billed cost per correct answer, and is non-inferior to full context on accuracy"
- `reports/Cutting edge context memory research.md`: "Tune on 200 questions, lock 300, and freeze the plan with a hash before the single run."
- `reports/Cutting edge context memory research.md`: "The compiler is dominated by BM25 at matched budget, or misses the non-inferiority margin against full context"

## Relations
- builds-on:: [[A billed cost-quality frontier for verbatim memory]]
- builds-on:: [[Locked holdout with a hashed analysis plan]]
- builds-on:: [[Non-inferiority needs a pre-set margin]]
- builds-on:: [[Judges need human calibration]]
- builds-on:: [[Minimum evidence package for a credible memory study]]
- builds-on:: [[R3 Whole history against BM25 against the current packet on answers]]
- source:: [[LongMemEval]]
