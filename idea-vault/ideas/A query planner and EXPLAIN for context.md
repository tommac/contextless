---
type: idea
status: designed
area: packets
needs: model-calls
---
The packet compiler would turn a message, workspace and route into a packet in passes, the way a database planner turns a query into a plan: classify the request, set a budget, check whether the whole history fits, gather candidates from lanes, fuse, expand for readability, select under the budget, add levels of detail, lay out for the cache, then size and route. A command, cx explain, would print the plan: each lane's top candidates, what was chosen and why, tokens per section, the length of the cache-stable prefix, and estimated cost and time per profile. It is designed as the main debugging tool and is not built; its cost model would be fitted from a few real runs.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "The compiler turns (message, workspace, route) into a packet in passes, like a database planner turns a query into a plan"
- `docs/CONTEXT-COMPILER.md`: "It is the main debugging tool, and the probes run on top of it."

## Relations
- builds-on:: [[Each call gets a fresh packet built by code]]
- builds-on:: [[Packets are sized before the model starts]]
- builds-on:: [[The engine never asks a model what to do next]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[Measure the bill not the tokens]]
- analog-of:: [[Query planning in databases]]
