---
type: idea
status: proposed
area: packets
needs: none
---
A good excerpt format separates the text used to find a passage from the evidence itself. Generated context or titles help retrieval find a passage, while the original span with its speaker, timestamp and thread pointer keeps it checkable. Rewriting for coreference should never overwrite the source text, since rewriting can introduce errors, and the notes found no accepted metric for whether an excerpt is readable out of context.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "A good excerpt format separates retrieval text from evidence: generated context or titles help find a passage, while the original span, speaker, timestamp and thread pointer keep it checkable."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "It should not overwrite the source text, since rewriting can introduce errors."
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "I found no accepted metric for whether an excerpt is readable out of context."

## Relations
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[The atom index is stemmed and searches card text]]
- supports:: [[Idea notes as verified keys over verbatim paragraphs]]
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- analog-of:: [[Decontextualization]]
- source:: [[Anthropic contextual retrieval]]
- tested-by:: [[Contextual retrieval cut the top-20 miss rate from 5.7 to 1.9 percent]]
