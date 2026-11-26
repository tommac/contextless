---
type: source
status: external
area: evaluation
needs: none
year: 2026
url: https://github.com/dial481/locomo-audit
---
The locomo-audit repository is a third-party audit of the LoCoMo benchmark carried out from February to April 2026. It finds wrong gold answers, a lenient judge and comparisons that cannot be told apart, and it records disputed vendor scores such as one system's claimed 92.32 percent against 38.38 reported by a third party. It is not peer reviewed, and the notes did not check two of its figures at their detailed source files.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "99 of 1,540 gold answers (6.4%) are wrong, giving a ceiling of 93.57%"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "The locomo-audit 92.62% judge and the third-party 38.38% setup were not checked at their detailed source files."
- `reports/Cutting edge context memory research.md`: "92.32% claimed, 38.38% reported by a third party"

## Relations
- supports:: [[Benchmark scores are not comparable across papers]]
- contradicts:: [[A LoCoMo leaderboard position as a headline metric]]
- tested-by:: [[LoCoMo audit finds 6.4 percent of gold answers wrong]]
- tested-by:: [[A LoCoMo judge accepted 62.81 percent of vague wrong answers]]
