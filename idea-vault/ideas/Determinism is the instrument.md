---
type: idea
status: built
area: evaluation
needs: none
---
A deterministic compiler over immutable text means two arms of an experiment differ only by a known set of atoms. Most memory systems change extraction, retrieval and wording together and so cannot isolate one change, and the notes found no study of byte-level context reproducibility. Reproducibility alone is not a paper. Its value is that it makes removal, swap and replay experiments clean, while the reader and the judge stay variable.

## Evidence
- `reports/Cutting edge context memory research.md`: "A deterministic compiler over immutable text means two arms differ only by a known diff of atoms."
- `reports/Cutting edge context memory research.md`: "Reproducibility alone is not a paper. Its value is that it makes removal, swap and replay experiments clean."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "fixed retrieval and packing let answer changes be traced to the context diff, reducing one major source of evaluation noise"

## Relations
- builds-on:: [[Each call gets a fresh packet built by code]]
- builds-on:: [[Turns are stored verbatim and never rewritten]]
- builds-on:: [[The engine never asks a model what to do next]]
- supports:: [[Atom-level causal evaluation]]
- supports:: [[Verified citations and recall faults as ranking labels]]
- supports:: [[Write-time intelligence and read-time determinism]]
