---
type: field
status: external
area: index
needs: none
---
Succinct indexes such as the FM-index and suffix arrays, from string algorithms, support exact substring search over a text in compact space. The notes suggest one could add exact substring search for repetitive tool output, but a targeted search found no agent-memory work that uses them. Whether such an index would improve recall over BM25 with regular expressions is unknown.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "A suffix or FM-index could add exact substring search with compact storage, especially for repetitive tool output. Whether it improves recall over BM25 plus regex is unknown."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "A targeted search for FM-index or suffix-array use in agent memory found only general compressed-index research"

## Relations
- analog-of:: [[Fused ranking over exact full-text and value lanes]]
- analog-of:: [[The value lane finds stated values near the question's words]]
- analog-of:: [[Quotes are pinned to the paragraphs that hold them]]
