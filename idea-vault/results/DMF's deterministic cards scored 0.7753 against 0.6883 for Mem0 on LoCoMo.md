---
type: result
status: preprint
area: index
needs: none
measured-by: others
---
DMF archives raw records and derives memory cards with no LLM in memory management; the cards are retrieval aids, not replacements for raw records, and are expanded to raw turns before rendering. On LoCoMo its judge score was 0.7753 against 0.6883 for Mem0 on 1,540 questions, with temporal at 0.7072 against 0.1526, and zero memory tokens against 49.98M. On a 60-question LongMemEval sample Mem0 led 0.767 to 0.717. There is no ablation, and only narrow patterns that can be normalised deterministically become cards, so the value of the cards alone is unmeasured.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "LoCoMo judge score 0.7753 against Mem0 0.6883 (n = 1,540): temporal 0.7072 against 0.1526, multi-hop tied at 0.7943, single-hop lower (0.8181 against 0.8716)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "On a 60-question LongMemEval sample Mem0 led 0.767 to 0.717, and 0.700 to 0.500 on multi-session."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "retrieval aids, not replacements for raw records"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Zero memory tokens against 49.98M for Mem0 on LoCoMo."

## Relations
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Derived validity records over immutable turns]]
- part-of:: [[Materialized views]]
- source:: [[DMF]]
