---
type: source
status: external
area: storage
needs: none
year: 2026
url: https://arxiv.org/html/2609.34242v2
---
Stashbird (Biringa et al., Microsoft; arXiv v2 17 Oct 2026) keeps, for every derived entity and relation, the set of supporting episodes, and reconciles updates and deletions against those sets. It extracts with an LLM at write time and uses one at query time. It reports LongMemEval-S 81.0 against 78.3 for Hindsight and 70.0 for Mem0, LoCoMo 87.5 against 89.1 for Hindsight, and ingestion with 484K prompt tokens against 37M for Graphiti. It cites database provenance work, not view maintenance.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "keeps for every derived entity and relation the set of supporting episodes"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Ingestion used 484K prompt tokens against 37M for Graphiti. It cites database provenance work, not view maintenance."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LongMemEval-S 81.0 +/- 0.8 against Hindsight 78.3 +/- 2.6 and Mem0 70.0 +/- 0.6 (GPT-4.1-mini)"

## Relations
- part-of:: [[Provenance and lineage]]
- supports:: [[Facts supersede facts]]
- supports:: [[Derived validity records over immutable turns]]
- alternative-to:: [[Zep and Graphiti]]
- tested-by:: [[Support sets gave 20 of 20 correct final states against 8 of 20]]
