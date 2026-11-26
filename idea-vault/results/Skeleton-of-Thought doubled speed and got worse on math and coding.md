---
type: result
status: measured
area: plans
needs: none
measured-by: others
---
Skeleton-of-Thought (ICLR 2024) writes a skeleton of points and expands each in a parallel call, assuming the points are independent. It was more than 2 times faster on 8 of 12 models, with quality at least as good as normal generation in about 60% of cases and poor on writing, Fermi, math and coding questions. Its authors added a router that decides per question whether to decompose at all. It is the measured case for independence assumed and quality lost.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Speedup above 2× on 8 of 12 models, maximum 2.39×. Quality at least as good as normal generation in about 60% of cases"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "speed rises about 2× but math and coding answers get worse, and the paper's remedy is to route such questions away from decomposition"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Skeleton-of-Thought explicitly assumes the points are independent and ignores dependencies between them; the authors name a graph structure as the fix, and add a router"

## Relations
- supports:: [[Plans that buy nothing are answered in one call]]
- supports:: [[A task list with dependency placeholders resolved by code]]
- supports:: [[A cascade tries one agent first and escalates on hard cases]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
