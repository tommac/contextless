---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
In MEMTIER, adapting fusion weights with reinforcement learning gave no gain at all, because raw BM25 scores dominated the bounded signals, and removing individual signals changed accuracy by only 0.012 to 0.014. The authors name score normalization as a precondition for learning weights. For this project it is a warning about learning from signals, and it sets the order of work: normalize per query first, then learn a handful of global weights.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "PPO weight adaptation gave no gain (0.382, ±0.000) because raw BM25 dominated bounded signals."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "MEMTIER traced its failed weight learning to raw BM25 dominance, naming score normalization as a precondition"

## Relations
- contradicts:: [[Learning from recall requests and citations]]
- supports:: [[Fused ranking over exact full-text and value lanes]]
- supports:: [[Full text alone matched all lanes on real probes]]
