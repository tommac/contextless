---
type: result
status: measured
area: evaluation
needs: none
measured-by: others
---
The LoCoMo audit found that a gpt-4o-mini judge accepted 62.81 percent of answers that were vague, on topic and wrong, and 10.61 percent of answers that were specific and wrong. A judge this lenient rewards hedged answers. A vendor's stricter rubric reported no false positives on 100 probes, which still allows a false-positive rate of about 3 to 4 percent.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "The gpt-4o-mini judge accepted 62.81% of vague but topical wrong answers and 10.61% of specific wrong answers"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "The AgentOS 0-of-100 result still allows a false-positive rate up to about 3 to 4% at 95% confidence, so it does not show the judge is error-free."

## Relations
- supports:: [[LLM reviewers are lenient or noisy]]
- supports:: [[Judges need human calibration]]
- supports:: [[Benchmark scores are not comparable across papers]]
- source:: [[locomo-audit]]
