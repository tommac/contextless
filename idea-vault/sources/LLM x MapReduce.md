---
type: source
status: external
area: plans
needs: none
year: 2024
url: https://arxiv.org/html/2410.09342v1
---
A 2024 system from Tsinghua for answering over very long texts: each chunk is mapped to a four-field record (extracted information, rationale, answer or a literal no-information value, and a confidence score), and the records are collapsed and reduced. Its ablation is the strongest measured evidence that a fixed-shape record per unit beats free text, with the record contributing far more than the confidence calibration. ContextlessAI's unit record is this map output almost field for field. The results are the paper's own, on 70B open models, with no independent replication.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Each map call returns four fields: Extracted Information (query-relevant facts), Rationale (reasoning from the facts to the answer), Answer"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "The structured record contributes far more than the calibration."
- `reports/Parallel AI agents data structures.md`: "this record is the LLM×MapReduce map output almost field for field"

## Relations
- supports:: [[Units return typed records with checked claims]]
- supports:: [[No information is a result not a failure]]
- supports:: [[Code compares unit records before the assembler writes]]
