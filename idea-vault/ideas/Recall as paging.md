---
type: idea
status: designed
area: recall
needs: none
---
A needs_recall is treated as a page fault: the packet lacked something the model needed. That suggests three mechanisms from virtual memory: a working set, where atoms recalled or cited in the last few calls stay in the next packets; prefetch, where the graph neighbours of a faulted atom join the next packet; and the fault rate per call as a health metric. It is designed and not built, and it would add two recall verbs, expand and zoom out.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "A `needs_recall` is a page fault: the packet lacked something the model needed."
- `docs/CONTEXT-COMPILER.md`: "A rising fault rate means the selection is too stingy; zero faults with large packets means it is too generous."

## Relations
- builds-on:: [[Recall by id or by search replaces memory]]
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[The link graph and its expansion lane]]
- analog-of:: [[Virtual memory paging]]
- tested-by:: [[Pichay faulted on 0.0254 percent of evictions offline and thrashed live]]
