---
type: result
status: preprint
area: index
needs: none
measured-by: others
---
In AgentMemBench, a third-party benchmark of low power, dense key-value retrieval had a macro recall at 5 of 0.792, against 0.556 for summarization, 0.468 for in-context windowing and 0.361 for an entity graph built with spaCy. Building structure without a model is therefore not automatically good, and the author attributes the weak entity graph largely to noisy entity recognition. BM25, grep, full context, Mem0 and Zep were not tested, so the result does not rank dense retrieval against the simple baselines.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "dense key-value retrieval had macro R@5 0.792 vs summarization 0.556, in-context windowing 0.468 and spaCy entity graph 0.361"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Non-LLM construction is not automatically good: AgentMemBench's spaCy entity graph was its weakest retriever, and the author attributes this largely to noisy NER"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "BM25, grep, full context, Mem0 and Zep were not tested"

## Relations
- contradicts:: [[Model-free hubs over verbatim turns]]
- contradicts:: [[LLM summaries as memory]]
- supports:: [[An embedding lane beside the exact full-text and value lanes]]
