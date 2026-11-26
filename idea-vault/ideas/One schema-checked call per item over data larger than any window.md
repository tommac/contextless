---
type: idea
status: proposed
area: plans
needs: model-calls
---
LCM has an operation called llm_map: one schema-checked model call per item, over data larger than any window. It was considered in Stage 2b and not taken yet. The roadmap's reason is that the chunked path for an oversized message and plan units cover today's cases.

## Evidence
- `docs/ROADMAP.md`: "**LCM's `llm_map`** (one schema-checked call per item over data larger than any window). The chunked path and plan units cover today's cases."

## Relations
- alternative-to:: [[A message too large for any window is split at paragraph boundaries]]
- alternative-to:: [[Plans run in waves with dependencies and shared decisions]]
- source:: [[LCM Lossless Context Management]]
