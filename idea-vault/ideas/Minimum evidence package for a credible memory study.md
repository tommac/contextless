---
type: idea
status: proposed
area: evaluation
needs: download
---
The notes set out, as the author's judgment and not a venue rule, the least evidence a credible memory paper needs. It includes two independent datasets, two readers from different providers, a locked test split, the standard baselines, human-calibrated judging, paired clustered intervals, cost and latency, per-item outputs and a one-command rerun. The likely reviewer objections are the mirror image: no same-harness baselines, tuning on the test data, only LoCoMo, an uncalibrated judge, and intervals that ignore conversation clustering.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "at least two independent datasets, two readers from different providers, a locked test split, full-context, oracle, closed-book, BM25 and embedding baselines, human-calibrated judging, paired clustered intervals, cost and latency, per-item outputs, and a one-command rerun"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "No same-harness full-context, oracle, closed-book or BM25 baseline."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Item-level intervals that ignore conversation clustering."

## Relations
- builds-on:: [[Same-harness baselines every new structure must beat]]
- builds-on:: [[Paired tests with conversation-level resampling]]
- builds-on:: [[Judges need human calibration]]
- builds-on:: [[Locked holdout with a hashed analysis plan]]
- builds-on:: [[Three layers reached used and correct]]
- supports:: [[Evaluation-led papers do not need a new top score]]
- part-of:: [[R9 One locked external benchmark]]
