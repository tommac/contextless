---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
On HotpotQA (7,405 validation queries, 66,581 passages) adding RM3 feedback to BM25 lowered recall at 10 from 0.742 to 0.675, with 227 wins, 5,998 ties and 1,180 losses across queries, and it tripled median latency. An earlier study on the full-wiki setting found RM3 and Rocchio below BM25 as well, and its authors explain that feedback relies on term frequency in the top documents and misses entities that are not frequent there. The headline numbers come from a single-author preprint whose 2Wiki figures the research note treats as weak. For this project it is evidence that a classic feedback pass can lower recall on questions that need a second piece of evidence.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Classic pseudo-relevance feedback is measured to hurt bridge-style multi-hop retrieval at small cutoffs (HotpotQA R@10 0.742 to 0.675 with RM3) and adds at most 1.6 points at depth 500"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "BM25 R@5 0.631, R@10 0.742, MRR 0.784; BM25+RM3 0.531/0.675/0.662"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "RM3 (PRF-RM) = 0.083/0.175/0.242/0.296, MAP 0.406"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "single author; the 2Wiki BM25 R@10 is 0.643 in the main table but 0.512 in the robustness table"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "RM3 also tripled median latency (0.169 s against 0.057 s on HotpotQA)."

## Relations
- contradicts:: [[Feedback expansion from first-pass hits needs no model]]
- supports:: [[Feedback links and coverage did not beat search]]
- part-of:: [[Pseudo-relevance feedback]]
- source:: [[SPRIG]]
