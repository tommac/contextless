---
type: source
status: external
area: evaluation
needs: none
year: 2026
url: https://aclanthology.org/2026.aacl-long.211.pdf
---
This peer-reviewed paper ran twelve memory baselines across five families of scenario and found an unindexed agent that greps and reads among the strongest. It introduced the structural rate, the share of questions no single retrieved unit can answer, and showed that the value of an index tracks it. Its ablation of an index view bundles the index with selection controlled by the agent, as the authors note.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "A cross-scenario evaluation at AACL 2026 ran twelve baselines across five scenario families."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "The ablation bundles the index with agent-controlled selection, as the authors note"

## Relations
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[Raw text plus an index the agent may choose to use]]
- tested-by:: [[Graph methods gained 13.0 points on LoCoMo and lost 15.4 on HotpotQA]]
- tested-by:: [[HippoRAG 2 stored form could answer about 10 percent of AMABench questions]]
- tested-by:: [[Grep beat vectors in all 10 inline pairs and scored 17.0 elsewhere]]
