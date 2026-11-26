---
type: source
status: external
area: structure
needs: none
year: 2026
url: https://arxiv.org/html/2601.02744
---
Synapse, in Findings of ACL 2026, retrieves by spreading activation over a graph of episodes. Its ablation is the clearest evidence that some of what a graph seems to add is chronology, because removing the graph hurt temporal questions and left multi-hop questions flat. Its edges mix temporal links between consecutive episodes with abstraction links and high-similarity semantic links.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "LoCoMo average F1: full 40.5, no graph structure 32.9, no activation dynamics 30.5, vectors only 25.2"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Its edges include consecutive-episode temporal links, abstraction links, and semantic links above 0.92 similarity"

## Relations
- supports:: [[Adjacent-turn expansion is the control for any link graph]]
- alternative-to:: [[HippoRAG 2]]
- tested-by:: [[Removing graph structure halved temporal F1 and left multi-hop flat]]
