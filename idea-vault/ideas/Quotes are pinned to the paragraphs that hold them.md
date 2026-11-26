---
type: idea
status: built
area: plans
needs: none
---
When a unit's quote occurs in the text it cites, the engine pins the claim to the paragraph ids that hold the quote, also when the unit cited only the whole turn. A quote found elsewhere in a cited turn is pinned where it is, with a note, and a quote found in none of its texts lowers the claim to inferred. Pinning turns a unit's evidence into atom addresses, so code can see which claims of different units rest on the same paragraphs. The survey notes that no surveyed system can pass evidence by reference and verify it by code as directly.

## Evidence
- `docs/ARCHITECTURE.md`: "A quote that does is pinned to the paragraphs that hold it (`t000123#a4`), also when the unit cited the whole turn."
- `docs/ARCHITECTURE.md`: "A quoted claim whose quote is in none of its texts, or that gives no quote, counts as inferred."
- `reports/Parallel AI agents data structures.md`: "because atoms are stable addresses into immutable turns, evidence can be passed by reference and verified by code"

## Relations
- part-of:: [[Units return typed records with checked claims]]
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- analog-of:: [[Proof-carrying memory]]
