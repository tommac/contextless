---
type: result
status: preprint
area: plans
needs: none
measured-by: others
---
LongAgent reports that members fabricate answers when their chunk lacks the information. Training members on unanswerable examples raised correct refusals from 51.0% to 78.6%. LLM x MapReduce handles the same problem with an explicit no-information answer that later stages drop. The measurement is from 2024 on a fine-tuned 7B model.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "mitigations were training on unanswerable examples (correct refusals 51.0% to 78.6%) and pairwise chunk exchange between disagreeing members"
- `reports/Parallel AI agents data structures.md`: "A unit whose input lacks the answer tends to invent one; training members to refuse raised correct refusals from 51.0% to 78.6%"

## Relations
- supports:: [[No information is a result not a failure]]
- supports:: [[Units return typed records with checked claims]]
- source:: [[LongAgent]]
