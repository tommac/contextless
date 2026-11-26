---
type: idea
status: designed
area: index
needs: model-calls
---
The indexer could write, per atom, a one-line claim as its title and one to three questions the atom answers, so the title works like an API for the paragraph. Titles and questions are index entries: searched, shown on catalog lines with the atom id, never shown as facts and never sent in place of the atom. A cheap model would write them at index time, so the step is opt-in and kept only if fresh probes improve. The card result is the first evidence for it, while a no-model variant, indexing each assistant atom with the question it answers, did not help.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "They are index entries: searched, shown on catalog lines with the atom id, never shown as facts and never sent in place of the atom."
- `docs/CONTEXT-COMPILER.md`: "The indexer can write, per atom, a one-line claim and one to three questions the atom answers (doc2query)."

## Relations
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- source:: [[doc2query]]
- alternative-to:: [[A-MEM]]
- analog-of:: [[Evergreen note titles]]
- part-of:: [[Write-time intelligence and read-time determinism]]
- tested-by:: [[Atoms lost a planted decision until searched with card text]]
- tested-by:: [[Indexing atoms with the question they answer did not help overall]]
