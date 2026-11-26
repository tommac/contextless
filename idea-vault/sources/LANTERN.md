---
type: source
status: external
area: evaluation
needs: none
year: 2026
---
LANTERN is a 2026 paper from Cisco. It archives every turn, uses hybrid retrieval, and evaluates with recovery probes. ContextlessAI took its recovery probes without the LLM judge, its RRF fusion of ranked lanes, and a value lane made for the config-value blind spot in LANTERN's failure analysis. Two of its own numbers are behind what was not taken: its misses were mostly coverage, and an LLM reranker gained it 1.9 points, not significant.

## Evidence
- `docs/ROADMAP.md`: "LANTERN (Cisco: archive every turn, hybrid retrieval, recovery probes)"
- `docs/ROADMAP.md`: "LANTERN's evaluation, without its LLM judge"
- `docs/ROADMAP.md`: "LANTERN's RRF; its config-value blind spot"
- `docs/ARCHITECTURE.md`: "LANTERN (Cisco, 2026) and its failure analysis"

## Relations
- alternative-to:: [[LCM Lossless Context Management]]
- alternative-to:: [[ARC Addressable Recall Compaction]]
- alternative-to:: [[CogCanvas]]
- tested-by:: [[LANTERN misses were 361 of 420 coverage and 59 of 420 ranking]]
- tested-by:: [[LANTERN measured 1.9 points from an LLM reranker, not significant]]
