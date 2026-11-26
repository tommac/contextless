---
type: idea
status: rejected
area: storage
needs: none
---
Keeping memory as Zettelkasten notes or a compiled wiki written by a model is rejected as storage for this design. A-MEM lost to the plain baseline on single-hop questions with GPT-4o, and a compiled wiki used about 21 times the tokens per query while a decomposition baseline recovered most of its advantage. Linked notes do have measured support on cross-document synthesis, but at large cost and with links confounded with other changes, so the safe reading is to keep source pointers beside any note.

## Evidence
- `reports/Cutting edge context memory research.md`: "a wiki used about 21 times the tokens per query, and decomposition RAG recovered most of its advantage"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Measured support exists for LLM-compiled, linked notes on cross-document synthesis and multi-hop questions, but it comes with large build or query costs and usually confounds links with other changes."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Keeping source pointers alongside atomic notes is the safest reading of the evidence, because the failures reported are mostly dropped specifics and over-broad linking."

## Relations
- builds-on:: [[Raw text beats rewritten memory]]
- alternative-to:: [[Idea notes as verified keys over verbatim paragraphs]]
- tested-by:: [[A-MEM lost to the plain baseline on single-hop with GPT-4o]]
- tested-by:: [[A compiled wiki used about 21 times the tokens per query of RAG]]
- tested-by:: [[Verbatim files scored 78.1 against 37.5 for curated memory on PersonaMem]]
- source:: [[A-MEM]]
- source:: [[Karpathy LLM Wiki]]
