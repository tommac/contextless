---
type: idea
status: designed
area: index
needs: download
---
The hypothesis is that one small embedder, or token matching limited to a lexical shortlist, adds at least 5 points on paraphrased probes without hurting exact and value lookups. Retrieval is tested first, with the model hash pinned and exact search used so that packets stay reproducible. It is falsified if the gain is under 5 points, if exact lookup falls, or if the slowest queries exceed about 50 ms. The published gains are the authors' own, and hybrid retrieval sometimes adds nothing.

## Evidence
- `reports/Cutting edge context memory research.md`: "One small embedder, or token matching limited to a lexical shortlist, adds at least 5 points on paraphrased probes without hurting exact and value lookups"
- `reports/Cutting edge context memory research.md`: "Pin the model hash and use exact search, so packets stay reproducible"
- `reports/Cutting edge context memory research.md`: "The gain is under 5 points, or exact lookup falls, or p95 exceeds about 50 ms."

## Relations
- builds-on:: [[Hybrid lexical and embedding retrieval beats either alone]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[The value lane finds stated values near the question's words]]
- builds-on:: [[Every derived index version is an input to the context build]]
- tested-by:: [[BM25 plus a small embedder gained 9 recall points on LongMemEval]]
- builds-on:: [[An embedding lane beside the exact full-text and value lanes]]
- source:: [[Luo et al 2026]]
