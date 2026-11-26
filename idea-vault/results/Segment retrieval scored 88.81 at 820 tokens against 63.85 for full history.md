---
type: result
status: measured
area: selection
needs: none
measured-by: others
---
In the peer-reviewed SeCom paper, retrieving topical segments scored 88.81 at 820 tokens on Long-MT-Bench+, against 63.85 for full history at 19,287 tokens. Topical segments also beat the best turn-level baseline on LoCoMo. The report takes this as support for small verbatim units, measured on older readers.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "full history got GPT4Score 63.85 at 19,287 tokens, versus 88.81 for segment retrieval at 820 tokens"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "SeCom's topical segments beat the best turn-level baseline on LoCoMo (71.57 vs 65.58, BM25)"
- `reports/Cutting edge context memory research.md`: "Supports small verbatim units, on older readers"

## Relations
- supports:: [[Unrelated context can hurt answers]]
- supports:: [[Two opposite slogans about context size both fail]]
- supports:: [[Atoms are paragraph-sized addresses into a turn]]
- supports:: [[Threads group turns by topic]]
- source:: [[SeCom]]
