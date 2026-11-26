---
type: idea
status: designed
area: selection
needs: none
---
Typed, weighted edges between atoms would be derived by code and stored in SQLite: next and previous, answers, same-value, same-entity, supersedes, quotes and thread. An expansion lane would spread score from the top hits one or two hops along same-value, same-entity and answers edges with personalized PageRank, as HippoRAG does over a knowledge graph, and supersession closure would make a selected value travel with its newest replacement. It is phase 3 of the build plan, kept only if multi-hop probes improve and nothing else drops.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "Edges are typed, weighted, derived by code, and stored in SQLite (`edges(src, dst, type, weight)`)"
- `docs/CONTEXT-COMPILER.md`: "| 3 | link graph, expansion lane, readability rules, budgeted selection with redundancy penalty | none | multi-hop probes improve; nothing else drops |"

## Relations
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Facts supersede facts]]
- builds-on:: [[Threads group turns by topic]]
- builds-on:: [[Links help when questions need several pieces of evidence]]
- tested-by:: [[Feedback links and coverage did not beat search]]
