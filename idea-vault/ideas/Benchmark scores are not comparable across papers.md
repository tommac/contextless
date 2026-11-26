---
type: idea
status: external
area: evaluation
needs: none
---
Scores from different papers cannot be compared, because the reader, the judge, the prompt and the question subset move scores by more than most claimed gains. On LoCoMo alone the same full-context baseline ranges from 56.92 to 92.62 percent. The report therefore sorts every outside result by how far it can be trusted, and holds the reader and judge fixed across arms in anything it plans to run itself.

## Evidence
- `reports/Cutting edge context memory research.md`: "Scores from different papers cannot be compared."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "the reader, judge, prompt and question subset move scores by more than most claimed gains"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Reader and judge must be held fixed across arms."

## Relations
- supports:: [[Same-harness baselines every new structure must beat]]
- tested-by:: [[LoCoMo full-context baseline ranges from 56.92 to 92.62 percent]]
- tested-by:: [[Reader choice alone moved one system by 10.6 points]]
- tested-by:: [[One system scored 61.01 or 77.86 depending on the judge]]
- tested-by:: [[Mem0 and Zep LoCoMo scores were disputed and never reconciled]]
- tested-by:: [[LoCoMo audit finds 6.4 percent of gold answers wrong]]
- source:: [[locomo-audit]]
