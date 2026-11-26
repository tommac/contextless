---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
In an audit with 108 users over 43 TREC DL 2019 queries, 20.9% of queries benefited from feedback and 25.6% were degraded, and avoiding harm was valued nearly twice as much as a successful expansion. The classic study of Collins-Thompson shows the same hidden harm: on the wt10g collection more queries were hurt than helped even though mean MAP rose, and a risk-constrained choice of expansion terms cut the loss. The research note concludes that any feedback lane should be reported with per-question wins, ties and losses and a robustness index, since mean gains hide the harmed share. The audit is a workshop abstract, and the harm it shows is the reason a feedback lane is proposed only behind a gate.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "in an audit with 108 users over 43 TREC DL 2019 queries, 20.9% of queries benefited from feedback and 25.6% were degraded; avoiding harm was valued nearly twice as much as exploiting a successful expansion"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "On wt10g more queries were hurt than helped even though mean MAP rose."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Any feedback lane should be reported with per-question wins, ties and losses and a robustness index, since mean gains hide the harmed share"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "paired bootstrap intervals with wins, ties and losses per query (SPRIG), paired conversation-cluster bootstrap with 10 clusters (EdgeMem), and a robustness index for any expansion lane (Collins-Thompson)"

## Relations
- contradicts:: [[Feedback expansion from first-pass hits needs no model]]
- supports:: [[Feedback gated by the top-score gap]]
- supports:: [[Feedback links and coverage did not beat search]]
- supports:: [[Paired tests with conversation-level resampling]]
- part-of:: [[Pseudo-relevance feedback]]
