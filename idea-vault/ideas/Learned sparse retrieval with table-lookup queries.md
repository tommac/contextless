---
type: idea
status: external
area: index
needs: download
---
Li-LSR, peer reviewed at SIGIR 2025, makes query encoding a table lookup, so a model is needed only to encode documents and to train; at query time the system does arithmetic over an inverted index. It reports MS MARCO MRR@10 of 38.8 against 37.8 for SPLADE-v3-Doc and BEIR nDCG@10 of 48.8 against 47.0. For this project it would buy semantic matching at write time while keeping the read path free of models, at the price of a model download. Learned sparse models have not been evaluated on LongMemEval or LoCoMo in what the notes found.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Li-LSR makes query encoding a **table lookup**, so a model is needed only for documents and training."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "MS MARCO MRR@10 38.8 (Big) against 37.8 for SPLADE-v3-Doc; BEIR nDCG@10 48.8 against 47.0."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Learned sparse models have not, in what I found, been evaluated on LongMemEval or LoCoMo."

## Relations
- part-of:: [[Write-time intelligence and read-time determinism]]
- alternative-to:: [[Hybrid lexical and embedding retrieval beats either alone]]
- alternative-to:: [[Atom titles as APIs]]
