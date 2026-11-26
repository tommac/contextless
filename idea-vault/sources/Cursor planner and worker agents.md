---
type: source
status: external
area: plans
needs: none
year: 2026
url: https://cursor.com/blog/self-driving-codebases
---
Two Cursor posts (January and February 2026, Wilson Lin) on running hundreds of coding agents on one codebase. Flat peers with a shared lock file, then optimistic concurrency, then a standalone integrator were each tried and removed. The final design is a tree of planners over workers that are unaware of the larger system, work on their own copy of the repository and return one handoff upward. Its numbers are volume, about 1,000 commits per hour, with quality deferred, and the February post revises the January one.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Cursor tried flat peers with a shared lock file, then optimistic concurrency, then a standalone integrator, and removed each"
- `research_notes/Parallel AI agents data structures/production_systems.md`: "Cursor's January 2026 post (planner/worker/judge) is superseded by its February 2026 post (judge and standalone integrator removed, recursive planners, handoff documents)."
- `reports/Parallel AI agents data structures.md`: "Vendor volume metrics; quality deferred"

## Relations
- supports:: [[Isolated workspaces beat shared mutable state]]
- supports:: [[One writer per surface]]
- supports:: [[The handoff document is a free-text field in the typed record]]
- contradicts:: [[Locks held by agents]]
- contradicts:: [[An LLM integrator that merges coder branches on its own]]
