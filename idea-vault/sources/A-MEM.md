---
type: source
status: external
area: structure
needs: none
year: 2025
url: https://arxiv.org/html/2502.12110
---
A-MEM builds Zettelkasten-style notes written by a model, with generated links and notes that evolve, and was presented at NeurIPS 2025. With GPT-4o its own table shows it losing to the plain baseline on single-hop and adversarial questions while winning multi-hop and temporal ones. Its score moves sharply when other groups rerun it, which makes it a weak base for claims about linked notes.

## Evidence
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "builds Zettelkasten-style notes with link generation and memory evolution"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "With GPT-4o, A-MEM's own table shows it **losing** to the plain baseline on single-hop (48.43 vs 61.56) and adversarial (36.35 vs 52.61)"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "A-MEM's score moves sharply when others rerun it"

## Relations
- part-of:: [[Model-written notes lose as storage]]
- analog-of:: [[Zettelkasten]]
- alternative-to:: [[Karpathy LLM Wiki]]
- tested-by:: [[A-MEM lost to the plain baseline on single-hop with GPT-4o]]
- tested-by:: [[A-MEM links raised multi-hop F1 from 9.65 to 21.35 in its own ablation]]
- tested-by:: [[Memory frameworks scored 5 to 18 percent on STALE]]
