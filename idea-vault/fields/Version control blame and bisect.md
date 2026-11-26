---
type: field
status: external
area: structure
needs: none
---
Version control keeps every change, so blame can point from a line to the change that last set it, a line-history command can list the changes to a range in order, and bisect can find by binary search the change where a property flipped. Agent systems have borrowed commit, branch, merge and rollback, but a targeted search found no system that applies a blame or bisect analogue to facts changing within a conversation. The research note maps it directly onto the value lane: blame is a pointer from the current value of a key to the span that last set it, and the line history is the ordered chain of spans where the value changed. Bisect is only worth having when testing a point in history is expensive; with a sorted value ledger the change points are read off directly.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Blame is a pointer from the current value of a key to the span that last set it."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A targeted search found no system that applies a blame or bisect analogue to facts changing within a conversation."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "git bisect finds the commit where a property changed by binary search in about log2(N) tests"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "the abstract reports over 13% better task resolution than long-context baselines and over 80% on SWE-Bench Verified"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Bisect is only worth having when testing a point in history is expensive; with a sorted value ledger the change points are read off directly."

## Relations
- analog-of:: [[Facts supersede facts]]
- analog-of:: [[Derived validity records over immutable turns]]
- analog-of:: [[Quotes are pinned to the paragraphs that hold them]]
- part-of:: [[Provenance and lineage]]
