---
type: source
status: external
area: evaluation
needs: none
year: 2026
url: https://arxiv.org/html/2605.14498
---
GroupMemBench is a third-party benchmark of memory over group conversations, first posted in May 2026. Its headline finding is that a plain BM25 baseline matches or beats most agent memory systems, which the authors attribute to ingestion that erases thread structure, speaker context and wording. A later paper reran BM25 on it with a changed method and a newer reader and got a higher score, so its numbers move with the setup too.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Average accuracy: Hindsight 46.01, BM25 43.22"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The authors attribute this to ingestion erasing thread structure, speaker context and wording"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "BM25 moved from 43.22 in GroupMemBench to 49.26 for T-Mem's expanded BM25 on the same benchmark, with different reader versions and a changed retrieval method."

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Threads group turns by topic]]
- tested-by:: [[BM25 at 43.22 beat Mem0 and GraphRAG on GroupMemBench]]
- tested-by:: [[T-Mem beat BM25 with expansion by 13.83 points at 5.6 times the tokens]]
