---
type: source
status: external
area: selection
needs: none
year: 2026
url: https://arxiv.org/pdf/2609.34227
---
A pre-registered preprint of 28 September 2026 on memory selection under a budget, on LoCoMo and LongMemEval. Reranking's gain over cosine order is large when 3 of 30 candidates are kept and nearly gone when 20 of 30 are kept. At the tight budget raw turns picked by one selector call were non-inferior to an LLM-extraction memory and cost 3,061 times less to write, while at generous budgets extraction systems were more accurate. Its selector is a model called per query over an embedding shortlist.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "At the tight budget raw turns picked by one selector call were non-inferior to an LLM-extraction memory (one-sided 95% bound −3.0 against a −5 margin) and cost 3,061× less to write."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "At generous budgets extraction systems were more accurate."

## Relations
- supports:: [[Reranking pays only when the budget is tight]]
- supports:: [[Raw text beats rewritten memory]]
- tested-by:: [[Reranking gains fall from 17.4 points to 1.5 as the budget loosens]]
