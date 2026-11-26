---
type: idea
status: built
area: plans
needs: none
---
Each kind of work is a role: responder, coder, indexer, long reader, researcher, reviewer, editor and assembler. A role maps to an ordered list of provider profiles, and the order is config, not code. The first choices were seeded from BookWriter's measurements on a book, for example codex medium for the indexer because 61 of 61 summaries and 130 of 130 annotations were integrated. The cx stats command reports runs, success rate, time and tokens per profile and kind, which is how routing is meant to move from reputation to measurement.

## Evidence
- `docs/ARCHITECTURE.md`: "Roles are seeded from BookWriter's measurements on a book-length manuscript and are config, not code"
- `docs/ROADMAP.md`: "Per profile and kind: runs, success rate, median wall time, planned and peak tokens, failure classes. This is how routing moves from reputation to measurement."
- `docs/ARCHITECTURE.md`: "each doing the kind of work it is measurably good at"

## Relations
- builds-on:: [[Provider CLIs run as stateless workers]]
- source:: [[BookWriter]]
- source:: [[MultiAIWorkspace]]
- tested-by:: [[Codex medium integrated 61 of 61 summaries and 130 of 130 annotations]]
