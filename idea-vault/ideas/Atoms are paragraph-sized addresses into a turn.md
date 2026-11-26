---
type: idea
status: built
area: structure
needs: none
---
Every turn is cut by code into atoms, paragraph-sized spans of about 60 to 720 tokens, each with a stable id such as t000123#a4. An atom is an address into the stored turn and never a rewrite, so the atoms of a turn put back together give the body byte for byte. Atoms are the unit that is ranked, sent and recalled, which lets a packet carry the one paragraph that matters without the long turn around it.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "An atom is an address into the immutable turn, never a rewrite or a summary."
- `docs/ARCHITECTURE.md`: "The ranked unit is the **atom**: a paragraph-sized piece of a turn (about 60 to 720 tokens), cut by code at blank lines."
- `docs/CONTEXT-COMPILER.md`: "The concatenation of a turn's atoms gives back the body byte for byte."

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- source:: [[Dense X Retrieval]]
- tested-by:: [[Atoms at 8k found 107 of 120 against 103 for whole turns]]
- tested-by:: [[Atoms reach the answer after 267 to 339 tokens against 1718 to 2077]]
- analog-of:: [[Evergreen note titles]]
- analog-of:: [[Zettelkasten]]
- source:: [[SeCom]]
