---
type: idea
status: built
area: storage
needs: none
---
The records of a conversation are Markdown files with YAML front matter in a git repository, and they are the truth about what was said. A SQLite database beside them holds what the machine needs: the index, full-text search, runs and provider health. Models never write either one; they return a handoff and the engine applies it. The paragraph index in the database is a cache that reading commands may bring up to date.

## Evidence
- `docs/ARCHITECTURE.md`: "Files are the truth for the conversation; SQLite is the truth for the machine."
- `docs/ARCHITECTURE.md`: "Records are Markdown with YAML frontmatter in a git repository. `.cx/state.db` holds the index, full-text search, runs, and provider health."
- `docs/ARCHITECTURE.md`: "The atom index is a cache in `state.db`, rebuilt by any process that opens the workspace and finds turns missing from it"

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[The engine never asks a model what to do next]]
- part-of:: [[The vault layers records from session down to verbatim turn]]
- source:: [[BookWriter]]
