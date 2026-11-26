---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
In SmartSearch's oracle analysis retrieval recall was 98.6%, but only 22.5% of gold evidence survived truncation to the token budget. The authors call this the compilation bottleneck and answer it with neural rerankers, which moved the mean rank of the first gold passage among substring candidates from 195 to 8. Its budget curves show gold-evidence recall on LoCoMo rising from .927 at 1K words to .979 at 4K. A lexical-only system cannot borrow that answer, so its equivalent must be better lexical ranking signals and coverage-aware packing.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "retrieval recall was 98.6% but only 22.5% of gold evidence survived truncation to the token budget, which the authors call the compilation bottleneck and answer with neural rerankers"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "gold-evidence recall at 1K/2K/3K/4K words is .927/.958/.972/.979 on LoCoMo and .850/.928/.959/.973 on LongMemEval-S"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Its equivalent must be better lexical ranking signals and coverage-aware packing, which is where mechanisms 1 to 3 act."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the mean rank of the first gold passage among substring candidates from 195 to 8"

## Relations
- supports:: [[Reranking pays only when the budget is tight]]
- supports:: [[More budget buys little after 8k]]
- supports:: [[Complete-evidence coverage at a fixed budget is the outcome measure]]
- supports:: [[Budget-matched baselines]]
- source:: [[SmartSearch]]
