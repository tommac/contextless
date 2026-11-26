---
type: idea
status: designed
area: review
needs: none
---
Checking is ordered from cheapest and most reliable to least. Schema, evidence ids and quotes, citation coverage, the allowlist and the verify command for coder units run first, and the cross-provider reviewer runs second with the engine's comparison so that it checks the join. Checks grounded in something external have the strongest measured effect: executable feedback added 4.2 and 5.4 points for MetaGPT, and deterministic rules reached AUROC 0.95 to 0.98 on local violations. Rules and LLM checks fail in complementary places, which is why both are kept.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Checking should be ordered from cheapest and most reliable to least."
- `reports/Parallel AI agents data structures.md`: "executable feedback added 4.2 and 5.4 points for MetaGPT, and a high-level verification step lifted ChatDev from 25.0 to 40.6"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "Rule-based checks and LLM checks fail in complementary places: rules miss cross-event relations unless written at trace level; LLM verifiers cover more but with many false alarms."

## Relations
- builds-on:: [[LLM reviewers are lenient or noisy]]
- builds-on:: [[Independent review on another provider]]
- builds-on:: [[Units return typed records with checked claims]]
- tested-by:: [[LLM verifiers flagged 22.5 to 70.4 percent of benign traces]]
- part-of:: [[Maker-checker separation]]
