---
type: idea
status: built
area: plans
needs: none
---
An answer too long for one call ends with the status partial. The engine stores what is done, and a continuation run gets the work so far verbatim, the units that remain and the next step, and picks up there. The handoff fields for this are BookWriter's handoff contract.

## Evidence
- `docs/ARCHITECTURE.md`: "store what is done; a continuation run gets `units_remaining` and `next_step`"
- `docs/ARCHITECTURE.md`: "Handoff with `units_done`, `units_remaining`, `next_step`, partial runs, continuations"

## Relations
- builds-on:: [[The engine never asks a model what to do next]]
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- source:: [[BookWriter]]
