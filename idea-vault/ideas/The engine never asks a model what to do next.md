---
type: idea
status: built
area: plans
needs: none
---
Control flow is code. A model returns a handoff that is checked against a schema, and its status decides what the engine does next: store the answer, pin recalled text and run again, continue a partial answer, run a plan, or put a question back to the user. A model may propose a plan of units with roles, but code decides how to run it. Models also never write the vault or the database themselves; they return a handoff and the engine applies it.

## Evidence
- `docs/ARCHITECTURE.md`: "Control flow is code over schema-validated handoffs. A model can propose a plan (units with roles); code decides how to run it."
- `docs/ARCHITECTURE.md`: "The handoff decides what happens next"
- `docs/ARCHITECTURE.md`: "Models never write either directly; they return a handoff and the engine applies it."

## Relations
- part-of:: [[Determinism is the instrument]]
- builds-on:: [[Provider CLIs run as stateless workers]]
- source:: [[BookWriter]]
