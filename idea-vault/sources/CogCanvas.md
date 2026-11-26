---
type: source
status: external
area: index
needs: none
year: 2026
---
CogCanvas is a 2026 paper that keeps decisions, facts and reminders in a graph, each grounded in verbatim text. From it, and from the gaps the paper states for itself, ContextlessAI took the rule that a fact supersedes an older fact and that sources carry dates. Its embedding lane, LLM reranker and reminder records were considered and not taken yet. The reranker is its largest ablation, 11.5 points.

## Evidence
- `docs/ROADMAP.md`: "CogCanvas (verbatim-grounded decisions, facts and reminders in a graph)."
- `docs/ARCHITECTURE.md`: "CogCanvas (2026) and its stated gaps; LongMemEval's knowledge updates"
- `docs/ROADMAP.md`: "CogCanvas's largest ablation (-11.5 points without it)"

## Relations
- alternative-to:: [[LCM Lossless Context Management]]
- alternative-to:: [[ARC Addressable Recall Compaction]]
- alternative-to:: [[LANTERN]]
- tested-by:: [[CogCanvas lost 11.5 points without its LLM reranker]]
