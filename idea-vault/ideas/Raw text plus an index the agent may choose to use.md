---
type: idea
status: external
area: recall
needs: none
---
On files as agent memory, the best-supported design gives the agent raw text plus an index it can choose to use. Agents that grep and read are strong baselines but no method wins everywhere: literal questions local to a session favour grep, while paraphrase and distant evidence favour an index. The harness and the format of tool results alone can move scores by tens of points, so any internal comparison must fix both.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "The best-supported design gives the agent raw text plus an index it can choose to use."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Literal, session-local questions favor grep; paraphrase and distant evidence favor an index."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Any internal comparison should fix the harness and the tool-result format, because those alone can move scores by tens of points."

## Relations
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[Recall by id or by search replaces memory]]
- supports:: [[Hybrid lexical and embedding retrieval beats either alone]]
- tested-by:: [[Grep beat vectors in all 10 inline pairs and scored 17.0 elsewhere]]
- source:: [[MemGPT and Letta]]
- source:: [[Chen et al AACL 2026 cross-scenario memory evaluation]]
