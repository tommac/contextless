---
type: idea
status: external
area: structure
needs: none
---
The strongest pattern that transfers from tree and DAG memories is to keep the leaves immutable and to treat every derived node as disposable. LCM, MemForest's session trees and the raw-dialogue lesson of a cross-method study all point that way. A plain time or session hierarchy can be built without a model, and the summaries are the expensive and non-deterministic part, so structure and summary should be separate layers.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The strongest transferable pattern is **immutable leaves plus disposable derived nodes**."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "A plain time or session hierarchy can be built without a model. The summaries are the expensive and non-deterministic part, so structure and summary should be separate layers."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "The authors conclude hierarchy helps, while keeping raw dialogue is essential to avoid semantic loss"

## Relations
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[Digests label blocks of omitted turns]]
- supports:: [[Views with lineage instead of summaries]]
- builds-on:: [[Summary trees help only on very long histories]]
- source:: [[LCM Lossless Context Management]]
- analog-of:: [[Append-only storage]]
