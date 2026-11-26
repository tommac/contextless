---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
Training-free fusion of BM25 with late-interaction scores gave 8.8 to 17.2 points of Hit at 1 over late interaction alone on LoCoMo and 11.2 over BM25 at best, reaching 0.752. On LongMemEval-S, where BM25 is already strong, the gain over BM25 was small and not significant, and a web-search cross-encoder applied to the fused top 10 lowered Hit at 1 by 6.9. In SeCom, BM25 alone beat MPNet embeddings on LoCoMo at both turn and session level. So fusion helped where the lexical lane was weak and added little where it was already strong, and a reranker trained for web search made the top result worse.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "training-free fusion of BM25 with late-interaction scores gave +8.8 to +17.2 Hit@1 over late interaction alone on LoCoMo and +11.2 over BM25 at best (Hit@1 0.752)"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "on LongMemEval-S, where BM25 is already strong, the gain over BM25 was small and not significant, and a web-search cross-encoder applied to the fused top 10 lowered Hit@1 by 6.9"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "on LoCoMo BM25 beat MPNet at turn level (GPT4Score 65.58 against 57.99) and session level (63.16 against 51.18)"

## Relations
- supports:: [[Hybrid lexical and embedding retrieval beats either alone]]
- supports:: [[An embedding lane beside the exact full-text and value lanes]]
- contradicts:: [[Reranking pays only when the budget is tight]]
- source:: [[SeCom]]
