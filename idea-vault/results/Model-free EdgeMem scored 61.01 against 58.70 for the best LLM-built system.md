---
type: result
status: preprint
area: structure
needs: none
measured-by: others
---
EdgeMem stores verbatim turns and builds its hubs with spaCy, BM25 and a date parser only, with no LLM and no embeddings. Under one shared answer prompt and a strict judge on LoCoMo it scored 61.01 against 58.70 for CompassMem and 54.94 for CAM, among six reproduced systems built with an LLM. Its build cost was 0 LLM tokens and 1.8 minutes, against 0.46M to 11.77M tokens and 17 minutes to 10.1 hours for those systems. On LongMemEval-S its baseline rows are transcribed from another paper and include no BM25 row, so its advantage over plain BM25 there is unknown.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Against six reproduced systems under the same prompt: 61.01 against CompassMem 58.70 and CAM 54.94."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Build cost for LoCoMo: 0 LLM tokens and 1.8 minutes single-threaded, against 0.46M to 11.77M tokens and 17 minutes to 10.1 hours for six reproduced LLM-built systems."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "EdgeMem's LongMemEval baselines are transcribed from another paper and include no BM25 row, so the size of its advantage over plain BM25 on that benchmark is unknown."

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Model-free hubs over verbatim turns]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Novelty must come from one mechanism with an isolated measurement]]
- source:: [[EdgeMem]]
