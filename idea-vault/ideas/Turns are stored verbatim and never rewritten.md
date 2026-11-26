---
type: idea
status: built
area: storage
needs: none
---
Every user message and every answer is kept word for word as its own Markdown file, append-only and committed to git. A turn is never rewritten, so any turn can be put back into any later packet exactly as it was said. This is the base of the whole project: the model's context window is never the memory, the vault of verbatim records is.

## Evidence
- `docs/ARCHITECTURE.md`: "Every user and assistant turn is stored verbatim, append-only, one Markdown file per turn, committed to git."
- `docs/ARCHITECTURE.md`: "Any turn can be put back into any later packet word for word."

## Relations
- builds-on:: [[Raw text beats rewritten memory]]
- analog-of:: [[Append-only storage]]
- source:: [[BookWriter]]
- tested-by:: [[At 1,200 turns every turn is in the packet verbatim or named in it]]
