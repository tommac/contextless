---
type: result
status: preprint
area: plans
needs: none
measured-by: others
---
A July 2026 preprint on a five-role pipeline with Qwen2.5-7B-Instruct on GSM8K reports multi-agent accuracy falling from 75.0% to 45.0% with JSON messages between agents and recovering to 82.0% with plain text. It was read at abstract level only, uses a single 7B model and is not peer reviewed. It is the one result against structured handoffs. The survey reads it as a warning to validate at the boundary, retry with the error and keep the record small.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "reports multi-agent accuracy falling from 75.0% to 45.0% with JSON inter-agent messages and recovering to 82.0% with plain text"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "(Abstract only; single 7B model, not peer reviewed.)"
- `reports/Parallel AI agents data structures.md`: "The 7B-model result against JSON handoffs is a warning to validate at the boundary and retry with the error, as LCM does, and to keep the record small."

## Relations
- contradicts:: [[Units return typed records with checked claims]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- supports:: [[A cascade tries one agent first and escalates on hard cases]]
