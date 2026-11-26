---
type: idea
status: built
area: packets
needs: none
---
While the earlier turns total at most whole_history_tokens, 60,000 by default, and fit the role's first profile, all of them are sent verbatim in order and nothing is ranked. The first run on real transcripts motivated it: selection that leaves turns out when everything fits loses to a plain transcript. The 60,000 limit is a cost choice and not a measured optimum, and plan units always get a selected packet.

## Evidence
- `docs/ARCHITECTURE.md`: "While the earlier turns total at most `whole_history_tokens` (60,000 by default, at the first profile's token density) and fit the role's first profile, all of them are sent verbatim in order, nothing is ranked"
- `docs/CONTEXT-COMPILER.md`: "The default limit of 60,000 is a cost choice, not a measured optimum."
- `docs/ARCHITECTURE.md`: "Selection that leaves turns out when everything fits loses to a plain transcript."

## Relations
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[Packets are sized before the model starts]]
- alternative-to:: [[Ranked atoms are sent as excerpts under a token budget]]
- tested-by:: [[The 140 turn conversation sent whole held all 30 sources at 83100 tokens]]
- tested-by:: [[Floor whole history held all 120 answers at 297525 tokens]]
