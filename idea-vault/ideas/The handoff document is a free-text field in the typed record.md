---
type: idea
status: built
area: plans
needs: none
---
A one-way handoff sent upward carries what a schema misses: concerns, deviations from the brief, and findings a sibling should know. It is field practice in Cursor's final design, and Cognition's manager pattern fails without it. The survey's verdict is to adopt it as a free-text field inside the typed record, stored verbatim and never summarized. In ContextlessAI the unit record carries the choices another unit could have made differently, and notes.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Adopt as a free-text field inside the typed record, stored verbatim"
- `reports/Parallel AI agents data structures.md`: "The `assumptions` and `notes` fields are the handoff document, the one place where Cursor's and Cognition's field experience says a schema alone falls short."
- `docs/ARCHITECTURE.md`: "the choices it made that another unit could have made differently"

## Relations
- part-of:: [[Units return typed records with checked claims]]
- alternative-to:: [[Actor mailboxes between agents]]
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- source:: [[Cursor planner and worker agents]]
- source:: [[Cognition Dont Build Multi-Agents]]
