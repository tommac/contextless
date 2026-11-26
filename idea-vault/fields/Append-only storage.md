---
type: field
status: external
area: storage
needs: none
---
An append-only store only ever adds records and never changes one that is already written. It is the storage rule ContextlessAI shares with the 2026 systems it compares itself with: ARC has an append-only store, LCM an immutable store, and LANTERN archives every turn. In ContextlessAI each turn is appended as its own Markdown file with an immutable body and committed to git, which is what lets any turn be put back into a later packet word for word.

## Evidence
- `docs/ARCHITECTURE.md`: "Every user and assistant turn is stored verbatim, append-only, one Markdown file per turn, committed to git."
- `docs/ROADMAP.md`: "ARC (Fujitsu: append-only store, a view rebuilt before every call, recall by address)"
- `docs/ROADMAP.md`: "LCM (Voltropy: immutable store, summary DAG, `lcm_grep`/`lcm_expand`)"
- `docs/ARCHITECTURE.md`: "L5 the verbatim turn (immutable body)"

## Relations
- source:: [[ARC Addressable Recall Compaction]]
- source:: [[LCM Lossless Context Management]]
- source:: [[LANTERN]]
