---
type: result
status: preprint
area: evaluation
needs: none
measured-by: others
---
EdgeMem scored 61.01 on LoCoMo under a strict gpt-4o judge and 77.86 under a lenient gpt-4o-mini judge on the same setup, a gap of 16.85 points. In a separate audit, re-judging unchanged answers flipped three labels. The judge is part of the measurement, so it has to be named, held fixed and calibrated.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "EdgeMem scored 61.01 under a strict gpt-4o judge and 77.86 under a lenient gpt-4o-mini judge on the same LoCoMo setup, a 16.85-point gap"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "re-judging unchanged answers flipped three labels"

## Relations
- supports:: [[Benchmark scores are not comparable across papers]]
- supports:: [[Judges need human calibration]]
- supports:: [[LLM reviewers are lenient or noisy]]
- source:: [[EdgeMem]]
