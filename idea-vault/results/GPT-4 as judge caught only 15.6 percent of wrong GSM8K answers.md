---
type: result
status: measured
area: review
needs: none
measured-by: others
---
Wang and colleagues (EMNLP 2024) measured GPT-4-0613 judging answers. On GSM8K it accepted correct answers 99.2% of the time and identified incorrect ones only 15.6%, and on HotpotQA the figures were 99.4% and 2.9%. A judge left to its default is lenient. The model is from 2023 and the tasks are question answering, so nothing here measures a 2026 reviewer on assembled answers.

## Evidence
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "on GSM8K it accepts correct answers 99.2% of the time but identifies incorrect ones only 15.6%; MATH 91.1% / 46.1%; TheoremQA 94.5% / 23.2%; HotpotQA 99.4% / 2.9%; CSQA 98.7% / 6.0%"
- `reports/Parallel AI agents data structures.md`: "GPT-4 as a judge accepted correct GSM8K answers 99.2% of the time and identified incorrect ones only 15.6%"

## Relations
- supports:: [[LLM reviewers are lenient or noisy]]
- supports:: [[Deterministic checks run before the reviewer]]
- supports:: [[Seeded errors measure a reviewer's catch and false-positive rates]]
