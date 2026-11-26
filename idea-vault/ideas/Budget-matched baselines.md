---
type: idea
status: proposed
area: evaluation
needs: none
---
Baselines should get the same budget of delivered tokens as the system under test, and the budget should then be swept from small to full. A result with a large top-k approaches full context, so it has to be reported with the tokens delivered to the reader. LANTERN is the precedent: every method got the same restore budget, and recovery rose steeply up to 6,000 characters and then levelled off.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "give BM25, embedding RAG, recency truncation and a simple summary the same delivered-token budget as the system, then sweep budgets"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Any large top-k result should be reported with tokens delivered to the reader, because large k approaches full context."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "recovery rose from 36.6% at 2,000 characters to 73.2% at 6,000 and 77.7% at 8,000 before plateauing"

## Relations
- builds-on:: [[Packets are sized before the model starts]]
- builds-on:: [[Recovery probes measure selection without a model]]
- supports:: [[The offline floor is the baseline every selector must beat]]
- part-of:: [[Minimum evidence package for a credible memory study]]
- source:: [[LANTERN]]
