---
type: source
status: external
area: index
needs: none
year: 2026
url: https://arxiv.org/html/2606.03463
---
DMF (Stabile and Zimuel, Roma Tre; arXiv 2 Jun 2026) archives raw records and derives memory cards with kind, subject, predicate, object, qualifiers, time anchor, validity and a source record id, with no LLM in memory management. The cards are retrieval aids, not replacements for raw records, and are expanded to raw turns before rendering. Decay is by interaction count, so memory state is a function of the message sequence alone, and a supersession penalty of 0.35 demotes replaced entries. It uses bge-small embeddings, reports no ablation, and turns only narrow patterns that can be normalised deterministically into cards.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Decay is by interaction count, so memory state is a function of the message sequence alone; a supersession penalty of 0.35 demotes replaced entries."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "and are expanded to raw turns before rendering"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Zero memory tokens against 49.98M for Mem0 on LoCoMo."

## Relations
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[Facts supersede facts]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- part-of:: [[Materialized views]]
- alternative-to:: [[Mem0]]
- tested-by:: [[DMF's deterministic cards scored 0.7753 against 0.6883 for Mem0 on LoCoMo]]
