---
type: idea
status: rejected
area: plans
needs: none
---
In the actor model each agent has private state and a mailbox, and others reach it only by asynchronous messages, with a supervisor that restarts failures. AutoGen Core adopted it with no scaling data and its successor moved to supersteps for checkpointable boundaries. Claude Code agent teams use JSON mailboxes and remain experimental, with lagging task status and early completion. The survey's verdict is to take the supervisor (deadlines, restart limits) as code and to skip peer mailboxes.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Take the supervisor (deadlines, restart limits) as code; skip peer mailboxes"
- `reports/Parallel AI agents data structures.md`: "Adopted by AutoGen Core with no scaling data; its successor moved to supersteps for checkpointable boundaries."
- `research_notes/Parallel AI agents data structures/production_systems.md`: "task status can lag because teammates fail to mark tasks complete"

## Relations
- analog-of:: [[Actor model and supervision]]
- alternative-to:: [[One writer per surface]]
- alternative-to:: [[The handoff document is a free-text field in the typed record]]
- alternative-to:: [[An open blackboard that every agent reads and writes]]
