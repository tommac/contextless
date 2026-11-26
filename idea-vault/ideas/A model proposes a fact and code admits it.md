---
type: idea
status: external
area: storage
needs: none
---
MemTxn suggests a split that suits a deterministic engine: a model may propose a fact, but whether the fact is admitted and whether it is visible can be decided by code. Admission checks that the quoted evidence exists in the source, and visibility follows a declared chronological rule. The gate does not establish that a fact is true, only that its evidence is really in the text.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "MemTxn suggests a useful split for determinism: an LLM may propose a fact, but admission (quoted evidence exists in the source) and visibility (declared chronological rule) can be pure code."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The audit used template-generated corruptions, and the gate does not establish semantic truth"

## Relations
- supports:: [[The indexer writes cards facts and decisions with checked quotes]]
- supports:: [[Units return typed records with checked claims]]
- supports:: [[Quotes are pinned to the paragraphs that hold them]]
- supports:: [[Write-time intelligence and read-time determinism]]
- supports:: [[Proof-carrying memory]]
- source:: [[MemTxn]]
