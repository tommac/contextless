---
type: idea
status: external
area: review
needs: model-calls
---
A model used as a judge accepts almost every correct answer and catches few wrong ones, and when told to hunt for violations it swings to many false alarms. Neither mode is a reliable gate on its own. Checks grounded in something external, such as running code, deterministic rules or an authoritative source, have the strongest measured effect. No peer-reviewed study reports both the catch rate and the false-positive rate of an independent reviewer on realistic work with 2026 models, so the project's own reviewer is unmeasured too.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Review is weaker than its reputation."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "LLM judges are lenient by default: very high acceptance of correct answers and poor rejection of wrong ones (2.9 to 46% on 2023 GPT-4)."
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Neither mode is a reliable gate on its own."
- `reports/Parallel AI agents data structures.md`: "No peer-reviewed study reports both catch rate and false-positive rate for an independent reviewer on realistic work with 2026 models."

## Relations
- tested-by:: [[GPT-4 as judge caught only 15.6 percent of wrong GSM8K answers]]
- tested-by:: [[LLM verifiers flagged 22.5 to 70.4 percent of benign traces]]
- tested-by:: [[MAST census puts 44 percent of failures in system design and 24 in verification]]
- source:: [[MAST failure taxonomy]]
- part-of:: [[Maker-checker separation]]
