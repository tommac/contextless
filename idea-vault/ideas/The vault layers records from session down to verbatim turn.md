---
type: idea
status: built
area: structure
needs: none
---
Each conversation has one vault, a git repository of small Markdown records in layers. At the top is the session record with purpose and standing instructions, then threads, then entities, facts and decisions, then a catalog card per turn, and at the bottom the verbatim turn. Edges between records are stored on one side and indexed both ways, so a turn names its threads and a fact names its source turn. The packet builder follows these edges, and the layered design with one-sided edges came from BookWriter's vault.

## Evidence
- `docs/ARCHITECTURE.md`: "The conversation lives in a vault of small records on disk."
- `docs/ARCHITECTURE.md`: "Edges are stored on one side and indexed both ways"
- `docs/ARCHITECTURE.md`: "session.md L0 purpose, standing instructions, preferences"
- `docs/ARCHITECTURE.md`: "Layered records with one-sided edges, sha-pinned summaries"

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[Links help when questions need several pieces of evidence]]
- analog-of:: [[The conversation vault as an Obsidian vault]]
- source:: [[BookWriter]]
