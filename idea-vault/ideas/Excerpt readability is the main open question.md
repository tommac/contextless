---
type: idea
status: proposed
area: evaluation
needs: model-calls
---
The probes show that the answer is in the packet, not that a model answers correctly: a paragraph out of context can lose what a pronoun refers to. Excerpts carry the turn, the paragraph numbers and the message they answer in their header, and the paragraphs around one come back by id, but whether a model reads an excerpt as well as a whole turn is unmeasured. Only an end-to-end run with a model settles it, and that run is the next step in the roadmap, with model calls to be asked for first.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "Still unmeasured: whether a model reads an excerpt as well as a whole turn."
- `docs/CONTEXT-COMPILER.md`: "That readability risk is the main open question, and only an end-to-end test with a model settles it."

## Relations
- part-of:: [[No model has yet answered from a packet]]
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
