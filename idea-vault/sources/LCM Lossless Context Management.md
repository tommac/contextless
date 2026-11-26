---
type: source
status: external
area: recall
needs: none
year: 2026
---
LCM is a 2026 paper from Voltropy, one of four that year that do in part what ContextlessAI does. It has an immutable store, a summary DAG, and the tools lcm_grep and lcm_expand. ContextlessAI took recall by search from lcm_grep, and the idea of digest lines from the summary DAG, though its digests are built by code. It did not take llm_map, and the roadmap notes that LCM leaves embeddings out on purpose.

## Evidence
- `docs/ROADMAP.md`: "Four 2026 papers do what ContextlessAI does in part: LCM (Voltropy: immutable store, summary DAG, `lcm_grep`/`lcm_expand`)"
- `docs/ROADMAP.md`: "(LCM `lcm_grep`). `recall_queries` in the handoff; the engine searches every earlier turn and reports what each search found."
- `docs/ARCHITECTURE.md`: "LCM's summary DAG, built by code as in ARC"
- `docs/ROADMAP.md`: "LCM leaves embeddings out on purpose"

## Relations
- alternative-to:: [[ARC Addressable Recall Compaction]]
- alternative-to:: [[LANTERN]]
- alternative-to:: [[CogCanvas]]
- contradicts:: [[An embedding lane beside the exact full-text and value lanes]]
