---
type: source
status: external
area: evaluation
needs: none
year: 2024
url: https://aclanthology.org/2024.acl-long.747/
---
LoCoMo is an ACL 2024 benchmark of 10 long conversations with 1,540 answerable questions and a further set of adversarial ones. An audit finds 6.4 percent of its gold answers wrong and its usual judge lenient, and its conversations are short enough to fit in a context window. The report therefore rules that a position on a LoCoMo leaderboard is never a headline metric.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "10 conversations; 1,540 answerable questions in categories 1-4"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "99 of 1,540 gold answers (6.4%) are wrong, giving a ceiling of 93.57%"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "A LoCoMo result without a same-harness full-context row is uninterpretable."

## Relations
- supports:: [[Benchmark scores are not comparable across papers]]
- contradicts:: [[A LoCoMo leaderboard position as a headline metric]]
- alternative-to:: [[LongMemEval]]
- tested-by:: [[LoCoMo audit finds 6.4 percent of gold answers wrong]]
- tested-by:: [[LoCoMo full-context baseline ranges from 56.92 to 92.62 percent]]
