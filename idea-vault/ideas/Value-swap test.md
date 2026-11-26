---
type: idea
status: proposed
area: evaluation
needs: model-calls
---
The value-swap test changes the value that carries the answer in the source, such as a port number or a date, and checks that the answer follows. It separates use of memory from prior knowledge and is the cleanest fit for closed APIs, where methods that need a model's internals are unavailable. The notes found no standard, validated value-swap protocol for conversational memory, so the steps are a synthesis that would need piloting.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "change the answer-bearing value in the source (a port number, a date) and check that the answer follows"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "This separates use of memory from prior knowledge and is the cleanest fit for closed APIs, where AT2-style internals are unavailable."
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "I found no standard, validated value-swap protocol for conversational memory."

## Relations
- part-of:: [[Atom-level causal evaluation]]
- part-of:: [[R7 Test whether answers depend on retrieved memory]]
- builds-on:: [[Determinism is the instrument]]
- alternative-to:: [[ContextCite]]
