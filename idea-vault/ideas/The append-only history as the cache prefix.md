---
type: idea
status: proposed
area: caching
needs: model-calls
---
While the whole history fits, sending it verbatim as an append-only prefix may be both the best and the cheapest packet once it is cached, since each call would pay full price only for the new turns. That would let the full transcript beat any selection on quality and cost at once. Whether the vendor CLIs keep the prompt cached across separate headless runs is not known, so this is a proposal waiting on the same live test as the rest of the cache work.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "sending it verbatim as an append-only prefix may be both the best and the cheapest packet once it is cached, since each call pays full price only for the new turns"
- `docs/CONTEXT-COMPILER.md`: "Whether the CLIs keep our prompt cached across separate headless runs is not known yet"

## Relations
- builds-on:: [[A short history is sent whole]]
- builds-on:: [[The brief runs from stable to volatile]]
- builds-on:: [[Prompt cache savings are a ceiling until a live test]]
- alternative-to:: [[Ranked atoms are sent as excerpts under a token budget]]
- analog-of:: [[Append-only storage]]
- source:: [[TokenPilot]]
