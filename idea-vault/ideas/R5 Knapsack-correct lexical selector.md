---
type: idea
status: designed
area: selection
needs: none
---
The hypothesis is that lexical coverage chosen by gain per token, with concave saturation, an RRF term and a fallback to the best single atom, finds more answers at tight budgets than filling in rank order, with no model. The notes found no such study on transcripts, and the report calls the contribution incremental: worth building, not a headline. It is falsified if it does not beat BM25 fill by at least 8 points at 2k or 4k on locked probes, or if it loses at 8k.

## Evidence
- `reports/Cutting edge context memory research.md`: "Lexical coverage selected by gain per token, with concave saturation, an RRF modular term and a best-single-atom fallback, finds more answers at tight budgets than rank-order filling, with no model."
- `reports/Cutting edge context memory research.md`: "It does not beat BM25 fill by at least 8 points at 2k or 4k on locked probes, or it loses at 8k"
- `reports/Cutting edge context memory research.md`: "**Incremental.** Worth building, not a headline"

## Relations
- builds-on:: [[Budgeted selection needs a monotone submodular objective]]
- builds-on:: [[The offline floor is the baseline every selector must beat]]
- builds-on:: [[R1 Repair the instrument and record the offline floor]]
- alternative-to:: [[Ranked atoms are sent as excerpts under a token budget]]
- source:: [[Lin and Bilmes 2011]]
