---
type: idea
status: proposed
area: evaluation
needs: none
---
With small samples the analysis should pair each item across conditions, test the pairs, and build intervals by resampling conversations and not questions. Questions that share a conversation are not independent, so intervals computed over items overstate the evidence. The report applies this to itself: its local results come from only three conversations, two of which overlap, so they cannot support general claims however many probes are added.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Use paired per-item tests, resample at the conversation level, freeze a held-out set, and correct for judge error."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "paired bootstrap of the accuracy difference, resampling conversations rather than questions, with 10,000 resamples"
- `reports/Cutting edge context memory research.md`: "Local results come from only three conversations, two of which overlap, so they cannot support general claims however many probes are added: questions from one conversation are not independent."

## Relations
- analog-of:: [[McNemar test for paired outcomes]]
- part-of:: [[Minimum evidence package for a credible memory study]]
- builds-on:: [[Recovery probes measure selection without a model]]
- tested-by:: [[LoCoMo 1,540 items in 10 conversations act like about 94 to 178]]
- tested-by:: [[At 100 items and 20 percent discordance only 13 points are detectable]]
- source:: [[BFR-MV]]
