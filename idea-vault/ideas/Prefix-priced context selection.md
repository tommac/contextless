---
type: idea
status: proposed
area: cost
needs: model-calls
---
An item in the packet gets the cached price only if every earlier item in the prefix is unchanged, so the set of cheaply kept items is prefix-closed and the problem is not an ordinary knapsack with a discount. A practical solver would enumerate candidate reuse boundaries, keep the prefix through each boundary at the read price, and solve a knapsack for the suffix at the write or uncached price, with a quality cost per retained token. The research notes found no paper that solves selection jointly with cache state, so this is an open problem and a derivation by the notes' author, not a published result.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "An item gets the cached price only if every earlier item in the prefix is unchanged, so the feasible cached set is prefix-closed."
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "No paper was found that jointly optimizes context selection, ordering and provider cache state with measured answer quality."

## Relations
- builds-on:: [[Budgeted selection needs a monotone submodular objective]]
- builds-on:: [[The brief runs from stable to volatile]]
- builds-on:: [[Prompt cache savings are a ceiling until a live test]]
- builds-on:: [[Measure the bill not the tokens]]
- part-of:: [[A query planner and EXPLAIN for context]]
- tested-by:: [[A cache write at 1.25 pays back when reuse probability passes 28 percent]]
- analog-of:: [[Rent-or-buy decisions]]
