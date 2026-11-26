---
type: field
status: external
area: storage
needs: none
---
Content-addressed storage names each stored item by a hash of its bytes, which gives deduplication, integrity checks and stable references. A store of this kind in about 300 lines cut the retained storage of seven agent frameworks by 4.8 to 32.7 times, and every retained file passed its restore check. It fits immutable turns directly, but deletion and privacy need a separate design, because immutable stores linked by hashes make erasure difficult.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Across 70 runs of seven frameworks, retained storage fell by 4.8× to 32.7×, and all 310 retained files passed restore checks"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Content addressing fits immutable turns directly: hashes give deduplication, integrity and stable references in summary provenance."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Deletion and privacy need separate design, because hash-linked immutable stores make erasure difficult."

## Relations
- analog-of:: [[Turns are stored verbatim and never rewritten]]
- analog-of:: [[Recall by id or by search replaces memory]]
- part-of:: [[Append-only storage]]
