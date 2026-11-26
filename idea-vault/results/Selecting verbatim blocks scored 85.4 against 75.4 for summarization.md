---
type: result
status: preprint
area: selection
needs: none
measured-by: others
---
ReCAP keeps a graph of atomic history blocks with importance and dependency edges derived from attention, and selects blocks with no model call, though use with a closed model needs a pass by a proxy model. On Lost-in-Conversation it scored 85.4 against 65.6 for full history and 75.4 for summarization, and on SWE-Together it scored 28.5 against 24.4 and 28.7. In this author-run preprint, selecting verbatim blocks beat or matched summarizing them.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "On Lost-in-Conversation with Qwen: 85.4 vs full history 65.6 and summarization 75.4. On SWE-Together: 28.5 vs full history 24.4 and summarization 28.7"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "selection needs no model call, though black-box use needs a proxy model pass"

## Relations
- supports:: [[Raw text beats rewritten memory]]
- supports:: [[Unrelated context can hurt answers]]
- supports:: [[Ranked atoms are sent as excerpts under a token budget]]
- contradicts:: [[LLM summaries as memory]]
