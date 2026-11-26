---
type: idea
status: proposed
area: evaluation
needs: download
---
The research note found no study that compares a chronology-only expansion (the seed, k neighbours and the rest of its session) with an entity-link expansion at an equal token budget across LongMemEval question types. EdgeMem, the nearest, does not separate neighbour reading from session-level BM25 mixing and distinct-session sampling. The gap matters here because the published ablations give most of the gain on conversations to session and neighbour context and to time, with co-occurrence links adding about 2 points on top.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No study compares a chronology-only expansion (seed plus k neighbours plus same session) with an entity-link expansion at equal token budget across LongMemEval question types."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "EdgeMem does not separate neighbour reading from session-level BM25 mixing and distinct-session sampling."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the published ablations assign most of the gain to session and neighbour context and to time, with co-occurrence links adding about 2 points on top"

## Relations
- part-of:: [[R12 Link graph against the adjacency control]]
- builds-on:: [[Adjacent-turn expansion is the control for any link graph]]
- builds-on:: [[Budget-matched baselines]]
- source:: [[EdgeMem]]
- source:: [[LongMemEval]]
