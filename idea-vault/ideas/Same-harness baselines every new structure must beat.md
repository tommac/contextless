---
type: idea
status: proposed
area: evaluation
needs: none
---
Any new structure in the engine should have to beat four baselines on the same reader, judge and context budget: recency only, BM25 over verbatim turns, BM25 plus dense retrieval, and full context where it fits. Evaluation should report evidence recall and tokens used beside answer accuracy, because controlling the budget changed the conclusions of one study. A closed-book score is a cheap further control that the notes found unpublished for the two main benchmarks.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Any new structure in the engine should have to beat, on the same reader, judge and context budget: (a) recency only, (b) BM25 over verbatim turns, (c) BM25 plus dense retrieval, and (d) full context where it fits."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Budget control changed conclusions in Beyond Memory Leaderboards."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "I found no published closed-book (no-history) score for LoCoMo or LongMemEval-S with named reader and judge."

## Relations
- builds-on:: [[The offline floor is the baseline every selector must beat]]
- builds-on:: [[Benchmark scores are not comparable across papers]]
- builds-on:: [[Budget-matched baselines]]
- part-of:: [[Minimum evidence package for a credible memory study]]
