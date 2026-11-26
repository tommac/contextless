---
type: source
status: external
area: plans
needs: none
year: 2026
url: https://www.anthropic.com/engineering/building-c-compiler
---
Nicholas Carlini's February 2026 field report: 16 agents over nearly 2,000 sessions and about two weeks produced about 100,000 lines of Rust for just under $20,000. There was no orchestrator: each agent claimed a task by writing a file, each had its own container and clone, and git push was the arbiter. Parallelism worked while there were many independent failing tests and failed on one monolithic task, where every agent fixed the same bug and overwrote the others. It returned only when a known-good oracle let each agent work on a different subset of files.

## Evidence
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "16 agents, nearly 2,000 sessions, about two weeks, just under $20,000, about 100,000 lines of Rust."
- `reports/Parallel AI agents data structures.md`: "Parallelism returned only when a known-good oracle let each agent work on a different subset of files"
- `research_notes/Parallel AI agents data structures/classic_parallel_lessons.md`: "The Carlini oracle trick is a method for manufacturing partitionability."

## Relations
- supports:: [[Claims and leases enforced by a tool or store]]
- supports:: [[Isolated workspaces beat shared mutable state]]
- supports:: [[Plans that buy nothing are answered in one call]]
