---
type: idea
status: built
area: packets
needs: none
---
Sizing has a last step for a new message that is itself too big for any window. The engine splits it at paragraph boundaries into chunk units, every part is read by a model, and the outputs are then assembled. An engine test splits a 400 KB message and checks that every paragraph reaches a model, using scripted providers and no real model.

## Evidence
- `docs/ARCHITECTURE.md`: "then, if the focus itself is too big, split it into chunk units."
- `internal/engine/engine_test.go`: "// shrink every window so a 400 KB message fits none"

## Relations
- part-of:: [[Packets are sized before the model starts]]
- builds-on:: [[Drop tiers set the order in which records are left out]]
- builds-on:: [[Units return typed records with checked claims]]
- alternative-to:: [[One schema-checked call per item over data larger than any window]]
