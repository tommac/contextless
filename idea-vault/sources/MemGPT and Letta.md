---
type: source
status: external
area: recall
needs: none
year: 2025
url: https://www.letta.com/blog/benchmarking-ai-agent-memory/
---
Letta, a vendor of agent memory, reported that an agent working over one uploaded LoCoMo file scored 74.0 with GPT-4o mini, above the 68.5 reported for Mem0. The result is vendor-reported and was not grep-only, because the files were embedded and the agent had to call a search tool first. The notes read it as evidence for an agent that searches raw files step by step, not for grep or Markdown as such. In a third-party benchmark MemGPT itself scored 28.15, below BM25 at 43.22.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "reported 74.0 on LoCoMo with a file-based agent vs Mem0's reported 68.5. This is not a pure grep result"
- `research_notes/Cutting edge context memory research/atomic_notes_and_tangents.md`: "Letta's result is better read as evidence for iterative agent retrieval than for Markdown or grep specifically."
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Hindsight 46.01, BM25 43.22, HippoRAG 39.72, embeddings 38.04, A-Mem 35.10, MemGPT 28.15, Mem0 25.73, GraphRAG 20.56"

## Relations
- supports:: [[Recall by id or by search replaces memory]]
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Benchmark scores are not comparable across papers]]
- alternative-to:: [[Mem0]]
- tested-by:: [[BM25 at 43.22 beat Mem0 and GraphRAG on GroupMemBench]]
