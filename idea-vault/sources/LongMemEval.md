---
type: source
status: external
area: evaluation
needs: none
year: 2024
url: https://arxiv.org/abs/2410.10813
---
LongMemEval is a peer-reviewed benchmark of 500 questions, each with its own history, covering extraction, multi-session reasoning, knowledge updates, temporal reasoning and abstention. It ships a full-context baseline and an oracle, and its own paper shows evidence-only sessions beating full history. Its histories were cleaned in a 2025 update, so the file version must be pinned, and the report picks it for the one locked external benchmark.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "500 questions, each with its own history; S is about 115k tokens, M about 500 sessions; 30 abstention questions"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Histories were cleaned in a 2025/09 update"
- `reports/Cutting edge context memory research.md`: "Add a LongMemEval importer with no model at ingest."

## Relations
- supports:: [[Unrelated context can hurt answers]]
- part-of:: [[R9 One locked external benchmark]]
- alternative-to:: [[LoCoMo]]
- tested-by:: [[Evidence sessions scored 0.870 against 0.606 for full history on LongMemEval]]
- tested-by:: [[Fact-augmented keys raised recall 9.4 percent on LongMemEval]]
