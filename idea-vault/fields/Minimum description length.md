---
type: field
status: external
area: compression
needs: none
---
Minimum description length chooses the model that minimises the cost of stating the model plus the cost of stating the data given the model. The research note found no paper applying it to context selection, only software heuristics, and a related tool, the normalised compression distance, has been measured on text classification and not on retrieval. For this project a compression distance is a model-free novelty estimate, the extra compressed bytes a span adds to the set already selected, and could serve as the redundancy term in greedy packing where MMR would use embedding similarity. Short spans are a known weak spot of compressors.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No paper applying minimum description length to context selection was found. A Rust crate documents a stage selector that minimises model cost plus data-given-model cost, citing Rissanen (1978)"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A compression distance is a model-free novelty estimate: the extra compressed bytes a span adds to the already selected set."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No retrieval measurement exists, and short spans are a known weak spot of compressors."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "competitive with non-pretrained neural models on six of seven in-distribution sets and ahead of BERT on several low-resource languages; limited by gzip's sliding window"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "removing it lowered GPT4Score from 69.33 to 59.87 on LoCoMo"

## Relations
- analog-of:: [[Budgeted selection needs a monotone submodular objective]]
- alternative-to:: [[Submodular maximization under a knapsack budget]]
- part-of:: [[Result diversification]]
