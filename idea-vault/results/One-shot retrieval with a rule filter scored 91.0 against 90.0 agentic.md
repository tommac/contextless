---
type: result
status: preprint
area: recall
needs: none
measured-by: others
---
On 1,000 questions over more than 40,000 government documents with an LLM judge, basic top-5 retrieval scored 81.0, one-shot filling of a token budget 87.5, and one-shot with a rule-based chunk filter 91.0; a basic agentic loop scored 83.5 and its best variant, with chunk deletion, 90.0. Combining a large initial context with iteration was worse than iteration alone, because bigger starts encouraged what the authors call retrieval laziness. The authors prefer one-shot when tokens are sufficient and iteration when tokens are limited.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Basic top-5 81.0. One-shot filling a token budget 87.5, with a rule-based chunk filter **91.0**. Basic agentic 83.5, with fallback search 87.5, with chunk deletion **90.0**."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "The authors prefer one-shot when tokens are sufficient and iteration when tokens are limited"

## Relations
- supports:: [[Each call gets a fresh packet built by code]]
- supports:: [[The engine never asks a model what to do next]]
- supports:: [[Recall by id or by search replaces memory]]
