---
type: result
status: measured
area: plans
needs: none
measured-by: others
---
In 1,642 failed traces from seven frameworks, about 44% of labelled failures are system design, about 32% inter-agent misalignment and about 24% verification, with incorrect verification alone at 9.1%. The category sums are computed from per-mode figures and are not stated by the paper. Outright refusal to share is rare, and the common failures are unclear unit contracts and missing stop conditions. Prompt and topology fixes recovered only 9 to 16 points, and the newest model in the dataset is Claude 3.7 Sonnet.

## Evidence
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "summing the v3 modes gives about 44.2% / 32.4% / 23.5% (my arithmetic, not the paper's)"
- `reports/Parallel AI agents data structures.md`: "Outright refusal to share is rare (information withholding 0.85%); the common failures are unclear unit contracts and missing stop conditions."
- `reports/Parallel AI agents data structures.md`: "Prompt and topology fixes recovered only 9 to 16 points."

## Relations
- supports:: [[LLM reviewers are lenient or noisy]]
- supports:: [[A supervisor in code sets deadlines restart limits and backup units]]
- supports:: [[Units return typed records with checked claims]]
- source:: [[MAST failure taxonomy]]
