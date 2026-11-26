---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
Re:CAP used RM3 as its classical control on multi-hop sets. Across six dataset and retriever cells of 98 to 100 queries each, RM3 lifted top-500 recall by at most 1.6 points over flat BM25, and at a matched budget of about 240 documents it reached 0.636 on MuSiQue against 0.610 for flat BM25. Re:CAP's own gain comes from an LLM loop of 250 to 500 calls per audited query, and its leave-one-out table attributes the lift to an entity ledger. So even deep in the ranking, classic feedback adds little to multi-hop recall, while tracking entities adds a lot.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "RM3 lifts top-500 recall by at most 1.6 pp over flat BM25"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "At a matched budget of about 240 documents RM3 reached 0.636 on MuSiQue"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "dropping it costs 8.58 points on MuSiQue and 4.17 on HotpotQA, and changes MultiHop-RAG by -1.47 (the lever hurt slightly there)"

## Relations
- contradicts:: [[Feedback expansion from first-pass hits needs no model]]
- supports:: [[Propagation from lexical seeds over a rare-key graph]]
- supports:: [[Links help when questions need several pieces of evidence]]
- part-of:: [[Pseudo-relevance feedback]]
