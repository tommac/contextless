---
type: idea
status: measured-kept
area: packets
needs: model-calls
---
The packet shows every checked idea note as one line that names the paragraphs holding its quotes. The model reads the map and asks for the paragraphs its answer rests on, and the engine sends them word for word within a token budget per request. A note is a pointer and is never sent in place of the text. Built on 2026-11-25 after the answer test, and confirmed the same day on 120 other held-out questions: the engine's packet with the map stated 0.882 of the needed passages against 0.706 with the map off. Still open: notes written by the indexer, a model that is not offered the request as a step, and a search arm of the same size.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "The pull arm is now a path of the engine, called the atlas:"
- `docs/CONTEXT-COMPILER.md`: "- Q2: THE MAP HELPS THE ENGINE. The live packet with the map is ahead of the"

## Relations
- builds-on:: [[Idea notes as verified keys over verbatim paragraphs]]
- tested-by:: [[The live packet with the map scored 0.882 against 0.706 with the map off]]
- tested-by:: [[Note lines plus requested paragraphs beat same-size search again by 0.090]]
- tested-by:: [[One request from the map raised held passages from 0.327 to 0.840]]
