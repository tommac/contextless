---
type: result
status: measured
area: index
needs: none
measured-by: us
---
Indexing each assistant atom together with the user question it answers, which needs no model, was a little better on one conversation and worse on another. The stated reason is that these probe questions already share words with their answers. It is weak evidence against the question half of atom titles and against deterministic keys, and it leaves open whether written questions would help on paraphrased probes.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "**Indexing each assistant atom with the question it answers** did not help overall (a little better on one conversation, worse on another)"

## Relations
- contradicts:: [[Atom titles as APIs]]
- contradicts:: [[Deterministic keys on spans with verbatim output]]
- supports:: [[Held-out probes written from memory without the transcript]]
