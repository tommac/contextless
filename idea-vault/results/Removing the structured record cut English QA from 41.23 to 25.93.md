---
type: result
status: preprint
area: plans
needs: none
measured-by: others
---
In the LLM x MapReduce ablation on InfiniteBench with Llama3-70B, the full method scored 41.23 on English QA and 91.43 on Math.Find. Without confidence calibration it scored 39.18 and 90.00, and without calibration and without the four-field record 25.93 and 56.00. The record is worth 15 to 35 points and calibration only 2 to 5. It is an ablation inside one system on 70B open models, the paper's own measurement with no independent replication.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Ablation (Llama3-70B): full method En.Avg 41.23, Code.Debug 62.94, Math.Find 91.43; without confidence calibration 39.18 / 58.12 / 90.00; without calibration and without the structured protocol 25.93 / 46.45 / 56.00."
- `reports/Parallel AI agents data structures.md`: "Measured ablation inside one system, 70B open models"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "LLM×MapReduce's ablation attributes roughly 15 points of English QA and 35 points of Math.Find to the four-field record, and only 2 to 5 points to calibration"

## Relations
- supports:: [[Units return typed records with checked claims]]
- contradicts:: [[Debate rounds voting and confidence scores as a merge rule]]
- source:: [[LLM x MapReduce]]
