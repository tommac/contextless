---
type: source
status: external
area: plans
needs: none
year: 2025
url: https://arxiv.org/abs/2503.13657
---
The paper Why Do Multi-Agent LLM Systems Fail (Cemri and colleagues, UC Berkeley, 2025) built a taxonomy of 14 failure modes in 3 categories from annotated traces of seven open-source frameworks, 1,642 traces in its third version. About 44% of labelled failures are system design, 32% inter-agent misalignment and 24% verification, and prompt and topology fixes recovered only part of them. The newest model in the dataset is Claude 3.7 Sonnet, so nothing is known for 2026 models.

## Evidence
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "MAST is a 14-mode, 3-category failure taxonomy built from annotated traces of 7 open-source frameworks running GPT-4-class and Claude-3.x-class models"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "No MAST results on 2026-generation models were found; the newest model in the dataset is Claude-3.7-Sonnet."

## Relations
- supports:: [[LLM reviewers are lenient or noisy]]
- supports:: [[A supervisor in code sets deadlines restart limits and backup units]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
