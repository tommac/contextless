---
type: result
status: vendor-reported
area: plans
needs: none
measured-by: others
---
Cursor reports that with flat peers coordinating through a shared file and locks, twenty agents slowed to the effective throughput of two or three, restated in the February post as one to three. Agents held locks too long, forgot to release them, crashed while holding them, or edited the file without locking. It is an informal field measurement by the vendor and not a controlled study. Read through Amdahl's formula it implies a serial fraction of roughly 30% to 47%, a derived illustration that Cursor did not publish.

## Evidence
- `research_notes/Parallel AI agents data structures/production_systems.md`: "twenty agents slowed to the effective throughput of two or three; agents could crash while holding locks, re-acquire locks they held, or edit the file without locking"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "read through Amdahl's formula, implies a serial (lock-waiting) fraction of roughly 30% to 47%"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "Cursor did not publish this figure; it is a derived illustration."

## Relations
- contradicts:: [[Locks held by agents]]
- contradicts:: [[An open blackboard that every agent reads and writes]]
- supports:: [[One writer per surface]]
- supports:: [[Isolated workspaces beat shared mutable state]]
- source:: [[Cursor planner and worker agents]]
