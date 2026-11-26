---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
SPRIG builds an entity to passage graph weighted by TF-IDF with no LLM and runs personalized PageRank over it. Seeded from the top BM25 passages it reached recall at 10 of 0.775 on HotpotQA against 0.742 for BM25, with 739 wins, 6,411 ties and 255 losses, while seeding from query entities alone fell to 0.464. Downstream answer gain with a small extractive reader was slight, from 0.173 to 0.180 exact match. It is the best measurement found for propagation from lexical seeds, on a document benchmark and not on conversations, from a single-author preprint with stated inconsistencies.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Seeding from query entities alone was far worse than BM25 (HotpotQA R@10 0.464 against 0.742)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "gave 0.775 on HotpotQA, delta +0.0330"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Downstream answer gain with a small extractive reader was slight (exact match 0.173 to 0.180)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "HotpotQA R@10 +0.033, 2Wiki +0.100 over BM25, Dec 2025."

## Relations
- supports:: [[Links help when questions need several pieces of evidence]]
- supports:: [[The link graph and its expansion lane]]
- supports:: [[Propagation from lexical seeds over a rare-key graph]]
- supports:: [[Search first then bounded query-filtered expansion]]
- part-of:: [[Personalized PageRank]]
- source:: [[SPRIG]]
