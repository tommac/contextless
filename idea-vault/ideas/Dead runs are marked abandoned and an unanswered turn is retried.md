---
type: idea
status: built
area: storage
needs: none
---
One process writes a workspace at a time, held by a lock file. When a workspace is opened, any run that a dead process left marked as running is marked abandoned. A user turn that was stored but never answered is answered with cx retry, without appending the message again. The message itself is already in the vault, because appending it is the first step of handling it. The recovery step came from BookWriter.

## Evidence
- `docs/ARCHITECTURE.md`: "On open, runs a dead process left `running` are marked `abandoned`; a user turn left unanswered is answered by `cx retry`."
- `docs/ARCHITECTURE.md`: "One process writes a workspace at a time (a lock file under `.cx/`)."
- `docs/ROADMAP.md`: "A user turn left unanswered can be answered with `cx retry` or `/retry`, without appending it again."
- `docs/ROADMAP.md`: "(BookWriter `recover.go`). Runs left `running` by a dead process are marked `abandoned` on open."

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[Files are the truth for the conversation and SQLite for the machine]]
- part-of:: [[One writer per surface]]
- source:: [[BookWriter]]
