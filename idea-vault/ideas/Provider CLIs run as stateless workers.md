---
type: idea
status: built
area: packets
needs: none
---
Each provider's command-line tool is spawned headless in the packet directory with its model and effort pinned, session persistence off and compaction disabled. Nothing is inherited from a previous session, so the packet is all the model knows. The provider layer, with adapters for seven CLIs, the macOS sandbox, watchdogs, compaction switches and usage parsing, was copied whole from BookWriter.

## Evidence
- `docs/ARCHITECTURE.md`: "Nothing is inherited from a previous session; every CLI runs with `--no-session-persistence` (or its equivalent) and compaction disabled."
- `docs/ARCHITECTURE.md`: "The CLI is spawned headless in the packet directory, model and effort pinned, compaction disabled, usage streamed"
- `docs/ARCHITECTURE.md`: "BookWriter `internal/bw/providers` (copied whole)"

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[Isolated workspaces beat shared mutable state]]
- source:: [[BookWriter]]
