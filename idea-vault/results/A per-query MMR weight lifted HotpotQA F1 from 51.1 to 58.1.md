---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
DF-RAG is an MMR variant whose relevance weight is chosen per query by an LLM planner and evaluator. At about 500 words with Llama 3.3 70B, HotpotQA F1 was 51.1 for plain retrieval, 53.3 with a fixed weight, 58.1 with a per-query weight and 69.5 with an oracle weight, and MuSiQue and 2Wiki show the same order. The oracle weight is worth up to 18 F1 over plain retrieval and the best fixed weight differs by dataset, so the right balance between relevance and diversity depends on the query. The study has 200 queries per set, single runs and no evidence recall, and it needs embeddings and an LLM at query time.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "F1 at about 500 words with Llama 3.3 70B, plain retrieval, fixed weight, per-query weight, oracle weight: HotpotQA 51.1, 53.3, 58.1, 69.5; MuSiQue 29.9, 30.9, 33.3, 46.2; 2Wiki 49.8, 54.8, 58.1, 65.3."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "DF-RAG's oracle result says the best relevance-diversity balance is query dependent."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "200 queries per set, single runs, no evidence recall reported"

## Relations
- supports:: [[A query planner and EXPLAIN for context]]
- supports:: [[Slot allocation over deterministic aspects]]
- part-of:: [[Result diversification]]
